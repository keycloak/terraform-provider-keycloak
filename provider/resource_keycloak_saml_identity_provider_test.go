package provider

import (
	"fmt"
	"github.com/keycloak/terraform-provider-keycloak/keycloak/types"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakSamlIdentityProvider_basic(t *testing.T) {
	t.Parallel()

	samlName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakSamlIdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakSamlIdentityProvider_basic(samlName),
				Check:  testAccCheckKeycloakSamlIdentityProviderExists("keycloak_saml_identity_provider.saml"),
			},
		},
	})
}

func TestAccKeycloakSamlIdentityProvider_customProviderId(t *testing.T) {
	t.Parallel()

	samlName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakSamlIdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakSamlIdentityProvider_customProviderId(samlName, "saml"), //actually needs to be something that exists
				Check:  testAccCheckKeycloakSamlIdentityProviderExists("keycloak_saml_identity_provider.saml"),
			},
		},
	})
}

func TestAccKeycloakSamlIdentityProvider_nameIdPolicyFormatTransient(t *testing.T) {
	t.Parallel()

	samlName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakSamlIdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakSamlIdentityProvider_withNameIdPolicyFormat(samlName, "Transient"),
				Check:  testAccCheckKeycloakSamlIdentityProviderHasNameIdPolicyFormatValue("keycloak_saml_identity_provider.saml", "urn:oasis:names:tc:SAML:2.0:nameid-format:transient"),
			},
		},
	})
}

func TestAccKeycloakSamlIdentityProvider_extraConfig(t *testing.T) {
	t.Parallel()

	samlName := acctest.RandomWithPrefix("tf-acc")
	customConfigValue := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakSamlIdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakSamlIdentityProvider_extra_config(samlName, "dummyConfig", customConfigValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakSamlIdentityProviderHasCustomConfigValue("keycloak_saml_identity_provider.saml", customConfigValue),
				),
			},
			{
				Config: testKeycloakSamlIdentityProvider_extra_config(samlName, "another-test-config", customConfigValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakSamlIdentityProviderHasNameIdPolicyFormatValue("keycloak_saml_identity_provider.saml", nameIdPolicyFormats["Email"]),
				),
			},
		},
	})
}

// ensure that extra_config keys which are covered by top-level attributes are not allowed
func TestAccKeycloakSamlIdentityProvider_extraConfigInvalid(t *testing.T) {
	t.Parallel()

	samlName := acctest.RandomWithPrefix("tf-acc")
	customConfigValue := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakSamlIdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakSamlIdentityProvider_extra_config(samlName, "syncMode", customConfigValue),
				ExpectError: regexp.MustCompile("extra_config key \"syncMode\" is not allowed"),
			},
		},
	})
}

func TestAccKeycloakSamlIdentityProvider_createAfterManualDestroy(t *testing.T) {
	t.Parallel()

	var saml = &keycloak.IdentityProvider{}

	samlName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakSamlIdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakSamlIdentityProvider_basic(samlName),
				Check:  testAccCheckKeycloakSamlIdentityProviderFetch("keycloak_saml_identity_provider.saml", saml),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteIdentityProvider(testCtx, saml.Realm, saml.Alias)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakSamlIdentityProvider_basic(samlName),
				Check:  testAccCheckKeycloakSamlIdentityProviderExists("keycloak_saml_identity_provider.saml"),
			},
		},
	})
}

