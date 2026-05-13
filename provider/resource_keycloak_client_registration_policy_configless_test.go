package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// runConfiglessBasic exercises the create + import-verify flow for a
// config-less policy resource (consent_required, full_scope_disallowed,
// client_disabled). They share the same helpers under the hood, so the
// test bodies are identical apart from the resource type.
func runConfiglessBasic(t *testing.T, resourceType, addressLocal string) {
	name := acctest.RandomWithPrefix("tf-acc-cfg")
	address := fmt.Sprintf("%s.%s", resourceType, addressLocal)

	hcl := fmt.Sprintf(`
resource "%s" "%s" {
  realm_id = "%s"
  name     = "%s"
  sub_type = "authenticated"
}
`, resourceType, addressLocal, testAccRealm.Realm, name)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             crpDestroy(resourceType),
		Steps: []resource.TestStep{
			{
				Config: hcl,
				Check: resource.ComposeTestCheckFunc(
					crpExists(address),
					resource.TestCheckResourceAttr(address, "name", name),
					resource.TestCheckResourceAttr(address, "sub_type", "authenticated"),
				),
			},
			{
				ResourceName:      address,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: crpImportId(address),
			},
		},
	})
}

func TestAccKeycloakClientRegistrationPolicyConsentRequired_basic(t *testing.T) {
	t.Parallel()
	runConfiglessBasic(t, "keycloak_client_registration_policy_consent_required", "cr")
}

func TestAccKeycloakClientRegistrationPolicyFullScopeDisallowed_basic(t *testing.T) {
	t.Parallel()
	runConfiglessBasic(t, "keycloak_client_registration_policy_full_scope_disallowed", "fsd")
}

func TestAccKeycloakClientRegistrationPolicyClientDisabled_basic(t *testing.T) {
	t.Parallel()
	runConfiglessBasic(t, "keycloak_client_registration_policy_client_disabled", "cd")
}

// runConfiglessUpdate verifies a config-less policy can be renamed in place
// (the only mutable schema field on these resources — realm_id and sub_type
// are ForceNew).
func runConfiglessUpdate(t *testing.T, resourceType, addressLocal string) {
	original := acctest.RandomWithPrefix("tf-acc-cfgu")
	renamed := original + "-renamed"
	address := fmt.Sprintf("%s.%s", resourceType, addressLocal)

	configFor := func(name string) string {
		return fmt.Sprintf(`
resource "%s" "%s" {
  realm_id = "%s"
  name     = "%s"
  sub_type = "authenticated"
}
`, resourceType, addressLocal, testAccRealm.Realm, name)
	}

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             crpDestroy(resourceType),
		Steps: []resource.TestStep{
			{
				Config: configFor(original),
				Check: resource.ComposeTestCheckFunc(
					crpExists(address),
					resource.TestCheckResourceAttr(address, "name", original),
				),
			},
			{
				Config: configFor(renamed),
				Check: resource.ComposeTestCheckFunc(
					crpExists(address),
					resource.TestCheckResourceAttr(address, "name", renamed),
				),
			},
		},
	})
}

func TestAccKeycloakClientRegistrationPolicyConsentRequired_update(t *testing.T) {
	t.Parallel()
	runConfiglessUpdate(t, "keycloak_client_registration_policy_consent_required", "cr")
}

func TestAccKeycloakClientRegistrationPolicyFullScopeDisallowed_update(t *testing.T) {
	t.Parallel()
	runConfiglessUpdate(t, "keycloak_client_registration_policy_full_scope_disallowed", "fsd")
}

func TestAccKeycloakClientRegistrationPolicyClientDisabled_update(t *testing.T) {
	t.Parallel()
	runConfiglessUpdate(t, "keycloak_client_registration_policy_client_disabled", "cd")
}
