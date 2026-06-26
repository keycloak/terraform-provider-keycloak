package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakUsersAdminPermissions_basic(t *testing.T) {
	skipIfVersionIsLessThan(testCtx, t, keycloakClient, keycloak.Version_26_2)
	skipIfFGAPv2NotEnabled(testCtx, t, keycloakClient)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakUsersAdminPermissions_basic("view"),
				Check:  testAccCheckKeycloakUsersAdminPermissionsExists("keycloak_users_admin_permissions.test"),
			},
			{
				Config: testKeycloakUsersAdminPermissions_basic("manage"),
				Check:  testAccCheckKeycloakUsersAdminPermissionsExists("keycloak_users_admin_permissions.test"),
			},
			{
				ResourceName:      "keycloak_users_admin_permissions.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccKeycloakUsersAdminPermissions_withScopes(t *testing.T) {
	skipIfVersionIsLessThan(testCtx, t, keycloakClient, keycloak.Version_26_2)
	skipIfFGAPv2NotEnabled(testCtx, t, keycloakClient)

	groupName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakUsersAdminPermissions_withScopes(groupName),
				Check:  testAccCheckKeycloakUsersAdminPermissionsExists("keycloak_users_admin_permissions.test"),
			},
		},
	})
}

func testAccCheckKeycloakUsersAdminPermissionsExists(resourceName string) resource.TestCheckFunc {
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

func testKeycloakUsersAdminPermissions_basic(scope string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_users_admin_permissions" "test" {
	realm_id = data.keycloak_realm.realm.id
	name     = "test-users-permission"
	scopes   = ["%s"]
}
`, testAccRealmFGAPv2.Realm, scope)
}

func testKeycloakUsersAdminPermissions_withScopes(groupName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

data "keycloak_openid_client" "admin_permissions" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = "admin-permissions"
}

resource "keycloak_group" "group" {
	realm_id = data.keycloak_realm.realm.id
	name     = "%s"
}

resource "keycloak_openid_client_group_policy" "test" {
	realm_id           = data.keycloak_realm.realm.id
	resource_server_id = data.keycloak_openid_client.admin_permissions.id
	name               = "users_group_policy_test_%s"
	groups {
		id              = keycloak_group.group.id
		path            = keycloak_group.group.path
		extend_children = false
	}
	logic             = "POSITIVE"
	decision_strategy = "UNANIMOUS"
}

resource "keycloak_users_admin_permissions" "test" {
	realm_id          = data.keycloak_realm.realm.id
	name              = "test-users-manage-%s"
	scopes            = ["view", "manage"]
	policies          = [keycloak_openid_client_group_policy.test.id]
	description       = "test users permission with policy"
	decision_strategy = "UNANIMOUS"
}
`, testAccRealmFGAPv2.Realm, groupName, groupName, groupName)
}
