package provider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakOpenidClientAuthorizationClientScopePolicy_basic(t *testing.T) {
	t.Parallel()

	clientId := acctest.RandomWithPrefix("tf-acc")
	clientScopeName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testResourceKeycloakOpenidClientAuthorizationClientScopePolicyDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testResourceKeycloakOpenidClientAuthorizationClientScopePolicy_basic(clientScopeName, clientId),
				Check:  testResourceKeycloakOpenidClientAuthorizationClientScopePolicyExists("keycloak_openid_client_authorization_client_scope_policy.test"),
			},
		},
	})
}

func TestAccKeycloakOpenidClientAuthorizationClientScopePolicy_multiple(t *testing.T) {
	t.Parallel()

	clientId := acctest.RandomWithPrefix("tf-acc")
	var clientScopeNames []string
	for i := 0; i < acctest.RandIntRange(7, 12); i++ {
		clientScopeNames = append(clientScopeNames, acctest.RandomWithPrefix("tf-acc"))
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testResourceKeycloakOpenidClientAuthorizationClientScopePolicyDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testResourceKeycloakOpenidClientAuthorizationClientScopePolicy_multipleClientScopes(clientScopeNames, clientId),
				Check:  testResourceKeycloakOpenidClientAuthorizationClientScopePolicyExists("keycloak_openid_client_authorization_client_scope_policy.test"),
			},
		},
	})
}

func getResourceKeycloakOpenidClientAuthorizationClientScopePolicyFromState(s *terraform.State, resourceName string) (*keycloak.OpenidClientAuthorizationClientScopePolicy, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	realm := rs.Primary.Attributes["realm_id"]
	resourceServerId := rs.Primary.Attributes["resource_server_id"]
	policyId := rs.Primary.ID

	policy, err := keycloakClient.GetOpenidClientAuthorizationClientScopePolicy(testCtx, realm, resourceServerId, policyId)
	if err != nil {
		return nil, fmt.Errorf("error getting openid client auth client scope policy config with alias %s: %s", resourceServerId, err)
	}

	return policy, nil
}

func testResourceKeycloakOpenidClientAuthorizationClientScopePolicyDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_openid_client_authorization_client_scope_policy" {
				continue
			}

			realm := rs.Primary.Attributes["realm_id"]
			resourceServerId := rs.Primary.Attributes["resource_server_id"]
			policyId := rs.Primary.ID

			policy, _ := keycloakClient.GetOpenidClientAuthorizationClientScopePolicy(testCtx, realm, resourceServerId, policyId)
			if policy != nil {
				return fmt.Errorf("policy config with id %s still exists", policyId)
			}
		}

		return nil
	}
}

func testResourceKeycloakOpenidClientAuthorizationClientScopePolicyExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getResourceKeycloakOpenidClientAuthorizationClientScopePolicyFromState(s, resourceName)

		if err != nil {
			return err
		}

		return nil
	}
}

func testResourceKeycloakOpenidClientAuthorizationClientScopePolicy_basic(clientScopeName, clientId string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "test" {
	client_id                = "%s"
	realm_id                 = data.keycloak_realm.realm.id
	access_type              = "CONFIDENTIAL"
	service_accounts_enabled = true
	authorization {
		policy_enforcement_mode = "ENFORCING"
	}
}

resource "keycloak_openid_client_scope" "test" {
    realm_id               = data.keycloak_realm.realm.id
    name                   = "%s"
    description            = "test"
}

resource "keycloak_openid_client_authorization_client_scope_policy" "test" {
    resource_server_id = keycloak_openid_client.test.resource_server_id
    realm_id           = data.keycloak_realm.realm.id
    name               = "keycloak_openid_client_authorization_client_scope_policy"
    description        = "test"
    decision_strategy  = "AFFIRMATIVE"
    logic              = "POSITIVE"

    scope {
      id       = keycloak_openid_client_scope.test.id
      required = false
    }
}
	`, testAccRealm.Realm, clientScopeName, clientId)
}

func testResourceKeycloakOpenidClientAuthorizationClientScopePolicy_multipleClientScopes(clientScopeNames []string, clientId string) string {
	var (
		clientScopes        strings.Builder
		clientScopePolicies strings.Builder
	)
	for i, clientScopeName := range clientScopeNames {
		clientScopes.WriteString(fmt.Sprintf(`
resource "keycloak_openid_client_scope" "scope_%d" {
	realm_id    = data.keycloak_realm.realm.id
	name        = "%s"
}
`, i, clientScopeName))
		clientScopePolicies.WriteString(fmt.Sprintf(`
	scope  {
		id = keycloak_openid_client_scope.scope_%d.id
		required = false
	}
`, i))
	}

	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "test" {
	client_id                = "%s"
	realm_id                 = data.keycloak_realm.realm.id
	access_type              = "CONFIDENTIAL"
	service_accounts_enabled = true
	authorization {
		policy_enforcement_mode = "ENFORCING"
	}
}

%s

resource "keycloak_openid_client_authorization_client_scope_policy" "test" {
    resource_server_id = keycloak_openid_client.test.resource_server_id
    realm_id           = data.keycloak_realm.realm.id
    name               = "keycloak_openid_client_authorization_client_scope_policy"
    description        = "test"
    decision_strategy  = "AFFIRMATIVE"
    logic              = "POSITIVE"

%s

}
	`, testAccRealm.Realm, clientId, clientScopes.String(), clientScopePolicies.String())
}
