package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccKeycloakClientRegistrationPolicyMaxClients_basic(t *testing.T) {
	t.Parallel()

	name := acctest.RandomWithPrefix("tf-acc-mc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckClientRegistrationPolicyMaxClientsDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakClientRegistrationPolicyMaxClients_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientRegistrationPolicyMaxClientsExists("keycloak_client_registration_policy_max_clients.mc"),
					resource.TestCheckResourceAttr("keycloak_client_registration_policy_max_clients.mc", "max_clients", "5000"),
					resource.TestCheckResourceAttr("keycloak_client_registration_policy_max_clients.mc", "sub_type", "anonymous"),
				),
			},
			{
				ResourceName:      "keycloak_client_registration_policy_max_clients.mc",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["keycloak_client_registration_policy_max_clients.mc"]
					if !ok {
						return "", fmt.Errorf("not found in state")
					}
					return fmt.Sprintf("%s/%s", rs.Primary.Attributes["realm_id"], rs.Primary.ID), nil
				},
			},
		},
	})
}

func testAccCheckClientRegistrationPolicyMaxClientsExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found in state: %s", name)
		}
		_, err := keycloakClient.GetClientRegistrationPolicy(testCtx, rs.Primary.Attributes["realm_id"], rs.Primary.ID)
		return err
	}
}

func testAccCheckClientRegistrationPolicyMaxClientsDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_client_registration_policy_max_clients" {
				continue
			}
			policy, _ := keycloakClient.GetClientRegistrationPolicy(testCtx, rs.Primary.Attributes["realm_id"], rs.Primary.ID)
			if policy != nil {
				return fmt.Errorf("max_clients policy %s still exists", rs.Primary.ID)
			}
		}
		return nil
	}
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
