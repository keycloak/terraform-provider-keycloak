package provider

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccKeycloakRealmClientPolicyProfile_basic(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")
	resourceName := "test-profile"
	description := "Test description"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
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
	configuration := map[string]interface{}{
		"auto-configure": "true",
	}

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmClientPolicyProfile_basicWithExecutor(realmName, resourceName, description, executorName, testKeycloakRealmClientPolicyProfile_mapConfig(configuration)),
				Check:  testAccCheckKeycloakRealmClientPolicyProfileWithExecutorExists(realmName, resourceName, executorName),
			},
		},
	})
}

func TestAccKeycloakRealmClientPolicyProfile_basicWithExecutorAndJSON(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")
	resourceName := "test-profile-with-executor-and-configuration"
	description := "Test description with executor and configuration"
	executorName := "secure-client-authenticator"
	configuration := map[string]interface{}{
		"allowed-client-authenticators": []string{"client-secret", "client-secret-jwt"},
		"default-client-authenticator":  "client-secret",
	}

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmClientPolicyProfile_basicWithExecutor(realmName, resourceName, description, executorName, testKeycloakRealmClientPolicyProfile_mapConfig(configuration)),
				Check:  testAccCheckKeycloakRealmClientPolicyProfileWithExecutorMatches(realmName, resourceName, executorName, configuration),
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
	conditionName := "client-attributes"
	configuration := map[string]interface{}{
		"is_negative_logic": false,
		"attributes": []map[string]string{
			{
				"key":   "test-key",
				"value": "test-value",
			},
		},
	}

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmClientPolicyProfile_basicWithPolicy(realmName, profileName, profileDescription, policyName, policyDescription, conditionName, testKeycloakRealmClientPolicyProfile_mapConfig(configuration)),
				Check:  testAccCheckKeycloakRealmClientPolicyProfilePolicyExists(realmName, policyName),
			},
		},
	})
}

func TestAccKeycloakRealmClientPolicyProfile_basicWithPolicyAndJSON(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")
	profileName := "test-profile"
	profileDescription := "Test profile description"
	policyName := "test-policy"
	policyDescription := "Test policy description"
	conditionName := "client-updater-context"
	configuration := map[string]interface{}{
		"is_negative_logic":    false,
		"update-client-source": []string{"ByInitialAccessToken", "ByRegistrationAccessToken"},
	}

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmClientPolicyProfile_basicWithPolicy(realmName, profileName, profileDescription, policyName, policyDescription, conditionName, testKeycloakRealmClientPolicyProfile_mapConfig(configuration)),
				Check:  testAccCheckKeycloakRealmClientPolicyProfilePolicyMatches(realmName, policyName, conditionName, configuration),
			},
		},
	})
}

func testKeycloakRealmClientPolicyProfile_mapConfig(configuration map[string]interface{}) string {
	var s string = "{"
	for k, v := range configuration {
		switch reflect.TypeOf(v).Kind() {
		case reflect.Map, reflect.Slice:
			jsonStr, _ := json.Marshal(v)
			s += fmt.Sprintf("%s = jsonencode(%s)\n", k, string(jsonStr))
		case reflect.String:
			s += fmt.Sprintf("%s = \"%v\"\n", k, v)
		default:
			s += fmt.Sprintf("%s = %v\n", k, v)
		}
	}
	s += "}"
	return s
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

func testKeycloakRealmClientPolicyProfile_basicWithExecutor(realm string, name string, description string, executorName string, configuration string) string {
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
		configuration = %s
	}
}

resource "keycloak_openid_client" "policy-client" {
  client_id   = "policy-test-client"
  realm_id    = keycloak_realm.realm.realm
  access_type = "PUBLIC"

  depends_on = [
    keycloak_realm_client_policy_profile.profile
  ]
}
	`, realm, name, description, executorName, configuration)
}

func testKeycloakRealmClientPolicyProfile_basicWithPolicy(realm string, profileName string, profileDescription string, policyName string, policyDescription string, conditionName string, configuration string) string {
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
    configuration = %s
  }
}

resource "keycloak_openid_client" "policy-client" {
  client_id   = "policy-test-client"
  realm_id    = keycloak_realm.realm.realm
  access_type = "PUBLIC"

  depends_on = [
    keycloak_realm_client_policy_profile.profile
  ]
}
	`, realm, profileName, profileDescription, policyName, policyDescription, conditionName, configuration)
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

