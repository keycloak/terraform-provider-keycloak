package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccKeycloakDataSourceOpenidClientAuthorizationScope_basic(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	scopeName := acctest.RandomWithPrefix("tf-acc")
	dataSourceName := "data.keycloak_openid_client_authorization_scope.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccKeycloakOpenidClientAuthorizationScopeDataSourceConfig(clientId, scopeName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "id", regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")),
					resource.TestCheckResourceAttr(dataSourceName, "realm_id", testAccRealm.Realm),
					resource.TestCheckResourceAttr(dataSourceName, "name", scopeName),
				),
			},
		},
	})
}

func testAccKeycloakOpenidClientAuthorizationScopeDataSourceConfig(clientId, scopeName string) string {
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

resource "keycloak_openid_client_authorization_scope" "test" {
	resource_server_id = keycloak_openid_client.test.resource_server_id
	realm_id           = data.keycloak_realm.realm.id
	name               = "%s"
}

data "keycloak_openid_client_authorization_scope" "test" {
	resource_server_id = keycloak_openid_client.test.resource_server_id
	realm_id           = data.keycloak_realm.realm.id
	name               = keycloak_openid_client_authorization_scope.test.name
	depends_on         = [keycloak_openid_client_authorization_scope.test]
}
`, testAccRealm.Realm, clientId, scopeName)
}
