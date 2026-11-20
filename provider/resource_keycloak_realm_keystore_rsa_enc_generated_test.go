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

func TestAccKeycloakRealmKeystoreRsaEncGenerated_basic(t *testing.T) {
	t.Parallel()

	rsaName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRealmKeystoreRsaEncGeneratedDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmKeystoreRsaEncGenerated_basic(rsaName),
				Check:  testAccCheckRealmKeystoreRsaEncGeneratedExists("keycloak_realm_keystore_rsa_enc_generated.realm_rsa"),
			},
			{
				ResourceName:      "keycloak_realm_keystore_rsa_enc_generated.realm_rsa",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getRealmKeystoreGenericImportId("keycloak_realm_keystore_rsa_enc_generated.realm_rsa"),
			},
		},
	})
}

func TestAccKeycloakRealmKeystoreRsaEncGenerated_createAfterManualDestroy(t *testing.T) {
	t.Parallel()

	var rsa = &keycloak.RealmKeystoreRsaEncGenerated{}

	fullNameKeystoreName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRealmKeystoreRsaEncGeneratedDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmKeystoreRsaEncGenerated_basic(fullNameKeystoreName),
				Check:  testAccCheckRealmKeystoreRsaEncGeneratedFetch("keycloak_realm_keystore_rsa_enc_generated.realm_rsa", rsa),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteRealmKeystoreRsaEncGenerated(testCtx, rsa.RealmId, rsa.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakRealmKeystoreRsaEncGenerated_basic(fullNameKeystoreName),
				Check:  testAccCheckRealmKeystoreRsaEncGeneratedFetch("keycloak_realm_keystore_rsa_enc_generated.realm_rsa", rsa),
			},
		},
	})
}

func TestAccKeycloakRealmKeystoreRsaEncGenerated_keySizeValidation(t *testing.T) {
	t.Parallel()

	rsaName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRealmKeystoreRsaEncGeneratedDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmKeystoreRsaEncGenerated_basicWithAttrValidation(rsaName, "key_size",
					strconv.Itoa(acctest.RandIntRange(0, 1000)*2+1)),
				ExpectError: regexp.MustCompile("expected key_size to be one of .+ got .+"),
			},
			{
				Config: testKeycloakRealmKeystoreRsaEncGenerated_basicWithAttrValidation(rsaName, "key_size", "2048"),
				Check:  testAccCheckRealmKeystoreRsaEncGeneratedExists("keycloak_realm_keystore_rsa_enc_generated.realm_rsa"),
			},
		},
	})
}

func TestAccKeycloakRealmKeystoreRsaEncGenerated_algorithmValidation(t *testing.T) {
	t.Parallel()

	algorithm := randomStringInSlice(keycloakRealmKeystoreRsaEncGeneratedAlgorithm)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRealmKeystoreRsaEncGeneratedDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmKeystoreRsaEncGenerated_basicWithAttrValidation(algorithm, "algorithm",
					acctest.RandString(10)),
				ExpectError: regexp.MustCompile("expected algorithm to be one of .+ got .+"),
			},
			{
				Config: testKeycloakRealmKeystoreRsaEncGenerated_basicWithAttrValidation(algorithm, "algorithm", algorithm),
				Check:  testAccCheckRealmKeystoreRsaEncGeneratedExists("keycloak_realm_keystore_rsa_enc_generated.realm_rsa"),
			},
		},
	})
}

func TestAccKeycloakRealmKeystoreRsaEncGenerated_updateRsaKeystoreGenerated(t *testing.T) {
	t.Parallel()

	enabled := randomBool()
	active := randomBool()

	groupKeystoreOne := &keycloak.RealmKeystoreRsaEncGenerated{
		Name:      acctest.RandString(10),
		RealmId:   testAccRealmUserFederation.Realm,
		Enabled:   enabled,
		Active:    active,
		Priority:  acctest.RandIntRange(0, 100),
		KeySize:   1024,
		Algorithm: randomStringInSlice(keycloakRealmKeystoreRsaEncGeneratedAlgorithm),
	}

	groupKeystoreTwo := &keycloak.RealmKeystoreRsaEncGenerated{
		Name:      acctest.RandString(10),
		RealmId:   testAccRealmUserFederation.Realm,
		Enabled:   enabled,
		Active:    active,
		Priority:  acctest.RandIntRange(0, 100),
		KeySize:   2048,
		Algorithm: randomStringInSlice(keycloakRealmKeystoreRsaEncGeneratedAlgorithm),
	}

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRealmKeystoreRsaEncGeneratedDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmKeystoreRsaEncGenerated_basicFromInterface(groupKeystoreOne),
				Check:  testAccCheckRealmKeystoreRsaEncGeneratedMatches("keycloak_realm_keystore_rsa_enc_generated.realm_rsa", groupKeystoreOne),
			},
			{
				Config: testKeycloakRealmKeystoreRsaEncGenerated_basicFromInterface(groupKeystoreTwo),
				Check:  testAccCheckRealmKeystoreRsaEncGeneratedMatches("keycloak_realm_keystore_rsa_enc_generated.realm_rsa", groupKeystoreTwo),
			},
		},
	})
}

