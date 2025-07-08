package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"regexp"
	"testing"
)

func TestAccKeycloakDataSourceOpenidClientServiceAccountUser_basic(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.keycloak_openid_client_service_account_user.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccKeycloakOpenidClientServiceAccountUserConfig(clientId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "realm_id", testAccRealm.Realm),
					resource.TestMatchResourceAttr(dataSourceName, "client_id", regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")),
					resource.TestCheckResourceAttr(dataSourceName, "username", "service-account-"+clientId),
				),
			},
		},
	})
}

func testAccKeycloakOpenidClientServiceAccountUserConfig(clientId string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "test" {
	name                  	 = "%s"
	client_id                = "%s"
	realm_id              	 = data.keycloak_realm.realm.id
	description           	 = "a test openid client"
	standard_flow_enabled    = true
	access_type              = "CONFIDENTIAL"
	service_accounts_enabled = true
	client_secret            = "secret"
	valid_redirect_uris      = [
		"http://localhost:5555/callback",
	]
	authorization {
		policy_enforcement_mode = "ENFORCING"
	}
	web_origins              = [
		"http://localhost"
	]
}

data keycloak_openid_client_service_account_user test {
	realm_id  = data.keycloak_realm.realm.id
	client_id = keycloak_openid_client.test.id
}
`, testAccRealm.Realm, clientId, clientId)
}
