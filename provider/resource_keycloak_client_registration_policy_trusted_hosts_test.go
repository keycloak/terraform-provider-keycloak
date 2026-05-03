package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccKeycloakClientRegistrationPolicyTrustedHosts_basic(t *testing.T) {
	t.Parallel()

	name := acctest.RandomWithPrefix("tf-acc-th")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             crpDestroy("keycloak_client_registration_policy_trusted_hosts"),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakClientRegistrationPolicyTrustedHosts_basic(name),
				Check: resource.ComposeTestCheckFunc(
					crpExists("keycloak_client_registration_policy_trusted_hosts.th"),
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
				ImportStateIdFunc: crpImportId("keycloak_client_registration_policy_trusted_hosts.th"),
			},
		},
	})
}

// TestAccKeycloakClientRegistrationPolicyTrustedHosts_update mutates every
// per-resource field and verifies the diff is applied.
func TestAccKeycloakClientRegistrationPolicyTrustedHosts_update(t *testing.T) {
	t.Parallel()

	name := acctest.RandomWithPrefix("tf-acc-thu")
	updatedName := name + "-renamed"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             crpDestroy("keycloak_client_registration_policy_trusted_hosts"),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakClientRegistrationPolicyTrustedHosts_basic(name),
				Check: resource.ComposeTestCheckFunc(
					crpExists("keycloak_client_registration_policy_trusted_hosts.th"),
					resource.TestCheckResourceAttr("keycloak_client_registration_policy_trusted_hosts.th", "name", name),
					resource.TestCheckResourceAttr("keycloak_client_registration_policy_trusted_hosts.th", "trusted_hosts.#", "2"),
					resource.TestCheckResourceAttr("keycloak_client_registration_policy_trusted_hosts.th", "host_sending_registration_request_must_match", "false"),
					resource.TestCheckResourceAttr("keycloak_client_registration_policy_trusted_hosts.th", "client_uris_must_match", "true"),
				),
			},
			{
				Config: testKeycloakClientRegistrationPolicyTrustedHosts_updated(updatedName),
				Check: resource.ComposeTestCheckFunc(
					crpExists("keycloak_client_registration_policy_trusted_hosts.th"),
					resource.TestCheckResourceAttr("keycloak_client_registration_policy_trusted_hosts.th", "name", updatedName),
					resource.TestCheckResourceAttr("keycloak_client_registration_policy_trusted_hosts.th", "trusted_hosts.#", "1"),
					resource.TestCheckResourceAttr("keycloak_client_registration_policy_trusted_hosts.th", "host_sending_registration_request_must_match", "true"),
					resource.TestCheckResourceAttr("keycloak_client_registration_policy_trusted_hosts.th", "client_uris_must_match", "false"),
				),
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
				Check:  crpExists("keycloak_client_registration_policy_max_clients.mc"),
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
				ImportState:       true,
				ImportStateIdFunc: crpImportId("keycloak_client_registration_policy_max_clients.mc"),
				ExpectError:       regexp.MustCompile(`providerId="max-clients".*expects providerId="trusted-hosts"`),
			},
		},
	})
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

func testKeycloakClientRegistrationPolicyTrustedHosts_updated(name string) string {
	return fmt.Sprintf(`
resource "keycloak_client_registration_policy_trusted_hosts" "th" {
  realm_id = "%s"
  name     = "%s"
  sub_type = "anonymous"

  trusted_hosts                                = ["10.0.0.1"]
  host_sending_registration_request_must_match = true
  client_uris_must_match                       = false
}
`, testAccRealm.Realm, name)
}
