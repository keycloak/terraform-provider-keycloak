package provider

import (
	"fmt"
	"github.com/keycloak/terraform-provider-keycloak/keycloak/types"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakSamlClient_basic(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakSamlClientDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakSamlClient_basic(clientId),
				Check:  testAccCheckKeycloakSamlClientExistsWithCorrectProtocol("keycloak_saml_client.saml_client"),
			},
			{
				ResourceName:        "keycloak_saml_client.saml_client",
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: testAccRealm.Realm + "/",
			},
		},
	})
}

func TestAccKeycloakSamlClient_generatedCertificate(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakSamlClientDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakSamlClient_basic(clientId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakSamlClientExistsWithCorrectProtocol("keycloak_saml_client.saml_client"),
					resource.TestCheckResourceAttrSet("keycloak_saml_client.saml_client", "signing_certificate"),
					resource.TestCheckResourceAttrSet("keycloak_saml_client.saml_client", "signing_private_key"),
					resource.TestCheckResourceAttrSet("keycloak_saml_client.saml_client", "signing_certificate_sha1"),
					resource.TestCheckResourceAttrSet("keycloak_saml_client.saml_client", "signing_private_key_sha1"),
				),
			},
		},
	})
}

func TestAccKeycloakSamlClient_createAfterManualDestroy(t *testing.T) {
	t.Parallel()
	var client = &keycloak.SamlClient{}

	clientId := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakSamlClientDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakSamlClient_basic(clientId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakSamlClientExistsWithCorrectProtocol("keycloak_saml_client.saml_client"),
					testAccCheckKeycloakSamlClientFetch("keycloak_saml_client.saml_client", client),
				),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteSamlClient(testCtx, client.RealmId, client.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakSamlClient_basic(clientId),
				Check:  testAccCheckKeycloakSamlClientExistsWithCorrectProtocol("keycloak_saml_client.saml_client"),
			},
		},
	})
}

func TestAccKeycloakSamlClient_updateRealm(t *testing.T) {
	t.Parallel()

	clientId := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakSamlClientDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakSamlClient_updateRealmBefore(clientId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakSamlClientExistsWithCorrectProtocol("keycloak_saml_client.saml_client"),
					resource.TestCheckResourceAttr("keycloak_saml_client.saml_client", "realm_id", testAccRealm.Realm),
				),
			},
			{
				Config: testKeycloakSamlClient_updateRealmAfter(clientId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakSamlClientExistsWithCorrectProtocol("keycloak_saml_client.saml_client"),
					resource.TestCheckResourceAttr("keycloak_saml_client.saml_client", "realm_id", testAccRealmTwo.Realm),
				),
			},
		},
	})
}

