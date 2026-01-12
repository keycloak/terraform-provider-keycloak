package provider

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/keycloak/terraform-provider-keycloak/keycloak/types"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

var (
	keycloakSamlClientNameIdFormats                    = []string{"username", "email", "transient", "persistent"}
	keycloakSamlClientSignatureAlgorithms              = []string{"RSA_SHA1", "RSA_SHA256", "RSA_SHA256_MGF1", "RSA_SHA512", "RSA_SHA512_MGF1", "DSA_SHA1"}
	keycloakSamlClientEncryptionAlgorithmFriendlyToURI = map[string]string{
		"AES_256_GCM": "http://www.w3.org/2009/xmlenc11#aes256-gcm",
		"AES_192_GCM": "http://www.w3.org/2009/xmlenc11#aes192-gcm",
		"AES_128_GCM": "http://www.w3.org/2009/xmlenc11#aes128-gcm",
		"AES_256_CBC": "http://www.w3.org/2001/04/xmlenc#aes256-cbc",
		"AES_192_CBC": "http://www.w3.org/2001/04/xmlenc#aes192-cbc",
		"AES_128_CBC": "http://www.w3.org/2001/04/xmlenc#aes128-cbc",
	}
	keycloakSamlClientEncryptionAlgorithmURIToFriendly    = reverseStringMap(keycloakSamlClientEncryptionAlgorithmFriendlyToURI)
	keycloakSamlClientEncryptionKeyAlgorithmFriendlyToURI = map[string]string{
		"RSA-OAEP-11":    "http://www.w3.org/2009/xmlenc11#rsa-oaep",
		"RSA-OAEP-MGF1P": "http://www.w3.org/2001/04/xmlenc#rsa-oaep-mgf1p",
		"RSA1_5":         "http://www.w3.org/2001/04/xmlenc#rsa-1_5",
	}
	keycloakSamlClientEncryptionKeyAlgorithmURIToFriendly = reverseStringMap(keycloakSamlClientEncryptionKeyAlgorithmFriendlyToURI)
	keycloakSamlClientEncryptionDigestMethodFriendlyToURI = map[string]string{
		"SHA-512": "http://www.w3.org/2001/04/xmlenc#sha512",
		"SHA-256": "http://www.w3.org/2001/04/xmlenc#sha256",
		"SHA-1":   "http://www.w3.org/2000/09/xmldsig#sha1",
	}
	keycloakSamlClientEncryptionDigestMethodURIToFriendly           = reverseStringMap(keycloakSamlClientEncryptionDigestMethodFriendlyToURI)
	keycloakSamlClientEncryptionMaskGenerationFunctionFriendlyToURI = map[string]string{
		"mgf1sha1":   "http://www.w3.org/2009/xmlenc11#mgf1sha1",
		"mgf1sha224": "http://www.w3.org/2009/xmlenc11#mgf1sha224",
		"mgf1sha256": "http://www.w3.org/2009/xmlenc11#mgf1sha256",
		"mgf1sha384": "http://www.w3.org/2009/xmlenc11#mgf1sha384",
		"mgf1sha512": "http://www.w3.org/2009/xmlenc11#mgf1sha512",
	}
	keycloakSamlClientEncryptionMaskGenerationFunctionURIToFriendly = reverseStringMap(keycloakSamlClientEncryptionMaskGenerationFunctionFriendlyToURI)
	keycloakSamlClientSignatureKeyNames                             = []string{"NONE", "KEY_ID", "CERT_SUBJECT"}
	keycloakSamlClientCanonicalizationMethods                       = map[string]string{
		"EXCLUSIVE":               "http://www.w3.org/2001/10/xml-exc-c14n#",
		"EXCLUSIVE_WITH_COMMENTS": "http://www.w3.org/2001/10/xml-exc-c14n#WithComments",
		"INCLUSIVE":               "http://www.w3.org/TR/2001/REC-xml-c14n-20010315",
		"INCLUSIVE_WITH_COMMENTS": "http://www.w3.org/TR/2001/REC-xml-c14n-20010315#WithComments",
	}
)

