package provider

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakRealmKeystoreEcdhGenerated_basic(t *testing.T) {
	t.Parallel()

	ecdhName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRealmKeystoreEcdhGeneratedDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmKeystoreEcdhGenerated_basic(ecdhName),
				Check:  testAccCheckRealmKeystoreEcdhGeneratedExists("keycloak_realm_keystore_ecdh_generated.realm_ecdh"),
			},
			{
				ResourceName:      "keycloak_realm_keystore_ecdh_generated.realm_ecdh",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getRealmKeystoreGenericImportId("keycloak_realm_keystore_ecdh_generated.realm_ecdh"),
			},
		},
	})
}

func TestAccKeycloakRealmKeystoreEcdhGenerated_createAfterManualDestroy(t *testing.T) {
	t.Parallel()

	var ecdh = &keycloak.RealmKeystoreEcdhGenerated{}

	fullNameKeystoreName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRealmKeystoreEcdhGeneratedDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmKeystoreEcdhGenerated_basic(fullNameKeystoreName),
				Check:  testAccCheckRealmKeystoreEcdhGeneratedFetch("keycloak_realm_keystore_ecdh_generated.realm_ecdh", ecdh),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteRealmKeystoreEcdhGenerated(testCtx, ecdh.RealmId, ecdh.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakRealmKeystoreEcdhGenerated_basic(fullNameKeystoreName),
				Check:  testAccCheckRealmKeystoreEcdhGeneratedFetch("keycloak_realm_keystore_ecdh_generated.realm_ecdh", ecdh),
			},
		},
	})
}

func TestAccKeycloakRealmKeystoreEcdhGenerated_ellipticCurveValidation(t *testing.T) {
	t.Parallel()

	ecdhName := acctest.RandomWithPrefix("tf-acc")
	ellipticCurve := randomStringInSlice(keycloakRealmKeystoreEcdhGeneratedEllipticCurve)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRealmKeystoreEcdhGeneratedDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakRealmKeystoreEcdhGenerated_basicWithAttrValidation(ecdhName, "elliptic_curve_key", acctest.RandString(10)),
				ExpectError: regexp.MustCompile("expected elliptic_curve_key to be one of .+ got .+"),
			},
			{
				Config: testKeycloakRealmKeystoreEcdhGenerated_basicWithAttrValidation(ecdhName, "elliptic_curve_key", ellipticCurve),
				Check:  testAccCheckRealmKeystoreEcdhGeneratedExists("keycloak_realm_keystore_ecdh_generated.realm_ecdh"),
			},
		},
	})
}

func TestAccKeycloakRealmKeystoreEcdhGenerated_algorithmValidation(t *testing.T) {
	t.Parallel()

	ecdhName := acctest.RandomWithPrefix("tf-acc")
	algorithm := randomStringInSlice(keycloakRealmKeystoreEcdhGeneratedAlgorithm)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRealmKeystoreEcdhGeneratedDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakRealmKeystoreEcdhGenerated_basicWithAttrValidation(ecdhName, "algorithm", acctest.RandString(10)),
				ExpectError: regexp.MustCompile("expected algorithm to be one of .+ got .+"),
			},
			{
				Config: testKeycloakRealmKeystoreEcdhGenerated_basicWithAttrValidation(ecdhName, "algorithm", algorithm),
				Check:  testAccCheckRealmKeystoreEcdhGeneratedExists("keycloak_realm_keystore_ecdh_generated.realm_ecdh"),
			},
		},
	})
}

func TestAccKeycloakRealmKeystoreEcdhGenerated_updateRealmKeystoreEcdhGenerated(t *testing.T) {
	t.Parallel()

	enabled := randomBool()
	active := randomBool()

	groupKeystoreOne := &keycloak.RealmKeystoreEcdhGenerated{
		Name:          acctest.RandString(10),
		RealmId:       testAccRealmUserFederation.Realm,
		Enabled:       enabled,
		Active:        active,
		Priority:      acctest.RandIntRange(0, 100),
		EllipticCurve: randomStringInSlice(keycloakRealmKeystoreEcdhGeneratedEllipticCurve),
		Algorithm:     randomStringInSlice(keycloakRealmKeystoreEcdhGeneratedAlgorithm),
	}

	groupKeystoreTwo := &keycloak.RealmKeystoreEcdhGenerated{
		Name:          acctest.RandString(10),
		RealmId:       testAccRealmUserFederation.Realm,
		Enabled:       !enabled,
		Active:        !active,
		Priority:      acctest.RandIntRange(0, 100),
		EllipticCurve: randomStringInSlice(keycloakRealmKeystoreEcdhGeneratedEllipticCurve),
		Algorithm:     randomStringInSlice(keycloakRealmKeystoreEcdhGeneratedAlgorithm),
	}

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRealmKeystoreEcdhGeneratedDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmKeystoreEcdhGenerated_basicFromInterface(groupKeystoreOne),
				Check:  testAccCheckRealmKeystoreEcdhGeneratedMatches("keycloak_realm_keystore_ecdh_generated.realm_ecdh", groupKeystoreOne),
			},
			{
				Config: testKeycloakRealmKeystoreEcdhGenerated_basicFromInterface(groupKeystoreTwo),
				Check:  testAccCheckRealmKeystoreEcdhGeneratedMatches("keycloak_realm_keystore_ecdh_generated.realm_ecdh", groupKeystoreTwo),
			},
		},
	})
}

