package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakGroupPermission_basic(t *testing.T) {
	groupName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakGroupPermission_basic(groupName),
				Check:  testAccCheckKeycloakGroupPermissionExists("keycloak_group_permissions.test"),
			},
			{
				ResourceName:      "keycloak_group_permissions.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckKeycloakGroupPermissionExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		permissions, err := getGroupPermissionsFromState(s, resourceName)
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
		// manage_members_scope
		manageMembersScopePolicyId := rs.Primary.Attributes["manage_members_scope.0.policies.0"]
		manageMembersScopeDescription := rs.Primary.Attributes["manage_members_scope.0.description"]
		manageMembersScopeDecisionStrategy := rs.Primary.Attributes["manage_members_scope.0.decision_strategy"]

		authzClientManageMembersScope, err := keycloakClient.GetOpenidClientAuthorizationPermission(testCtx, permissions.RealmId, realmManagementId, permissions.ScopePermissions["manage-members"].(string))
		if err != nil {
			return err
		}
		policyId := authzClientManageMembersScope.Policies[0]

		if manageMembersScopePolicyId != policyId {
			return fmt.Errorf("computed manageMembersScopePolicyId %s was not equal to policyId %s", manageMembersScopePolicyId, policyId)
		}

		if authzClientManageMembersScope.Description != manageMembersScopeDescription {
			return fmt.Errorf("DecisionStrategy %s was not equal to %s", authzClientManageMembersScope.DecisionStrategy, manageMembersScopeDescription)
		}

		if authzClientManageMembersScope.DecisionStrategy != manageMembersScopeDecisionStrategy {
			return fmt.Errorf("DecisionStrategy %s was not equal to %s", authzClientManageMembersScope.DecisionStrategy, manageMembersScopeDecisionStrategy)
		}

		manageMembershipScope := rs.Primary.Attributes["manage_membership_scope"]

		if manageMembershipScope != "" {
			return fmt.Errorf("manage_membership_scope found")
		}

		return nil
	}
}

func getGroupPermissionsFromState(s *terraform.State, resourceName string) (*keycloak.GroupPermissions, error) {
	keycloakClient := testAccProvider.Meta().(*keycloak.KeycloakClient)

	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	realmId := rs.Primary.Attributes["realm_id"]
	groupId := rs.Primary.Attributes["group_id"]

	permissions, err := keycloakClient.GetGroupPermissions(testCtx, realmId, groupId)
	if err != nil {
		return nil, fmt.Errorf("error getting group permissions with realm id %s and group id %s : %s", realmId, groupId, err)
	}

	return permissions, nil
}

func testKeycloakGroupPermission_basic(groupName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

data "keycloak_openid_client" "realm_management" {
  realm_id  = data.keycloak_realm.realm.id
  client_id = "realm-management"
}

resource "keycloak_openid_client_permissions" "realm-management_permission" {
	realm_id   = data.keycloak_realm.realm.id
	client_id  = data.keycloak_openid_client.realm_management.id
}

resource "keycloak_group" "group" {
	realm_id = data.keycloak_realm.realm.id
	name     = "%s"
}

resource "keycloak_openid_client_group_policy" "test" {
	realm_id           = data.keycloak_realm.realm.id
	resource_server_id = data.keycloak_openid_client.realm_management.id
	name 			   = "client_group_policy_test"
	groups {
		id              = keycloak_group.group.id
		path            = keycloak_group.group.path
		extend_children = false
	}
	logic             = "POSITIVE"
	decision_strategy = "UNANIMOUS"
	depends_on = [
		keycloak_openid_client_permissions.realm-management_permission,
	]
}

resource "keycloak_group_permissions" "test" {
	realm_id                               = data.keycloak_realm.realm.id
	group_id                               = keycloak_group.group.id
	manage_members_scope {
		policies          = [
			keycloak_openid_client_group_policy.test.id
		]
		description       = "mangage_members_scope"
		decision_strategy = "UNANIMOUS"
	}

}
	`, testAccRealm.Realm, groupName)
}