func resourceKeycloakSamlClient() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakSamlClientCreate,
		ReadContext:   resourceKeycloakSamlClientRead,
		DeleteContext: resourceKeycloakSamlClientDelete,
		UpdateContext: resourceKeycloakSamlClientUpdate,
		// This resource can be imported using {{realm}}/{{client_id}}. The Client ID is displayed in the GUI
		Importer: &schema.ResourceImporter{
			StateContext: resourceKeycloakSamlClientImport,
		},
		Schema: map[string]*schema.Schema{
			"client_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressDiffWhenNotInConfig("name"),
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"description": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressDiffWhenNotInConfig("description"),
			},
			"include_authn_statement": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"sign_documents": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"sign_assertions": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"encrypt_assertions": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"encryption_algorithm": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateSamlEncryptionAlgorithm,
			},
			"encryption_key_algorithm": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateSamlEncryptionKeyAlgorithm,
			},
			"encryption_digest_method": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateSamlEncryptionDigestMethod,
			},
			"encryption_mask_generation_function": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateSamlEncryptionMaskGenerationFunction,
			},
			"client_signature_required": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"force_post_binding": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"consent_required": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"front_channel_logout": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"force_name_id_format": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"signature_algorithm": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(keycloakSamlClientSignatureAlgorithms, false),
			},
			"signature_key_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "KEY_ID",
				ValidateFunc: validation.StringInSlice(keycloakSamlClientSignatureKeyNames, false),
			},
			"canonicalization_method": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "EXCLUSIVE",
				ValidateFunc: validation.StringInSlice(keys(keycloakSamlClientCanonicalizationMethods), false),
			},
			"name_id_format": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(keycloakSamlClientNameIdFormats, false),
			},
			"root_url": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"valid_redirect_uris": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Optional: true,
			},
			"base_url": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"login_theme": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressDiffWhenNotInConfig("login_theme"),
			},
			"master_saml_processing_url": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"encryption_certificate": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				DiffSuppressFunc: func(_, old, new string, _ *schema.ResourceData) bool {
					return old == formatCertificate(new)
				},
			},
			"signing_certificate": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				DiffSuppressFunc: func(_, old, new string, _ *schema.ResourceData) bool {
					return old == formatCertificate(new)
				},
			},
			"signing_private_key": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				DiffSuppressFunc: func(_, old, new string, _ *schema.ResourceData) bool {
					return old == formatSigningPrivateKey(new)
				},
			},
			"encryption_certificate_sha1": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"signing_certificate_sha1": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"signing_private_key_sha1": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"idp_initiated_sso_url_name": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressDiffWhenNotInConfig("idp_initiated_sso_url_name"),
			},
			"idp_initiated_sso_relay_state": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressDiffWhenNotInConfig("idp_initiated_sso_relay_state"),
			},
			"assertion_consumer_post_url": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressDiffWhenNotInConfig("assertion_consumer_post_url"),
			},
			"assertion_consumer_redirect_url": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressDiffWhenNotInConfig("assertion_consumer_redirect_url"),
			},
			"logout_service_post_binding_url": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressDiffWhenNotInConfig("logout_service_post_binding_url"),
			},
			"logout_service_redirect_binding_url": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressDiffWhenNotInConfig("logout_service_redirect_binding_url"),
			},
			"full_scope_allowed": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"authentication_flow_binding_overrides": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"browser_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"direct_grant_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"extra_config": {
				Type:             schema.TypeMap,
				Optional:         true,
				ValidateDiagFunc: validateExtraConfig(reflect.ValueOf(&keycloak.SamlClientAttributes{}).Elem()),
			},
			"always_display_in_console": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
		CustomizeDiff: validateKeycloakSamlClientEncryptionSettings(),
	}
}

func formatCertificate(signingCertificate string) string {
	r := strings.NewReplacer(
		"-----BEGIN CERTIFICATE-----", "",
		"-----END CERTIFICATE-----", "",
		"\n", "",
	)

	return r.Replace(signingCertificate)
}

func formatSigningPrivateKey(signingPrivateKey string) string {
	r := strings.NewReplacer(
		"-----BEGIN PRIVATE KEY-----", "",
		"-----END PRIVATE KEY-----", "",
		"\n", "",
	)

	return r.Replace(signingPrivateKey)
}

