package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccKeycloakRealmAttributes(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmAttributes_basic(realmName, "foo", "bar"),
				Check:  testAccCheckKeycloakRealmAttributesExists(realmName),
			},
		},
	})
}

func testKeycloakRealmAttributes_basic(realm string, name string, description string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_realm_attributes" "attributes" {
	realm_id = keycloak_realm.realm.id
	attributes = {
		name        = "%s"
		description = "%s"
	}
}`, realm, name, description)
}

func testAccCheckKeycloakRealmAttributesExists(realm string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := keycloakClient.GetRealmAttributes(testCtx, realm)
		if err != nil {
			return fmt.Errorf("Realm attributes not found: %s", realm)
		}

		return nil
	}
}
