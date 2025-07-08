package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"regexp"
	"testing"
)

func TestAccKeycloakDataSourceOpenidClientAuthorizationPolicy_basic(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.keycloak_openid_client_authorization_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccKeycloakOpenidClientAuthorizationPolicyConfig(clientId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "resource_server_id", regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")),
					resource.TestCheckResourceAttr(dataSourceName, "realm_id", testAccRealm.Realm),
					resource.TestCheckResourceAttr(dataSourceName, "name", "default"),
					resource.TestCheckResourceAttr(dataSourceName, "decision_strategy", "UNANIMOUS"),
					resource.TestCheckResourceAttr(dataSourceName, "logic", "POSITIVE"),
					resource.TestCheckResourceAttr(dataSourceName, "type", "resource"),
				),
			},
		},
	})
}

func testAccKeycloakOpenidClientAuthorizationPolicyConfig(clientId string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "test" {
	client_id                      = "%s"
	name                           = "%s"
	realm_id                       = data.keycloak_realm.realm.id
	description                    = "a test openid client"
	standard_flow_enabled          = true
	service_accounts_enabled       = true
	access_type                    = "CONFIDENTIAL"
	client_secret                  = "secret"
	valid_redirect_uris            = [
		"http://localhost:5555/callback",
	]
	authorization {
		policy_enforcement_mode = "ENFORCING"
	}
}

data "keycloak_openid_client_authorization_policy" "test" {
	resource_server_id = keycloak_openid_client.test.resource_server_id
	realm_id           = data.keycloak_realm.realm.id
	name               = "default"

	depends_on = [
		keycloak_openid_client.test,
	]
}
`, testAccRealm.Realm, clientId, clientId)
}
