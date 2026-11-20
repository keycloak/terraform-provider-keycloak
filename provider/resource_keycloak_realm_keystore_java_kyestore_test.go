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

func TestAccKeycloakRealmKeystoreJava_basic(t *testing.T) {
	t.Parallel()

	javaKeystoreName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRealmKeystoreJavaDestroy(),
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

	var javaKeystore = &keycloak.RealmKeystoreJavaKeystore{}

	fullNameKeystoreName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRealmKeystoreJavaDestroy(),
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

	keystoreName := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRealmKeystoreJavaDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakRealmKeystoreJava_basicWithAttrValidation(keystoreName, "sig-es256", "algorithm", acctest.RandString(10)),
				ExpectError: regexp.MustCompile("expected algorithm to be one of .+ got .+"),
			},
			{
				Config: testKeycloakRealmKeystoreJava_basicWithAttrValidation(keystoreName, "sig-es256", "algorithm", "ES256"),
				Check:  testAccCheckRealmKeystoreJavaExists("keycloak_realm_keystore_java_keystore.realm_java_keystore"),
			},
		},
	})
}

func TestAccKeycloakRealmKeystoreJava_updateRsaKeystoreGenerated(t *testing.T) {
	t.Parallel()

	enabled := randomBool()
	active := randomBool()

	groupKeystoreOne := &keycloak.RealmKeystoreJavaKeystore{
		Name:     acctest.RandString(10),
		RealmId:  testAccRealmKeystore.Realm,
		Enabled:  enabled,
		Active:   active,
		Priority: acctest.RandIntRange(0, 100),
	}

	groupKeystoreTwo := &keycloak.RealmKeystoreJavaKeystore{
		Name:      acctest.RandString(10),
		RealmId:   testAccRealmKeystore.Realm,
		Enabled:   enabled,
		Active:    active,
		Priority:  acctest.RandIntRange(0, 100),
		Algorithm: "ES256",
	}

	groupKeystoreThree := &keycloak.RealmKeystoreJavaKeystore{
		Name:      acctest.RandString(10),
		RealmId:   testAccRealmKeystore.Realm,
		Enabled:   enabled,
		Active:    active,
		Priority:  acctest.RandIntRange(0, 100),
		Algorithm: "AES",
		KeyUse:    "enc",
	}

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRealmKeystoreJavaDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmKeystoreJava_basicFromInterface(groupKeystoreOne, "sig-rs256"),
				Check:  testAccCheckRealmKeystoreJavaExists("keycloak_realm_keystore_java_keystore.realm_java_keystore"),
			},
			{
				Config: testKeycloakRealmKeystoreJava_basicFromInterface(groupKeystoreTwo, "sig-es256"),
				Check:  testAccCheckRealmKeystoreJavaExists("keycloak_realm_keystore_java_keystore.realm_java_keystore"),
			},
			{
				Config: testKeycloakRealmKeystoreJava_basicFromInterface(groupKeystoreThree, "enc-aes256"),
				Check:  testAccCheckRealmKeystoreJavaExists("keycloak_realm_keystore_java_keystore.realm_java_keystore"),
			},
		},
	})
}