func TestAccKeycloakSamlIdentityProvider_basicUpdateAll(t *testing.T) {
	t.Parallel()

	firstEnabled := randomBool()
	firstBackchannel := randomBool()
	firstValidateSignature := randomBool()
	firstHideOnLogin := randomBool()
	firstForceAuthn := randomBool()
	firstAssertionsEncrypted := randomBool()
	firstAssertionsSigned := randomBool()
	firstPostBindingLogout := randomBool()
	firstPostBindingResponse := randomBool()
	firstPostBindingRequest := randomBool()
	firstLoginHint := randomBool()

	firstSaml := &keycloak.IdentityProvider{
		Alias:   acctest.RandString(10),
		Enabled: firstEnabled,
		Config: &keycloak.IdentityProviderConfig{
			EntityId:                        "https://example.com/entity_id/1",
			SingleSignOnServiceUrl:          "https://example.com/signon/1",
			BackchannelSupported:            types.KeycloakBoolQuoted(firstBackchannel),
			ValidateSignature:               types.KeycloakBoolQuoted(firstValidateSignature),
			HideOnLoginPage:                 types.KeycloakBoolQuoted(firstHideOnLogin),
			NameIDPolicyFormat:              "Email",
			SingleLogoutServiceUrl:          "https://example.com/logout/1",
			SigningCertificate:              acctest.RandString(10),
			SignatureAlgorithm:              "RSA_SHA512",
			XmlSigKeyInfoKeyNameTransformer: "KEY_ID",
			PostBindingAuthnRequest:         types.KeycloakBoolQuoted(firstPostBindingRequest),
			PostBindingResponse:             types.KeycloakBoolQuoted(firstPostBindingResponse),
			PostBindingLogout:               types.KeycloakBoolQuoted(firstPostBindingLogout),
			ForceAuthn:                      types.KeycloakBoolQuoted(firstForceAuthn),
			WantAssertionsSigned:            types.KeycloakBoolQuoted(firstAssertionsSigned),
			WantAssertionsEncrypted:         types.KeycloakBoolQuoted(firstAssertionsEncrypted),
			LoginHint:                       strconv.Quote(strconv.FormatBool(firstLoginHint)),
			GuiOrder:                        strconv.Itoa(acctest.RandIntRange(1, 3)),
			SyncMode:                        randomStringInSlice(syncModes),
			AuthnContextClassRefs:           types.KeycloakSliceQuoted{"foo", "bar"},
			AuthnContextDeclRefs:            types.KeycloakSliceQuoted{"foo"},
			AuthnContextComparisonType:      "exact",
		},
	}

	secondSaml := &keycloak.IdentityProvider{
		Alias:   acctest.RandString(10),
		Enabled: !firstEnabled,
		Config: &keycloak.IdentityProviderConfig{
			EntityId:                        "https://example.com/entity_id/2",
			SingleSignOnServiceUrl:          "https://example.com/signon/2",
			BackchannelSupported:            types.KeycloakBoolQuoted(!firstBackchannel),
			ValidateSignature:               types.KeycloakBoolQuoted(!firstValidateSignature),
			HideOnLoginPage:                 types.KeycloakBoolQuoted(!firstHideOnLogin),
			NameIDPolicyFormat:              "Persistent",
			SingleLogoutServiceUrl:          "https://example.com/logout/2",
			SigningCertificate:              acctest.RandString(10),
			SignatureAlgorithm:              "RSA_SHA256",
			XmlSigKeyInfoKeyNameTransformer: "NONE",
			PostBindingAuthnRequest:         types.KeycloakBoolQuoted(!firstPostBindingRequest),
			PostBindingResponse:             types.KeycloakBoolQuoted(!firstPostBindingResponse),
			PostBindingLogout:               types.KeycloakBoolQuoted(!firstPostBindingLogout),
			ForceAuthn:                      types.KeycloakBoolQuoted(!firstForceAuthn),
			WantAssertionsSigned:            types.KeycloakBoolQuoted(!firstAssertionsSigned),
			WantAssertionsEncrypted:         types.KeycloakBoolQuoted(!firstAssertionsEncrypted),
			LoginHint:                       strconv.Quote(strconv.FormatBool(!firstLoginHint)),
			GuiOrder:                        strconv.Itoa(acctest.RandIntRange(1, 3)),
			SyncMode:                        randomStringInSlice(syncModes),
			AuthnContextClassRefs:           types.KeycloakSliceQuoted{"foo", "hello"},
			AuthnContextDeclRefs:            types.KeycloakSliceQuoted{"baz"},
			AuthnContextComparisonType:      "exact",
		},
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakSamlIdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakSamlIdentityProvider_basicFromInterface(firstSaml),
				Check:  testAccCheckKeycloakSamlIdentityProviderExists("keycloak_saml_identity_provider.saml"),
			},
			{
				Config: testKeycloakSamlIdentityProvider_basicFromInterface(secondSaml),
				Check:  testAccCheckKeycloakSamlIdentityProviderExists("keycloak_saml_identity_provider.saml"),
			},
		},
	})
}

func testAccCheckKeycloakSamlIdentityProviderExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getKeycloakSamlIdentityProviderFromState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckKeycloakSamlIdentityProviderFetch(resourceName string, saml *keycloak.IdentityProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedSaml, err := getKeycloakSamlIdentityProviderFromState(s, resourceName)
		if err != nil {
			return err
		}

		saml.Alias = fetchedSaml.Alias
		saml.Realm = fetchedSaml.Realm

		return nil
	}
}

func testAccCheckKeycloakSamlIdentityProviderHasCustomConfigValue(resourceName, customConfigValue string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedSaml, err := getKeycloakSamlIdentityProviderFromState(s, resourceName)
		if err != nil {
			return err
		}

		if fetchedSaml.Config.ExtraConfig["dummyConfig"].(string) != customConfigValue {
			return fmt.Errorf("expected custom saml provider to have config with a custom key 'dummyConfig' with a value %s, but value was %s", customConfigValue, fetchedSaml.Config.ExtraConfig["dummyConfig"].(string))
		}

		return nil
	}
}

func testAccCheckKeycloakSamlIdentityProviderHasNameIdPolicyFormatValue(resourceName, nameIdPolicyFormatValue string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedSaml, err := getKeycloakSamlIdentityProviderFromState(s, resourceName)
		if err != nil {
			return err
		}

		if fetchedSaml.Config.NameIDPolicyFormat != nameIdPolicyFormatValue {
			return fmt.Errorf("expected saml provider to have config with nameIdPolicyFormat with a value %s, but value was %s", nameIdPolicyFormatValue, fetchedSaml.Config.NameIDPolicyFormat)
		}

		return nil
	}
}

