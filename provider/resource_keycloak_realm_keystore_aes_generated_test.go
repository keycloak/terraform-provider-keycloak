package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"regexp"
	"strconv"
	"testing"
)

func TestAccKeycloakRealmKeystoreAesGenerated_basic(t *testing.T) {
	t.Parallel()

	aesName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckRealmKeystoreAesGeneratedDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmKeystoreAesGenerated_basic(aesName),
				Check:  testAccCheckRealmKeystoreAesGeneratedExists("keycloak_realm_keystore_aes_generated.realm_aes"),
			},
			{
				ResourceName:      "keycloak_realm_keystore_aes_generated.realm_aes",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getRealmKeystoreGenericImportId("keycloak_realm_keystore_aes_generated.realm_aes"),
			},
		},
	})
}

func TestAccKeycloakRealmKeystoreAesGenerated_createAfterManualDestroy(t *testing.T) {
	t.Parallel()

	var aes = &keycloak.RealmKeystoreAesGenerated{}

	fullNameKeystoreName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckRealmKeystoreAesGeneratedDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmKeystoreAesGenerated_basic(fullNameKeystoreName),
				Check:  testAccCheckRealmKeystoreAesGeneratedFetch("keycloak_realm_keystore_aes_generated.realm_aes", aes),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteRealmKeystoreAesGenerated(testCtx, aes.RealmId, aes.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakRealmKeystoreAesGenerated_basic(fullNameKeystoreName),
				Check:  testAccCheckRealmKeystoreAesGeneratedFetch("keycloak_realm_keystore_aes_generated.realm_aes", aes),
			},
		},
	})
}

func TestAccKeycloakRealmKeystoreAesGenerated_secretSizeValidation(t *testing.T) {
	t.Parallel()

	aesName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckRealmKeystoreAesGeneratedDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmKeystoreAesGenerated_basicWithAttrValidation(aesName, "secret_size",
					strconv.Itoa(acctest.RandIntRange(0, 1000)*2+1)),
				ExpectError: regexp.MustCompile("expected secret_size to be one of .+ got .+"),
			},
			{
				Config: testKeycloakRealmKeystoreAesGenerated_basicWithAttrValidation(aesName, "secret_size", "16"),
				Check:  testAccCheckRealmKeystoreAesGeneratedExists("keycloak_realm_keystore_aes_generated.realm_aes"),
			},
		},
	})
}

func TestAccKeycloakRealmKeystoreAesGenerated_updateRealmKeystoreAesGenerated(t *testing.T) {
	t.Parallel()

	enabled := randomBool()
	active := randomBool()

	groupKeystoreOne := &keycloak.RealmKeystoreAesGenerated{
		Name:       acctest.RandString(10),
		RealmId:    testAccRealmUserFederation.Realm,
		Enabled:    enabled,
		Active:     active,
		Priority:   acctest.RandIntRange(0, 100),
		SecretSize: 16,
	}

	groupKeystoreTwo := &keycloak.RealmKeystoreAesGenerated{
		Name:       acctest.RandString(10),
		RealmId:    testAccRealmUserFederation.Realm,
		Enabled:    enabled,
		Active:     active,
		Priority:   acctest.RandIntRange(0, 100),
		SecretSize: 32,
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckRealmKeystoreAesGeneratedDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmKeystoreAesGenerated_basicFromInterface(groupKeystoreOne),
				Check:  testAccCheckRealmKeystoreAesGeneratedExists("keycloak_realm_keystore_aes_generated.realm_aes"),
			},
			{
				Config: testKeycloakRealmKeystoreAesGenerated_basicFromInterface(groupKeystoreTwo),
				Check:  testAccCheckRealmKeystoreAesGeneratedExists("keycloak_realm_keystore_aes_generated.realm_aes"),
			},
		},
	})
}

func testAccCheckRealmKeystoreAesGeneratedExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getKeycloakRealmKeystoreAesGeneratedFromState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckRealmKeystoreAesGeneratedFetch(resourceName string, keystore *keycloak.RealmKeystoreAesGenerated) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedKeystore, err := getKeycloakRealmKeystoreAesGeneratedFromState(s, resourceName)
		if err != nil {
			return err
		}

		keystore.Id = fetchedKeystore.Id
		keystore.RealmId = fetchedKeystore.RealmId

		return nil
	}
}

func testAccCheckRealmKeystoreAesGeneratedDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_realm_keystore_aes_generated" {
				continue
			}

			id := rs.Primary.ID
			realm := rs.Primary.Attributes["realm_id"]

			ldapGroupKeystore, _ := keycloakClient.GetRealmKeystoreAesGenerated(testCtx, realm, id)
			if ldapGroupKeystore != nil {
				return fmt.Errorf("aes keystore with id %s still exists", id)
			}
		}

		return nil
	}
}

func getKeycloakRealmKeystoreAesGeneratedFromState(s *terraform.State,
	resourceName string) (*keycloak.RealmKeystoreAesGenerated,
	error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm_id"]

	realmKeystore, err := keycloakClient.GetRealmKeystoreAesGenerated(testCtx, realm, id)
	if err != nil {
		return nil, fmt.Errorf("error getting aes keystore with id %s: %s", id, err)
	}

	return realmKeystore, nil
}

func testKeycloakRealmKeystoreAesGenerated_basic(aesName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_realm_keystore_aes_generated" "realm_aes" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id

    priority           = 100
}
	`, testAccRealmUserFederation.Realm, aesName)
}

func testKeycloakRealmKeystoreAesGenerated_basicWithAttrValidation(aesName, attr, val string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_realm_keystore_aes_generated" "realm_aes" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id

	%s        = "%s"
}
	`, testAccRealmUserFederation.Realm, aesName, attr, val)
}

func testKeycloakRealmKeystoreAesGenerated_basicFromInterface(keystore *keycloak.RealmKeystoreAesGenerated) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_realm_keystore_aes_generated" "realm_aes" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id

    priority    = "%s"
    secret_size = "%s"
}
	`, testAccRealmUserFederation.Realm, keystore.Name, strconv.Itoa(keystore.Priority), strconv.Itoa(keystore.SecretSize))
}