func TestAccKeycloakRealmKeystoreRsaEncGenerated_parentIdDifferentFromRealmName(t *testing.T) {
	t.Parallel()

	realmName := acctest.RandomWithPrefix("tf-acc")
	internalId := acctest.RandomWithPrefix("tf-acc")
	rsaName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRealmKeystoreRsaEncGeneratedDestroy(),
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
				Config: testKeycloakRealmKeystoreRsaEncGenerated_parentId(realmName, rsaName, "100"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRealmKeystoreRsaEncGeneratedExists("keycloak_realm_keystore_rsa_enc_generated.realm_rsa"),
					testAccCheckRealmKeystoreRsaEncGeneratedParentId("keycloak_realm_keystore_rsa_enc_generated.realm_rsa", internalId),
				),
			},
			{
				// update: the omitted parentId must leave the stored parent untouched
				Config: testKeycloakRealmKeystoreRsaEncGenerated_parentId(realmName, rsaName, "200"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRealmKeystoreRsaEncGeneratedExists("keycloak_realm_keystore_rsa_enc_generated.realm_rsa"),
					testAccCheckRealmKeystoreRsaEncGeneratedParentId("keycloak_realm_keystore_rsa_enc_generated.realm_rsa", internalId),
				),
			},
		},
	})
}

func testAccCheckRealmKeystoreRsaEncGeneratedExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getKeycloakRealmKeystoreRsaEncGeneratedFromState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckRealmKeystoreRsaEncGeneratedFetch(resourceName string, keystore *keycloak.RealmKeystoreRsaEncGenerated) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedKeystore, err := getKeycloakRealmKeystoreRsaEncGeneratedFromState(s, resourceName)
		if err != nil {
			return err
		}

		keystore.Id = fetchedKeystore.Id
		keystore.RealmId = fetchedKeystore.RealmId

		return nil
	}
}

func testAccCheckRealmKeystoreRsaEncGeneratedParentId(resourceName, expectedParentId string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		keystore, err := getKeycloakRealmKeystoreRsaEncGeneratedFromState(s, resourceName)
		if err != nil {
			return err
		}

		if keystore.ParentId != expectedParentId {
			return fmt.Errorf("expected rsa generated keystore %s to have parent id %s, but got %s", keystore.Id, expectedParentId, keystore.ParentId)
		}

		return nil
	}
}

func testAccCheckRealmKeystoreRsaEncGeneratedMatches(resourceName string, expected *keycloak.RealmKeystoreRsaEncGenerated) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetched, err := getKeycloakRealmKeystoreRsaEncGeneratedFromState(s, resourceName)
		if err != nil {
			return err
		}

		if fetched.Name != expected.Name {
			return fmt.Errorf("expected name %s, got %s", expected.Name, fetched.Name)
		}
		if fetched.Priority != expected.Priority {
			return fmt.Errorf("expected priority %d, got %d", expected.Priority, fetched.Priority)
		}
		if fetched.KeySize != expected.KeySize {
			return fmt.Errorf("expected key size %d, got %d", expected.KeySize, fetched.KeySize)
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

func testAccCheckRealmKeystoreRsaEncGeneratedDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_realm_keystore_rsa_enc_generated" {
				continue
			}

			id := rs.Primary.ID
			realm := rs.Primary.Attributes["realm_id"]

			ldapGroupKeystore, err := keycloakClient.GetRealmKeystoreRsaEncGenerated(testCtx, realm, id)
			if err != nil && !keycloak.ErrorIs404(err) {
				return err
			}
			if ldapGroupKeystore != nil {
				return fmt.Errorf("rsa keystore with id %s still exists", id)
			}
		}

		return nil
	}
}

func getKeycloakRealmKeystoreRsaEncGeneratedFromState(s *terraform.State,
	resourceName string) (*keycloak.RealmKeystoreRsaEncGenerated,
	error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm_id"]

	realmKeystore, err := keycloakClient.GetRealmKeystoreRsaEncGenerated(testCtx, realm, id)
	if err != nil {
		return nil, fmt.Errorf("error getting rsa keystore with id %s: %s", id, err)
	}

	return realmKeystore, nil
}

func testKeycloakRealmKeystoreRsaEncGenerated_basic(rsaName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_realm_keystore_rsa_enc_generated" "realm_rsa" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id

    priority  = 100
    algorithm = "RSA-OAEP-256"
}
	`, testAccRealmUserFederation.Realm, rsaName)
}

func testKeycloakRealmKeystoreRsaEncGenerated_parentId(realmName, rsaName, priority string) string {
	return fmt.Sprintf(`
resource "keycloak_realm_keystore_rsa_enc_generated" "realm_rsa" {
	name      = "%s"
	realm_id  = "%s"

    priority  = %s
    algorithm = "RSA-OAEP-256"
}
	`, rsaName, realmName, priority)
}

func testKeycloakRealmKeystoreRsaEncGenerated_basicWithAttrValidation(rsaName, attr, val string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_realm_keystore_rsa_enc_generated" "realm_rsa" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id

	%s        = "%s"
}
	`, testAccRealmUserFederation.Realm, rsaName, attr, val)
}

func testKeycloakRealmKeystoreRsaEncGenerated_basicFromInterface(keystore *keycloak.RealmKeystoreRsaEncGenerated) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_realm_keystore_rsa_enc_generated" "realm_rsa" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id

    enabled   = %t
    active    = %t
    priority  = %s
    algorithm = "%s"
    key_size  = %s
}
	`, testAccRealmUserFederation.Realm, keystore.Name, keystore.Enabled, keystore.Active,
		strconv.Itoa(keystore.Priority), keystore.Algorithm, strconv.Itoa(keystore.KeySize))
}
