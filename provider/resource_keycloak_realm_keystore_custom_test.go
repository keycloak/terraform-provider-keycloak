package provider

import (
	"fmt"
	"regexp"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

// The tests use built-in key providers without a typed resource, a custom SPI behaves the same way.

func TestAccKeycloakRealmKeystoreCustom_basic(t *testing.T) {
	t.Parallel()

	keystoreName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRealmKeystoreCustomDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmKeystoreCustom_ecdh(testAccRealmUserFederation.Realm, keystoreName, map[string]string{
					"priority":             "100",
					"ecdhEllipticCurveKey": "P-384",
					"ecdhAlgorithm":        "ECDH-ES+A192KW",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRealmKeystoreCustomConfig("keycloak_realm_keystore_custom.custom", map[string]string{
						"priority":             "100",
						"ecdhEllipticCurveKey": "P-384",
						"ecdhAlgorithm":        "ECDH-ES+A192KW",
					}, nil),
					resource.TestCheckResourceAttr("keycloak_realm_keystore_custom.custom", "provider_id", "ecdh-generated"),
					resource.TestCheckResourceAttr("keycloak_realm_keystore_custom.custom", "config.%", "3"),
				),
			},
			{
				ResourceName:      "keycloak_realm_keystore_custom.custom",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getRealmKeystoreGenericImportId("keycloak_realm_keystore_custom.custom"),
			},
			{
				ResourceName:      "keycloak_realm_keystore_custom.custom",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     fmt.Sprintf("%s/%s/ecdh-generated", testAccRealmUserFederation.Realm, keystoreName),
			},
		},
	})
}

func TestAccKeycloakRealmKeystoreCustom_eddsaGenerated(t *testing.T) {
	t.Parallel()

	keystoreName := acctest.RandomWithPrefix("tf-acc")
	config := map[string]string{
		"priority":              "100",
		"eddsaEllipticCurveKey": "Ed448",
	}

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRealmKeystoreCustomDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmKeystoreCustom_provider(testAccRealmUserFederation.Realm, keystoreName, "eddsa-generated", config),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRealmKeystoreCustomConfig("keycloak_realm_keystore_custom.custom", config, nil),
					testAccCheckRealmKeystoreCustomKeyStatus("keycloak_realm_keystore_custom.custom", "ACTIVE"),
				),
			},
		},
	})
}

func TestAccKeycloakRealmKeystoreCustom_update(t *testing.T) {
	t.Parallel()

	keystoreName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRealmKeystoreCustomDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmKeystoreCustom_ecdh(testAccRealmUserFederation.Realm, keystoreName, map[string]string{
					"priority":             "100",
					"enabled":              "true",
					"active":               "true",
					"ecdhEllipticCurveKey": "P-256",
					"ecdhAlgorithm":        "ECDH-ES",
				}),
				Check: testAccCheckRealmKeystoreCustomConfig("keycloak_realm_keystore_custom.custom", map[string]string{
					"priority":             "100",
					"enabled":              "true",
					"active":               "true",
					"ecdhEllipticCurveKey": "P-256",
					"ecdhAlgorithm":        "ECDH-ES",
				}, nil),
			},
			{
				// change values and remove keys: keycloak must drop the removed keys
				Config: testKeycloakRealmKeystoreCustom_ecdh(testAccRealmUserFederation.Realm, keystoreName+"-renamed", map[string]string{
					"priority":             "200",
					"active":               "false",
					"ecdhEllipticCurveKey": "P-521",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRealmKeystoreCustomConfig("keycloak_realm_keystore_custom.custom", map[string]string{
						"priority":             "200",
						"active":               "false",
						"ecdhEllipticCurveKey": "P-521",
					}, []string{"enabled", "ecdhAlgorithm"}),
					testAccCheckRealmKeystoreCustomName("keycloak_realm_keystore_custom.custom", keystoreName+"-renamed"),
					resource.TestCheckResourceAttr("keycloak_realm_keystore_custom.custom", "name", keystoreName+"-renamed"),
					resource.TestCheckResourceAttr("keycloak_realm_keystore_custom.custom", "config.%", "3"),
				),
			},
		},
	})
}

