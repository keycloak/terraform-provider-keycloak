package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccKeycloakRealmClientPolicyProfile_basic(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")
	resourceName := "test-profile"
	description := "Test description"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmClientPolicyProfile_basic(realmName, resourceName, description),
				Check:  testAccCheckKeycloakRealmClientPolicyProfileExists(realmName, resourceName),
			},
		},
	})
}

func TestAccKeycloakRealmClientPolicyProfile_basicWithExecutor(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")
	resourceName := "test-profile-with-executor"
	description := "Test description with executor"
	executorName := "pkce-enforcer"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmClientPolicyProfile_basicWithExecutor(realmName, resourceName, description, executorName),
				Check:  testAccCheckKeycloakRealmClientPolicyProfileWithExecutorExists(realmName, resourceName, executorName),
			},
		},
	})
}

func TestAccKeycloakRealmClientPolicyProfile_basicWithPolicy(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")
	profileName := "test-profile"
	profileDescription := "Test profile description"
	policyName := "test-policy"
	policyDescription := "Test policy description"
	conditionName := "client-updater-source-roles"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmClientPolicyProfile_basicWithPolicy(realmName, profileName, profileDescription, policyName, policyDescription, conditionName),
				Check:  testAccCheckKeycloakRealmClientPolicyProfilePolicyExists(realmName, policyName),
			},
		},
	})
}

func testKeycloakRealmClientPolicyProfile_basic(realm string, name string, description string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm = "%s"
}
	
resource "keycloak_realm_client_policy_profile" "profile" {
	realm_id      = keycloak_realm.realm.realm
	name          = "%s"
	description   = "%s"
}
	`, realm, name, description)
}

func testKeycloakRealmClientPolicyProfile_basicWithExecutor(realm string, name string, description string, executorName string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_realm_client_policy_profile" "profile" {
	realm_id    = keycloak_realm.realm.realm
	name        = "%s"
	description = "%s"

	executor {
		name = "%s"
		configuration = {
			auto-configure = "true"
		}
	}
}
	`, realm, name, description, executorName)
}

func testKeycloakRealmClientPolicyProfile_basicWithPolicy(realm string, profileName string, profileDescription string, policyName string, policyDescription string, conditionName string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm = "%s"
}

	resource "keycloak_realm_client_policy_profile" "profile" {
	realm_id    = keycloak_realm.realm.realm
	name        = "%s"
	description	= "%s"
}

resource "keycloak_realm_client_policy_profile_policy" "policy" {
	realm_id    = keycloak_realm.realm.realm
  name        = "%s"
  description = "%s"

  profiles = [
    keycloak_realm_client_policy_profile.profile.name
  ]

  condition {
    name = "%s"
    configuration = {
			is_negative_logic = false
			attributes        = jsonencode([{"key": "test-key", "value": "test-value"}])
			}
  }
}
	`, realm, profileName, profileDescription, policyName, policyDescription, conditionName)
}

func testAccCheckKeycloakRealmClientPolicyProfileExists(realm string, profileName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := keycloakClient.GetRealmClientPolicyProfileByName(testCtx, realm, profileName)
		if err != nil {
			return fmt.Errorf("Client policy profile not found: %s", profileName)
		}

		return nil
	}
}

func testAccCheckKeycloakRealmClientPolicyProfileWithExecutorExists(realm string, profileName string, executorName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		profile, err := keycloakClient.GetRealmClientPolicyProfileByName(testCtx, realm, profileName)
		if err != nil {
			return fmt.Errorf("Client policy profile not found: %s", profileName)
		}

		if profile.Executors[0].Name != executorName {
			return fmt.Errorf("Client policy profile executor not found: %s", executorName)
		}

		return nil
	}
}

func testAccCheckKeycloakRealmClientPolicyProfilePolicyExists(realm string, policyName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := keycloakClient.GetRealmClientPolicyProfilePolicyByName(testCtx, realm, policyName)
		if err != nil {
			return fmt.Errorf("Client policy profile policy not found: %s", policyName)
		}

		return nil
	}
}
