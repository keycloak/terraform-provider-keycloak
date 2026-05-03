package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccKeycloakClientRegistrationPolicyTrustedHosts_basic(t *testing.T) {
	t.Parallel()

	name := acctest.RandomWithPrefix("tf-acc-th")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckClientRegistrationPolicyTrustedHostsDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakClientRegistrationPolicyTrustedHosts_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientRegistrationPolicyTrustedHostsExists("keycloak_client_registration_policy_trusted_hosts.th"),
					resource.TestCheckResourceAttr("keycloak_client_registration_policy_trusted_hosts.th", "name", name),
					resource.TestCheckResourceAttr("keycloak_client_registration_policy_trusted_hosts.th", "sub_type", "anonymous"),
					resource.TestCheckResourceAttr("keycloak_client_registration_policy_trusted_hosts.th", "client_uris_must_match", "true"),
					resource.TestCheckResourceAttr("keycloak_client_registration_policy_trusted_hosts.th", "host_sending_registration_request_must_match", "false"),
					resource.TestCheckResourceAttr("keycloak_client_registration_policy_trusted_hosts.th", "trusted_hosts.#", "2"),
				),
			},
			{
				ResourceName:      "keycloak_client_registration_policy_trusted_hosts.th",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["keycloak_client_registration_policy_trusted_hosts.th"]
					if !ok {
						return "", fmt.Errorf("not found in state")
					}
					return fmt.Sprintf("%s/%s", rs.Primary.Attributes["realm_id"], rs.Primary.ID), nil
				},
			},
		},
	})
}

func TestAccKeycloakClientRegistrationPolicyTrustedHosts_importMismatch(t *testing.T) {
	t.Parallel()

	name := acctest.RandomWithPrefix("tf-acc-mm")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Create a max_clients policy, then attempt to import it as
				// trusted_hosts. The provider_id guard should fail.
				Config: testKeycloakClientRegistrationPolicyMaxClients_basic(name),
				Check:  testAccCheckClientRegistrationPolicyMaxClientsExists("keycloak_client_registration_policy_max_clients.mc"),
			},
			{
				ResourceName: "keycloak_client_registration_policy_trusted_hosts.th",
				Config: testKeycloakClientRegistrationPolicyMaxClients_basic(name) + `
resource "keycloak_client_registration_policy_trusted_hosts" "th" {
  realm_id = "` + testAccRealm.Realm + `"
  name     = "ignored"
  sub_type = "anonymous"
}
`,
				ImportState: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["keycloak_client_registration_policy_max_clients.mc"]
					if !ok {
						return "", fmt.Errorf("not found in state")
					}
					return fmt.Sprintf("%s/%s", rs.Primary.Attributes["realm_id"], rs.Primary.ID), nil
				},
				ExpectError: regexp.MustCompile(`providerId="max-clients".*expects providerId="trusted-hosts"`),
			},
		},
	})
}

func testAccCheckClientRegistrationPolicyTrustedHostsExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found in state: %s", name)
		}
		_, err := keycloakClient.GetClientRegistrationPolicy(testCtx, rs.Primary.Attributes["realm_id"], rs.Primary.ID)
		return err
	}
}

func testAccCheckClientRegistrationPolicyTrustedHostsDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_client_registration_policy_trusted_hosts" {
				continue
			}
			policy, _ := keycloakClient.GetClientRegistrationPolicy(testCtx, rs.Primary.Attributes["realm_id"], rs.Primary.ID)
			if policy != nil {
				return fmt.Errorf("trusted_hosts policy %s still exists", rs.Primary.ID)
			}
		}
		return nil
	}
}

func testKeycloakClientRegistrationPolicyTrustedHosts_basic(name string) string {
	return fmt.Sprintf(`
resource "keycloak_client_registration_policy_trusted_hosts" "th" {
  realm_id = "%s"
  name     = "%s"
  sub_type = "anonymous"

  trusted_hosts                                = ["127.0.0.1", "localhost"]
  host_sending_registration_request_must_match = false
  client_uris_must_match                       = true
}
`, testAccRealm.Realm, name)
}