func TestAccKeycloakSamlClient_updateInPlace(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	enabled := randomBool()
	frontChannelLogout := randomBool()

	encryptionCertificateBefore := acctest.RandomWithPrefix("tf-acc")
	encryptionCertificateAfter := acctest.RandomWithPrefix("tf-acc")
	signingCertificateBefore := acctest.RandomWithPrefix("tf-acc")
	signingCertificateAfter := acctest.RandomWithPrefix("tf-acc")
	signingPrivateKeyBefore := acctest.RandomWithPrefix("tf-acc")
	signingPrivateKeyAfter := acctest.RandomWithPrefix("tf-acc")

	samlClientBefore := &keycloak.SamlClient{
		RealmId:  testAccRealm.Realm,
		ClientId: clientId,
		Name:     acctest.RandString(10),

		Enabled:     enabled,
		Description: acctest.RandString(50),

		FrontChannelLogout: frontChannelLogout,

		RootUrl: "http://localhost:2222/" + acctest.RandString(20),
		ValidRedirectUris: []string{
			acctest.RandString(20),
			acctest.RandString(20),
			acctest.RandString(20),
		},
		BaseUrl:                 "http://localhost:2222/" + acctest.RandString(20),
		MasterSamlProcessingUrl: acctest.RandString(20),

		Attributes: &keycloak.SamlClientAttributes{
			IncludeAuthnStatement:           types.KeycloakBoolQuoted(randomBool()),
			SignDocuments:                   types.KeycloakBoolQuoted(randomBool()),
			SignAssertions:                  types.KeycloakBoolQuoted(randomBool()),
			EncryptAssertions:               types.KeycloakBoolQuoted(randomBool()),
			ClientSignatureRequired:         true,
			ForcePostBinding:                types.KeycloakBoolQuoted(randomBool()),
			ForceNameIdFormat:               types.KeycloakBoolQuoted(randomBool()),
			SignatureAlgorithm:              randomStringInSlice(keycloakSamlClientSignatureAlgorithms),
			SignatureKeyName:                randomStringInSlice(keycloakSamlClientSignatureKeyNames),
			NameIdFormat:                    randomStringInSlice(keycloakSamlClientNameIdFormats),
			EncryptionCertificate:           encryptionCertificateBefore,
			SigningCertificate:              signingCertificateBefore,
			SigningPrivateKey:               signingPrivateKeyBefore,
			IDPInitiatedSSOURLName:          acctest.RandString(20),
			IDPInitiatedSSORelayState:       acctest.RandString(20),
			AssertionConsumerPostURL:        acctest.RandString(20),
			AssertionConsumerRedirectURL:    acctest.RandString(20),
			LogoutServicePostBindingURL:     acctest.RandString(20),
			LogoutServiceRedirectBindingURL: acctest.RandString(20),
			LoginTheme:                      "keycloak",
		},
	}

	samlClientAfter := &keycloak.SamlClient{
		RealmId:  testAccRealm.Realm,
		ClientId: clientId,
		Name:     acctest.RandString(10),

		Enabled:     !enabled,
		Description: acctest.RandString(50),

		FrontChannelLogout: !frontChannelLogout,

		RootUrl: "http://localhost:2222/" + acctest.RandString(20),
		ValidRedirectUris: []string{
			acctest.RandString(20),
		},
		BaseUrl:                 "http://localhost:2222/" + acctest.RandString(20),
		MasterSamlProcessingUrl: acctest.RandString(20),

		Attributes: &keycloak.SamlClientAttributes{
			IncludeAuthnStatement:           types.KeycloakBoolQuoted(randomBool()),
			SignDocuments:                   types.KeycloakBoolQuoted(randomBool()),
			SignAssertions:                  types.KeycloakBoolQuoted(randomBool()),
			EncryptAssertions:               types.KeycloakBoolQuoted(randomBool()),
			ClientSignatureRequired:         true,
			ForcePostBinding:                types.KeycloakBoolQuoted(randomBool()),
			ForceNameIdFormat:               types.KeycloakBoolQuoted(randomBool()),
			SignatureAlgorithm:              randomStringInSlice(keycloakSamlClientSignatureAlgorithms),
			SignatureKeyName:                randomStringInSlice(keycloakSamlClientSignatureKeyNames),
			NameIdFormat:                    randomStringInSlice(keycloakSamlClientNameIdFormats),
			EncryptionCertificate:           encryptionCertificateAfter,
			SigningCertificate:              signingCertificateAfter,
			SigningPrivateKey:               signingPrivateKeyAfter,
			IDPInitiatedSSOURLName:          acctest.RandString(20),
			IDPInitiatedSSORelayState:       acctest.RandString(20),
			AssertionConsumerPostURL:        acctest.RandString(20),
			AssertionConsumerRedirectURL:    acctest.RandString(20),
			LogoutServicePostBindingURL:     acctest.RandString(20),
			LogoutServiceRedirectBindingURL: acctest.RandString(20),
			LoginTheme:                      "keycloak",
		},
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakSamlClientDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakSamlClient_fromInterface(samlClientBefore),
				Check:  testAccCheckKeycloakSamlClientExistsWithCorrectProtocol("keycloak_saml_client.saml_client"),
			},
			{
				Config: testKeycloakSamlClient_fromInterface(samlClientAfter),
				Check:  testAccCheckKeycloakSamlClientExistsWithCorrectProtocol("keycloak_saml_client.saml_client"),
			},
		},
	})
}

