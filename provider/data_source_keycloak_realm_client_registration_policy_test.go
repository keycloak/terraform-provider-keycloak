package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccKeycloakDataSourceRealmClientRegistrationPolicy_basic(t *testing.T) {
	t.Parallel()

	policyName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRealmClientRegistrationPolicyDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testDataSourceKeycloakRealmClientRegistrationPolicy(policyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRealmClientRegistrationPolicyExists("keycloak_realm_client_registration_policy.policy"),
					resource.TestCheckResourceAttrPair(
						"data.keycloak_realm_client_registration_policy.policy", "id",
						"keycloak_realm_client_registration_policy.policy", "id",
					),
					resource.TestCheckResourceAttr("data.keycloak_realm_client_registration_policy.policy", "provider_id", "max-clients"),
					resource.TestCheckResourceAttr("data.keycloak_realm_client_registration_policy.policy", "sub_type", "anonymous"),
					resource.TestCheckResourceAttr("data.keycloak_realm_client_registration_policy.policy", "config.max-clients", "50"),
				),
			},
		},
	})
}

func testDataSourceKeycloakRealmClientRegistrationPolicy(name string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_realm_client_registration_policy" "policy" {
	realm_id    = data.keycloak_realm.realm.id
	name        = "%s"
	provider_id = "max-clients"
	sub_type    = "anonymous"
	config = {
		"max-clients" = "50"
	}
}

data "keycloak_realm_client_registration_policy" "policy" {
	realm_id    = data.keycloak_realm.realm.id
	name        = keycloak_realm_client_registration_policy.policy.name
	provider_id = "max-clients"
	sub_type    = "anonymous"
}
`, testAccRealm.Realm, name)
}
