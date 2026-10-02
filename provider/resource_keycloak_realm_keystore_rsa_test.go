package provider

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakRealmKeystoreRsa_basic(t *testing.T) {
	t.Parallel()

	rsaName := acctest.RandomWithPrefix("tf-acc")
	privateKey, certificate := generateKeyAndCert(2048)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRealmKeystoreRsaDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmKeystoreRsa_basic(rsaName, privateKey, certificate),
				Check:  testAccCheckRealmKeystoreRsaExists("keycloak_realm_keystore_rsa.realm_rsa"),
			},
			// we can't verify this import test because there's no way to get the private key / cert from the Keycloak API
			{
				ResourceName:      "keycloak_realm_keystore_rsa.realm_rsa",
				ImportState:       true,
				ImportStateIdFunc: getRealmKeystoreGenericImportId("keycloak_realm_keystore_rsa.realm_rsa"),
			},
		},
	})
}

func TestAccKeycloakRealmKeystoreRsa_createAfterManualDestroy(t *testing.T) {
	t.Parallel()

	var keystoreRsa = &keycloak.RealmKeystoreRsa{}

	fullNameKeystoreName := acctest.RandomWithPrefix("tf-acc")
	privateKey, certificate := generateKeyAndCert(2048)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRealmKeystoreRsaDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmKeystoreRsa_basic(fullNameKeystoreName, privateKey, certificate),
				Check:  testAccCheckRealmKeystoreRsaFetch("keycloak_realm_keystore_rsa.realm_rsa", keystoreRsa),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteRealmKeystoreRsa(testCtx, keystoreRsa.RealmId, keystoreRsa.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakRealmKeystoreRsa_basic(fullNameKeystoreName, privateKey, certificate),
				Check:  testAccCheckRealmKeystoreRsaFetch("keycloak_realm_keystore_rsa.realm_rsa", keystoreRsa),
			},
		},
	})
}

func TestAccKeycloakRealmKeystoreRsa_algorithmValidation(t *testing.T) {
	t.Parallel()

	rsaAlgorithm := randomStringInSlice(keycloakRealmKeystoreRsaAlgorithm)
	rsaEncAlgorithm := randomStringInSlice(keycloakRealmKeystoreRsaEncAlgorithm)
	privateKey, certificate := generateKeyAndCert(2048)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRealmKeystoreRsaDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmKeystoreRsa_basicWithAttrValidation("rsa", rsaAlgorithm, "algorithm",
					acctest.RandString(10), privateKey, certificate),
				ExpectError: regexp.MustCompile("expected algorithm to be one of .+ got .+"),
			},
			{
				Config: testKeycloakRealmKeystoreRsa_basicWithAttrValidation("rsa", rsaAlgorithm, "algorithm", rsaAlgorithm,
					privateKey, certificate),
				Check: testAccCheckRealmKeystoreRsaExists("keycloak_realm_keystore_rsa.realm_rsa"),
			},
			{
				Config: testKeycloakRealmKeystoreRsa_basicWithAttrValidation("rsa-enc", rsaEncAlgorithm, "algorithm",
					acctest.RandString(10), privateKey, certificate),
				ExpectError: regexp.MustCompile("expected algorithm to be one of .+ got .+"),
			},
			{
				Config: testKeycloakRealmKeystoreRsa_basicWithAttrValidation("rsa-enc", rsaEncAlgorithm, "algorithm", rsaEncAlgorithm,
					privateKey, certificate),
				Check: testAccCheckRealmKeystoreRsaExists("keycloak_realm_keystore_rsa.realm_rsa"),
			},
		},
	})
}

