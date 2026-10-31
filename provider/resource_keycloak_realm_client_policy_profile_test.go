package provider

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
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
			{
				ResourceName:      "keycloak_realm_client_policy_profile.profile",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getRealmClientPolicyImportId("keycloak_realm_client_policy_profile.profile"),
			},
			{
				ResourceName:  "keycloak_realm_client_policy_profile.profile",
				ImportState:   true,
				ImportStateId: realmName + "/does-not-exist",
				ExpectError:   regexp.MustCompile("profile with name: does-not-exist not found"),
			},
			{
				ResourceName:  "keycloak_realm_client_policy_profile.profile",
				ImportState:   true,
				ImportStateId: realmName,
				ExpectError:   regexp.MustCompile("Invalid import"),
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
			{
				ResourceName:      "keycloak_realm_client_policy_profile.profile",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getRealmClientPolicyImportId("keycloak_realm_client_policy_profile.profile"),
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
			{
				ResourceName:      "keycloak_realm_client_policy_profile.profile",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getRealmClientPolicyImportId("keycloak_realm_client_policy_profile.profile"),
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
			{
				ResourceName:      "keycloak_realm_client_policy_profile_policy.policy",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getRealmClientPolicyImportId("keycloak_realm_client_policy_profile_policy.policy"),
			},
			{
				ResourceName:  "keycloak_realm_client_policy_profile_policy.policy",
				ImportState:   true,
				ImportStateId: realmName + "/does-not-exist",
				ExpectError:   regexp.MustCompile("policy with name: does-not-exist not found"),
			},
			{
				ResourceName:  "keycloak_realm_client_policy_profile_policy.policy",
				ImportState:   true,
				ImportStateId: realmName,
				ExpectError:   regexp.MustCompile("Invalid import"),
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
			{
				ResourceName:      "keycloak_realm_client_policy_profile_policy.policy",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getRealmClientPolicyImportId("keycloak_realm_client_policy_profile_policy.policy"),
			},
		},
	})
}

func TestAccKeycloakRealmClientPolicyProfile_createAfterManualDestroy(t *testing.T) {
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
			{
				PreConfig: func() {
					profiles, err := keycloakClient.GetAllRealmClientPolicyProfiles(testCtx, realmName)
					if err != nil {
						t.Fatal(err)
					}

					profiles.Profiles = slices.DeleteFunc(profiles.Profiles, func(p keycloak.RealmClientPolicyProfile) bool {
						return p.Name == resourceName
					})

					err = keycloakClient.UpdateRealmClientPolicyProfiles(testCtx, realmName, profiles)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakRealmClientPolicyProfile_basic(realmName, resourceName, description),
				Check:  testAccCheckKeycloakRealmClientPolicyProfileExists(realmName, resourceName),
			},
		},
	})
}

func TestAccKeycloakRealmClientPolicyProfile_policyCreateAfterManualDestroy(t *testing.T) {
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
				Check:  testAccCheckKeycloakRealmClientPolicyProfilePolicyExists(realmName, policyName),
			},
			{
				PreConfig: func() {
					policies, err := keycloakClient.GetAllRealmClientPolicyProfilePolices(testCtx, realmName)
					if err != nil {
						t.Fatal(err)
					}

					policies.Policies = slices.DeleteFunc(policies.Policies, func(p keycloak.RealmClientPolicyProfilePolicy) bool {
						return p.Name == policyName
					})

					err = keycloakClient.UpdateRealmClientPolicyProfilePolicies(testCtx, realmName, policies)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakRealmClientPolicyProfile_basicWithPolicy(realmName, profileName, profileDescription, policyName, policyDescription, conditionName, testKeycloakRealmClientPolicyProfile_mapConfig(configuration)),
				Check:  testAccCheckKeycloakRealmClientPolicyProfilePolicyExists(realmName, policyName),
			},
		},
	})
}