func validateSamlEncryptionAlgorithm(v interface{}, _ string) (ws []string, es []error) {
	value := v.(string)
	if value == "" {
		return
	}

	if _, ok := keycloakSamlClientEncryptionAlgorithmFriendlyToURI[value]; ok {
		return
	}

	if _, ok := keycloakSamlClientEncryptionAlgorithmURIToFriendly[value]; ok {
		return
	}

	es = append(es, fmt.Errorf(`Invalid encryption_algorithm: value must be one of %s`, strings.Join(keys(keycloakSamlClientEncryptionAlgorithmFriendlyToURI), ", ")))
	return
}

func validateSamlEncryptionKeyAlgorithm(v interface{}, _ string) (ws []string, es []error) {
	value := v.(string)
	if value == "" {
		return
	}

	if _, ok := keycloakSamlClientEncryptionKeyAlgorithmFriendlyToURI[value]; ok {
		return
	}

	if _, ok := keycloakSamlClientEncryptionKeyAlgorithmURIToFriendly[value]; ok {
		return
	}

	es = append(es, fmt.Errorf(`Invalid encryption_key_algorithm: value must be one of %s`, strings.Join(keys(keycloakSamlClientEncryptionKeyAlgorithmFriendlyToURI), ", ")))
	return
}

func validateSamlEncryptionDigestMethod(v interface{}, _ string) (ws []string, es []error) {
	value := v.(string)
	if value == "" {
		return
	}

	if _, ok := keycloakSamlClientEncryptionDigestMethodFriendlyToURI[value]; ok {
		return
	}

	if _, ok := keycloakSamlClientEncryptionDigestMethodURIToFriendly[value]; ok {
		return
	}

	es = append(es, fmt.Errorf(`Invalid encryption_digest_method: value must be one of %s`, strings.Join(keys(keycloakSamlClientEncryptionDigestMethodFriendlyToURI), ", ")))
	return
}

func validateSamlEncryptionMaskGenerationFunction(v interface{}, _ string) (ws []string, es []error) {
	value := v.(string)
	if value == "" {
		return
	}

	if _, ok := keycloakSamlClientEncryptionMaskGenerationFunctionFriendlyToURI[value]; ok {
		return
	}

	normalized := normalizeSamlEncryptionMaskGenerationFunction(value)
	if _, ok := keycloakSamlClientEncryptionMaskGenerationFunctionFriendlyToURI[normalized]; ok {
		return
	}

	if _, ok := keycloakSamlClientEncryptionMaskGenerationFunctionURIToFriendly[value]; ok {
		return
	}

	es = append(es, fmt.Errorf(`Invalid encryption_mask_generation_function: value must be one of %s`, strings.Join(keys(keycloakSamlClientEncryptionMaskGenerationFunctionFriendlyToURI), ", ")))
	return
}

func convertSamlEncryptionAlgorithmToAPI(value string) string {
	if uri, ok := keycloakSamlClientEncryptionAlgorithmFriendlyToURI[value]; ok {
		return uri
	}

	return value
}

func convertSamlEncryptionKeyAlgorithmToAPI(value string) string {
	if uri, ok := keycloakSamlClientEncryptionKeyAlgorithmFriendlyToURI[value]; ok {
		return uri
	}

	return value
}

func convertSamlEncryptionDigestMethodToAPI(value string) string {
	if uri, ok := keycloakSamlClientEncryptionDigestMethodFriendlyToURI[value]; ok {
		return uri
	}

	return value
}

func normalizeSamlEncryptionMaskGenerationFunction(value string) string {
	normalized := strings.ReplaceAll(value, "_", "")
	return strings.ToLower(normalized)
}

func convertSamlEncryptionMaskGenerationFunctionToAPI(value string) string {
	if uri, ok := keycloakSamlClientEncryptionMaskGenerationFunctionFriendlyToURI[value]; ok {
		return uri
	}

	if uri, ok := keycloakSamlClientEncryptionMaskGenerationFunctionFriendlyToURI[normalizeSamlEncryptionMaskGenerationFunction(value)]; ok {
		return uri
	}

	return value
}

func convertSamlEncryptionAlgorithmToState(value string) string {
	if friendly, ok := keycloakSamlClientEncryptionAlgorithmURIToFriendly[value]; ok {
		return friendly
	}

	return value
}