func TestAccKeycloakRealmKeystoreCustom_createAfterManualDestroy(t *testing.T) {
	t.Parallel()

	var keystore = &keycloak.RealmKeystoreCustom{}

	keystoreName := acctest.RandomWithPrefix("tf-acc")
	config := map[string]string{"priority": "100"}

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRealmKeystoreCustomDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmKeystoreCustom_ecdh(testAccRealmUserFederation.Realm, keystoreName, config),
				Check:  testAccCheckRealmKeystoreCustomFetch("keycloak_realm_keystore_custom.custom", keystore),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteRealmKeystoreCustom(testCtx, keystore.RealmId, keystore.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakRealmKeystoreCustom_ecdh(testAccRealmUserFederation.Realm, keystoreName, config),
				Check:  testAccCheckRealmKeystoreCustomFetch("keycloak_realm_keystore_custom.custom", keystore),
			},
		},
	})
}

func TestAccKeycloakRealmKeystoreCustom_driftIsReverted(t *testing.T) {
	t.Parallel()

	var keystore = &keycloak.RealmKeystoreCustom{}

	keystoreName := acctest.RandomWithPrefix("tf-acc")
	config := map[string]string{
		"priority":      "100",
		"ecdhAlgorithm": "ECDH-ES+A128KW",
	}

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRealmKeystoreCustomDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmKeystoreCustom_ecdh(testAccRealmUserFederation.Realm, keystoreName, config),
				Check:  testAccCheckRealmKeystoreCustomFetch("keycloak_realm_keystore_custom.custom", keystore),
			},
			{
				PreConfig: func() {
					keystore.Name = keystoreName
					keystore.ProviderId = "ecdh-generated"
					keystore.Config = map[string]string{"priority": "999"}

					err := keycloakClient.UpdateRealmKeystoreCustom(testCtx, keystore, []string{"ecdhAlgorithm"})
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakRealmKeystoreCustom_ecdh(testAccRealmUserFederation.Realm, keystoreName, config),
				Check:  testAccCheckRealmKeystoreCustomConfig("keycloak_realm_keystore_custom.custom", config, nil),
			},
		},
	})
}

func TestAccKeycloakRealmKeystoreCustom_validation(t *testing.T) {
	t.Parallel()

	keystoreName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRealmKeystoreCustomDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakRealmKeystoreCustom_provider(testAccRealmUserFederation.Realm, keystoreName, "tf-acc-not-installed", map[string]string{"priority": "100"}),
				ExpectError: regexp.MustCompile("key provider with id tf-acc-not-installed is not installed on the server"),
			},
			{
				Config:      testKeycloakRealmKeystoreCustom_ecdh(testAccRealmUserFederation.Realm, keystoreName, map[string]string{"keySize": "2048"}),
				ExpectError: regexp.MustCompile(`config key "keySize" is not declared by key provider ecdh-generated`),
			},
			{
				Config:      testKeycloakRealmKeystoreCustom_ecdh(testAccRealmUserFederation.Realm, keystoreName, map[string]string{"priority": ""}),
				ExpectError: regexp.MustCompile(`config key "priority" must not be empty`),
			},
		},
	})
}