// Profiles created outside of terraform (e.g. in the admin console) use json booleans instead of strings
func TestAccKeycloakRealmClientPolicyProfile_importNonStringConfiguration(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")
	resourceName := "test-profile-non-string"
	description := "Test description"
	executorName := "pkce-enforcer"
	configuration := map[string]interface{}{
		"auto-configure": true,
	}
	config := testKeycloakRealmClientPolicyProfile_basicWithExecutor(realmName, resourceName, description, executorName, testKeycloakRealmClientPolicyProfile_mapConfig(configuration))

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmClientPolicyProfile_realm(realmName),
			},
			{
				PreConfig: func() {
					testAccCreateKeycloakRealmClientPolicyProfile(t, realmName, keycloak.RealmClientPolicyProfile{
						Name:        resourceName,
						Description: description,
						Executors: []keycloak.RealmClientPolicyProfileExecutor{
							{Name: executorName, Configuration: configuration},
						},
					})
				},
				Config:             config,
				ResourceName:       "keycloak_realm_client_policy_profile.profile",
				ImportState:        true,
				ImportStateId:      realmName + "/" + resourceName,
				ImportStatePersist: true,
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					state, err := getImportedInstanceState(states, fmt.Sprintf("%s/realm-client-policy-profiles/%s", realmName, resourceName))
					if err != nil {
						return err
					}
					if got := state.Attributes["executor.0.configuration.auto-configure"]; got != "true" {
						return fmt.Errorf("expected executor configuration auto-configure to be \"true\", got %q", got)
					}
					return nil
				},
			},
			{
				Config: config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("keycloak_realm_client_policy_profile.profile", plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

// Policies created outside of terraform (e.g. in the admin console) use json booleans instead of strings
func TestAccKeycloakRealmClientPolicyProfile_policyImportNonStringConfiguration(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")
	profileName := "test-profile"
	profileDescription := "Test profile description"
	policyName := "test-policy-non-string"
	policyDescription := "Test policy description"
	conditionName := "client-access-type"
	configuration := map[string]interface{}{
		"is_negative_logic": false,
		"type":              []interface{}{"confidential"},
	}
	config := testKeycloakRealmClientPolicyProfile_basicWithPolicy(realmName, profileName, profileDescription, policyName, policyDescription, conditionName, testKeycloakRealmClientPolicyProfile_mapConfig(configuration))

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmClientPolicyProfile_realm(realmName),
			},
			{
				PreConfig: func() {
					testAccCreateKeycloakRealmClientPolicyProfile(t, realmName, keycloak.RealmClientPolicyProfile{
						Name:        profileName,
						Description: profileDescription,
						Executors:   []keycloak.RealmClientPolicyProfileExecutor{},
					})

					policies, err := keycloakClient.GetAllRealmClientPolicyProfilePolices(testCtx, realmName)
					if err != nil {
						t.Fatal(err)
					}

					policies.Policies = append(policies.Policies, keycloak.RealmClientPolicyProfilePolicy{
						Name:        policyName,
						Description: policyDescription,
						Enabled:     true,
						Profiles:    []string{profileName},
						Conditions: []keycloak.RealmClientPolicyProfilePolicyCondition{
							{Name: conditionName, Configuration: configuration},
						},
					})

					err = keycloakClient.UpdateRealmClientPolicyProfilePolicies(testCtx, realmName, policies)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config:             config,
				ResourceName:       "keycloak_realm_client_policy_profile.profile",
				ImportState:        true,
				ImportStateId:      realmName + "/" + profileName,
				ImportStatePersist: true,
			},
			{
				Config:             config,
				ResourceName:       "keycloak_realm_client_policy_profile_policy.policy",
				ImportState:        true,
				ImportStateId:      realmName + "/" + policyName,
				ImportStatePersist: true,
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					state, err := getImportedInstanceState(states, fmt.Sprintf("%s/realm-client-policy-profile-policies/%s", realmName, policyName))
					if err != nil {
						return err
					}
					if got := state.Attributes["condition.0.configuration.is_negative_logic"]; got != "false" {
						return fmt.Errorf("expected condition configuration is_negative_logic to be \"false\", got %q", got)
					}
					if got := state.Attributes["condition.0.configuration.type"]; got != `["confidential"]` {
						return fmt.Errorf("expected condition configuration type to be [\"confidential\"], got %q", got)
					}
					return nil
				},
			},
			{
				Config: config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("keycloak_realm_client_policy_profile.profile", plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction("keycloak_realm_client_policy_profile_policy.policy", plancheck.ResourceActionNoop),
					},
				},
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

func testKeycloakRealmClientPolicyProfile_realm(realm string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm = "%s"
}
	`, realm)
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

func testAccCreateKeycloakRealmClientPolicyProfile(t *testing.T, realm string, profile keycloak.RealmClientPolicyProfile) {
	profiles, err := keycloakClient.GetAllRealmClientPolicyProfiles(testCtx, realm)
	if err != nil {
		t.Fatal(err)
	}

	profiles.Profiles = append(profiles.Profiles, profile)

	err = keycloakClient.UpdateRealmClientPolicyProfiles(testCtx, realm, profiles)
	if err != nil {
		t.Fatal(err)
	}
}

func getRealmClientPolicyImportId(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}

		return rs.Primary.Attributes["realm_id"] + "/" + rs.Primary.Attributes["name"], nil
	}
}

func getImportedInstanceState(states []*terraform.InstanceState, id string) (*terraform.InstanceState, error) {
	for _, state := range states {
		if state.ID == id {
			return state, nil
		}
	}

	return nil, fmt.Errorf("imported resource with id %s not found", id)
}
