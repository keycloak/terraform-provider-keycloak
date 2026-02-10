package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakOpenidClientAuthorizationRegexPolicy(t *testing.T) {
	skipIfVersionIsLessThan(testCtx, t, keycloakClient, keycloak.Version_24)

	clientId := acctest.RandomWithPrefix("tf-acc")
	pattern := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testResourceKeycloakOpenidClientAuthorizationRegexPolicyDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testResourceKeycloakOpenidClientAuthorizationRegexPolicy_basic(clientId, pattern, "sample-claim", true),
				Check:  testResourceKeycloakOpenidClientAuthorizationRegexPolicyExists("keycloak_openid_client_regex_policy.test"),
			},
		},
	})
}

func getResourceKeycloakOpenidClientAuthorizationRegexPolicyFromState(s *terraform.State, resourceName string) (*keycloak.OpenidClientAuthorizationRegexPolicy, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	realm := rs.Primary.Attributes["realm_id"]
	resourceServerId := rs.Primary.Attributes["resource_server_id"]
	policyId := rs.Primary.ID

	policy, err := keycloakClient.GetOpenidClientAuthorizationRegexPolicy(testCtx, realm, resourceServerId, policyId)
	if err != nil {
		return nil, fmt.Errorf("error getting openid client auth regex policy config with alias %s: %s", resourceServerId, err)
	}

	return policy, nil
}

func testResourceKeycloakOpenidClientAuthorizationRegexPolicyDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_openid_client_regex_policy" {
				continue
			}

			realm := rs.Primary.Attributes["realm_id"]
			resourceServerId := rs.Primary.Attributes["resource_server_id"]
			policyId := rs.Primary.ID

			policy, _ := keycloakClient.GetOpenidClientAuthorizationRegexPolicy(testCtx, realm, resourceServerId, policyId)
			if policy != nil {
				return fmt.Errorf("policy config with id %s still exists", policyId)
			}
		}

		return nil
	}
}

func testResourceKeycloakOpenidClientAuthorizationRegexPolicyExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getResourceKeycloakOpenidClientAuthorizationRegexPolicyFromState(s, resourceName)

		if err != nil {
			return err
		}

		return nil
	}
}

func testResourceKeycloakOpenidClientAuthorizationRegexPolicy_basic(clientId, pattern, targetClaim string, targetContextAttributes bool) string {
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

	resource keycloak_openid_client_regex_policy test {
		resource_server_id = "${keycloak_openid_client.test.resource_server_id}"
		realm_id = data.keycloak_realm.realm.id
		name = "client_regex_policy_test"
		pattern = "%s"
		logic = "POSITIVE"
		decision_strategy = "UNANIMOUS"
        target_claim = "%s"
        target_context_attributes = %t
	}
	`, testAccRealm.Realm, clientId, pattern, targetClaim, targetContextAttributes)
}
