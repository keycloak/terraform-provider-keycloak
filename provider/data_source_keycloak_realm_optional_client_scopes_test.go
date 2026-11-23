package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccKeycloakDataSourceRealmOptionalClientScopes_basic(t *testing.T) {
	t.Parallel()
	realmName := acctest.RandomWithPrefix("tf-acc")

	dataSourceName := "data.keycloak_realm_optional_client_scopes.optional_scopes"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testDataSourceKeycloakRealmOptionalClientScopes_basic(realmName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "realm_id", "keycloak_realm.realm", "id"),
					resource.TestCheckResourceAttr(dataSourceName, "optional_scopes.#", "1"),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "optional_scopes.*", "phone"),
				),
			},
		},
	})
}

func testDataSourceKeycloakRealmOptionalClientScopes_basic(realmName string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_realm_optional_client_scopes" "optional_scopes" {
	realm_id        = keycloak_realm.realm.id
	optional_scopes = ["phone"]
}

data "keycloak_realm_optional_client_scopes" "optional_scopes" {
	realm_id = keycloak_realm.realm.id

	depends_on = [keycloak_realm_optional_client_scopes.optional_scopes]
}`, realmName)
}