func TestAccKeycloakSamlClient_certificateAndKey(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakSamlClientDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakSamlClient_signingCertificateAndKey(clientId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakSamlClientExistsWithCorrectProtocol("keycloak_saml_client.saml_client"),
					testAccCheckKeycloakSamlClientHasSigningCertificate("keycloak_saml_client.saml_client"),
					testAccCheckKeycloakSamlClientHasPrivateKey("keycloak_saml_client.saml_client"),
					resource.TestCheckResourceAttrSet("keycloak_saml_client.saml_client", "signing_certificate_sha1"),
					resource.TestCheckResourceAttrSet("keycloak_saml_client.saml_client", "signing_private_key_sha1"),
				),
			},
		},
	})
}

func TestAccKeycloakSamlClient_encryptionCertificate(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakSamlClientDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakSamlClient_encryptionCertificate(clientId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakSamlClientExistsWithCorrectProtocol("keycloak_saml_client.saml_client"),
					testAccCheckKeycloakSamlClientHasEncryptionCertificate("keycloak_saml_client.saml_client"),
					resource.TestCheckResourceAttrSet("keycloak_saml_client.saml_client", "encryption_certificate_sha1"),
				),
			},
		},
	})
}

func TestAccCheckKeycloakSamlClient_authenticationFlowBindingOverrides(t *testing.T) {
	t.Parallel()

	clientId := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakSamlClientDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakSamlClient_authenticationFlowBindingOverrides(clientId),
				Check:  testAccCheckKeycloakSamlClientAuthenticationFlowBindingOverrides("keycloak_saml_client.client", "keycloak_authentication_flow.another_flow"),
			},
			{
				Config: testKeycloakSamlClient_withoutFlowBindingOverrides(clientId),
				Check:  testAccCheckKeycloakSamlClientAuthenticationFlowBindingOverrides("keycloak_saml_client.client", ""),
			},
		},
	})
}

func TestAccKeycloakSamlClient_extraConfig(t *testing.T) {
	t.Parallel()

	clientId := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakSamlClientDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakSamlClient_extraConfig(clientId, map[string]string{
					"key1": "value1",
					"key2": "value2",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakSamlClientExtraConfig("keycloak_saml_client.saml_client", "key1", "value1"),
					testAccCheckKeycloakSamlClientExtraConfig("keycloak_saml_client.saml_client", "key2", "value2"),
				),
			},
			{
				Config: testKeycloakSamlClient_extraConfig(clientId, map[string]string{
					"key2": "value2",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakSamlClientExtraConfig("keycloak_saml_client.saml_client", "key2", "value2"),
					testAccCheckKeycloakSamlClientExtraConfigMissing("keycloak_saml_client.saml_client", "key1"),
				),
			},
		},
	})
}

func TestAccKeycloakSamlClient_extraConfigInvalid(t *testing.T) {
	t.Parallel()

	clientId := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakSamlClientDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakSamlClient_extraConfig(clientId, map[string]string{"saml.signature.algorithm": "RSA_SHA1"}),
				ExpectError: regexp.MustCompile(`extra_config key "saml.signature.algorithm" is not allowed`),
			},
		},
	})
}

func testAccCheckKeycloakSamlClientExistsWithCorrectProtocol(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := getSamlClientFromState(s, resourceName)
		if err != nil {
			return err
		}

		if client.Protocol != "saml" {
			return fmt.Errorf("expected saml client to have saml protocol, but got %s", client.Protocol)
		}

		return nil
	}
}

func testAccCheckKeycloakSamlClientHasEncryptionCertificate(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := getSamlClientFromState(s, resourceName)
		if err != nil {
			return err
		}

		if client.Attributes.EncryptionCertificate == "" {
			return fmt.Errorf("expected saml client to have a encryption certificate")
		}

		if strings.Contains(client.Attributes.EncryptionCertificate, "-----BEGIN CERTIFICATE-----") || strings.Contains(client.Attributes.EncryptionCertificate, "-----END CERTIFICATE-----") {
			return fmt.Errorf("expected saml client encryption certificate to not contain headers")
		}

		if strings.ContainsAny(client.Attributes.EncryptionCertificate, "\n\r ") {
			return fmt.Errorf("expected saml client encryption certificate to not contain whitespace")
		}

		return nil
	}
}

