package provider

import (
	"fmt"
	"reflect"
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

	resourceName := "keycloak_generic_client_authorization_policy.test"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testResourceKeycloakGenericClientAuthorizationPolicyDestroy(),
		Steps: []resource.TestStep{
			{
				// The deployed JS policy in this test ignores config; it's set here purely
				// to verify the attribute round-trips through create/read/update/import.
				Config: testResourceKeycloakGenericClientAuthorizationPolicy_config(clientId, policyName, policyType, `{
					foo = "bar"
				}`),
				Check: resource.ComposeTestCheckFunc(
					testResourceKeycloakGenericClientAuthorizationPolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "config.foo", "bar"),
					testResourceKeycloakGenericClientAuthorizationPolicyHasConfig(resourceName, map[string]string{"foo": "bar"}),
				),
			},
			{
				// Update: change a value and add a key. Proves Update sends the new config,
				// not just Create.
				Config: testResourceKeycloakGenericClientAuthorizationPolicy_config(clientId, policyName, policyType, `{
					foo   = "baz"
					other = "value"
				}`),
				Check: resource.ComposeTestCheckFunc(
					testResourceKeycloakGenericClientAuthorizationPolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "config.foo", "baz"),
					resource.TestCheckResourceAttr(resourceName, "config.other", "value"),
					testResourceKeycloakGenericClientAuthorizationPolicyHasConfig(resourceName, map[string]string{"foo": "baz", "other": "value"}),
				),
			},
			{
				// Clear config entirely. Regression check for a real bug caught in review:
				// `omitempty` on the Go struct's Config field would drop the "config" key
				// from the update payload for an empty map, leaving the previous values
				// stuck on the server while Terraform reports a clean apply.
				Config: testResourceKeycloakGenericClientAuthorizationPolicy_config(clientId, policyName, policyType, `{}`),
				Check: resource.ComposeTestCheckFunc(
					testResourceKeycloakGenericClientAuthorizationPolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "config.%", "0"),
					testResourceKeycloakGenericClientAuthorizationPolicyHasConfig(resourceName, map[string]string{}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getResourceKeycloakGenericClientAuthorizationPolicyImportId(resourceName),
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

// testResourceKeycloakGenericClientAuthorizationPolicyHasConfig fetches the policy directly
// from the Keycloak Admin API (independent of Terraform's own Read path) and asserts its
// config matches expected. ImportStateVerify alone compares two reads that both go through
// the same Read function, so it can't catch a case where Terraform's in-memory state is
// correct but the value never actually reached (or was cleared on) the server.
func testResourceKeycloakGenericClientAuthorizationPolicyHasConfig(resourceName string, expected map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		policy, err := getResourceKeycloakGenericClientAuthorizationPolicyFromState(s, resourceName)
		if err != nil {
			return err
		}

		// A nil map (no config key returned by the server) and an empty map both mean
		// "no config", normalize before comparing so the empty-config case isn't a
		// false failure.
		actual := policy.Config
		if len(actual) == 0 {
			actual = map[string]string{}
		}
		if len(expected) == 0 {
			expected = map[string]string{}
		}

		if !reflect.DeepEqual(actual, expected) {
			return fmt.Errorf("expected policy config to be %v, got %v", expected, actual)
		}

		return nil
	}
}

func testResourceKeycloakGenericClientAuthorizationPolicy_config(clientId, policyName, policyType, configHcl string) string {
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
		config             = %s
	}
	`, testAccRealm.Realm, clientId, policyName, policyType, configHcl)
}
