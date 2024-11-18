package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakOpenidClientPermission_basic(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	username := acctest.RandomWithPrefix("tf-acc")
	email := acctest.RandomWithPrefix("tf-acc") + "@fakedomain.com"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClientPermission_basic(clientId, username, email),
				Check:  testAccCheckKeycloakOpenidClientPermissionExists("keycloak_openid_client_permissions.my_permission"),
			},
			{
				Config: testKeycloakOpenidClientPermissionDelete_basic(clientId, username, email),
				Check:  testAccCheckKeycloakOpenidClientPermissionsAreDisabled(clientId),
			},
		},
	})
}

func testAccCheckKeycloakOpenidClientPermissionExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		permissions, err := getOpenidClientPermissionsFromState(s, resourceName)
		if err != nil {
			return err
		}

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		authorizationResourceServerId := rs.Primary.Attributes["authorization_resource_server_id"]
		viewScopePolicyId := rs.Primary.Attributes["view_scope.0.policies.0"]
		viewScopeDescription := rs.Primary.Attributes["view_scope.0.description"]
		viewScopeDecisionStrategy := rs.Primary.Attributes["view_scope.0.decision_strategy"]

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

		authzClientView, err := keycloakClient.GetOpenidClientAuthorizationPermission(testCtx, permissions.RealmId, realmManagementId, permissions.ScopePermissions["view"])
		if err != nil {
			return err
		}

		if viewScopePolicyId != authzClientView.Policies[0] {
			return fmt.Errorf("computed view scope policy ID %s was not equal to %s", viewScopePolicyId, authzClientView.Policies[0])
		}
		if authzClientView.Description != viewScopeDescription {
			return fmt.Errorf("description %s was not equal to %s", authzClientView.DecisionStrategy, viewScopeDescription)
		}
		if authzClientView.DecisionStrategy != viewScopeDecisionStrategy {
			return fmt.Errorf("decision strategy %s was not equal to %s", authzClientView.DecisionStrategy, viewScopeDecisionStrategy)
		}

		return nil
	}
}

func testAccCheckKeycloakOpenidClientPermissionsAreDisabled(clientId string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := keycloakClient.GetOpenidClientByClientId(testCtx, testAccRealm.Realm, clientId)
		if err != nil {
			return err
		}

		permissions, err := keycloakClient.GetOpenidClientPermissions(testCtx, testAccRealm.Realm, client.Id)
		if err != nil {
			return fmt.Errorf("error getting openid_client permissions with realm id %s and client id %s: %s", testAccRealm.Realm, clientId, err)
		}

		if permissions.Enabled != false {
			return fmt.Errorf("expected openid client permission in Keycloak to be disabled")
		}

		return nil
	}
}

func getOpenidClientPermissionsFromState(s *terraform.State, resourceName string) (*keycloak.OpenidClientPermissions, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	realmId := rs.Primary.Attributes["realm_id"]
	clientId := rs.Primary.Attributes["client_id"]

	permissions, err := keycloakClient.GetOpenidClientPermissions(testCtx, testAccRealm.Realm, clientId)
	if err != nil {
		return nil, fmt.Errorf("error getting openid_client permissions with realm id %s and client id %s: %s", realmId, clientId, err)
	}

	return permissions, nil
}

func testKeycloakOpenidClientPermission_basic(clientId, username, email string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "openid_client" {
  realm_id              = data.keycloak_realm.realm.id
  name                  = "my_openid_client"
  client_id             = "%s"
  client_secret         = "secret"
  access_type           = "CONFIDENTIAL"
  standard_flow_enabled = true
  valid_redirect_uris   = [
    "http://localhost:8080/*",
  ]
}

data "keycloak_openid_client" "realm_management" {
  realm_id  = data.keycloak_realm.realm.id
  client_id = "realm-management"
}


resource keycloak_openid_client_permissions "realm-management_permission" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = data.keycloak_openid_client.realm_management.id
}

resource keycloak_user test {
	realm_id = data.keycloak_realm.realm.id
	username = "%s"

	email      = "%s"
	first_name = "Testy"
	last_name  = "Tester"
}

resource keycloak_openid_client_user_policy test {
	realm_id           = data.keycloak_realm.realm.id
	resource_server_id = data.keycloak_openid_client.realm_management.id

	name  = "client_user_policy_test"
	users = [
		keycloak_user.test.id
	]

	logic             = "POSITIVE"
	decision_strategy = "UNANIMOUS"

	depends_on = [
		keycloak_openid_client_permissions.realm-management_permission,
	]
}

resource "keycloak_openid_client_permissions" "my_permission" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = keycloak_openid_client.openid_client.id

	view_scope {
		policies          = [
			keycloak_openid_client_user_policy.test.id
		]
		description       = "view_scope"
		decision_strategy = "CONSENSUS"
	}
}`, testAccRealm.Realm, clientId, username, email)
}

func testKeycloakOpenidClientPermissionDelete_basic(clientId, username, email string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "openid_client" {
  realm_id              = data.keycloak_realm.realm.id
  name                  = "my_openid_client"
  client_id             = "%s"
  client_secret         = "secret"
  access_type           = "CONFIDENTIAL"
  standard_flow_enabled = true
  valid_redirect_uris   = [
    "http://localhost:8080/*",
  ]
}

data "keycloak_openid_client" "realm_management" {
  realm_id  = data.keycloak_realm.realm.id
  client_id = "realm-management"
}


resource keycloak_openid_client_permissions "realm-management_permission" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = data.keycloak_openid_client.realm_management.id
}

resource keycloak_user test {
	realm_id = data.keycloak_realm.realm.id
	username = "%s"

	email      = "%s"
	first_name = "Testy"
	last_name  = "Tester"
}

resource keycloak_openid_client_user_policy test {
	realm_id           = data.keycloak_realm.realm.id
	resource_server_id = data.keycloak_openid_client.realm_management.id

	name  = "client_user_policy_test"
	users = [
		keycloak_user.test.id
	]

	logic             = "POSITIVE"
	decision_strategy = "UNANIMOUS"

	depends_on = [
		keycloak_openid_client_permissions.realm-management_permission,
	]
}`, testAccRealm.Realm, clientId, username, email)
}
