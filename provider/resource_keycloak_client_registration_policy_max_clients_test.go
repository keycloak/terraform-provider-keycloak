package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccKeycloakClientRegistrationPolicyMaxClients_basic(t *testing.T) {
	t.Parallel()

	name := acctest.RandomWithPrefix("tf-acc-mc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             crpDestroy("keycloak_client_registration_policy_max_clients"),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakClientRegistrationPolicyMaxClients_basic(name),
				Check: resource.ComposeTestCheckFunc(
					crpExists("keycloak_client_registration_policy_max_clients.mc"),
					resource.TestCheckResourceAttr("keycloak_client_registration_policy_max_clients.mc", "max_clients", "5000"),
					resource.TestCheckResourceAttr("keycloak_client_registration_policy_max_clients.mc", "sub_type", "anonymous"),
				),
			},
			{
				ResourceName:      "keycloak_client_registration_policy_max_clients.mc",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: crpImportId("keycloak_client_registration_policy_max_clients.mc"),
			},
		},
	})
}

func TestAccKeycloakClientRegistrationPolicyMaxClients_update(t *testing.T) {
	t.Parallel()

	name := acctest.RandomWithPrefix("tf-acc-mcu")
	updatedName := name + "-renamed"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             crpDestroy("keycloak_client_registration_policy_max_clients"),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakClientRegistrationPolicyMaxClients_basic(name),
				Check: resource.ComposeTestCheckFunc(
					crpExists("keycloak_client_registration_policy_max_clients.mc"),
					resource.TestCheckResourceAttr("keycloak_client_registration_policy_max_clients.mc", "name", name),
					resource.TestCheckResourceAttr("keycloak_client_registration_policy_max_clients.mc", "max_clients", "5000"),
				),
			},
			{
				Config: testKeycloakClientRegistrationPolicyMaxClients_updated(updatedName),
				Check: resource.ComposeTestCheckFunc(
					crpExists("keycloak_client_registration_policy_max_clients.mc"),
					resource.TestCheckResourceAttr("keycloak_client_registration_policy_max_clients.mc", "name", updatedName),
					resource.TestCheckResourceAttr("keycloak_client_registration_policy_max_clients.mc", "max_clients", "1234"),
				),
			},
		},
	})
}

func testKeycloakClientRegistrationPolicyMaxClients_basic(name string) string {
	return fmt.Sprintf(`
resource "keycloak_client_registration_policy_max_clients" "mc" {
  realm_id    = "%s"
  name        = "%s"
  sub_type    = "anonymous"
  max_clients = 5000
}
`, testAccRealm.Realm, name)
}

func testKeycloakClientRegistrationPolicyMaxClients_updated(name string) string {
	return fmt.Sprintf(`
resource "keycloak_client_registration_policy_max_clients" "mc" {
  realm_id    = "%s"
  name        = "%s"
  sub_type    = "anonymous"
  max_clients = 1234
}
`, testAccRealm.Realm, name)
}