func testAccCheckKeycloakSamlClientHasSigningCertificate(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := getSamlClientFromState(s, resourceName)
		if err != nil {
			return err
		}

		if client.Attributes.SigningCertificate == "" {
			return fmt.Errorf("expected saml client to have a signing certificate")
		}

		if strings.Contains(client.Attributes.SigningCertificate, "-----BEGIN CERTIFICATE-----") || strings.Contains(client.Attributes.SigningCertificate, "-----END CERTIFICATE-----") {
			return fmt.Errorf("expected saml client signing certificate to not contain headers")
		}

		if strings.ContainsAny(client.Attributes.SigningCertificate, "\n\r ") {
			return fmt.Errorf("expected saml client signing certificate to not contain whitespace")
		}

		return nil
	}
}

func testAccCheckKeycloakSamlClientHasPrivateKey(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := getSamlClientFromState(s, resourceName)
		if err != nil {
			return err
		}

		if client.Attributes.SigningPrivateKey == "" {
			return fmt.Errorf("expected saml client to have a signing private key")
		}

		if strings.Contains(client.Attributes.SigningPrivateKey, "-----BEGIN PRIVATE KEY-----") || strings.Contains(client.Attributes.SigningPrivateKey, "-----END PRIVATE KEY-----") {
			return fmt.Errorf("expected saml client signing private key to not contain headers")
		}

		if strings.ContainsAny(client.Attributes.SigningPrivateKey, "\n\r ") {
			return fmt.Errorf("expected saml client signing private key to not contain whitespace")
		}

		return nil
	}
}

func testAccCheckKeycloakSamlClientFetch(resourceName string, client *keycloak.SamlClient) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedClient, err := getSamlClientFromState(s, resourceName)
		if err != nil {
			return err
		}

		client.Id = fetchedClient.Id
		client.RealmId = fetchedClient.RealmId

		return nil
	}
}

func testAccCheckKeycloakSamlClientDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_saml_client" {
				continue
			}

			id := rs.Primary.ID
			realm := rs.Primary.Attributes["realm_id"]

			client, _ := keycloakClient.GetSamlClient(testCtx, realm, id)
			if client != nil {
				return fmt.Errorf("saml client %s still exists", id)
			}
		}

		return nil
	}
}

func getSamlClientFromState(s *terraform.State, resourceName string) (*keycloak.SamlClient, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm_id"]

	client, err := keycloakClient.GetSamlClient(testCtx, realm, id)
	if err != nil {
		return nil, fmt.Errorf("error getting saml client %s: %s", id, err)
	}

	return client, nil
}

func testAccCheckKeycloakSamlClientAuthenticationFlowBindingOverrides(resourceName, flowResourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := getSamlClientFromState(s, resourceName)
		if err != nil {
			return err
		}

		if flowResourceName == "" {
			if client.AuthenticationFlowBindingOverrides.BrowserId != "" {
				return fmt.Errorf("expected openid client to have browserId set to empty, but got %s", client.AuthenticationFlowBindingOverrides.BrowserId)
			}

			if client.AuthenticationFlowBindingOverrides.DirectGrantId != "" {
				return fmt.Errorf("expected openid client to have directGrantId set to empty, but got %s", client.AuthenticationFlowBindingOverrides.DirectGrantId)
			}

		} else {
			flow, err := getAuthenticationFlowFromState(s, flowResourceName)
			if err != nil {
				return err
			}

			if client.AuthenticationFlowBindingOverrides.BrowserId != flow.Id {
				return fmt.Errorf("expected openid client to have browserId set to %s, but got %s", flow.Id, client.AuthenticationFlowBindingOverrides.BrowserId)
			}

			if client.AuthenticationFlowBindingOverrides.DirectGrantId != flow.Id {
				return fmt.Errorf("expected openid client to have directGrantId set to %s, but got %s", flow.Id, client.AuthenticationFlowBindingOverrides.DirectGrantId)
			}
		}

		return nil
	}
}