func TestAccKeycloakRealmKeystoreRsa_extraConfigKid(t *testing.T) {
	t.Parallel()

	rsaName := acctest.RandomWithPrefix("tf-acc")
	kid := acctest.RandomWithPrefix("tf-acc")
	privateKey, certificate := generateKeyAndCert(2048)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRealmKeystoreRsaDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmKeystoreRsa_withKidExtraConfig(rsaName, privateKey, certificate, kid),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRealmKeystoreRsaExists("keycloak_realm_keystore_rsa.realm_rsa"),
					testAccCheckRealmKeystoreRsaKidInRealmKeys("keycloak_realm_keystore_rsa.realm_rsa", kid),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccKeycloakRealmKeystoreRsa_parentIdDifferentFromRealmName(t *testing.T) {
	t.Parallel()

	realmName := acctest.RandomWithPrefix("tf-acc")
	internalId := acctest.RandomWithPrefix("tf-acc")
	rsaName := acctest.RandomWithPrefix("tf-acc")
	privateKey, certificate := generateKeyAndCert(2048)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRealmKeystoreRsaDestroy(),
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
				Config: testKeycloakRealmKeystoreRsa_parentId(realmName, rsaName, privateKey, certificate, "100"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRealmKeystoreRsaExists("keycloak_realm_keystore_rsa.realm_rsa"),
					testAccCheckRealmKeystoreRsaParentId("keycloak_realm_keystore_rsa.realm_rsa", internalId),
				),
			},
			{
				// update: the omitted parentId must leave the stored parent untouched
				Config: testKeycloakRealmKeystoreRsa_parentId(realmName, rsaName, privateKey, certificate, "200"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRealmKeystoreRsaExists("keycloak_realm_keystore_rsa.realm_rsa"),
					testAccCheckRealmKeystoreRsaParentId("keycloak_realm_keystore_rsa.realm_rsa", internalId),
				),
			},
		},
	})
}

func TestAccKeycloakRealmKeystoreRsa_writeOnly(t *testing.T) {
	t.Parallel()

	rsaName := acctest.RandomWithPrefix("tf-acc")
	privateKey, certificate := generateKeyAndCert(2048)
	newPrivateKey, newCertificate := generateKeyAndCert(2048)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRealmKeystoreRsaDestroy(),
		Steps: []resource.TestStep{
			{
				// CREATE using write-only attributes
				Config: testKeycloakRealmKeystoreRsa_writeOnly(rsaName, 100, privateKey, "v1", certificate, "v1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRealmKeystoreRsaExists("keycloak_realm_keystore_rsa.realm_rsa"),
					resource.TestCheckNoResourceAttr("keycloak_realm_keystore_rsa.realm_rsa", "private_key"),
					resource.TestCheckNoResourceAttr("keycloak_realm_keystore_rsa.realm_rsa", "private_key_wo"),
					resource.TestCheckNoResourceAttr("keycloak_realm_keystore_rsa.realm_rsa", "certificate"),
					resource.TestCheckNoResourceAttr("keycloak_realm_keystore_rsa.realm_rsa", "certificate_wo"),
					resource.TestCheckResourceAttr("keycloak_realm_keystore_rsa.realm_rsa", "private_key_wo_version", "v1"),
					resource.TestCheckResourceAttr("keycloak_realm_keystore_rsa.realm_rsa", "certificate_wo_version", "v1"),
				),
			},
			{
				// UPDATE of another attribute without changing the versions keeps the stored key
				Config: testKeycloakRealmKeystoreRsa_writeOnly(rsaName, 200, privateKey, "v1", certificate, "v1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRealmKeystoreRsaExists("keycloak_realm_keystore_rsa.realm_rsa"),
					resource.TestCheckResourceAttr("keycloak_realm_keystore_rsa.realm_rsa", "priority", "200"),
					testAccCheckRealmKeystoreRsaCertificate("keycloak_realm_keystore_rsa.realm_rsa", certificate),
				),
			},
			{
				// ROTATE the key by changing the versions
				Config: testKeycloakRealmKeystoreRsa_writeOnly(rsaName, 200, newPrivateKey, "v2", newCertificate, "v2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRealmKeystoreRsaExists("keycloak_realm_keystore_rsa.realm_rsa"),
					resource.TestCheckResourceAttr("keycloak_realm_keystore_rsa.realm_rsa", "private_key_wo_version", "v2"),
					resource.TestCheckResourceAttr("keycloak_realm_keystore_rsa.realm_rsa", "certificate_wo_version", "v2"),
					testAccCheckRealmKeystoreRsaCertificate("keycloak_realm_keystore_rsa.realm_rsa", newCertificate),
				),
			},
		},
	})
}