func TestAccKeycloakRealmKeystoreEcdhGenerated_updateKeepsKeyUnlessCurveChanges(t *testing.T) {
	t.Parallel()

	ecdhName := acctest.RandomWithPrefix("tf-acc")
	var kid string

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRealmKeystoreEcdhGeneratedDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmKeystoreEcdhGenerated_curve(ecdhName, "100", "P-256"),
				Check:  testAccCheckRealmKeystoreEcdhGeneratedKid("keycloak_realm_keystore_ecdh_generated.realm_ecdh", &kid, true),
			},
			{
				// update without a curve change: keycloak must keep the generated key pair
				Config: testKeycloakRealmKeystoreEcdhGenerated_curve(ecdhName, "200", "P-256"),
				Check:  testAccCheckRealmKeystoreEcdhGeneratedKid("keycloak_realm_keystore_ecdh_generated.realm_ecdh", &kid, true),
			},
			{
				// update with a curve change: keycloak must generate a new key pair
				Config: testKeycloakRealmKeystoreEcdhGenerated_curve(ecdhName, "200", "P-384"),
				Check:  testAccCheckRealmKeystoreEcdhGeneratedKid("keycloak_realm_keystore_ecdh_generated.realm_ecdh", &kid, false),
			},
		},
	})
}

func TestAccKeycloakRealmKeystoreEcdhGenerated_parentIdDifferentFromRealmName(t *testing.T) {
	t.Parallel()

	realmName := acctest.RandomWithPrefix("tf-acc")
	internalId := acctest.RandomWithPrefix("tf-acc")
	ecdhName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRealmKeystoreEcdhGeneratedDestroy(),
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
				Config: testKeycloakRealmKeystoreEcdhGenerated_parentId(realmName, ecdhName, "100"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRealmKeystoreEcdhGeneratedExists("keycloak_realm_keystore_ecdh_generated.realm_ecdh"),
					testAccCheckRealmKeystoreEcdhGeneratedParentId("keycloak_realm_keystore_ecdh_generated.realm_ecdh", internalId),
				),
			},
			{
				// update: the omitted parentId must leave the stored parent untouched
				Config: testKeycloakRealmKeystoreEcdhGenerated_parentId(realmName, ecdhName, "200"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRealmKeystoreEcdhGeneratedExists("keycloak_realm_keystore_ecdh_generated.realm_ecdh"),
					testAccCheckRealmKeystoreEcdhGeneratedParentId("keycloak_realm_keystore_ecdh_generated.realm_ecdh", internalId),
				),
			},
		},
	})
}

func testAccCheckRealmKeystoreEcdhGeneratedExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getKeycloakRealmKeystoreEcdhGeneratedFromState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckRealmKeystoreEcdhGeneratedFetch(resourceName string, keystore *keycloak.RealmKeystoreEcdhGenerated) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedKeystore, err := getKeycloakRealmKeystoreEcdhGeneratedFromState(s, resourceName)
		if err != nil {
			return err
		}

		keystore.Id = fetchedKeystore.Id
		keystore.RealmId = fetchedKeystore.RealmId

		return nil
	}
}

func testAccCheckRealmKeystoreEcdhGeneratedParentId(resourceName, expectedParentId string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		keystore, err := getKeycloakRealmKeystoreEcdhGeneratedFromState(s, resourceName)
		if err != nil {
			return err
		}

		if keystore.ParentId != expectedParentId {
			return fmt.Errorf("expected ecdh generated keystore %s to have parent id %s, but got %s", keystore.Id, expectedParentId, keystore.ParentId)
		}

		return nil
	}
}

func testAccCheckRealmKeystoreEcdhGeneratedMatches(resourceName string, expected *keycloak.RealmKeystoreEcdhGenerated) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetched, err := getKeycloakRealmKeystoreEcdhGeneratedFromState(s, resourceName)
		if err != nil {
			return err
		}

		if fetched.Name != expected.Name {
			return fmt.Errorf("expected name %s, got %s", expected.Name, fetched.Name)
		}
		if fetched.Priority != expected.Priority {
			return fmt.Errorf("expected priority %d, got %d", expected.Priority, fetched.Priority)
		}
		if fetched.EllipticCurve != expected.EllipticCurve {
			return fmt.Errorf("expected elliptic curve %s, got %s", expected.EllipticCurve, fetched.EllipticCurve)
		}
		if fetched.Algorithm != expected.Algorithm {
			return fmt.Errorf("expected algorithm %s, got %s", expected.Algorithm, fetched.Algorithm)
		}
		if fetched.Active != expected.Active {
			return fmt.Errorf("expected active %t, got %t", expected.Active, fetched.Active)
		}
		if fetched.Enabled != expected.Enabled {
			return fmt.Errorf("expected enabled %t, got %t", expected.Enabled, fetched.Enabled)
		}

		return nil
	}
}