func testAccCheckKeycloakRealmClientPolicyProfileWithExecutorMatches(realm string, profileName string, executorName string, configuration map[string]interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		profile, err := keycloakClient.GetRealmClientPolicyProfileByName(testCtx, realm, profileName)
		if err != nil {
			return fmt.Errorf("Client policy profile not found: %s", profileName)
		}

		if profile.Executors[0].Name != executorName {
			return fmt.Errorf("Client policy profile executor not found: %s", executorName)
		}

		for k, got := range profile.Executors[0].Configuration {
			want := configuration[k]

			if !equalsIgnoreType(got, want) {
				return fmt.Errorf("Client policy profile executor configuration does not match: want %v, got %v", want, got)
			}
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

func testAccCheckKeycloakRealmClientPolicyProfilePolicyMatches(realm string, policyName string, conditionName string, configuration map[string]interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		policy, err := keycloakClient.GetRealmClientPolicyProfilePolicyByName(testCtx, realm, policyName)
		if err != nil {
			return fmt.Errorf("Client policy profile policy not found: %s", policyName)
		}

		if policy.Conditions[0].Name != conditionName {
			return fmt.Errorf("Client policy profile policy condition not found: %s", conditionName)
		}

		for k, got := range policy.Conditions[0].Configuration {
			want := configuration[k]

			if !equalsIgnoreType(got, want) {
				return fmt.Errorf("Client policy profile policy condition configuration does not match: want %v, got %v", want, got)
			}
		}

		return nil
	}
}

// Creates and deletes several profiles and policies of the same realm in one apply. Terraform runs
// these operations in parallel; without a shared lock per realm they overwrite each other's list.
func TestAccKeycloakRealmClientPolicyProfile_parallel(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")
	firstNames := []string{"parallel-0", "parallel-1", "parallel-2", "parallel-3", "parallel-4"}
	secondNames := []string{"parallel-0", "parallel-1", "parallel-5", "parallel-6", "parallel-7"}

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmClientPolicyProfile_parallel(realmName, firstNames),
				Check:  testAccCheckKeycloakRealmClientPolicyProfilesAndPoliciesAre(realmName, "parallel-", firstNames),
			},
			{
				Config: testKeycloakRealmClientPolicyProfile_parallel(realmName, secondNames),
				Check:  testAccCheckKeycloakRealmClientPolicyProfilesAndPoliciesAre(realmName, "parallel-", secondNames),
			},
		},
	})
}

func testKeycloakRealmClientPolicyProfile_parallel(realm string, names []string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm = "%s"
}

locals {
	names = %s
}

resource "keycloak_realm_client_policy_profile" "profile" {
	count = length(local.names)

	realm_id = keycloak_realm.realm.realm
	name     = local.names[count.index]
}

resource "keycloak_realm_client_policy_profile_policy" "policy" {
	count = length(local.names)

	realm_id = keycloak_realm.realm.realm
	name     = local.names[count.index]
	profiles = [keycloak_realm_client_policy_profile.profile[count.index].name]

	condition {
		name = "client-type"
		configuration = {
			protocol = "openid-connect"
		}
	}
}
	`, realm, arrayOfStringsForTerraformResource(names))
}

func testAccCheckKeycloakRealmClientPolicyProfilesAndPoliciesAre(realm string, prefix string, expectedNames []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		profiles, err := keycloakClient.GetAllRealmClientPolicyProfiles(testCtx, realm)
		if err != nil {
			return err
		}
		var profileNames []string
		for _, profile := range profiles.Profiles {
			if strings.HasPrefix(profile.Name, prefix) {
				profileNames = append(profileNames, profile.Name)
			}
		}

		policies, err := keycloakClient.GetAllRealmClientPolicyProfilePolices(testCtx, realm)
		if err != nil {
			return err
		}
		var policyNames []string
		for _, policy := range policies.Policies {
			if strings.HasPrefix(policy.Name, prefix) {
				policyNames = append(policyNames, policy.Name)
			}
		}

		sort.Strings(profileNames)
		sort.Strings(policyNames)
		if !slices.Equal(profileNames, expectedNames) {
			return fmt.Errorf("expected client policy profiles %v, got %v", expectedNames, profileNames)
		}
		if !slices.Equal(policyNames, expectedNames) {
			return fmt.Errorf("expected client policies %v, got %v", expectedNames, policyNames)
		}

		return nil
	}
}
