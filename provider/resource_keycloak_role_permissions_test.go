package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakRolePermissions_basic(t *testing.T) {
	roleName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRolePermissions_basic(roleName),
				Check:  testAccCheckKeycloakRolePermissionsExists("keycloak_role_permissions.test"),
			},
			{
				ResourceName:      "keycloak_role_permissions.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccKeycloakRolePermissions_withScopes(t *testing.T) {
	roleName := acctest.RandomWithPrefix("tf-acc")
	groupName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRolePermissions_withScopes(roleName, groupName),
				Check:  testAccCheckKeycloakRolePermissionsExists("keycloak_role_permissions.test"),
			},
		},
	})
}

func testAccCheckKeycloakRolePermissionsExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		permissions, err := getRolePermissionsFromState(s, resourceName)
		if err != nil {
			return err
		}

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		authorizationResourceServerId := rs.Primary.Attributes["authorization_resource_server_id"]

		var realmManagementId string
		clients, _ := keycloakClient.GetOpenidClients(testCtx, permissions.RealmId, false)
		for _, client := range clients {
			if client.ClientId == "realm-management" {
				realmManagementId = client.Id
				break
			}
		}

		if authorizationResourceServerId != realmManagementId {
			return fmt.Errorf("computed authorizationResourceServerId %s was not equal to %s (the id of the realm-management client)", authorizationResourceServerId, realmManagementId)
		}

		if !permissions.Enabled {
			return fmt.Errorf("expected role permissions to be enabled")
		}

		return nil
	}
}

func getRolePermissionsFromState(s *terraform.State, resourceName string) (*keycloak.RolePermissions, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	realmId := rs.Primary.Attributes["realm_id"]
	roleId := rs.Primary.Attributes["role_id"]

	permissions, err := keycloakClient.GetRolePermissions(testCtx, realmId, roleId)
	if err != nil {
		return nil, fmt.Errorf("error getting role permissions with realm id %s and role id %s: %s", realmId, roleId, err)
	}

	return permissions, nil
}

func testKeycloakRolePermissions_basic(roleName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_role" "role" {
	realm_id = data.keycloak_realm.realm.id
	name     = "%s"
}

resource "keycloak_role_permissions" "test" {
	realm_id = data.keycloak_realm.realm.id
	role_id  = keycloak_role.role.id
}
`, testAccRealm.Realm, roleName)
}

func testKeycloakRolePermissions_withScopes(roleName, groupName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

data "keycloak_openid_client" "realm_management" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = "realm-management"
}

resource "keycloak_openid_client_permissions" "realm_management_permission" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = data.keycloak_openid_client.realm_management.id
}

resource "keycloak_role" "role" {
	realm_id = data.keycloak_realm.realm.id
	name     = "%s"
}

resource "keycloak_group" "group" {
	realm_id = data.keycloak_realm.realm.id
	name     = "%s"
}

resource "keycloak_openid_client_group_policy" "test" {
	realm_id           = data.keycloak_realm.realm.id
	resource_server_id = data.keycloak_openid_client.realm_management.id
	name               = "role_group_policy_test"
	groups {
		id              = keycloak_group.group.id
		path            = keycloak_group.group.path
		extend_children = false
	}
	logic             = "POSITIVE"
	decision_strategy = "UNANIMOUS"
	depends_on = [
		keycloak_openid_client_permissions.realm_management_permission,
	]
}

resource "keycloak_role_permissions" "test" {
	realm_id = data.keycloak_realm.realm.id
	role_id  = keycloak_role.role.id
	map_role_scope {
		policies          = [keycloak_openid_client_group_policy.test.id]
		description       = "map_role_scope"
		decision_strategy = "UNANIMOUS"
	}
}
`, testAccRealm.Realm, roleName, groupName)
}
