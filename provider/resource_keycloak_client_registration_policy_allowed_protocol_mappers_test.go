package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccKeycloakClientRegistrationPolicyAllowedProtocolMappers_basic(t *testing.T) {
	t.Parallel()

	name := acctest.RandomWithPrefix("tf-acc-apm")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             crpDestroy("keycloak_client_registration_policy_allowed_protocol_mappers"),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakClientRegistrationPolicyAllowedProtocolMappers_basic(name),
				Check: resource.ComposeTestCheckFunc(
					crpExists("keycloak_client_registration_policy_allowed_protocol_mappers.apm"),
					resource.TestCheckResourceAttr("keycloak_client_registration_policy_allowed_protocol_mappers.apm", "name", name),
					resource.TestCheckResourceAttr("keycloak_client_registration_policy_allowed_protocol_mappers.apm", "sub_type", "anonymous"),
					resource.TestCheckResourceAttr("keycloak_client_registration_policy_allowed_protocol_mappers.apm", "allowed_protocol_mapper_types.#", "5"),
				),
			},
			{
				ResourceName:      "keycloak_client_registration_policy_allowed_protocol_mappers.apm",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: crpImportId("keycloak_client_registration_policy_allowed_protocol_mappers.apm"),
			},
		},
	})
}

func TestAccKeycloakClientRegistrationPolicyAllowedProtocolMappers_update(t *testing.T) {
	t.Parallel()

	name := acctest.RandomWithPrefix("tf-acc-apmu")
	updatedName := name + "-renamed"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             crpDestroy("keycloak_client_registration_policy_allowed_protocol_mappers"),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakClientRegistrationPolicyAllowedProtocolMappers_basic(name),
				Check: resource.ComposeTestCheckFunc(
					crpExists("keycloak_client_registration_policy_allowed_protocol_mappers.apm"),
					resource.TestCheckResourceAttr("keycloak_client_registration_policy_allowed_protocol_mappers.apm", "name", name),
					resource.TestCheckResourceAttr("keycloak_client_registration_policy_allowed_protocol_mappers.apm", "allowed_protocol_mapper_types.#", "5"),
				),
			},
			{
				Config: testKeycloakClientRegistrationPolicyAllowedProtocolMappers_updated(updatedName),
				Check: resource.ComposeTestCheckFunc(
					crpExists("keycloak_client_registration_policy_allowed_protocol_mappers.apm"),
					resource.TestCheckResourceAttr("keycloak_client_registration_policy_allowed_protocol_mappers.apm", "name", updatedName),
					resource.TestCheckResourceAttr("keycloak_client_registration_policy_allowed_protocol_mappers.apm", "allowed_protocol_mapper_types.#", "2"),
				),
			},
		},
	})
}

func testKeycloakClientRegistrationPolicyAllowedProtocolMappers_basic(name string) string {
	return fmt.Sprintf(`
resource "keycloak_client_registration_policy_allowed_protocol_mappers" "apm" {
  realm_id = "%s"
  name     = "%s"
  sub_type = "anonymous"

  allowed_protocol_mapper_types = [
    "oidc-usermodel-property-mapper",
    "oidc-usermodel-attribute-mapper",
    "oidc-full-name-mapper",
    "oidc-sha256-pairwise-sub-mapper",
    "oidc-address-mapper",
  ]
}
`, testAccRealm.Realm, name)
}

func testKeycloakClientRegistrationPolicyAllowedProtocolMappers_updated(name string) string {
	return fmt.Sprintf(`
resource "keycloak_client_registration_policy_allowed_protocol_mappers" "apm" {
  realm_id = "%s"
  name     = "%s"
  sub_type = "anonymous"

  allowed_protocol_mapper_types = [
    "oidc-full-name-mapper",
    "oidc-address-mapper",
  ]
}
`, testAccRealm.Realm, name)
}
