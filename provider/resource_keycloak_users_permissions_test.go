package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakUsersPermission_basic(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")
	username := acctest.RandomWithPrefix("tf-acc")
	email := acctest.RandomWithPrefix("tf-acc") + "@fakedomain.com"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakUsersPermission_basic(realmName, username, email),
				Check:  testAccCheckKeycloakUsersPermissionExists("keycloak_users_permissions.my_permission"),
			},
			{
				ResourceName:      "keycloak_users_permissions.my_permission",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     realmName,
			},
			// check destroy
			{
				Config: testKeycloakUsersPermission_checkDestroy(realmName, username, email),
				Check:  testAccCheckKeycloakUsersPermissionsAreDisabled(realmName),
			},
		},
	})
}

func testAccCheckKeycloakUsersPermissionExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		permissions, err := getUsersPermissionsFromState(s, resourceName)
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

		viewScopePolicyId := rs.Primary.Attributes["view_scope.0.policies.0"]
		viewScopeDescription := rs.Primary.Attributes["view_scope.0.description"]
		viewScopeDecisionStrategy := rs.Primary.Attributes["view_scope.0.decision_strategy"]

		authzClientView, err := keycloakClient.GetOpenidClientAuthorizationPermission(testCtx, permissions.RealmId, realmManagementId, permissions.ScopePermissions["view"])
		if err != nil {
			return err
		}
		policyId := authzClientView.Policies[0]

		if viewScopePolicyId != policyId {
			return fmt.Errorf("computed view scope policy ID %s was not equal to %s", viewScopePolicyId, policyId)
		}

		if authzClientView.Description != viewScopeDescription {
			return fmt.Errorf("description %s was not equal to %s", authzClientView.DecisionStrategy, viewScopeDescription)
		}

		if authzClientView.DecisionStrategy != viewScopeDecisionStrategy {
			return fmt.Errorf("decision strategy %s was not equal to %s", authzClientView.DecisionStrategy, viewScopeDecisionStrategy)
		}

		authzClientManage, err := keycloakClient.GetOpenidClientAuthorizationPermission(testCtx, permissions.RealmId, realmManagementId, permissions.ScopePermissions["manage"])
		if err != nil {
			return err
		}
		policies := make([]interface{}, len(authzClientManage.Policies))
		for i := range authzClientManage.Policies {
			policies[i] = authzClientManage.Policies[i]
		}

		policyId = rs.Primary.Attributes["manage_scope.0.policies.0"]
		if !Contains(policies, policyId) {
			return fmt.Errorf("computed viewScopePolicyId %s was not equal to policyId %s", viewScopePolicyId, policyId)
		}
		policyId = rs.Primary.Attributes["manage_scope.0.policies.1"]
		if !Contains(policies, policyId) {
			return fmt.Errorf("computed viewScopePolicyId %s was not equal to policyId %s", viewScopePolicyId, policyId)
		}

		mapRolesScope := rs.Primary.Attributes["map_roles_scope"]

		if mapRolesScope != "" {
			return fmt.Errorf("map_roles_scope found")
		}

		return nil
	}
}

func testAccCheckKeycloakUsersPermissionsAreDisabled(realmId string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		permissions, err := keycloakClient.GetUsersPermissions(testCtx, realmId)
		if err != nil {
			return fmt.Errorf("error getting users permissions with realm id %s: %s", realmId, err)
		}

		if permissions.Enabled != false {
			return fmt.Errorf("expected users permission in Keycloak to be disabled")
		}

		return nil
	}
}

func getUsersPermissionsFromState(s *terraform.State, resourceName string) (*keycloak.UsersPermissions, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	realmId := rs.Primary.Attributes["realm_id"]

	permissions, err := keycloakClient.GetUsersPermissions(testCtx, realmId)
	if err != nil {
		return nil, fmt.Errorf("error getting users permissions with realm id %s: %s", realmId, err)

	}
	return permissions, nil
}

func testKeycloakUsersPermission_basic(realmId, username, email string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm = "%s"
}

data "keycloak_openid_client" "realm_management" {
	realm_id  = keycloak_realm.realm.id
	client_id = "realm-management"
}

resource "keycloak_openid_client_permissions" "realm_management_permission" {
	realm_id   = keycloak_realm.realm.id
	client_id  = data.keycloak_openid_client.realm_management.id
}

resource "keycloak_user" "test" {
	realm_id = keycloak_realm.realm.id
	username = "%s"

	email      = "%s"
	first_name = "Testy"
	last_name  = "Tester"
}

resource "keycloak_openid_client_user_policy" "test" {
	realm_id           = keycloak_realm.realm.id
	resource_server_id = data.keycloak_openid_client.realm_management.id
	name               = "client_user_policy_test"

	users             = [
		keycloak_user.test.id
	]
	logic             = "POSITIVE"
	decision_strategy = "UNANIMOUS"

	depends_on = [
		keycloak_openid_client_permissions.realm_management_permission,
	]
}
resource "keycloak_openid_client_user_policy" "test2" {
	realm_id           = keycloak_realm.realm.id
	resource_server_id = data.keycloak_openid_client.realm_management.id
	name               = "client_user_policy_test2"

	users             = [
		keycloak_user.test.id
	]
	logic             = "POSITIVE"
	decision_strategy = "UNANIMOUS"

	depends_on = [
		keycloak_openid_client_permissions.realm_management_permission,
	]
}

resource "keycloak_users_permissions" "my_permission" {
	realm_id = keycloak_realm.realm.id

	view_scope {
		policies          = [
			keycloak_openid_client_user_policy.test.id
		]
		description       = "view_scope"
		decision_strategy = "CONSENSUS"
	}

	manage_scope {
		policies          = [
			keycloak_openid_client_user_policy.test.id,
			keycloak_openid_client_user_policy.test2.id,
		]
		description       = "manage_scope"
		decision_strategy = "UNANIMOUS"
	}
}
	`, realmId, username, email)
}

func testKeycloakUsersPermission_checkDestroy(realmId, username, email string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm = "%s"
}

data "keycloak_openid_client" "realm_management" {
	realm_id  = keycloak_realm.realm.id
	client_id = "realm-management"
}

resource "keycloak_openid_client_permissions" "realm_management_permission" {
	realm_id   = keycloak_realm.realm.id
	client_id  = data.keycloak_openid_client.realm_management.id
}

resource "keycloak_user" "test" {
	realm_id = keycloak_realm.realm.id
	username = "%s"

	email      = "%s"
	first_name = "Testy"
	last_name  = "Tester"
}

resource "keycloak_openid_client_user_policy" "test" {
	realm_id           = keycloak_realm.realm.id
	resource_server_id = data.keycloak_openid_client.realm_management.id
	name               = "client_user_policy_test"

	users             = [
		keycloak_user.test.id
	]
	logic             = "POSITIVE"
	decision_strategy = "UNANIMOUS"

	depends_on = [
		keycloak_openid_client_permissions.realm_management_permission,
	]
}
resource "keycloak_openid_client_user_policy" "test2" {
	realm_id           = keycloak_realm.realm.id
	resource_server_id = data.keycloak_openid_client.realm_management.id
	name               = "client_user_policy_test2"

	users             = [
		keycloak_user.test.id
	]
	logic             = "POSITIVE"
	decision_strategy = "UNANIMOUS"

	depends_on = [
		keycloak_openid_client_permissions.realm_management_permission,
	]
}
	`, realmId, username, email)
}
