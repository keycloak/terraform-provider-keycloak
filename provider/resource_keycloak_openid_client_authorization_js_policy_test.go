package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakOpenidClientAuthorizationJSPolicy(t *testing.T) {
	t.Parallel()

	clientId := acctest.RandomWithPrefix("tf-acc")
	policyName := acctest.RandomWithPrefix("tf-acc")
	// Modern Keycloak no longer allows uploading JavaScript code through the API, so this
	// references the JavaScript policy deployed as a JAR by custom-authz-policy-example. Its
	// deployed provider id is "script-" + the fileName from META-INF/keycloak-scripts.json.
	code := "script-always-granting-policy.js"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testResourceKeycloakOpenidClientAuthorizationJSPolicyDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testResourceKeycloakOpenidClientAuthorizationJSPolicy_basic(clientId, policyName, code, "initial description"),
				Check:  testResourceKeycloakOpenidClientAuthorizationJSPolicyExists("keycloak_openid_client_js_policy.test"),
			},
			{
				// Update a non-code attribute to exercise the PUT /policy/js/{id} endpoint.
				Config: testResourceKeycloakOpenidClientAuthorizationJSPolicy_basic(clientId, policyName, code, "updated description"),
				Check: resource.ComposeTestCheckFunc(
					testResourceKeycloakOpenidClientAuthorizationJSPolicyExists("keycloak_openid_client_js_policy.test"),
					resource.TestCheckResourceAttr("keycloak_openid_client_js_policy.test", "description", "updated description"),
				),
			},
		},
	})
}

func getResourceKeycloakOpenidClientAuthorizationJSPolicyFromState(s *terraform.State, resourceName string) (*keycloak.OpenidClientAuthorizationJSPolicy, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	realm := rs.Primary.Attributes["realm_id"]
	resourceServerId := rs.Primary.Attributes["resource_server_id"]
	policyId := rs.Primary.ID

	policy, err := keycloakClient.GetOpenidClientAuthorizationJSPolicy(testCtx, realm, resourceServerId, policyId)
	if err != nil {
		return nil, fmt.Errorf("error getting openid client auth js policy config with alias %s: %s", resourceServerId, err)
	}

	return policy, nil
}

func testResourceKeycloakOpenidClientAuthorizationJSPolicyDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_openid_client_js_policy" {
				continue
			}

			realm := rs.Primary.Attributes["realm_id"]
			resourceServerId := rs.Primary.Attributes["resource_server_id"]
			policyId := rs.Primary.ID

			policy, _ := keycloakClient.GetOpenidClientAuthorizationJSPolicy(testCtx, realm, resourceServerId, policyId)
			if policy != nil {
				return fmt.Errorf("policy config with id %s still exists", policyId)
			}
		}

		return nil
	}
}

func testResourceKeycloakOpenidClientAuthorizationJSPolicyExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getResourceKeycloakOpenidClientAuthorizationJSPolicyFromState(s, resourceName)

		return err
	}
}

func testResourceKeycloakOpenidClientAuthorizationJSPolicy_basic(clientId, policyName, code, description string) string {
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

resource keycloak_openid_client_js_policy test {
	resource_server_id = keycloak_openid_client.test.resource_server_id
	realm_id           = data.keycloak_realm.realm.id
	name               = "%s"
	logic              = "POSITIVE"
	decision_strategy  = "UNANIMOUS"
	code               = "%s"
	description        = "%s"

	lifecycle {
		# Keycloak returns the deployed script's source in "code" on read; ignore that drift.
		ignore_changes = [code]
	}
}
	`, testAccRealm.Realm, clientId, policyName, code, description)
}
