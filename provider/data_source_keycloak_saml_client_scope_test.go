package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccKeycloakDataSourceSamlClientScope_basic(t *testing.T) {
	t.Parallel()
	clientScopeName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.keycloak_saml_client_scope.test"
	resourceName := "keycloak_saml_client_scope.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccKeycloakSamlClientScopeDataSourceConfig(clientScopeName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "realm_id", resourceName, "realm_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "consent_screen_text", resourceName, "consent_screen_text"),
				),
			},
		},
	})
}

func testAccKeycloakSamlClientScopeDataSourceConfig(name string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_saml_client_scope" "test" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id

	description         = "%s"
	consent_screen_text = "%s"
}

data "keycloak_saml_client_scope" "test" {
	name     = keycloak_saml_client_scope.test.name
	realm_id = data.keycloak_realm.realm.id
}
`, testAccRealm.Realm, name, acctest.RandString(10), acctest.RandString(10))
}