// TestAccKeycloakRealmKeystoreRsa_writeOnlyFromComputedValue covers the documented use case where
// `private_key_wo` and `certificate_wo` come from an ephemeral or computed source, so their values
// are unknown while the plan is created and only available during apply.
func TestAccKeycloakRealmKeystoreRsa_writeOnlyFromComputedValue(t *testing.T) {
	t.Parallel()

	rsaName := acctest.RandomWithPrefix("tf-acc")
	clientId := acctest.RandomWithPrefix("tf-acc")
	privateKey, certificate := generateKeyAndCert(2048)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRealmKeystoreRsaDestroy(),
		Steps: []resource.TestStep{
			{
				// CREATE from values that are unknown during plan
				Config: testKeycloakRealmKeystoreRsa_writeOnlyFromComputedValue(rsaName, clientId, privateKey, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRealmKeystoreRsaExists("keycloak_realm_keystore_rsa.realm_rsa"),
					testAccCheckRealmKeystoreRsaCertificate("keycloak_realm_keystore_rsa.realm_rsa", certificate),
					resource.TestCheckNoResourceAttr("keycloak_realm_keystore_rsa.realm_rsa", "private_key"),
					resource.TestCheckNoResourceAttr("keycloak_realm_keystore_rsa.realm_rsa", "private_key_wo"),
					resource.TestCheckNoResourceAttr("keycloak_realm_keystore_rsa.realm_rsa", "certificate"),
					resource.TestCheckNoResourceAttr("keycloak_realm_keystore_rsa.realm_rsa", "certificate_wo"),
					resource.TestCheckResourceAttr("keycloak_realm_keystore_rsa.realm_rsa", "private_key_wo_version", "v1"),
					resource.TestCheckResourceAttr("keycloak_realm_keystore_rsa.realm_rsa", "certificate_wo_version", "v1"),
				),
			},
			{
				// re-applying the same config results in an empty plan
				Config:             testKeycloakRealmKeystoreRsa_writeOnlyFromComputedValue(rsaName, clientId, privateKey, certificate),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccKeycloakRealmKeystoreRsa_writeOnlyValidation(t *testing.T) {
	t.Parallel()

	rsaName := acctest.RandomWithPrefix("tf-acc")
	privateKey, certificate := generateKeyAndCert(2048)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRealmKeystoreRsaDestroy(),
		Steps: []resource.TestStep{
			{
				// private_key and private_key_wo conflict
				Config: testKeycloakRealmKeystoreRsa_writeOnlyAttrs(rsaName, fmt.Sprintf(`
	private_key            = "%s"
	private_key_wo         = "%s"
	private_key_wo_version = "v1"
	certificate            = "%s"`, privateKey, privateKey, certificate)),
				ExpectError: regexp.MustCompile("conflicts with"),
			},
			{
				// private_key_wo requires private_key_wo_version
				Config: testKeycloakRealmKeystoreRsa_writeOnlyAttrs(rsaName, fmt.Sprintf(`
	private_key_wo = "%s"
	certificate    = "%s"`, privateKey, certificate)),
				ExpectError: regexp.MustCompile("private_key_wo_version"),
			},
			{
				// certificate or certificate_wo is required
				Config: testKeycloakRealmKeystoreRsa_writeOnlyAttrs(rsaName, fmt.Sprintf(`
	private_key = "%s"`, privateKey)),
				ExpectError: regexp.MustCompile("certificate"),
			},
		},
	})
}

func testAccCheckRealmKeystoreRsaCertificate(resourceName, expectedCertificate string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedKeystore, err := getKeycloakRealmKeystoreRsaFromState(s, resourceName)
		if err != nil {
			return err
		}

		if fetchedKeystore.Certificate != expectedCertificate {
			return fmt.Errorf("expected certificate %s but got %s", expectedCertificate, fetchedKeystore.Certificate)
		}

		return nil
	}
}