func testAccCheckKeycloakSamlClientExtraConfig(resourceName string, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := getSamlClientFromState(s, resourceName)
		if err != nil {
			return err
		}

		if client.Attributes.ExtraConfig[key] != value {
			return fmt.Errorf("expected saml client to have attribute %v set to %v, but got %v", key, value, client.Attributes.ExtraConfig[key])
		}

		return nil
	}
}

// check that a particular extra config key is missing
func testAccCheckKeycloakSamlClientExtraConfigMissing(resourceName string, key string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := getSamlClientFromState(s, resourceName)
		if err != nil {
			return err
		}

		if val, ok := client.Attributes.ExtraConfig[key]; ok {
			// keycloak 13+ will remove attributes if set to empty string. on older versions, we'll just check if this value is empty
			if versionOk, _ := keycloakClient.VersionIsGreaterThanOrEqualTo(testCtx, keycloak.Version_13); !versionOk {
				if val != "" {
					return fmt.Errorf("expected saml client to have empty attribute %v", key)
				}

				return nil
			}

			return fmt.Errorf("expected saml client to not have attribute %v", key)
		}

		return nil
	}
}

func testKeycloakSamlClient_basic(clientId string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_saml_client" "saml_client" {
	client_id = "%s"
	realm_id  = data.keycloak_realm.realm.id
}
	`, testAccRealm.Realm, clientId)
}

func testKeycloakSamlClient_generatedCertificate(clientId string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_saml_client" "saml_client" {
	client_id = "%s"
	realm_id  = data.keycloak_realm.realm.id

	sign_documents          = false
	sign_assertions         = true
	include_authn_statement = true
}
	`, testAccRealm.Realm, clientId)
}

func testKeycloakSamlClient_updateRealmBefore(clientId string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm_1" {
	realm = "%s"
}

data "keycloak_realm" "realm_2" {
	realm = "%s"
}

resource "keycloak_saml_client" "saml_client" {
	client_id = "%s"
	realm_id  = data.keycloak_realm.realm_1.id
}
	`, testAccRealm.Realm, testAccRealmTwo.Realm, clientId)
}

func testKeycloakSamlClient_updateRealmAfter(clientId string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm_1" {
	realm = "%s"
}

data "keycloak_realm" "realm_2" {
	realm = "%s"
}

resource "keycloak_saml_client" "saml_client" {
	client_id = "%s"
	realm_id  = data.keycloak_realm.realm_2.id
}
	`, testAccRealm.Realm, testAccRealmTwo.Realm, clientId)
}

func testKeycloakSamlClient_fromInterface(client *keycloak.SamlClient) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_saml_client" "saml_client" {
	realm_id    = data.keycloak_realm.realm.id
	client_id   = "%s"
	name        = "%s"
	description = "%s"
	enabled     = %t

	# below attributes are bools, but the model (and API) uses strings
	include_authn_statement    = %t
	sign_documents             = %t
	sign_assertions            = %t
	encrypt_assertions         = %t
	client_signature_required  = %t
	force_post_binding         = %t
	force_name_id_format       = %t

	front_channel_logout       = %t
	signature_algorithm        = "%s"
	signature_key_name         = "%s"
	name_id_format             = "%s"
	root_url                   = "%s"
	valid_redirect_uris        = %s
	base_url                   = "%s"
	master_saml_processing_url = "%s"

	encryption_certificate     = "%s"
	signing_certificate        = "%s"
	signing_private_key        = "%s"

	idp_initiated_sso_url_name    = "%s"
	idp_initiated_sso_relay_state = "%s"

	assertion_consumer_post_url         = "%s"
	assertion_consumer_redirect_url     = "%s"
	logout_service_post_binding_url     = "%s"
	logout_service_redirect_binding_url = "%s"
}
	`, client.RealmId,
		client.ClientId,
		client.Name,
		client.Description,
		client.Enabled,
		client.Attributes.IncludeAuthnStatement,
		client.Attributes.SignDocuments,
		client.Attributes.SignAssertions,
		client.Attributes.EncryptAssertions,
		client.Attributes.ClientSignatureRequired,
		client.Attributes.ForcePostBinding,
		client.Attributes.ForceNameIdFormat,
		client.FrontChannelLogout,
		client.Attributes.SignatureAlgorithm,
		client.Attributes.SignatureKeyName,
		client.Attributes.NameIdFormat,
		client.RootUrl,
		arrayOfStringsForTerraformResource(client.ValidRedirectUris),
		client.BaseUrl, client.MasterSamlProcessingUrl,
		client.Attributes.EncryptionCertificate,
		client.Attributes.SigningCertificate,
		client.Attributes.SigningPrivateKey,
		client.Attributes.IDPInitiatedSSOURLName,
		client.Attributes.IDPInitiatedSSORelayState,
		client.Attributes.AssertionConsumerPostURL,
		client.Attributes.AssertionConsumerRedirectURL,
		client.Attributes.LogoutServicePostBindingURL,
		client.Attributes.LogoutServiceRedirectBindingURL,
	)
}

func testKeycloakSamlClient_signingCertificateAndKey(clientId string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_saml_client" "saml_client" {
	client_id               = "%s"
	realm_id                = data.keycloak_realm.realm.id
	name                    = "test-saml-client"

	sign_documents          = false
	sign_assertions         = true
	encrypt_assertions      = false
	include_authn_statement = true

	signing_certificate     = file("misc/saml-cert.pem")
	signing_private_key     = file("misc/saml-key.pem")
}
	`, testAccRealm.Realm, clientId)
}