func TestAccKeycloakRealmKeystoreCustom_secretConfig(t *testing.T) {
	t.Parallel()

	keystoreName := acctest.RandomWithPrefix("tf-acc")
	privateKey, _ := generateKeyAndCert(2048)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRealmKeystoreCustomDestroy(),
		Steps: []resource.TestStep{
			{
				// keycloak masks the secret private key and generates the certificate, neither must cause a diff
				Config: testKeycloakRealmKeystoreCustom_provider(testAccRealmUserFederation.Realm, keystoreName, "rsa", map[string]string{
					"priority":   "100",
					"privateKey": privateKey,
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRealmKeystoreCustomConfig("keycloak_realm_keystore_custom.custom", map[string]string{
						"priority":   "100",
						"privateKey": keycloak.RealmKeystoreSecretValue,
					}, nil),
					resource.TestCheckResourceAttr("keycloak_realm_keystore_custom.custom", "config.privateKey", privateKey),
					resource.TestCheckNoResourceAttr("keycloak_realm_keystore_custom.custom", "config.certificate"),
					testAccCheckRealmKeystoreCustomKeyStatus("keycloak_realm_keystore_custom.custom", "ACTIVE"),
				),
			},
			{
				// an update sends the private key again, keycloak must keep using it
				Config: testKeycloakRealmKeystoreCustom_provider(testAccRealmUserFederation.Realm, keystoreName, "rsa", map[string]string{
					"priority":   "200",
					"privateKey": privateKey,
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRealmKeystoreCustomConfig("keycloak_realm_keystore_custom.custom", map[string]string{
						"priority": "200",
					}, nil),
					resource.TestCheckResourceAttr("keycloak_realm_keystore_custom.custom", "config.privateKey", privateKey),
					testAccCheckRealmKeystoreCustomKeyStatus("keycloak_realm_keystore_custom.custom", "ACTIVE"),
				),
			},
			{
				// the import can't read the private key, so it stores the masked value
				ResourceName:            "keycloak_realm_keystore_custom.custom",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       getRealmKeystoreGenericImportId("keycloak_realm_keystore_custom.custom"),
				ImportStateVerifyIgnore: []string{"config"},
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					if len(states) != 1 {
						return fmt.Errorf("expected 1 imported state, got %d", len(states))
					}

					attributes := states[0].Attributes
					if attributes["config.privateKey"] != keycloak.RealmKeystoreSecretValue {
						return fmt.Errorf("expected imported private key to be masked, got %q", attributes["config.privateKey"])
					}
					if attributes["config.certificate"] == "" {
						return fmt.Errorf("expected imported config to contain the generated certificate")
					}

					return nil
				},
			},
		},
	})
}

func TestAccKeycloakRealmKeystoreCustom_parentIdDifferentFromRealmName(t *testing.T) {
	t.Parallel()

	realmName := acctest.RandomWithPrefix("tf-acc")
	internalId := acctest.RandomWithPrefix("tf-acc")
	keystoreName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRealmKeystoreCustomDestroy(),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					err := keycloakClient.NewRealm(testCtx, &keycloak.Realm{
						Realm: realmName,
						Id:    internalId,
					})
					if err != nil {
						t.Fatal(err)
					}

					t.Cleanup(func() {
						if err := keycloakClient.DeleteRealm(testCtx, realmName); err != nil {
							t.Logf("failed to clean up realm %s: %s", realmName, err)
						}
					})
				},
				// create: keycloak must default the omitted parentId to the realm's internal id
				Config: testKeycloakRealmKeystoreCustom_ecdh(realmName, keystoreName, map[string]string{"priority": "100"}),
				Check:  testAccCheckRealmKeystoreCustomParentId("keycloak_realm_keystore_custom.custom", internalId),
			},
			{
				// update: the omitted parentId must leave the stored parent untouched
				Config: testKeycloakRealmKeystoreCustom_ecdh(realmName, keystoreName, map[string]string{"priority": "200"}),
				Check:  testAccCheckRealmKeystoreCustomParentId("keycloak_realm_keystore_custom.custom", internalId),
			},
		},
	})
}

func testAccCheckRealmKeystoreCustomFetch(resourceName string, keystore *keycloak.RealmKeystoreCustom) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedKeystore, err := getKeycloakRealmKeystoreCustomFromState(s, resourceName)
		if err != nil {
			return err
		}

		keystore.Id = fetchedKeystore.Id
		keystore.RealmId = fetchedKeystore.RealmId

		return nil
	}
}

