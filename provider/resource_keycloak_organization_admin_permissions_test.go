package provider

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakOrganizationAdminPermissions_basic(t *testing.T) {
	skipIfVersionIsLessThan(testCtx, t, keycloakClient, keycloak.Version_26_7)
	skipIfFGAPv2NotEnabled(testCtx, t, keycloakClient)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOrganizationAdminPermissions_basic("view"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOrganizationAdminPermissionsExists("keycloak_organization_admin_permissions.test"),
					testAccCheckKeycloakOrganizationAdminPermissionsTargets("keycloak_organization_admin_permissions.test", "Organizations"),
				),
			},
			{
				Config: testKeycloakOrganizationAdminPermissions_basic("manage"),
				Check:  testAccCheckKeycloakOrganizationAdminPermissionsExists("keycloak_organization_admin_permissions.test"),
			},
			{
				ResourceName:            "keycloak_organization_admin_permissions.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"organization_ids"},
			},
		},
	})
}

func TestAccKeycloakOrganizationAdminPermissions_withOrganizationsAndPolicy(t *testing.T) {
	skipIfVersionIsLessThan(testCtx, t, keycloakClient, keycloak.Version_26_7)
	skipIfFGAPv2NotEnabled(testCtx, t, keycloakClient)

	// The policy needs the admin-permissions client's id, which is not known to the configuration
	// before a permission exists; it is looked up here instead.
	adminPermissionsClientId, err := keycloakClient.GetAdminPermissionsClientId(testCtx, testAccRealmFGAPv2.Realm)
	if err != nil {
		t.Fatalf("error getting admin-permissions client id: %s", err)
	}

	organizationName := acctest.RandomWithPrefix("tf-acc")
	userName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOrganizationAdminPermissions_withOrganizationsAndPolicy(adminPermissionsClientId, organizationName, userName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOrganizationAdminPermissionsExists("keycloak_organization_admin_permissions.test"),
					testAccCheckKeycloakOrganizationAdminPermissionsTargetsOrganization("keycloak_organization_admin_permissions.test", "keycloak_organization.target"),
					resource.TestCheckResourceAttr("keycloak_organization_admin_permissions.test", "policies.#", "1"),
					resource.TestCheckResourceAttr("keycloak_organization_admin_permissions.test", "scopes.#", "2"),
				),
			},
		},
	})
}

func TestAccKeycloakOrganizationAdminPermissions_invalidScope(t *testing.T) {
	skipIfVersionIsLessThan(testCtx, t, keycloakClient, keycloak.Version_26_7)
	skipIfFGAPv2NotEnabled(testCtx, t, keycloakClient)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakOrganizationAdminPermissions_basic("manage-members"),
				ExpectError: regexp.MustCompile(`expected scopes\.\d+ to be one of`),
			},
		},
	})
}

func testAccCheckKeycloakOrganizationAdminPermissionsExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		realmId := rs.Primary.Attributes["realm_id"]
		permId := rs.Primary.Attributes["permission_id"]
		authorizationResourceServerId := rs.Primary.Attributes["authorization_resource_server_id"]

		adminPermissionsClientId, err := keycloakClient.GetAdminPermissionsClientId(testCtx, realmId)
		if err != nil {
			return fmt.Errorf("error getting admin-permissions client id: %s", err)
		}

		if authorizationResourceServerId != adminPermissionsClientId {
			return fmt.Errorf("computed authorizationResourceServerId %s was not equal to %s (the id of the admin-permissions client)", authorizationResourceServerId, adminPermissionsClientId)
		}

		perm, err := keycloakClient.GetOpenidClientAuthorizationPermission(testCtx, realmId, adminPermissionsClientId, permId)
		if err != nil {
			return fmt.Errorf("error fetching permission %s: %s", permId, err)
		}
		if perm == nil {
			return fmt.Errorf("permission %s not found in Keycloak", permId)
		}
		if perm.ResourceType != "Organizations" {
			return fmt.Errorf("permission %s has resource type %q, expected Organizations", permId, perm.ResourceType)
		}

		return nil
	}
}

// testAccCheckKeycloakOrganizationAdminPermissionsTargets checks the names of the authorization
// resources the permission applies to. Keycloak names a permission's resource after the entity it
// targets, or after the resource type when the permission targets every entity of that type.
func testAccCheckKeycloakOrganizationAdminPermissionsTargets(resourceName string, expectedNames ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		realmId := rs.Primary.Attributes["realm_id"]
		permId := rs.Primary.Attributes["permission_id"]
		resourceServerId := rs.Primary.Attributes["authorization_resource_server_id"]

		perm, err := keycloakClient.GetOpenidClientAuthorizationPermission(testCtx, realmId, resourceServerId, permId)
		if err != nil {
			return fmt.Errorf("error fetching permission %s: %s", permId, err)
		}

		var names []string
		for _, resourceId := range perm.Resources {
			r, err := keycloakClient.GetOpenidClientAuthorizationResource(testCtx, realmId, resourceServerId, resourceId)
			if err != nil {
				return fmt.Errorf("error fetching authorization resource %s: %s", resourceId, err)
			}
			names = append(names, r.Name)
		}

		sort.Strings(names)
		sort.Strings(expectedNames)
		if strings.Join(names, ",") != strings.Join(expectedNames, ",") {
			return fmt.Errorf("permission %s applies to %v, expected %v", permId, names, expectedNames)
		}

		return nil
	}
}

func testAccCheckKeycloakOrganizationAdminPermissionsTargetsOrganization(resourceName, organizationResourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		org, ok := s.RootModule().Resources[organizationResourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", organizationResourceName)
		}

		return testAccCheckKeycloakOrganizationAdminPermissionsTargets(resourceName, org.Primary.ID)(s)
	}
}

func testKeycloakOrganizationAdminPermissions_basic(scope string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_organization_admin_permissions" "test" {
	realm_id = data.keycloak_realm.realm.id
	name     = "test-organization-permission"
	scopes   = ["%s"]
}
`, testAccRealmFGAPv2.Realm, scope)
}

func testKeycloakOrganizationAdminPermissions_withOrganizationsAndPolicy(adminPermissionsClientId, organizationName, userName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_organization" "target" {
	realm = data.keycloak_realm.realm.id
	name  = "%s"

	domain {
		name = "%s.example.com"
	}
}

resource "keycloak_user" "admin" {
	realm_id = data.keycloak_realm.realm.id
	username = "%s"
}

resource "keycloak_openid_client_user_policy" "test" {
	realm_id           = data.keycloak_realm.realm.id
	resource_server_id = "%s"
	name               = "user_policy_test_%s"
	users              = [keycloak_user.admin.id]
	logic              = "POSITIVE"
	decision_strategy  = "UNANIMOUS"
}

resource "keycloak_organization_admin_permissions" "test" {
	realm_id          = data.keycloak_realm.realm.id
	name              = "test-organization-manage-%s"
	description       = "test organization permission with policy"
	decision_strategy = "UNANIMOUS"
	organization_ids  = [keycloak_organization.target.id]
	scopes            = ["view", "manage"]
	policies          = [keycloak_openid_client_user_policy.test.id]
}
`, testAccRealmFGAPv2.Realm, organizationName, organizationName, userName, adminPermissionsClientId, userName, organizationName)
}
