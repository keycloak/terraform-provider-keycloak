package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakGenericClientAuthorizationPolicy(t *testing.T) {
	t.Parallel()

	clientId := acctest.RandomWithPrefix("tf-acc")
	policyName := acctest.RandomWithPrefix("tf-acc")
	// The deployed JavaScript policy provided by custom-authz-policy-example. For deployed
	// scripts Keycloak generates the type as "script-" + the fileName declared in
	// META-INF/keycloak-scripts.json. A policy implemented as a Java SPI would instead use
	// the provider id returned by its PolicyProviderFactory.getId().
	policyType := "script-always-granting-policy.js"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testResourceKeycloakGenericClientAuthorizationPolicyDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testResourceKeycloakGenericClientAuthorizationPolicy_basic(clientId, policyName, policyType),
				Check:  testResourceKeycloakGenericClientAuthorizationPolicyExists("keycloak_generic_client_authorization_policy.test"),
			},
			{
				ResourceName:      "keycloak_generic_client_authorization_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getResourceKeycloakGenericClientAuthorizationPolicyImportId("keycloak_generic_client_authorization_policy.test"),
			},
		},
	})
}

func getResourceKeycloakGenericClientAuthorizationPolicyFromState(s *terraform.State, resourceName string) (*keycloak.GenericClientAuthorizationPolicy, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	realm := rs.Primary.Attributes["realm_id"]
	resourceServerId := rs.Primary.Attributes["resource_server_id"]
	policyId := rs.Primary.ID

	policy, err := keycloakClient.GetGenericClientAuthorizationPolicy(testCtx, realm, resourceServerId, policyId)
	if err != nil {
		return nil, fmt.Errorf("error getting generic client authorization policy with id %s: %s", policyId, err)
	}

	return policy, nil
}

func getResourceKeycloakGenericClientAuthorizationPolicyImportId(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}

		realm := rs.Primary.Attributes["realm_id"]
		resourceServerId := rs.Primary.Attributes["resource_server_id"]
		policyId := rs.Primary.ID

		return fmt.Sprintf("%s/%s/%s", realm, resourceServerId, policyId), nil
	}
}

func testResourceKeycloakGenericClientAuthorizationPolicyDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_generic_client_authorization_policy" {
				continue
			}

			realm := rs.Primary.Attributes["realm_id"]
			resourceServerId := rs.Primary.Attributes["resource_server_id"]
			policyId := rs.Primary.ID

			policy, _ := keycloakClient.GetGenericClientAuthorizationPolicy(testCtx, realm, resourceServerId, policyId)
			if policy != nil {
				return fmt.Errorf("policy config with id %s still exists", policyId)
			}
		}

		return nil
	}
}

func testResourceKeycloakGenericClientAuthorizationPolicyExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getResourceKeycloakGenericClientAuthorizationPolicyFromState(s, resourceName)

		return err
	}
}

func testResourceKeycloakGenericClientAuthorizationPolicy_basic(clientId, policyName, policyType string) string {
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

	resource keycloak_generic_client_authorization_policy test {
		resource_server_id = keycloak_openid_client.test.resource_server_id
		realm_id           = data.keycloak_realm.realm.id
		name               = "%s"
		type               = "%s"
		decision_strategy  = "UNANIMOUS"
		logic              = "POSITIVE"
	}
	`, testAccRealm.Realm, clientId, policyName, policyType)
}