// testAccCheckRealmKeystoreEcdhGeneratedKid looks up the key that keycloak generated for the keystore and compares its kid
// with the one recorded by a previous step, then records the current kid for the next step.
func testAccCheckRealmKeystoreEcdhGeneratedKid(resourceName string, kid *string, expectUnchanged bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		id := rs.Primary.ID
		realm := rs.Primary.Attributes["realm_id"]
		algorithm := rs.Primary.Attributes["algorithm"]

		keys, err := keycloakClient.GetRealmKeys(testCtx, realm)
		if err != nil {
			return fmt.Errorf("error getting keys of realm %s: %s", realm, err)
		}

		var key *keycloak.Key
		for i := range keys.Keys {
			if keys.Keys[i].ProviderId != nil && *keys.Keys[i].ProviderId == id {
				key = &keys.Keys[i]
				break
			}
		}
		if key == nil || key.Kid == nil {
			return fmt.Errorf("no key found for ecdh keystore with id %s", id)
		}
		if key.Type == nil || *key.Type != "EC" {
			return fmt.Errorf("expected key of ecdh keystore %s to have type EC, got %v", id, key.Type)
		}
		if key.Algorithm == nil || *key.Algorithm != algorithm {
			return fmt.Errorf("expected key of ecdh keystore %s to have algorithm %s, got %v", id, algorithm, key.Algorithm)
		}

		if *kid != "" {
			if expectUnchanged && *key.Kid != *kid {
				return fmt.Errorf("expected kid of ecdh keystore %s to stay %s, got %s", id, *kid, *key.Kid)
			}
			if !expectUnchanged && *key.Kid == *kid {
				return fmt.Errorf("expected kid of ecdh keystore %s to change, but it is still %s", id, *kid)
			}
		}

		*kid = *key.Kid

		return nil
	}
}

func testAccCheckRealmKeystoreEcdhGeneratedDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_realm_keystore_ecdh_generated" {
				continue
			}

			id := rs.Primary.ID
			realm := rs.Primary.Attributes["realm_id"]

			ecdhKeystore, err := keycloakClient.GetRealmKeystoreEcdhGenerated(testCtx, realm, id)
			if err != nil && !keycloak.ErrorIs404(err) {
				return err
			}
			if ecdhKeystore != nil {
				return fmt.Errorf("ecdh keystore with id %s still exists", id)
			}
		}

		return nil
	}
}

func getKeycloakRealmKeystoreEcdhGeneratedFromState(s *terraform.State,
	resourceName string) (*keycloak.RealmKeystoreEcdhGenerated,
	error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm_id"]

	realmKeystore, err := keycloakClient.GetRealmKeystoreEcdhGenerated(testCtx, realm, id)
	if err != nil {
		return nil, fmt.Errorf("error getting ecdh keystore with id %s: %s", id, err)
	}

	return realmKeystore, nil
}

func testKeycloakRealmKeystoreEcdhGenerated_basic(ecdhName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_realm_keystore_ecdh_generated" "realm_ecdh" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id

    priority           = 100
    elliptic_curve_key = "P-384"
    algorithm          = "ECDH-ES+A192KW"
}
	`, testAccRealmUserFederation.Realm, ecdhName)
}

func testKeycloakRealmKeystoreEcdhGenerated_curve(ecdhName, priority, ellipticCurve string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_realm_keystore_ecdh_generated" "realm_ecdh" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id

    priority           = %s
    elliptic_curve_key = "%s"
}
	`, testAccRealmUserFederation.Realm, ecdhName, priority, ellipticCurve)
}

func testKeycloakRealmKeystoreEcdhGenerated_parentId(realmName, ecdhName, priority string) string {
	return fmt.Sprintf(`
resource "keycloak_realm_keystore_ecdh_generated" "realm_ecdh" {
	name      = "%s"
	realm_id  = "%s"

    priority           = %s
    elliptic_curve_key = "P-384"
}
	`, ecdhName, realmName, priority)
}

func testKeycloakRealmKeystoreEcdhGenerated_basicWithAttrValidation(ecdhName, attr, val string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_realm_keystore_ecdh_generated" "realm_ecdh" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id

	%s         = "%s"
}
	`, testAccRealmUserFederation.Realm, ecdhName, attr, val)
}

func testKeycloakRealmKeystoreEcdhGenerated_basicFromInterface(keystore *keycloak.RealmKeystoreEcdhGenerated) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_realm_keystore_ecdh_generated" "realm_ecdh" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id

    enabled            = %t
    active             = %t
    priority           = %s
    elliptic_curve_key = "%s"
    algorithm          = "%s"
}
	`, testAccRealmUserFederation.Realm, keystore.Name, keystore.Enabled, keystore.Active,
		strconv.Itoa(keystore.Priority), keystore.EllipticCurve, keystore.Algorithm)
}