func TestAccKeycloakRealmKeystoreJava_parentIdDifferentFromRealmName(t *testing.T) {
	t.Parallel()

	// the keystore file must live under /opt/keycloak/data/<realm name>, so a random realm can't be used here
	realmName := testAccRealmKeystore.Realm
	internalId := testAccRealmKeystore.Id
	javaKeystoreName := acctest.RandomWithPrefix("tf-acc")

	if realmName == internalId {
		t.Fatalf("expected the shared keystore realm %s to have an internal id that differs from its name", realmName)
	}

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRealmKeystoreJavaDestroy(),
		Steps: []resource.TestStep{
			{
				// create: keycloak must default the omitted parentId to the realm's internal id
				Config: testKeycloakRealmKeystoreJava_parentId(realmName, javaKeystoreName, "100"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRealmKeystoreJavaExists("keycloak_realm_keystore_java_keystore.realm_java_keystore"),
					testAccCheckRealmKeystoreJavaParentId("keycloak_realm_keystore_java_keystore.realm_java_keystore", internalId),
					resource.TestCheckResourceAttr("keycloak_realm_keystore_java_keystore.realm_java_keystore", "parent_id", internalId),
				),
			},
			{
				// update: the omitted parentId must leave the stored parent untouched
				Config: testKeycloakRealmKeystoreJava_parentId(realmName, javaKeystoreName, "200"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRealmKeystoreJavaExists("keycloak_realm_keystore_java_keystore.realm_java_keystore"),
					testAccCheckRealmKeystoreJavaParentId("keycloak_realm_keystore_java_keystore.realm_java_keystore", internalId),
					resource.TestCheckResourceAttr("keycloak_realm_keystore_java_keystore.realm_java_keystore", "parent_id", internalId),
				),
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

func testAccCheckRealmKeystoreJavaFetch(resourceName string, keystore *keycloak.RealmKeystoreJavaKeystore) resource.TestCheckFunc {
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

func testAccCheckRealmKeystoreJavaParentId(resourceName, expectedParentId string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		keystore, err := getKeycloakRealmKeystoreJavaFromState(s, resourceName)
		if err != nil {
			return err
		}

		if keystore.ParentId != expectedParentId {
			return fmt.Errorf("expected java keystore %s to have parent id %s, but got %s", keystore.Id, expectedParentId, keystore.ParentId)
		}

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

    keystore          = "/opt/keycloak/data/tf-acc-keystore/keystore.jks"
    keystore_password = "12345678"
    key_alias    = "sig-rs256"
    key_password = "12345678"

    priority  = 100
    algorithm = "RS256"
}
	`, testAccRealmKeystore.Realm, javaKeystoreName)
}

func testKeycloakRealmKeystoreJava_parentId(realmName, javaKeystoreName, priority string) string {
	return fmt.Sprintf(`
resource "keycloak_realm_keystore_java_keystore" "realm_java_keystore" {
	name      = "%s"
	realm_id  = "%s"

    keystore          = "/opt/keycloak/data/tf-acc-keystore/keystore.jks"
    keystore_password = "12345678"
    key_alias    = "sig-rs256"
    key_password = "12345678"

    priority  = %s
    algorithm = "RS256"
}
	`, javaKeystoreName, realmName, priority)
}

func testKeycloakRealmKeystoreJava_basicWithAttrValidation(javaKeystoreName, keyAlias string, attr, val string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_realm_keystore_java_keystore" "realm_java_keystore" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id

    keystore          = "/opt/keycloak/data/tf-acc-keystore/keystore.jks"
    keystore_password = "12345678"
    key_alias    = "%s"
    key_password = "12345678"

	%s        = "%s"
}
	`, testAccRealmKeystore.Realm, javaKeystoreName, keyAlias, attr, val)
}

func testKeycloakRealmKeystoreJava_basicFromInterface(keystore *keycloak.RealmKeystoreJavaKeystore, keyAlias string) string {
	algorithmAttr := ""
	if keystore.Algorithm != "" {
		algorithmAttr = fmt.Sprintf(`algorithm = "%s"`, keystore.Algorithm)
	}

	keyUseAttr := ""
	if keystore.KeyUse != "" {
		keyUseAttr = fmt.Sprintf(`key_use = "%s"`, keystore.KeyUse)
	}

	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_realm_keystore_java_keystore" "realm_java_keystore" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id

    keystore          = "/opt/keycloak/data/tf-acc-keystore/keystore.jks"
    keystore_password = "12345678"
    key_alias    = "%s"
    key_password = "12345678"

    priority  = %s
	%s
    %s
}
	`, testAccRealmKeystore.Realm, keystore.Name, keyAlias, strconv.Itoa(keystore.Priority), algorithmAttr, keyUseAttr)
}
