package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccKeycloakClientRegistrationPolicyAllowedClientScopes_basic(t *testing.T) {
	t.Parallel()

	name := acctest.RandomWithPrefix("tf-acc-acs")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             crpDestroy("keycloak_client_registration_policy_allowed_client_scopes"),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakClientRegistrationPolicyAllowedClientScopes_basic(name),
				Check: resource.ComposeTestCheckFunc(
					crpExists("keycloak_client_registration_policy_allowed_client_scopes.acs"),
					resource.TestCheckResourceAttr("keycloak_client_registration_policy_allowed_client_scopes.acs", "name", name),
					resource.TestCheckResourceAttr("keycloak_client_registration_policy_allowed_client_scopes.acs", "sub_type", "anonymous"),
					resource.TestCheckResourceAttr("keycloak_client_registration_policy_allowed_client_scopes.acs", "allow_default_scopes", "false"),
					resource.TestCheckResourceAttr("keycloak_client_registration_policy_allowed_client_scopes.acs", "allowed_client_scopes.#", "3"),
				),
			},
			{
				ResourceName:      "keycloak_client_registration_policy_allowed_client_scopes.acs",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: crpImportId("keycloak_client_registration_policy_allowed_client_scopes.acs"),
			},
		},
	})
}

func testKeycloakClientRegistrationPolicyAllowedClientScopes_basic(name string) string {
	return fmt.Sprintf(`
resource "keycloak_client_registration_policy_allowed_client_scopes" "acs" {
  realm_id              = "%s"
  name                  = "%s"
  sub_type              = "anonymous"
  allow_default_scopes  = false
  allowed_client_scopes = ["openid", "email", "profile"]
}
`, testAccRealm.Realm, name)
}
