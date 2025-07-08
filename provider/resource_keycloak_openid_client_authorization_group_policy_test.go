package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakOpenidClientAuthorizationGroupPolicy(t *testing.T) {
	t.Parallel()

	clientId := acctest.RandomWithPrefix("tf-acc")

	var policyId string

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testResourceKeycloakOpenidClientAuthorizationGroupPolicy_basic(clientId),
				Check:  testResourceKeycloakOpenidClientAuthorizationGroupPolicyExists("keycloak_openid_client_group_policy.test", &policyId),
			},
			// we need a separate test step to verify destroy, since destroying the client will always lead to destroying the group policy
			{
				Config: testResourceKeycloakOpenidClientAuthorizationGroupPolicy_basicDestroy(clientId),
				Check:  testResourceKeycloakOpenidClientHasNoAuthorizationGroupPolicy("keycloak_openid_client.test", policyId),
			},
		},
	})
}

func getResourceKeycloakOpenidClientAuthorizationGroupPolicyFromState(s *terraform.State, resourceName string) (*keycloak.OpenidClientAuthorizationGroupPolicy, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	realm := rs.Primary.Attributes["realm_id"]
	resourceServerId := rs.Primary.Attributes["resource_server_id"]
	policyId := rs.Primary.ID

	policy, err := keycloakClient.GetOpenidClientAuthorizationGroupPolicy(testCtx, realm, resourceServerId, policyId)
	if err != nil {
		return nil, fmt.Errorf("error getting openid client auth role policy config with alias %s: %s", resourceServerId, err)
	}

	return policy, nil
}

func testResourceKeycloakOpenidClientHasNoAuthorizationGroupPolicy(resourceName string, policyId string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		realm := rs.Primary.Attributes["realm_id"]
		resourceServerId := rs.Primary.Attributes["resource_server_id"]

		_, err := keycloakClient.GetOpenidClientAuthorizationGroupPolicy(testCtx, realm, resourceServerId, policyId)
		if err == nil {
			return fmt.Errorf("policy with id %s still exists", policyId)
		}

		return nil
	}
}

func testResourceKeycloakOpenidClientAuthorizationGroupPolicyExists(resourceName string, policyId *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		policy, err := getResourceKeycloakOpenidClientAuthorizationGroupPolicyFromState(s, resourceName)

		if err != nil {
			return err
		}

		policyId = &policy.Id

		return nil
	}
}

func testResourceKeycloakOpenidClientAuthorizationGroupPolicy_basic(clientId string) string {
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

resource "keycloak_group" "test" {
	realm_id = data.keycloak_realm.realm.id
	name     = "foo"
}

resource keycloak_openid_client_group_policy test {
	resource_server_id = "${keycloak_openid_client.test.resource_server_id}"
	realm_id = data.keycloak_realm.realm.id
	name = "client_group_policy_test"
	groups {
		id = "${keycloak_group.test.id}"
		path = "${keycloak_group.test.path}"
		extend_children = false
	}
	logic = "POSITIVE"
	decision_strategy = "UNANIMOUS"
}
	`, testAccRealm.Realm, clientId)
}

func testResourceKeycloakOpenidClientAuthorizationGroupPolicy_basicDestroy(clientId string) string {
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
	`, testAccRealm.Realm, clientId)
}
