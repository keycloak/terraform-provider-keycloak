package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccKeycloakDataSourceRealmDefaultClientScopes_basic(t *testing.T) {
	t.Parallel()
	realmName := acctest.RandomWithPrefix("tf-acc")

	dataSourceName := "data.keycloak_realm_default_client_scopes.default_scopes"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testDataSourceKeycloakRealmDefaultClientScopes_basic(realmName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "realm_id", "keycloak_realm.realm", "id"),
					resource.TestCheckResourceAttr(dataSourceName, "default_scopes.#", "2"),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "default_scopes.*", "profile"),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "default_scopes.*", "email"),
				),
			},
		},
	})
}

func testDataSourceKeycloakRealmDefaultClientScopes_basic(realmName string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_realm_default_client_scopes" "default_scopes" {
	realm_id       = keycloak_realm.realm.id
	default_scopes = ["profile", "email"]
}

data "keycloak_realm_default_client_scopes" "default_scopes" {
	realm_id = keycloak_realm.realm.id

	depends_on = [keycloak_realm_default_client_scopes.default_scopes]
}`, realmName)
}
