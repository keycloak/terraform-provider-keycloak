package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccKeycloakClientRegistrationPolicies_fullStack declares all 8 typed
// policy resources in a single config (4 anonymous, 4 authenticated) and
// verifies that apply + destroy succeed with no resource interfering with
// any other.
func TestAccKeycloakClientRegistrationPolicies_fullStack(t *testing.T) {
	t.Parallel()

	prefix := acctest.RandomWithPrefix("tf-acc-fs")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			crpDestroy("keycloak_client_registration_policy_trusted_hosts"),
			crpDestroy("keycloak_client_registration_policy_max_clients"),
			crpDestroy("keycloak_client_registration_policy_allowed_client_scopes"),
			crpDestroy("keycloak_client_registration_policy_allowed_protocol_mappers"),
			crpDestroy("keycloak_client_registration_policy_web_origins"),
			crpDestroy("keycloak_client_registration_policy_consent_required"),
			crpDestroy("keycloak_client_registration_policy_full_scope_disallowed"),
			crpDestroy("keycloak_client_registration_policy_client_disabled"),
		),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakClientRegistrationPolicies_fullStack(prefix),
				Check: resource.ComposeAggregateTestCheckFunc(
					crpExists("keycloak_client_registration_policy_trusted_hosts.fs"),
					crpExists("keycloak_client_registration_policy_max_clients.fs"),
					crpExists("keycloak_client_registration_policy_allowed_client_scopes.fs"),
					crpExists("keycloak_client_registration_policy_allowed_protocol_mappers.fs"),
					crpExists("keycloak_client_registration_policy_web_origins.fs"),
					crpExists("keycloak_client_registration_policy_consent_required.fs"),
					crpExists("keycloak_client_registration_policy_full_scope_disallowed.fs"),
					crpExists("keycloak_client_registration_policy_client_disabled.fs"),
				),
			},
		},
	})
}

func testKeycloakClientRegistrationPolicies_fullStack(prefix string) string {
	return fmt.Sprintf(`
resource "keycloak_client_registration_policy_trusted_hosts" "fs" {
  realm_id                                     = "%[1]s"
  name                                         = "%[2]s-th"
  sub_type                                     = "anonymous"
  trusted_hosts                                = ["127.0.0.1", "localhost"]
  host_sending_registration_request_must_match = true
  client_uris_must_match                       = true
}

resource "keycloak_client_registration_policy_max_clients" "fs" {
  realm_id    = "%[1]s"
  name        = "%[2]s-mc"
  sub_type    = "anonymous"
  max_clients = 100
}

resource "keycloak_client_registration_policy_allowed_client_scopes" "fs" {
  realm_id              = "%[1]s"
  name                  = "%[2]s-acs"
  sub_type              = "anonymous"
  allow_default_scopes  = true
  allowed_client_scopes = ["openid", "email", "profile"]
}

resource "keycloak_client_registration_policy_allowed_protocol_mappers" "fs" {
  realm_id                      = "%[1]s"
  name                          = "%[2]s-apm"
  sub_type                      = "anonymous"
  allowed_protocol_mapper_types = ["oidc-usermodel-property-mapper", "oidc-full-name-mapper"]
}

resource "keycloak_client_registration_policy_web_origins" "fs" {
  realm_id    = "%[1]s"
  name        = "%[2]s-wo"
  sub_type    = "authenticated"
  web_origins = ["https://app.example.com"]
}

resource "keycloak_client_registration_policy_consent_required" "fs" {
  realm_id = "%[1]s"
  name     = "%[2]s-cr"
  sub_type = "authenticated"
}

resource "keycloak_client_registration_policy_full_scope_disallowed" "fs" {
  realm_id = "%[1]s"
  name     = "%[2]s-fsd"
  sub_type = "authenticated"
}

resource "keycloak_client_registration_policy_client_disabled" "fs" {
  realm_id = "%[1]s"
  name     = "%[2]s-cd"
  sub_type = "authenticated"
}
`, testAccRealm.Realm, prefix)
}