func testKeycloakRealmKeystoreRsa_writeOnly(rsaName string, priority int, privateKey, privateKeyVersion, certificate, certificateVersion string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_realm_keystore_rsa" "realm_rsa" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id

	priority  = %d
	algorithm = "RS384"

	private_key_wo         = "%s"
	private_key_wo_version = "%s"
	certificate_wo         = "%s"
	certificate_wo_version = "%s"
}
	`, testAccRealmUserFederation.Realm, rsaName, priority, privateKey, privateKeyVersion, certificate, certificateVersion)
}

func testKeycloakRealmKeystoreRsa_writeOnlyFromComputedValue(rsaName, clientId, privateKey, certificate string) string {
	// terraform_data.output is unknown during plan because its input depends on the id of a resource
	// that does not exist yet, which is the same situation as a value coming from another module
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "source" {
	realm_id    = data.keycloak_realm.realm.id
	client_id   = "%s"
	access_type = "CONFIDENTIAL"
}

resource "terraform_data" "private_key" {
	input = keycloak_openid_client.source.id != "" ? "%s" : ""
}

resource "terraform_data" "certificate" {
	input = keycloak_openid_client.source.id != "" ? "%s" : ""
}

resource "keycloak_realm_keystore_rsa" "realm_rsa" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id

	priority  = 100
	algorithm = "RS384"

	private_key_wo         = terraform_data.private_key.output
	private_key_wo_version = "v1"
	certificate_wo         = terraform_data.certificate.output
	certificate_wo_version = "v1"
}
	`, testAccRealmUserFederation.Realm, clientId, privateKey, certificate, rsaName)
}