// testAccCheckRealmKeystoreCustomConfig compares the config that keycloak returns with the expected values and
// checks that the absent keys are not set.
func testAccCheckRealmKeystoreCustomConfig(resourceName string, expected map[string]string, absent []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		keystore, err := getKeycloakRealmKeystoreCustomFromState(s, resourceName)
		if err != nil {
			return err
		}

		for k, v := range expected {
			if keystore.Config[k] != v {
				return fmt.Errorf("expected config key %s of keystore %s to be %q, got %q", k, keystore.Id, v, keystore.Config[k])
			}
		}

		for _, k := range absent {
			if v, ok := keystore.Config[k]; ok {
				return fmt.Errorf("expected config key %s of keystore %s to be removed, got %q", k, keystore.Id, v)
			}
		}

		return nil
	}
}

func testAccCheckRealmKeystoreCustomName(resourceName, expectedName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		keystore, err := getKeycloakRealmKeystoreCustomFromState(s, resourceName)
		if err != nil {
			return err
		}

		if keystore.Name != expectedName {
			return fmt.Errorf("expected custom keystore %s to have name %s, but got %s", keystore.Id, expectedName, keystore.Name)
		}

		return nil
	}
}

func testAccCheckRealmKeystoreCustomParentId(resourceName, expectedParentId string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		keystore, err := getKeycloakRealmKeystoreCustomFromState(s, resourceName)
		if err != nil {
			return err
		}

		if keystore.ParentId != expectedParentId {
			return fmt.Errorf("expected custom keystore %s to have parent id %s, but got %s", keystore.Id, expectedParentId, keystore.ParentId)
		}

		return nil
	}
}

// testAccCheckRealmKeystoreCustomKeyStatus checks the status of the key that the keystore provides to the realm.
func testAccCheckRealmKeystoreCustomKeyStatus(resourceName, expectedStatus string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		id := rs.Primary.ID
		realm := rs.Primary.Attributes["realm_id"]

		keys, err := keycloakClient.GetRealmKeys(testCtx, realm)
		if err != nil {
			return fmt.Errorf("error getting keys of realm %s: %s", realm, err)
		}

		for _, key := range keys.Keys {
			if key.ProviderId != nil && *key.ProviderId == id {
				if key.Status == nil || *key.Status != expectedStatus {
					return fmt.Errorf("expected key of custom keystore %s to have status %s, got %v", id, expectedStatus, key.Status)
				}

				return nil
			}
		}

		return fmt.Errorf("no key found for custom keystore with id %s", id)
	}
}

func testAccCheckRealmKeystoreCustomDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_realm_keystore_custom" {
				continue
			}

			id := rs.Primary.ID
			realm := rs.Primary.Attributes["realm_id"]

			keystore, err := keycloakClient.GetRealmKeystoreCustom(testCtx, realm, id)
			if err != nil && !keycloak.ErrorIs404(err) {
				return err
			}
			if keystore != nil {
				return fmt.Errorf("custom keystore with id %s still exists", id)
			}
		}

		return nil
	}
}

func getKeycloakRealmKeystoreCustomFromState(s *terraform.State, resourceName string) (*keycloak.RealmKeystoreCustom, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm_id"]

	keystore, err := keycloakClient.GetRealmKeystoreCustom(testCtx, realm, id)
	if err != nil {
		return nil, fmt.Errorf("error getting custom keystore with id %s: %s", id, err)
	}

	return keystore, nil
}

func testKeycloakRealmKeystoreCustom_ecdh(realm, keystoreName string, config map[string]string) string {
	return testKeycloakRealmKeystoreCustom_provider(realm, keystoreName, "ecdh-generated", config)
}

func testKeycloakRealmKeystoreCustom_provider(realm, keystoreName, providerId string, config map[string]string) string {
	keys := make([]string, 0, len(config))
	for k := range config {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	configHcl := ""
	for _, k := range keys {
		configHcl += fmt.Sprintf("\t\t%s = %q\n", k, config[k])
	}

	return fmt.Sprintf(`
resource "keycloak_realm_keystore_custom" "custom" {
	name        = "%s"
	realm_id    = "%s"
	provider_id = "%s"

	config = {
%s	}
}
	`, keystoreName, realm, providerId, configHcl)
}