func testAccCheckKeycloakSamlIdentityProviderDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_saml_identity_provider" {
				continue
			}

			id := rs.Primary.ID
			realm := rs.Primary.Attributes["realm"]

			saml, _ := keycloakClient.GetIdentityProvider(testCtx, realm, id)
			if saml != nil {
				return fmt.Errorf("saml config with id %s still exists", id)
			}
		}

		return nil
	}
}

func getKeycloakSamlIdentityProviderFromState(s *terraform.State, resourceName string) (*keycloak.IdentityProvider, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	realm := rs.Primary.Attributes["realm"]
	alias := rs.Primary.Attributes["alias"]

	saml, err := keycloakClient.GetIdentityProvider(testCtx, realm, alias)
	if err != nil {
		return nil, fmt.Errorf("error getting saml identity provider config with alias %s: %s", alias, err)
	}

	return saml, nil
}

func testKeycloakSamlIdentityProvider_basic(saml string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_saml_identity_provider" "saml" {
	realm             			= data.keycloak_realm.realm.id
	alias             			= "%s"
	entity_id					= "https://example.com/entity_id"
	single_sign_on_service_url  = "https://example.com/auth"
}
	`, testAccRealm.Realm, saml)
}

func testKeycloakSamlIdentityProvider_customProviderId(saml, providerId string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_saml_identity_provider" "saml" {
	realm             			= data.keycloak_realm.realm.id
	alias             			= "%s"
	provider_id       			= "%s"
	entity_id					= "https://example.com/entity_id"
	single_sign_on_service_url  = "https://example.com/auth"
}
	`, testAccRealm.Realm, saml, providerId)
}

func testKeycloakSamlIdentityProvider_withNameIdPolicyFormat(saml, nameIdPolicyFormat string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_saml_identity_provider" "saml" {
	realm             			= data.keycloak_realm.realm.id
	alias             			= "%s"
	name_id_policy_format		= "%s"
	principal_type				= "ATTRIBUTE"
	entity_id					= "https://example.com/entity_id"
	single_sign_on_service_url  = "https://example.com/auth"
}
	`, testAccRealm.Realm, saml, nameIdPolicyFormat)
}

func testKeycloakSamlIdentityProvider_extra_config(alias, configKey, configValue string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_saml_identity_provider" "saml" {
	realm             			= data.keycloak_realm.realm.id
	alias             			= "%s"
	entity_id					= "https://example.com/entity_id"
	single_sign_on_service_url  = "https://example.com/auth"
	extra_config                = {
		%s = "%s"
	}
	name_id_policy_format       = "Email"
}
	`, testAccRealm.Realm, alias, configKey, configValue)
}

func testKeycloakSamlIdentityProvider_basicFromInterface(saml *keycloak.IdentityProvider) string {
	var authnContextClassRefs []string
	for _, v := range saml.Config.AuthnContextClassRefs {
		authnContextClassRefs = append(authnContextClassRefs, v)
	}

	var authnContextDeclRefs []string
	for _, v := range saml.Config.AuthnContextDeclRefs {
		authnContextDeclRefs = append(authnContextDeclRefs, v)
	}

	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_saml_identity_provider" "saml" {
	realm             			= data.keycloak_realm.realm.id
	alias             			= "%s"
	enabled           			= %t
	entity_id					= "%s"
	single_sign_on_service_url = "%s"
	backchannel_supported      = %t
	validate_signature         = %t
	hide_on_login_page         = %t
	name_id_policy_format      = "%s"
	single_logout_service_url  = "%s"
	signing_certificate        = "%s"
	signature_algorithm        = "%s"
	xml_sign_key_info_key_name_transformer = "%s"
	post_binding_authn_request = %t
	post_binding_response      = %t
	post_binding_logout        = %t
	force_authn                = %t
	want_assertions_signed     = %t
	want_assertions_encrypted  = %t
	gui_order                  = %s
	sync_mode                  = "%s"

	authn_context_class_refs      = %v
	authn_context_decl_refs       = %v
	authn_context_comparison_type = "%s"
}
	`, testAccRealm.Realm, saml.Alias, saml.Enabled, saml.Config.EntityId, saml.Config.SingleSignOnServiceUrl, bool(saml.Config.BackchannelSupported), bool(saml.Config.ValidateSignature), bool(saml.Config.HideOnLoginPage), saml.Config.NameIDPolicyFormat, saml.Config.SingleLogoutServiceUrl, saml.Config.SigningCertificate, saml.Config.SignatureAlgorithm, saml.Config.XmlSigKeyInfoKeyNameTransformer, bool(saml.Config.PostBindingAuthnRequest), bool(saml.Config.PostBindingResponse), bool(saml.Config.PostBindingLogout), bool(saml.Config.ForceAuthn), bool(saml.Config.WantAssertionsSigned), bool(saml.Config.WantAssertionsEncrypted), saml.Config.GuiOrder, saml.Config.SyncMode, arrayOfStringsForTerraformResource(authnContextClassRefs), arrayOfStringsForTerraformResource(authnContextDeclRefs), saml.Config.AuthnContextComparisonType)
}