func testKeycloakSamlClient_signingCertificateNoKey(clientId string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_saml_client" "saml_client" {
	client_id               = "%s"
	realm_id                = data.keycloak_realm.realm.id
	name                    = "test-saml-client"

	sign_documents          = false
	sign_assertions         = true
	encrypt_assertions      = false
	include_authn_statement = true

	signing_certificate     = file("misc/saml-cert.pem")
}
	`, testAccRealm.Realm, clientId)
}

func testKeycloakSamlClient_encryptionCertificate(clientId string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_saml_client" "saml_client" {
	client_id               = "%s"
	realm_id                = data.keycloak_realm.realm.id
	name                    = "test-saml-client"

	encrypt_assertions      = true
	include_authn_statement = true

	encryption_certificate     = file("misc/saml-cert.pem")
}
	`, testAccRealm.Realm, clientId)
}

func testKeycloakSamlClient_NoEncryptionCertificate(clientId string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_saml_client" "saml_client" {
	client_id               = "%s"
	realm_id                = data.keycloak_realm.realm.id
	name                    = "test-saml-client"

	encrypt_assertions      = true
	include_authn_statement = true
}
	`, testAccRealm.Realm, clientId)
}

func testKeycloakSamlClient_authenticationFlowBindingOverrides(clientId string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_authentication_flow" "another_flow" {
  alias       = "anotherFlow"
  realm_id    = data.keycloak_realm.realm.id
  description = "this is another flow"
}

resource "keycloak_saml_client" "client" {
	client_id = "%s"
	realm_id  = data.keycloak_realm.realm.id
	name      = "test-saml-client"

	authentication_flow_binding_overrides {
		browser_id      = keycloak_authentication_flow.another_flow.id
		direct_grant_id = keycloak_authentication_flow.another_flow.id
	}
}
	`, testAccRealm.Realm, clientId)
}

func testKeycloakSamlClient_withoutFlowBindingOverrides(clientId string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_authentication_flow" "another_flow" {
  alias       = "anotherFlow"
  realm_id    = data.keycloak_realm.realm.id
  description = "this is another flow"
}

resource "keycloak_saml_client" "client" {
	client_id = "%s"
	realm_id  = data.keycloak_realm.realm.id
	name      = "test-saml-client"
}
	`, testAccRealm.Realm, clientId)
}

func testKeycloakSamlClient_extraConfig(clientId string, extraConfig map[string]string) string {
	var sb strings.Builder
	sb.WriteString("{\n")
	for k, v := range extraConfig {
		sb.WriteString(fmt.Sprintf("\t\t\"%s\" = \"%s\"\n", k, v))
	}
	sb.WriteString("}")

	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_saml_client" "saml_client" {
	client_id = "%s"
	realm_id  = data.keycloak_realm.realm.id

	extra_config = %s
}
	`, testAccRealm.Realm, clientId, sb.String())
}