func testKeycloakRealmKeystoreRsa_writeOnlyAttrs(rsaName, attrs string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_realm_keystore_rsa" "realm_rsa" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id
	%s
}
	`, testAccRealmUserFederation.Realm, rsaName, attrs)
}

func testAccCheckRealmKeystoreRsaExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getKeycloakRealmKeystoreRsaFromState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckRealmKeystoreRsaKidInRealmKeys(resourceName, expectedKid string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedKeystore, err := getKeycloakRealmKeystoreRsaFromState(s, resourceName)
		if err != nil {
			return err
		}

		keys, err := keycloakClient.GetRealmKeys(testCtx, fetchedKeystore.RealmId)
		if err != nil {
			return fmt.Errorf("error fetching realm keys: %w", err)
		}

		var candidates []keycloak.Key
		for _, k := range keys.Keys {
			if k.Algorithm != nil && *k.Algorithm == fetchedKeystore.Algorithm &&
				k.Certificate != nil && *k.Certificate == fetchedKeystore.Certificate {
				candidates = append(candidates, k)
			}
		}

		for _, c := range candidates {
			if c.Kid != nil && *c.Kid == expectedKid {
				return nil
			}
		}

		return fmt.Errorf("could not find expected kid in realm keys. expected kid=%s", expectedKid)
	}
}

func testAccCheckRealmKeystoreRsaFetch(resourceName string, keystore *keycloak.RealmKeystoreRsa) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedKeystore, err := getKeycloakRealmKeystoreRsaFromState(s, resourceName)
		if err != nil {
			return err
		}

		keystore.Id = fetchedKeystore.Id
		keystore.RealmId = fetchedKeystore.RealmId

		return nil
	}
}

func testAccCheckRealmKeystoreRsaParentId(resourceName, expectedParentId string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		keystore, err := getKeycloakRealmKeystoreRsaFromState(s, resourceName)
		if err != nil {
			return err
		}

		if keystore.ParentId != expectedParentId {
			return fmt.Errorf("expected rsa keystore %s to have parent id %s, but got %s", keystore.Id, expectedParentId, keystore.ParentId)
		}

		return nil
	}
}

func testAccCheckRealmKeystoreRsaDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_realm_keystore_rsa" {
				continue
			}

			id := rs.Primary.ID
			realm := rs.Primary.Attributes["realm_id"]
			keystoreRsa, _ := keycloakClient.GetRealmKeystoreRsa(testCtx, realm, id)
			if keystoreRsa != nil {
				return fmt.Errorf("rsa keystore with id %s still exists", id)
			}
		}

		return nil
	}
}

func getKeycloakRealmKeystoreRsaFromState(s *terraform.State,
	resourceName string) (*keycloak.RealmKeystoreRsa,
	error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm_id"]

	realmKeystore, err := keycloakClient.GetRealmKeystoreRsa(testCtx, realm, id)
	if err != nil {
		return nil, fmt.Errorf("error getting rsa keystore with id %s: %s", id, err)
	}

	return realmKeystore, nil
}

func generateKeyAndCert(bits int) (string, string) {
	key, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		log.Fatal("Private key cannot be created.", err.Error())
	}

	// Generate a pem block with the private key
	keyPem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	tml := x509.Certificate{
		// you can add any attr that you need
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(5, 0, 0),
		// you have to generate a different serial number each execution
		SerialNumber: big.NewInt(123123),
		Subject: pkix.Name{
			CommonName:   "New Name",
			Organization: []string{"New Org."},
		},
		BasicConstraintsValid: true,
	}
	cert, err := x509.CreateCertificate(rand.Reader, &tml, &tml, &key.PublicKey, key)
	if err != nil {
		log.Fatal("Certificate cannot be created.", err.Error())
	}

	// Generate a pem block with the certificate
	certPem := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert,
	})

	return parsePemRealmKeystoreRsa(string(keyPem)), parsePemRealmKeystoreRsa(string(certPem))
}

func parsePemRealmKeystoreRsa(input string) string {
	headersRegexp := regexp.MustCompile(`-----(BEGIN|END).+-----`) // Header and footer like "-----BEGIN RSA PRIVATE KEY-----"
	output := headersRegexp.ReplaceAllString(input, "")
	output = strings.ReplaceAll(output, "\n", "")

	return output
}

func testKeycloakRealmKeystoreRsa_basic(rsaName, privateKey, certificate string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_realm_keystore_rsa" "realm_rsa" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id

    priority    = 100
    algorithm   = "RS384"
    private_key = "%s"
    certificate = "%s"
}
	`, testAccRealmUserFederation.Realm, rsaName, privateKey, certificate)
}

func testKeycloakRealmKeystoreRsa_parentId(realmName, rsaName, privateKey, certificate, priority string) string {
	return fmt.Sprintf(`
resource "keycloak_realm_keystore_rsa" "realm_rsa" {
	name      = "%s"
	realm_id  = "%s"

    priority    = %s
    algorithm   = "RS384"
    private_key = "%s"
    certificate = "%s"
}
	`, rsaName, realmName, priority, privateKey, certificate)
}

func testKeycloakRealmKeystoreRsa_basicWithAttrValidation(provider, rsaName, attr, val, privateKey,
	certificate string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_realm_keystore_rsa" "realm_rsa" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id

	%s        = "%s"

    private_key = "%s"
    certificate = "%s"
	provider_id = "%s"
}
	`, testAccRealmUserFederation.Realm, rsaName, attr, val, privateKey, certificate, provider)
}

func testKeycloakRealmKeystoreRsa_withKidExtraConfig(rsaName, privateKey, certificate, kid string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
    realm = "%s"
}

resource "keycloak_realm_keystore_rsa" "realm_rsa" {
    name      = "%s"
    realm_id  = data.keycloak_realm.realm.id

    priority    = 100
    algorithm   = "RS256"
    private_key = "%s"
    certificate = "%s"
    provider_id = "rsa"

    extra_config = {
        "kid"   = "%s"
    }
}
    `, testAccRealmUserFederation.Realm, rsaName, privateKey, certificate, kid)
}
