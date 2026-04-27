package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakRealmClientRegistrationPolicy_basic(t *testing.T) {
	t.Parallel()

	policyName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRealmClientRegistrationPolicyDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmClientRegistrationPolicy_basic(policyName, "anonymous", "50"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRealmClientRegistrationPolicyExists("keycloak_realm_client_registration_policy.policy"),
					resource.TestCheckResourceAttr("keycloak_realm_client_registration_policy.policy", "name", policyName),
					resource.TestCheckResourceAttr("keycloak_realm_client_registration_policy.policy", "provider_id", "max-clients"),
					resource.TestCheckResourceAttr("keycloak_realm_client_registration_policy.policy", "sub_type", "anonymous"),
					resource.TestCheckResourceAttr("keycloak_realm_client_registration_policy.policy", "config.max-clients", "50"),
				),
			},
		},
	})
}

func TestAccKeycloakRealmClientRegistrationPolicy_update(t *testing.T) {
	t.Parallel()

	policyName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRealmClientRegistrationPolicyDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmClientRegistrationPolicy_basic(policyName, "anonymous", "50"),
				Check:  testAccCheckRealmClientRegistrationPolicyExists("keycloak_realm_client_registration_policy.policy"),
			},
			{
				Config: testKeycloakRealmClientRegistrationPolicy_basic(policyName, "anonymous", "100"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRealmClientRegistrationPolicyExists("keycloak_realm_client_registration_policy.policy"),
					resource.TestCheckResourceAttr("keycloak_realm_client_registration_policy.policy", "config.max-clients", "100"),
				),
			},
		},
	})
}

func TestAccKeycloakRealmClientRegistrationPolicy_import(t *testing.T) {
	t.Parallel()

	policyName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRealmClientRegistrationPolicyDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmClientRegistrationPolicy_basic(policyName, "anonymous", "50"),
				Check:  testAccCheckRealmClientRegistrationPolicyExists("keycloak_realm_client_registration_policy.policy"),
			},
			{
				ResourceName:      "keycloak_realm_client_registration_policy.policy",
				ImportState:       true,
				ImportStateIdFunc: testAccRealmClientRegistrationPolicyImportId("keycloak_realm_client_registration_policy.policy"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccKeycloakRealmClientRegistrationPolicy_createAfterManualDestroy(t *testing.T) {
	t.Parallel()

	var policy = &keycloak.RealmClientRegistrationPolicy{}

	policyName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRealmClientRegistrationPolicyDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmClientRegistrationPolicy_basic(policyName, "anonymous", "50"),
				Check:  testAccCheckRealmClientRegistrationPolicyFetch("keycloak_realm_client_registration_policy.policy", policy),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteRealmClientRegistrationPolicy(testCtx, policy.RealmId, policy.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakRealmClientRegistrationPolicy_basic(policyName, "anonymous", "50"),
				Check:  testAccCheckRealmClientRegistrationPolicyExists("keycloak_realm_client_registration_policy.policy"),
			},
		},
	})
}

func testAccCheckRealmClientRegistrationPolicyExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getRealmClientRegistrationPolicyFromState(s, resourceName)
		return err
	}
}

func testAccCheckRealmClientRegistrationPolicyFetch(resourceName string, policy *keycloak.RealmClientRegistrationPolicy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetched, err := getRealmClientRegistrationPolicyFromState(s, resourceName)
		if err != nil {
			return err
		}

		policy.Id = fetched.Id
		policy.RealmId = fetched.RealmId

		return nil
	}
}

func testAccCheckRealmClientRegistrationPolicyDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_realm_client_registration_policy" {
				continue
			}

			id := rs.Primary.ID
			realmId := rs.Primary.Attributes["realm_id"]

			policy, _ := keycloakClient.GetRealmClientRegistrationPolicy(testCtx, realmId, id)
			if policy != nil {
				return fmt.Errorf("client registration policy with id %s still exists", id)
			}
		}

		return nil
	}
}

func getRealmClientRegistrationPolicyFromState(s *terraform.State, resourceName string) (*keycloak.RealmClientRegistrationPolicy, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	id := rs.Primary.ID
	realmId := rs.Primary.Attributes["realm_id"]

	policy, err := keycloakClient.GetRealmClientRegistrationPolicy(testCtx, realmId, id)
	if err != nil {
		return nil, fmt.Errorf("error getting client registration policy with id %s: %s", id, err)
	}

	return policy, nil
}

func testAccRealmClientRegistrationPolicyImportId(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["realm_id"], rs.Primary.ID), nil
	}
}

func testKeycloakRealmClientRegistrationPolicy_basic(name, subType, maxClients string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_realm_client_registration_policy" "policy" {
	realm_id    = data.keycloak_realm.realm.id
	name        = "%s"
	provider_id = "max-clients"
	sub_type    = "%s"
	config = {
		"max-clients" = "%s"
	}
}
`, testAccRealm.Realm, name, subType, maxClients)
}
