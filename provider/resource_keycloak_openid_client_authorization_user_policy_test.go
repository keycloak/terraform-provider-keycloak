package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakOpenidClientAuthorizationUserPolicy(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	username := acctest.RandomWithPrefix("tf-acc")
	email := acctest.RandomWithPrefix("tf-acc") + "@fakedomain.com"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testResourceKeycloakOpenidClientAuthorizationUserPolicyDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testResourceKeycloakOpenidClientAuthorizationUserPolicy_basic(clientId, username, email),
				Check:  testResourceKeycloakOpenidClientAuthorizationUserPolicyExists("keycloak_openid_client_user_policy.test"),
			},
		},
	})
}

func getResourceKeycloakOpenidClientAuthorizationUserPolicyFromState(s *terraform.State, resourceName string) (*keycloak.OpenidClientAuthorizationUserPolicy, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	realm := rs.Primary.Attributes["realm_id"]
	resourceServerId := rs.Primary.Attributes["resource_server_id"]
	policyId := rs.Primary.ID

	policy, err := keycloakClient.GetOpenidClientAuthorizationUserPolicy(testCtx, realm, resourceServerId, policyId)
	if err != nil {
		return nil, fmt.Errorf("error getting openid client auth role policy config with alias %s: %s", resourceServerId, err)
	}

	return policy, nil
}

func testResourceKeycloakOpenidClientAuthorizationUserPolicyDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_openid_client_user_policy" {
				continue
			}

			realm := rs.Primary.Attributes["realm_id"]
			resourceServerId := rs.Primary.Attributes["resource_server_id"]
			policyId := rs.Primary.ID

			policy, _ := keycloakClient.GetOpenidClientAuthorizationUserPolicy(testCtx, realm, resourceServerId, policyId)
			if policy != nil {
				return fmt.Errorf("policy config with id %s still exists", policyId)
			}
		}

		return nil
	}
}

func testResourceKeycloakOpenidClientAuthorizationUserPolicyExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getResourceKeycloakOpenidClientAuthorizationUserPolicyFromState(s, resourceName)

		if err != nil {
			return err
		}

		return nil
	}
}

func testResourceKeycloakOpenidClientAuthorizationUserPolicy_basic(clientId, username, email string) string {
	return fmt.Sprintf(`
	data "keycloak_realm" "realm" {
		realm = "%s"
	}

	resource keycloak_openid_client test {
		client_id                = "%s"
		realm_id                 = data.keycloak_realm.realm.id
		access_type              = "CONFIDENTIAL"
		service_accounts_enabled = true
		authorization {
			policy_enforcement_mode = "ENFORCING"
		}
	}

	resource keycloak_user test {
		realm_id = data.keycloak_realm.realm.id
		username = "%s"

		email      = "%s"
		first_name = "Testy"
		last_name  = "Tester"
	}

	resource keycloak_openid_client_user_policy test {
		resource_server_id = "${keycloak_openid_client.test.resource_server_id}"
		realm_id = data.keycloak_realm.realm.id
		name = "client_user_policy_test"
		users = ["${keycloak_user.test.id}"]
		logic = "POSITIVE"
		decision_strategy = "UNANIMOUS"
	}
	`, testAccRealm.Realm, clientId, username, email)
}
