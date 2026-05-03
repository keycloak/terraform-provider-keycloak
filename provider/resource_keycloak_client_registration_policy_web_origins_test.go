package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccKeycloakClientRegistrationPolicyWebOrigins_basic(t *testing.T) {
	t.Parallel()

	name := acctest.RandomWithPrefix("tf-acc-wo")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             crpDestroy("keycloak_client_registration_policy_web_origins"),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakClientRegistrationPolicyWebOrigins_basic(name),
				Check: resource.ComposeTestCheckFunc(
					crpExists("keycloak_client_registration_policy_web_origins.wo"),
					resource.TestCheckResourceAttr("keycloak_client_registration_policy_web_origins.wo", "name", name),
					resource.TestCheckResourceAttr("keycloak_client_registration_policy_web_origins.wo", "sub_type", "anonymous"),
					resource.TestCheckResourceAttr("keycloak_client_registration_policy_web_origins.wo", "web_origins.#", "2"),
				),
			},
			{
				ResourceName:      "keycloak_client_registration_policy_web_origins.wo",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: crpImportId("keycloak_client_registration_policy_web_origins.wo"),
			},
		},
	})
}

func testKeycloakClientRegistrationPolicyWebOrigins_basic(name string) string {
	return fmt.Sprintf(`
resource "keycloak_client_registration_policy_web_origins" "wo" {
  realm_id    = "%s"
  name        = "%s"
  sub_type    = "anonymous"
  web_origins = ["https://app.example.com", "https://admin.example.com"]
}
`, testAccRealm.Realm, name)
}
