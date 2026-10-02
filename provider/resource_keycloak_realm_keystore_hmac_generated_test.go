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

func TestAccKeycloakRealmKeystoreHmacGenerated_basic(t *testing.T) {
	t.Parallel()

	hmacName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRealmKeystoreHmacGeneratedDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmKeystoreHmacGenerated_basic(hmacName),
				Check:  testAccCheckRealmKeystoreHmacGeneratedExists("keycloak_realm_keystore_hmac_generated.realm_hmac"),
			},
			{
				ResourceName:      "keycloak_realm_keystore_hmac_generated.realm_hmac",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getRealmKeystoreGenericImportId("keycloak_realm_keystore_hmac_generated.realm_hmac"),
			},
		},
	})
}

func TestAccKeycloakRealmKeystoreHmacGenerated_createAfterManualDestroy(t *testing.T) {
	t.Parallel()

	var hmac = &keycloak.RealmKeystoreHmacGenerated{}

	fullNameKeystoreName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRealmKeystoreHmacGeneratedDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmKeystoreHmacGenerated_basic(fullNameKeystoreName),
				Check:  testAccCheckRealmKeystoreHmacGeneratedFetch("keycloak_realm_keystore_hmac_generated.realm_hmac", hmac),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteRealmKeystoreHmacGenerated(testCtx, hmac.RealmId, hmac.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakRealmKeystoreHmacGenerated_basic(fullNameKeystoreName),
				Check:  testAccCheckRealmKeystoreHmacGeneratedFetch("keycloak_realm_keystore_hmac_generated.realm_hmac", hmac),
			},
		},
	})
}

func TestAccKeycloakRealmKeystoreHmacGenerated_algorithmValidation(t *testing.T) {
	t.Parallel()

	hmacName := acctest.RandomWithPrefix("tf-acc")
	algorithm := randomStringInSlice(keycloakRealmKeystoreHmacGeneratedAlgorithm)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRealmKeystoreHmacGeneratedDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmKeystoreHmacGenerated_basicWithAttrValidation(hmacName, "algorithm",
					acctest.RandString(10)),
				ExpectError: regexp.MustCompile("expected algorithm to be one of .+ got .+"),
			},
			{
				Config: testKeycloakRealmKeystoreHmacGenerated_basicWithAttrValidation(hmacName, "algorithm",
					algorithm),
				Check: testAccCheckRealmKeystoreHmacGeneratedExists("keycloak_realm_keystore_hmac_generated.realm_hmac"),
			},
		},
	})
}

func TestAccKeycloakRealmKeystoreHmacGenerated_updateRealmKeystoreHmacGenerated(t *testing.T) {
	t.Parallel()

	enabled := randomBool()
	active := randomBool()

	groupKeystoreOne := &keycloak.RealmKeystoreHmacGenerated{
		Name:       acctest.RandString(10),
		RealmId:    testAccRealmUserFederation.Realm,
		Enabled:    enabled,
		Active:     active,
		Priority:   acctest.RandIntRange(0, 100),
		SecretSize: 64,
		Algorithm:  randomStringInSlice(keycloakRealmKeystoreHmacGeneratedAlgorithm),
	}

	groupKeystoreTwo := &keycloak.RealmKeystoreHmacGenerated{
		Name:       acctest.RandString(10),
		RealmId:    testAccRealmUserFederation.Realm,
		Enabled:    enabled,
		Active:     active,
		Priority:   acctest.RandIntRange(0, 100),
		SecretSize: 32,
		Algorithm:  randomStringInSlice(keycloakRealmKeystoreHmacGeneratedAlgorithm),
	}

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRealmKeystoreHmacGeneratedDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmKeystoreHmacGenerated_basicFromInterface(groupKeystoreOne),
				Check:  testAccCheckRealmKeystoreHmacGeneratedExists("keycloak_realm_keystore_hmac_generated.realm_hmac"),
			},
			{
				Config: testKeycloakRealmKeystoreHmacGenerated_basicFromInterface(groupKeystoreTwo),
				Check:  testAccCheckRealmKeystoreHmacGeneratedExists("keycloak_realm_keystore_hmac_generated.realm_hmac"),
			},
		},
	})
}

