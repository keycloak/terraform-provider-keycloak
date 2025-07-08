package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccKeycloakDataSourceOpenidClientScope_basic(t *testing.T) {
	t.Parallel()
	clientScopeName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.keycloak_openid_client_scope.test"
	resourceName := "keycloak_openid_client_scope.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccKeycloakOpenidClientScopeConfig(clientScopeName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "realm_id", resourceName, "realm_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "consent_screen_text", resourceName, "consent_screen_text"),
					resource.TestCheckResourceAttrPair(dataSourceName, "include_in_token_scope", resourceName, "include_in_token_scope"),
				),
			},
		},
	})
}

func testAccKeycloakOpenidClientScopeConfig(name string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client_scope" "test" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id

	description            = "%s"
	consent_screen_text    = "%s"
	include_in_token_scope = %t
}

data "keycloak_openid_client_scope" "test" {
	name      = keycloak_openid_client_scope.test.name
	realm_id  = data.keycloak_realm.realm.id
}

resource "keycloak_openid_audience_protocol_mapper" "audience_mapper" {
	realm_id        = data.keycloak_realm.realm.id
	client_scope_id = data.keycloak_openid_client_scope.test.id
	name            = "audience-mapper"

	included_custom_audience = "foo"
}
`, testAccRealm.Realm, name, acctest.RandString(10), acctest.RandString(10), randomBool())
}