func convertSamlEncryptionKeyAlgorithmToState(value string) string {
	if friendly, ok := keycloakSamlClientEncryptionKeyAlgorithmURIToFriendly[value]; ok {
		return friendly
	}

	return value
}

func convertSamlEncryptionDigestMethodToState(value string) string {
	if friendly, ok := keycloakSamlClientEncryptionDigestMethodURIToFriendly[value]; ok {
		return friendly
	}

	return value
}

func convertSamlEncryptionMaskGenerationFunctionToState(value string) string {
	if friendly, ok := keycloakSamlClientEncryptionMaskGenerationFunctionURIToFriendly[value]; ok {
		return friendly
	}

	return value
}

func validateKeycloakSamlClientEncryptionSettings() schema.CustomizeDiffFunc {
	return func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
		keyAlgorithmRaw, keyAlgorithmOk := d.GetOkExists("encryption_key_algorithm")
		digestMethodRaw, digestMethodOk := d.GetOkExists("encryption_digest_method")
		maskGenerationFunctionRaw, maskGenerationFunctionOk := d.GetOkExists("encryption_mask_generation_function")

		if !digestMethodOk && !maskGenerationFunctionOk {
			return nil
		}

		if !keyAlgorithmOk {
			return fmt.Errorf("encryption_key_algorithm must be set when encryption_digest_method or encryption_mask_generation_function is set")
		}

		keyAlgorithm := keyAlgorithmRaw.(string)
		digestMethod := ""
		if digestMethodOk {
			digestMethod = digestMethodRaw.(string)
		}
		maskGenerationFunction := ""
		if maskGenerationFunctionOk {
			maskGenerationFunction = maskGenerationFunctionRaw.(string)
		}

		return validateSamlClientEncryptionKeySettings(keyAlgorithm, digestMethod, maskGenerationFunction)
	}
}

func validateSamlClientEncryptionKeySettings(keyAlgorithm, digestMethod, maskGenerationFunction string) error {
	if digestMethod != "" && keyAlgorithm != "RSA-OAEP-MGF1P" && keyAlgorithm != "RSA-OAEP-11" {
		return fmt.Errorf("encryption_digest_method is only valid when encryption_key_algorithm is RSA-OAEP-11 or RSA-OAEP-MGF1P")
	}

	if maskGenerationFunction != "" && keyAlgorithm != "RSA-OAEP-11" {
		return fmt.Errorf("encryption_mask_generation_function is only valid when encryption_key_algorithm is RSA-OAEP-11")
	}

	return nil
}