func TestAccKeycloakRealmKeystoreHmacGenerated_parentIdDifferentFromRealmName(t *testing.T) {
	t.Parallel()

	realmName := acctest.RandomWithPrefix("tf-acc")
	internalId := acctest.RandomWithPrefix("tf-acc")
	hmacName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRealmKeystoreHmacGeneratedDestroy(),
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
				Config: testKeycloakRealmKeystoreHmacGenerated_parentId(realmName, hmacName, "100"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRealmKeystoreHmacGeneratedExists("keycloak_realm_keystore_hmac_generated.realm_hmac"),
					testAccCheckRealmKeystoreHmacGeneratedParentId("keycloak_realm_keystore_hmac_generated.realm_hmac", internalId),
				),
			},
			{
				// update: the omitted parentId must leave the stored parent untouched
				Config: testKeycloakRealmKeystoreHmacGenerated_parentId(realmName, hmacName, "200"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRealmKeystoreHmacGeneratedExists("keycloak_realm_keystore_hmac_generated.realm_hmac"),
					testAccCheckRealmKeystoreHmacGeneratedParentId("keycloak_realm_keystore_hmac_generated.realm_hmac", internalId),
				),
			},
		},
	})
}

func testAccCheckRealmKeystoreHmacGeneratedExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getKeycloakRealmKeystoreHmacGeneratedFromState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckRealmKeystoreHmacGeneratedFetch(resourceName string, keystore *keycloak.RealmKeystoreHmacGenerated) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedKeystore, err := getKeycloakRealmKeystoreHmacGeneratedFromState(s, resourceName)
		if err != nil {
			return err
		}

		keystore.Id = fetchedKeystore.Id
		keystore.RealmId = fetchedKeystore.RealmId

		return nil
	}
}

func testAccCheckRealmKeystoreHmacGeneratedParentId(resourceName, expectedParentId string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		keystore, err := getKeycloakRealmKeystoreHmacGeneratedFromState(s, resourceName)
		if err != nil {
			return err
		}

		if keystore.ParentId != expectedParentId {
			return fmt.Errorf("expected hmac keystore %s to have parent id %s, but got %s", keystore.Id, expectedParentId, keystore.ParentId)
		}

		return nil
	}
}

func testAccCheckRealmKeystoreHmacGeneratedDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_realm_keystore_hmac_generated" {
				continue
			}

			id := rs.Primary.ID
			realm := rs.Primary.Attributes["realm_id"]

			ldapGroupKeystore, _ := keycloakClient.GetRealmKeystoreHmacGenerated(testCtx, realm, id)
			if ldapGroupKeystore != nil {
				return fmt.Errorf("hmac keystore with id %s still exists", id)
			}
		}

		return nil
	}
}

func getKeycloakRealmKeystoreHmacGeneratedFromState(s *terraform.State,
	resourceName string) (*keycloak.RealmKeystoreHmacGenerated,
	error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm_id"]

	realmKeystore, err := keycloakClient.GetRealmKeystoreHmacGenerated(testCtx, realm, id)
	if err != nil {
		return nil, fmt.Errorf("error getting hmac keystore with id %s: %s", id, err)
	}

	return realmKeystore, nil
}

func testKeycloakRealmKeystoreHmacGenerated_basic(hmacName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_realm_keystore_hmac_generated" "realm_hmac" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id

    priority    = 100
    secret_size = 32
    algorithm   = "HS384"
}
	`, testAccRealmUserFederation.Realm, hmacName)
}

func testKeycloakRealmKeystoreHmacGenerated_parentId(realmName, hmacName, priority string) string {
	return fmt.Sprintf(`
resource "keycloak_realm_keystore_hmac_generated" "realm_hmac" {
	name      = "%s"
	realm_id  = "%s"

    priority    = %s
    secret_size = 32
    algorithm   = "HS384"
}
	`, hmacName, realmName, priority)
}

func testKeycloakRealmKeystoreHmacGenerated_basicWithAttrValidation(hmacName, attr, val string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_realm_keystore_hmac_generated" "realm_hmac" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id

	%s        = "%s"
}
	`, testAccRealmUserFederation.Realm, hmacName, attr, val)
}

func testKeycloakRealmKeystoreHmacGenerated_basicFromInterface(keystore *keycloak.RealmKeystoreHmacGenerated) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_realm_keystore_hmac_generated" "realm_hmac" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id

    priority    = "%s"
    secret_size = "%s"
    algorithm   = "%s"
}
	`, testAccRealmUserFederation.Realm, keystore.Name, strconv.Itoa(keystore.Priority), strconv.Itoa(keystore.SecretSize), keystore.Algorithm)
}
