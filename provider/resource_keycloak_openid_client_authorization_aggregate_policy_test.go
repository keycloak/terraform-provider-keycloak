package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakOpenidClientAuthorizationAggregatePolicy(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testResourceKeycloakOpenidClientAuthorizationAggregatePolicyDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testResourceKeycloakOpenidClientAuthorizationAggregatePolicy_basic(clientId),
				Check:  testResourceKeycloakOpenidClientAuthorizationAggregatePolicyExists("keycloak_openid_client_aggregate_policy.test"),
			},
		},
	})
}

func getResourceKeycloakOpenidClientAuthorizationAggregatePolicyFromState(s *terraform.State, resourceName string) (*keycloak.OpenidClientAuthorizationAggregatePolicy, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	realm := rs.Primary.Attributes["realm_id"]
	resourceServerId := rs.Primary.Attributes["resource_server_id"]
	policyId := rs.Primary.ID

	policy, err := keycloakClient.GetOpenidClientAuthorizationAggregatePolicy(testCtx, realm, resourceServerId, policyId)
	if err != nil {
		return nil, fmt.Errorf("error getting openid client auth aggregate policy config with alias %s: %s", resourceServerId, err)
	}

	return policy, nil
}

func testResourceKeycloakOpenidClientAuthorizationAggregatePolicyDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_openid_client_aggregate_policy" {
				continue
			}

			realm := rs.Primary.Attributes["realm"]
			resourceServerId := rs.Primary.Attributes["resource_server_id"]
			policyId := rs.Primary.ID

			policy, _ := keycloakClient.GetOpenidClientAuthorizationAggregatePolicy(testCtx, realm, resourceServerId, policyId)
			if policy != nil {
				return fmt.Errorf("policy config with id %s still exists", policyId)
			}
		}

		return nil
	}
}

func testResourceKeycloakOpenidClientAuthorizationAggregatePolicyExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getResourceKeycloakOpenidClientAuthorizationAggregatePolicyFromState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testResourceKeycloakOpenidClientAuthorizationAggregatePolicy_basic(clientId string) string {
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

resource "keycloak_role" "test" {
	realm_id    = data.keycloak_realm.realm.id
	name        = "aggregate_policy_role"
}

resource keycloak_openid_client_role_policy test {
	resource_server_id = "${keycloak_openid_client.test.resource_server_id}"
	realm_id = data.keycloak_realm.realm.id
	name = "keycloak_openid_client_role_policy"
	decision_strategy = "UNANIMOUS"
	logic = "POSITIVE"
	type = "role"
	role  {
		id = "${keycloak_role.test.id}"
		required = false
	}
}

resource keycloak_openid_client_aggregate_policy test {
	resource_server_id = "${keycloak_openid_client.test.resource_server_id}"
	realm_id = data.keycloak_realm.realm.id
	name = "keycloak_openid_client_aggregate_policy"
	decision_strategy = "UNANIMOUS"
	logic = "POSITIVE"
	policies = ["${keycloak_openid_client_role_policy.test.id}"]
}
	`, testAccRealm.Realm, clientId)
}
