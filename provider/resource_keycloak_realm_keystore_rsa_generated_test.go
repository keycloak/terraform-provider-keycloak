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

func TestAccKeycloakRealmKeystoreRsaGenerated_basic(t *testing.T) {
	t.Parallel()

	rsaName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckRealmKeystoreRsaGeneratedDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmKeystoreRsaGenerated_basic(rsaName),
				Check:  testAccCheckRealmKeystoreRsaGeneratedExists("keycloak_realm_keystore_rsa_generated.realm_rsa"),
			},
			{
				ResourceName:      "keycloak_realm_keystore_rsa_generated.realm_rsa",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getRealmKeystoreGenericImportId("keycloak_realm_keystore_rsa_generated.realm_rsa"),
			},
		},
	})
}

func TestAccKeycloakRealmKeystoreRsaGenerated_createAfterManualDestroy(t *testing.T) {
	t.Parallel()

	var rsa = &keycloak.RealmKeystoreRsaGenerated{}

	fullNameKeystoreName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckRealmKeystoreRsaGeneratedDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmKeystoreRsaGenerated_basic(fullNameKeystoreName),
				Check:  testAccCheckRealmKeystoreRsaGeneratedFetch("keycloak_realm_keystore_rsa_generated.realm_rsa", rsa),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteRealmKeystoreRsaGenerated(testCtx, rsa.RealmId, rsa.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakRealmKeystoreRsaGenerated_basic(fullNameKeystoreName),
				Check:  testAccCheckRealmKeystoreRsaGeneratedFetch("keycloak_realm_keystore_rsa_generated.realm_rsa", rsa),
			},
		},
	})
}

func TestAccKeycloakRealmKeystoreRsaGenerated_keySizeValidation(t *testing.T) {
	t.Parallel()

	rsaName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckRealmKeystoreRsaGeneratedDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmKeystoreRsaGenerated_basicWithAttrValidation(rsaName, "key_size",
					strconv.Itoa(acctest.RandIntRange(0, 1000)*2+1)),
				ExpectError: regexp.MustCompile("expected key_size to be one of .+ got .+"),
			},
			{
				Config: testKeycloakRealmKeystoreRsaGenerated_basicWithAttrValidation(rsaName, "key_size", "2048"),
				Check:  testAccCheckRealmKeystoreRsaGeneratedExists("keycloak_realm_keystore_rsa_generated.realm_rsa"),
			},
		},
	})
}

func TestAccKeycloakRealmKeystoreRsaGenerated_algorithmValidation(t *testing.T) {
	t.Parallel()

	algorithm := randomStringInSlice(keycloakRealmKeystoreRsaAlgorithm)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckRealmKeystoreRsaGeneratedDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmKeystoreRsaGenerated_basicWithAttrValidation(algorithm, "algorithm",
					acctest.RandString(10)),
				ExpectError: regexp.MustCompile("expected algorithm to be one of .+ got .+"),
			},
			{
				Config: testKeycloakRealmKeystoreRsaGenerated_basicWithAttrValidation(algorithm, "algorithm", algorithm),
				Check:  testAccCheckRealmKeystoreRsaGeneratedExists("keycloak_realm_keystore_rsa_generated.realm_rsa"),
			},
		},
	})
}

func TestAccKeycloakRealmKeystoreRsaGenerated_updateRsaKeystoreGenerated(t *testing.T) {
	t.Parallel()

	enabled := randomBool()
	active := randomBool()

	groupKeystoreOne := &keycloak.RealmKeystoreRsaGenerated{
		Name:      acctest.RandString(10),
		RealmId:   testAccRealmUserFederation.Realm,
		Enabled:   enabled,
		Active:    active,
		Priority:  acctest.RandIntRange(0, 100),
		KeySize:   1024,
		Algorithm: randomStringInSlice(keycloakRealmKeystoreRsaAlgorithm),
	}

	groupKeystoreTwo := &keycloak.RealmKeystoreRsaGenerated{
		Name:      acctest.RandString(10),
		RealmId:   testAccRealmUserFederation.Realm,
		Enabled:   enabled,
		Active:    active,
		Priority:  acctest.RandIntRange(0, 100),
		KeySize:   2048,
		Algorithm: randomStringInSlice(keycloakRealmKeystoreRsaAlgorithm),
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckRealmKeystoreRsaGeneratedDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmKeystoreRsaGenerated_basicFromInterface(groupKeystoreOne),
				Check:  testAccCheckRealmKeystoreRsaGeneratedExists("keycloak_realm_keystore_rsa_generated.realm_rsa"),
			},
			{
				Config: testKeycloakRealmKeystoreRsaGenerated_basicFromInterface(groupKeystoreTwo),
				Check:  testAccCheckRealmKeystoreRsaGeneratedExists("keycloak_realm_keystore_rsa_generated.realm_rsa"),
			},
		},
	})
}

func testAccCheckRealmKeystoreRsaGeneratedExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getKeycloakRealmKeystoreRsaGeneratedFromState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckRealmKeystoreRsaGeneratedFetch(resourceName string, keystore *keycloak.RealmKeystoreRsaGenerated) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedKeystore, err := getKeycloakRealmKeystoreRsaGeneratedFromState(s, resourceName)
		if err != nil {
			return err
		}

		keystore.Id = fetchedKeystore.Id
		keystore.RealmId = fetchedKeystore.RealmId

		return nil
	}
}

func testAccCheckRealmKeystoreRsaGeneratedDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_realm_keystore_rsa_generated" {
				continue
			}

			id := rs.Primary.ID
			realm := rs.Primary.Attributes["realm_id"]

			ldapGroupKeystore, _ := keycloakClient.GetRealmKeystoreRsaGenerated(testCtx, realm, id)
			if ldapGroupKeystore != nil {
				return fmt.Errorf("rsa keystore with id %s still exists", id)
			}
		}

		return nil
	}
}

func getKeycloakRealmKeystoreRsaGeneratedFromState(s *terraform.State,
	resourceName string) (*keycloak.RealmKeystoreRsaGenerated,
	error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm_id"]

	realmKeystore, err := keycloakClient.GetRealmKeystoreRsaGenerated(testCtx, realm, id)
	if err != nil {
		return nil, fmt.Errorf("error getting rsa keystore with id %s: %s", id, err)
	}

	return realmKeystore, nil
}

func testKeycloakRealmKeystoreRsaGenerated_basic(rsaName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_realm_keystore_rsa_generated" "realm_rsa" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id

    priority  = 100
    algorithm = "RS384"
}
	`, testAccRealmUserFederation.Realm, rsaName)
}

func testKeycloakRealmKeystoreRsaGenerated_basicWithAttrValidation(rsaName, attr, val string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_realm_keystore_rsa_generated" "realm_rsa" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id

	%s        = "%s"
}
	`, testAccRealmUserFederation.Realm, rsaName, attr, val)
}

func testKeycloakRealmKeystoreRsaGenerated_basicFromInterface(keystore *keycloak.RealmKeystoreRsaGenerated) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_realm_keystore_rsa_generated" "realm_rsa" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id

    priority  = %s
    algorithm = "%s"
    key_size  = %s
}
	`, testAccRealmUserFederation.Realm, keystore.Name, strconv.Itoa(keystore.Priority), keystore.Algorithm,
		strconv.Itoa(keystore.KeySize))
}