func mapToSamlClientFromData(data *schema.ResourceData) *keycloak.SamlClient {
	var validRedirectUris []string

	if v, ok := data.GetOk("valid_redirect_uris"); ok {
		for _, validRedirectUri := range v.(*schema.Set).List() {
			validRedirectUris = append(validRedirectUris, validRedirectUri.(string))
		}
	}

	keyAlgorithm := ""
	if value, ok := data.GetOkExists("encryption_key_algorithm"); ok {
		keyAlgorithm = value.(string)
	} else if data.Id() == "" {
		keyAlgorithm = "RSA-OAEP-11"
	}

	digestMethod := ""
	if value, ok := data.GetOkExists("encryption_digest_method"); ok {
		digestMethod = value.(string)
	} else if data.Id() == "" && (keyAlgorithm == "RSA-OAEP-11" || keyAlgorithm == "RSA-OAEP-MGF1P") {
		digestMethod = "SHA-256"
	}

	maskGenerationFunction := ""
	if value, ok := data.GetOkExists("encryption_mask_generation_function"); ok {
		maskGenerationFunction = value.(string)
	} else if data.Id() == "" && keyAlgorithm == "RSA-OAEP-11" {
		maskGenerationFunction = "mgf1sha256"
	}

	// Use GetOkExists for string fields to preserve empty strings
	idpInitiatedSSOURLName, idpInitiatedSSOURLNameOk := data.GetOkExists("idp_initiated_sso_url_name")
	idpInitiatedSSORelayState, idpInitiatedSSORelayStateOk := data.GetOkExists("idp_initiated_sso_relay_state")
	assertionConsumerPostURL, assertionConsumerPostURLOk := data.GetOkExists("assertion_consumer_post_url")
	assertionConsumerRedirectURL, assertionConsumerRedirectURLOk := data.GetOkExists("assertion_consumer_redirect_url")
	logoutServicePostBindingURL, logoutServicePostBindingURLOk := data.GetOkExists("logout_service_post_binding_url")
	logoutServiceRedirectBindingURL, logoutServiceRedirectBindingURLOk := data.GetOkExists("logout_service_redirect_binding_url")
	loginTheme, loginThemeOk := data.GetOkExists("login_theme")

	samlAttributes := &keycloak.SamlClientAttributes{
		IncludeAuthnStatement:            types.KeycloakBoolQuoted(data.Get("include_authn_statement").(bool)),
		ForceNameIdFormat:                types.KeycloakBoolQuoted(data.Get("force_name_id_format").(bool)),
		SignDocuments:                    types.KeycloakBoolQuoted(data.Get("sign_documents").(bool)),
		SignAssertions:                   types.KeycloakBoolQuoted(data.Get("sign_assertions").(bool)),
		EncryptAssertions:                types.KeycloakBoolQuoted(data.Get("encrypt_assertions").(bool)),
		EncryptionAlgorithm:              convertSamlEncryptionAlgorithmToAPI(data.Get("encryption_algorithm").(string)),
		EncryptionKeyAlgorithm:           convertSamlEncryptionKeyAlgorithmToAPI(keyAlgorithm),
		EncryptionDigestMethod:           convertSamlEncryptionDigestMethodToAPI(digestMethod),
		EncryptionMaskGenerationFunction: convertSamlEncryptionMaskGenerationFunctionToAPI(maskGenerationFunction),
		ClientSignatureRequired:          types.KeycloakBoolQuoted(data.Get("client_signature_required").(bool)),
		ForcePostBinding:                 types.KeycloakBoolQuoted(data.Get("force_post_binding").(bool)),
		SignatureAlgorithm:               data.Get("signature_algorithm").(string),
		SignatureKeyName:                 data.Get("signature_key_name").(string),
		CanonicalizationMethod:           keycloakSamlClientCanonicalizationMethods[data.Get("canonicalization_method").(string)],
		NameIdFormat:                     data.Get("name_id_format").(string),
		ExtraConfig:                      getExtraConfigFromData(data),
	}

	// Set string fields only if explicitly provided, preserving empty strings
	if idpInitiatedSSOURLNameOk {
		samlAttributes.IDPInitiatedSSOURLName = idpInitiatedSSOURLName.(string)
	}
	if idpInitiatedSSORelayStateOk {
		samlAttributes.IDPInitiatedSSORelayState = idpInitiatedSSORelayState.(string)
	}
	if assertionConsumerPostURLOk {
		samlAttributes.AssertionConsumerPostURL = assertionConsumerPostURL.(string)
	}
	if assertionConsumerRedirectURLOk {
		samlAttributes.AssertionConsumerRedirectURL = assertionConsumerRedirectURL.(string)
	}
	if logoutServicePostBindingURLOk {
		samlAttributes.LogoutServicePostBindingURL = logoutServicePostBindingURL.(string)
	}
	if logoutServiceRedirectBindingURLOk {
		samlAttributes.LogoutServiceRedirectBindingURL = logoutServiceRedirectBindingURL.(string)
	}
	if loginThemeOk {
		samlAttributes.LoginTheme = loginTheme.(string)
	}

	if encryptionCertificate, ok := data.GetOk("encryption_certificate"); ok {
		samlAttributes.EncryptionCertificate = formatCertificate(encryptionCertificate.(string))
	}

	if signingCertificate, ok := data.GetOk("signing_certificate"); ok {
		samlAttributes.SigningCertificate = formatCertificate(signingCertificate.(string))
	}

	if signingPrivateKey, ok := data.GetOk("signing_private_key"); ok {
		samlAttributes.SigningPrivateKey = formatSigningPrivateKey(signingPrivateKey.(string))
	}

	// Use GetOkExists for client-level string fields to preserve empty strings
	name, nameOk := data.GetOkExists("name")
	description, descriptionOk := data.GetOkExists("description")
	rootUrl, rootUrlOk := data.GetOkExists("root_url")
	baseUrl, baseUrlOk := data.GetOkExists("base_url")
	masterSamlProcessingUrl, masterSamlProcessingUrlOk := data.GetOkExists("master_saml_processing_url")

	samlClient := &keycloak.SamlClient{
		Id:                     data.Id(),
		ClientId:               data.Get("client_id").(string),
		RealmId:                data.Get("realm_id").(string),
		Enabled:                data.Get("enabled").(bool),
		FrontChannelLogout:     data.Get("front_channel_logout").(bool),
		ValidRedirectUris:      validRedirectUris,
		FullScopeAllowed:       data.Get("full_scope_allowed").(bool),
		ConsentRequired:        data.Get("consent_required").(bool),
		AlwaysDisplayInConsole: data.Get("always_display_in_console").(bool),
		Attributes:             samlAttributes,
	}

	// Set client-level string fields only if explicitly provided, preserving empty strings
	if nameOk {
		samlClient.Name = name.(string)
	}
	if descriptionOk {
		samlClient.Description = description.(string)
	}
	if rootUrlOk {
		samlClient.RootUrl = rootUrl.(string)
	}
	if baseUrlOk {
		samlClient.BaseUrl = baseUrl.(string)
	}
	if masterSamlProcessingUrlOk {
		samlClient.MasterSamlProcessingUrl = masterSamlProcessingUrl.(string)
	}

	if v, ok := data.GetOk("authentication_flow_binding_overrides"); ok {
		authenticationFlowBindingOverridesData := v.(*schema.Set).List()[0]
		authenticationFlowBindingOverrides := authenticationFlowBindingOverridesData.(map[string]interface{})
		samlClient.AuthenticationFlowBindingOverrides = keycloak.SamlAuthenticationFlowBindingOverrides{
			BrowserId:     authenticationFlowBindingOverrides["browser_id"].(string),
			DirectGrantId: authenticationFlowBindingOverrides["direct_grant_id"].(string),
		}
	}

	return samlClient
}

