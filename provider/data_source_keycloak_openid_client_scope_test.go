package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccKeycloakDataSourceOpenidClientScope_basic(t *testing.T) {
	t.Parallel()
	clientScopeName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.keycloak_openid_client_scope.test"
	resourceName := "keycloak_openid_client_scope.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccKeycloakOpenidClientScopeConfig(clientScopeName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "realm_id", resourceName, "realm_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "consent_screen_text", resourceName, "consent_screen_text"),
					resource.TestCheckResourceAttrPair(dataSourceName, "display_on_consent_screen", resourceName, "display_on_consent_screen"),
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

// Keycloak creates the built-in offline_access scope without these attributes and treats a missing one as true.
func TestAccKeycloakDataSourceOpenidClientScope_missingAttributesAreTrue(t *testing.T) {
	t.Parallel()
	realmName := acctest.RandomWithPrefix("tf-acc")
	dataSourceName := "data.keycloak_openid_client_scope.offline_access"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeycloakRealmDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccKeycloakOpenidClientScopeConfig_builtIn(realmName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "include_in_token_scope", "true"),
					resource.TestCheckResourceAttr(dataSourceName, "include_in_openid_provider_metadata", "true"),
				),
			},
		},
	})
}

func testAccKeycloakOpenidClientScopeConfig_builtIn(realmName string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm = "%s"
}

data "keycloak_openid_client_scope" "offline_access" {
	name     = "offline_access"
	realm_id = keycloak_realm.realm.id
}
`, realmName)
}
