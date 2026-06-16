package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakOpenidClientAdminPermissions_basic(t *testing.T) {
	skipIfVersionIsLessThan(testCtx, t, keycloakClient, keycloak.Version_26_2)
	skipIfFGAPv2NotEnabled(testCtx, t, keycloakClient)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClientAdminPermissions_basic("view"),
				Check:  testAccCheckKeycloakOpenidClientAdminPermissionsExists("keycloak_openid_client_admin_permissions.test"),
			},
			{
				Config: testKeycloakOpenidClientAdminPermissions_basic("manage"),
				Check:  testAccCheckKeycloakOpenidClientAdminPermissionsExists("keycloak_openid_client_admin_permissions.test"),
			},
			{
				ResourceName:            "keycloak_openid_client_admin_permissions.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"client_ids"},
			},
		},
	})
}

func TestAccKeycloakOpenidClientAdminPermissions_withScopes(t *testing.T) {
	skipIfVersionIsLessThan(testCtx, t, keycloakClient, keycloak.Version_26_2)
	skipIfFGAPv2NotEnabled(testCtx, t, keycloakClient)

	clientName := acctest.RandomWithPrefix("tf-acc")
	groupName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClientAdminPermissions_withScopes(clientName, groupName),
				Check:  testAccCheckKeycloakOpenidClientAdminPermissionsExists("keycloak_openid_client_admin_permissions.test"),
			},
		},
	})
}

func testAccCheckKeycloakOpenidClientAdminPermissionsExists(resourceName string) resource.TestCheckFunc {
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

		return nil
	}
}

func testKeycloakOpenidClientAdminPermissions_basic(scope string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client_admin_permissions" "test" {
	realm_id = data.keycloak_realm.realm.id
	name     = "test-client-permission"
	scopes   = ["%s"]
}
`, testAccRealmFGAPv2.Realm, scope)
}

func testKeycloakOpenidClientAdminPermissions_withScopes(clientName, groupName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

data "keycloak_openid_client" "admin_permissions" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = "admin-permissions"
}

resource "keycloak_openid_client" "target" {
	realm_id    = data.keycloak_realm.realm.id
	client_id   = "%s"
	access_type = "CONFIDENTIAL"
}

resource "keycloak_group" "group" {
	realm_id = data.keycloak_realm.realm.id
	name     = "%s"
}

resource "keycloak_openid_client_group_policy" "test" {
	realm_id           = data.keycloak_realm.realm.id
	resource_server_id = data.keycloak_openid_client.admin_permissions.id
	name               = "client_group_policy_test_%s"
	groups {
		id              = keycloak_group.group.id
		path            = keycloak_group.group.path
		extend_children = false
	}
	logic             = "POSITIVE"
	decision_strategy = "UNANIMOUS"
}

resource "keycloak_openid_client_admin_permissions" "test" {
	realm_id          = data.keycloak_realm.realm.id
	name              = "test-client-manage-%s"
	client_ids        = [keycloak_openid_client.target.id]
	scopes            = ["view", "manage"]
	policies          = [keycloak_openid_client_group_policy.test.id]
	description       = "test client permission with policy"
	decision_strategy = "UNANIMOUS"
}
`, testAccRealmFGAPv2.Realm, clientName, groupName, groupName, clientName)
}