func mapToDataFromSamlClient(ctx context.Context, data *schema.ResourceData, client *keycloak.SamlClient) error {
	data.SetId(client.Id)

	data.Set("include_authn_statement", client.Attributes.IncludeAuthnStatement)
	data.Set("force_name_id_format", client.Attributes.ForceNameIdFormat)
	data.Set("sign_documents", client.Attributes.SignDocuments)
	data.Set("sign_assertions", client.Attributes.SignAssertions)
	data.Set("encrypt_assertions", client.Attributes.EncryptAssertions)
	data.Set("encryption_algorithm", convertSamlEncryptionAlgorithmToState(client.Attributes.EncryptionAlgorithm))
	data.Set("encryption_key_algorithm", convertSamlEncryptionKeyAlgorithmToState(client.Attributes.EncryptionKeyAlgorithm))
	data.Set("encryption_digest_method", convertSamlEncryptionDigestMethodToState(client.Attributes.EncryptionDigestMethod))
	data.Set("encryption_mask_generation_function", convertSamlEncryptionMaskGenerationFunctionToState(client.Attributes.EncryptionMaskGenerationFunction))
	data.Set("client_signature_required", client.Attributes.ClientSignatureRequired)
	data.Set("force_post_binding", client.Attributes.ForcePostBinding)

	if (keycloak.SamlAuthenticationFlowBindingOverrides{}) == client.AuthenticationFlowBindingOverrides {
		data.Set("authentication_flow_binding_overrides", nil)
	} else {
		authenticationFlowBindingOverridesSettings := make(map[string]interface{})
		authenticationFlowBindingOverridesSettings["browser_id"] = client.AuthenticationFlowBindingOverrides.BrowserId
		authenticationFlowBindingOverridesSettings["direct_grant_id"] = client.AuthenticationFlowBindingOverrides.DirectGrantId
		data.Set("authentication_flow_binding_overrides", []interface{}{authenticationFlowBindingOverridesSettings})
	}

	data.Set("client_id", client.ClientId)
	data.Set("realm_id", client.RealmId)
	data.Set("name", client.Name)
	data.Set("enabled", client.Enabled)
	data.Set("description", client.Description)
	data.Set("front_channel_logout", client.FrontChannelLogout)
	data.Set("root_url", client.RootUrl)
	data.Set("valid_redirect_uris", client.ValidRedirectUris)
	data.Set("base_url", client.BaseUrl)
	data.Set("master_saml_processing_url", client.MasterSamlProcessingUrl)
	data.Set("signature_algorithm", client.Attributes.SignatureAlgorithm)
	data.Set("signature_key_name", client.Attributes.SignatureKeyName)
	data.Set("name_id_format", client.Attributes.NameIdFormat)
	data.Set("idp_initiated_sso_url_name", client.Attributes.IDPInitiatedSSOURLName)
	data.Set("idp_initiated_sso_relay_state", client.Attributes.IDPInitiatedSSORelayState)
	data.Set("assertion_consumer_post_url", client.Attributes.AssertionConsumerPostURL)
	data.Set("assertion_consumer_redirect_url", client.Attributes.AssertionConsumerRedirectURL)
	data.Set("logout_service_post_binding_url", client.Attributes.LogoutServicePostBindingURL)
	data.Set("logout_service_redirect_binding_url", client.Attributes.LogoutServiceRedirectBindingURL)
	data.Set("full_scope_allowed", client.FullScopeAllowed)
	data.Set("login_theme", client.Attributes.LoginTheme)
	data.Set("consent_required", client.ConsentRequired)
	data.Set("always_display_in_console", client.AlwaysDisplayInConsole)

	if canonicalizationMethod, ok := mapKeyFromValue(keycloakSamlClientCanonicalizationMethods, client.Attributes.CanonicalizationMethod); ok {
		data.Set("canonicalization_method", canonicalizationMethod)
	}

	setExtraConfigData(data, client.Attributes.ExtraConfig)

	data.Set("encryption_certificate", client.Attributes.EncryptionCertificate)
	data.Set("signing_certificate", client.Attributes.SigningCertificate)
	data.Set("signing_private_key", client.Attributes.SigningPrivateKey)
	resourceKeycloakSamlClientSetSha1(ctx, data, "encryption_certificate_sha1", client.Attributes.EncryptionCertificate)
	resourceKeycloakSamlClientSetSha1(ctx, data, "signing_certificate_sha1", client.Attributes.SigningCertificate)
	resourceKeycloakSamlClientSetSha1(ctx, data, "signing_private_key_sha1", client.Attributes.SigningPrivateKey)

	return nil
}

