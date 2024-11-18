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

func TestAccKeycloakRealmKeystoreJava_basic(t *testing.T) {
	t.Parallel()

	skipIfEnvSet(t, "CI") // temporary while I figure out how to put java keystore file to keycloak container in CI

	javaKeystoreName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckRealmKeystoreJavaDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmKeystoreJava_basic(javaKeystoreName),
				Check:  testAccCheckRealmKeystoreJavaExists("keycloak_realm_keystore_java_keystore.realm_java_keystore"),
			},
			{
				ResourceName:      "keycloak_realm_keystore_java_keystore.realm_java_keystore",
				ImportState:       true,
				ImportStateVerify: false, //OOTB verify doesnt work here since secrets are not returned when reading
				ImportStateIdFunc: getRealmKeystoreGenericImportId("keycloak_realm_keystore_java_keystore.realm_java_keystore"),
			},
		},
	})
}

func TestAccKeycloakRealmKeystoreJava_createAfterManualDestroy(t *testing.T) {
	t.Parallel()

	skipIfEnvSet(t, "CI") // temporary while I figure out how to put java keystore file to keycloak container in CI

	var javaKeystore = &keycloak.RealmKeystoreJavaKeystore{}

	fullNameKeystoreName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckRealmKeystoreJavaDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmKeystoreJava_basic(fullNameKeystoreName),
				Check:  testAccCheckRealmKeystoreJavaFetch("keycloak_realm_keystore_java_keystore.realm_java_keystore", javaKeystore),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteRealmKeystoreJavaKeystore(testCtx, javaKeystore.RealmId, javaKeystore.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakRealmKeystoreJava_basic(fullNameKeystoreName),
				Check:  testAccCheckRealmKeystoreJavaFetch("keycloak_realm_keystore_java_keystore.realm_java_keystore", javaKeystore),
			},
		},
	})
}

func TestAccKeycloakRealmKeystoreJava_algorithmValidation(t *testing.T) {
	t.Parallel()

	skipIfEnvSet(t, "CI") // temporary while I figure out how to put java keystore file to keycloak container in CI

	algorithm := randomStringInSlice(keycloakRealmKeystoreRsaAlgorithm)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckRealmKeystoreJavaDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmKeystoreJava_basicWithAttrValidation(algorithm, "algorithm",
					acctest.RandString(10)),
				ExpectError: regexp.MustCompile("expected algorithm to be one of .+ got .+"),
			},
			{
				Config: testKeycloakRealmKeystoreJava_basicWithAttrValidation(algorithm, "algorithm", algorithm),
				Check:  testAccCheckRealmKeystoreJavaExists("keycloak_realm_keystore_java_keystore.realm_java_keystore"),
			},
		},
	})
}

func TestAccKeycloakRealmKeystoreJava_updateRsaKeystoreGenerated(t *testing.T) {
	t.Parallel()

	skipIfEnvSet(t, "CI") // temporary while I figure out how to put java keystore file to keycloak container in CI

	enabled := randomBool()
	active := randomBool()

	groupKeystoreOne := &keycloak.RealmKeystoreJavaKeystore{
		Name:      acctest.RandString(10),
		RealmId:   testAccRealmUserFederation.Realm,
		Enabled:   enabled,
		Active:    active,
		Priority:  acctest.RandIntRange(0, 100),
		Algorithm: randomStringInSlice(keycloakRealmKeystoreJavaKeystoreAlgorithm),
	}

	groupKeystoreTwo := &keycloak.RealmKeystoreJavaKeystore{
		Name:      acctest.RandString(10),
		RealmId:   testAccRealmUserFederation.Realm,
		Enabled:   enabled,
		Active:    active,
		Priority:  acctest.RandIntRange(0, 100),
		Algorithm: randomStringInSlice(keycloakRealmKeystoreJavaKeystoreAlgorithm),
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckRealmKeystoreJavaDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmKeystoreJava_basicFromInterface(groupKeystoreOne),
				Check:  testAccCheckRealmKeystoreJavaExists("keycloak_realm_keystore_java_keystore.realm_java_keystore"),
			},
			{
				Config: testKeycloakRealmKeystoreJava_basicFromInterface(groupKeystoreTwo),
				Check:  testAccCheckRealmKeystoreJavaExists("keycloak_realm_keystore_java_keystore.realm_java_keystore"),
			},
		},
	})
}

func testAccCheckRealmKeystoreJavaExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getKeycloakRealmKeystoreJavaFromState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckRealmKeystoreJavaFetch(resourceName string, keystore *keycloak.RealmKeystoreJavaKeystore) resource.
	TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedKeystore, err := getKeycloakRealmKeystoreJavaFromState(s, resourceName)
		if err != nil {
			return err
		}

		keystore.Id = fetchedKeystore.Id
		keystore.RealmId = fetchedKeystore.RealmId

		return nil
	}
}

func testAccCheckRealmKeystoreJavaDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_realm_keystore_java_keystore" {
				continue
			}

			id := rs.Primary.ID
			realm := rs.Primary.Attributes["realm_id"]

			ldapGroupKeystore, _ := keycloakClient.GetRealmKeystoreJavaKeystore(testCtx, realm, id)
			if ldapGroupKeystore != nil {
				return fmt.Errorf("rsa keystore with id %s still exists", id)
			}
		}

		return nil
	}
}

func getKeycloakRealmKeystoreJavaFromState(s *terraform.State,
	resourceName string) (*keycloak.RealmKeystoreJavaKeystore,
	error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm_id"]

	realmKeystore, err := keycloakClient.GetRealmKeystoreJavaKeystore(testCtx, realm, id)
	if err != nil {
		return nil, fmt.Errorf("error getting rsa keystore with id %s: %s", id, err)
	}

	return realmKeystore, nil
}

func testKeycloakRealmKeystoreJava_basic(javaKeystoreName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_realm_keystore_java_keystore" "realm_java_keystore" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id

    keystore          = "/opt/keycloak/misc/keystore.jks"
    keystore_password = "12345678"
    key_alias    = "test"
    key_password = "12345678"

    priority  = 100
    algorithm = "RS256"
}
	`, testAccRealmUserFederation.Realm, javaKeystoreName)
}

func testKeycloakRealmKeystoreJava_basicWithAttrValidation(javaKeystoreName, attr, val string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_realm_keystore_java_keystore" "realm_java_keystore" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id

    keystore          = "/opt/keycloak/misc/keystore.jks"
    keystore_password = "12345678"
    key_alias    = "test"
    key_password = "12345678"

	%s        = "%s"
}
	`, testAccRealmUserFederation.Realm, javaKeystoreName, attr, val)
}

func testKeycloakRealmKeystoreJava_basicFromInterface(keystore *keycloak.RealmKeystoreJavaKeystore) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_realm_keystore_java_keystore" "realm_java_keystore" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id

    keystore          = "/opt/keycloak/misc/keystore.jks"
    keystore_password = "12345678"
    key_alias    = "test"
    key_password = "12345678"

    priority  = %s
    algorithm = "%s"
}
	`, testAccRealmUserFederation.Realm, keystore.Name, strconv.Itoa(keystore.Priority), keystore.Algorithm)
}