func resourceKeycloakSamlClientSetSha1(ctx context.Context, data *schema.ResourceData, attribute, value string) {
	if value != "" {
		bytes, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			tflog.Warn(ctx, "Cannot compute sha1sum", map[string]interface{}{
				"error":     err.Error(),
				"attribute": attribute,
			})
			data.Set(attribute, "")

			return
		}

		hash := sha1.New()
		hash.Write(bytes)

		data.Set(attribute, hex.EncodeToString(hash.Sum(nil)))
	} else {
		data.Set(attribute, "")
	}
}

func resourceKeycloakSamlClientCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	client := mapToSamlClientFromData(data)

	err := keycloakClient.NewSamlClient(ctx, client)
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(client.Id)

	return resourceKeycloakSamlClientRead(ctx, data, meta)
}

func resourceKeycloakSamlClientRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	client, err := keycloakClient.GetSamlClient(ctx, realmId, id)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	err = mapToDataFromSamlClient(ctx, data, client)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakSamlClientUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	client := mapToSamlClientFromData(data)

	err := keycloakClient.UpdateSamlClient(ctx, client)
	if err != nil {
		return diag.FromErr(err)
	}

	err = mapToDataFromSamlClient(ctx, data, client)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakSamlClientDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	return diag.FromErr(keycloakClient.DeleteSamlClient(ctx, realmId, id))
}

func resourceKeycloakSamlClientImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("Invalid import. Supported import formats: {{realmId}}/{{samlClientId}}")
	}

	_, err := keycloakClient.GetSamlClient(ctx, parts[0], parts[1])
	if err != nil {
		return nil, err
	}

	d.Set("realm_id", parts[0])
	d.SetId(parts[1])

	diagnostics := resourceKeycloakSamlClientRead(ctx, d, meta)
	if diagnostics.HasError() {
		return nil, errors.New(diagnostics[0].Summary)
	}

	return []*schema.ResourceData{d}, nil
}
