package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/meta"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func KeycloakProvider(client *keycloak.KeycloakClient) *schema.Provider {
	provider := &schema.Provider{
		DataSourcesMap: map[string]*schema.Resource{
			"keycloak_group":                              dataSourceKeycloakGroup(),
			"keycloak_openid_client":                      dataSourceKeycloakOpenidClient(),
			"keycloak_openid_client_authorization_policy": dataSourceKeycloakOpenidClientAuthorizationPolicy(),
			"keycloak_openid_client_scope":                dataSourceKeycloakOpenidClientScope(),
			"keycloak_openid_client_service_account_user": dataSourceKeycloakOpenidClientServiceAccountUser(),
			"keycloak_realm":                              dataSourceKeycloakRealm(),
			"keycloak_realm_keys":                         dataSourceKeycloakRealmKeys(),
			"keycloak_role":                               dataSourceKeycloakRole(),
			"keycloak_user":                               dataSourceKeycloakUser(),
			"keycloak_user_realm_roles":                   dataSourceKeycloakUserRealmRoles(),
			"keycloak_saml_client_installation_provider":  dataSourceKeycloakSamlClientInstallationProvider(),
			"keycloak_saml_client":                        dataSourceKeycloakSamlClient(),
			"keycloak_authentication_execution":           dataSourceKeycloakAuthenticationExecution(),
			"keycloak_authentication_flow":                dataSourceKeycloakAuthenticationFlow(),
			"keycloak_client_description_converter":       dataSourceKeycloakClientDescriptionConverter(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"keycloak_realm":                                             resourceKeycloakRealm(),
			"keycloak_realm_events":                                      resourceKeycloakRealmEvents(),
			"keycloak_realm_keystore_aes_generated":                      resourceKeycloakRealmKeystoreAesGenerated(),
			"keycloak_realm_keystore_ecdsa_generated":                    resourceKeycloakRealmKeystoreEcdsaGenerated(),
			"keycloak_realm_keystore_hmac_generated":                     resourceKeycloakRealmKeystoreHmacGenerated(),
			"keycloak_realm_keystore_java_keystore":                      resourceKeycloakRealmKeystoreJavaKeystore(),
			"keycloak_realm_keystore_rsa":                                resourceKeycloakRealmKeystoreRsa(),
			"keycloak_realm_keystore_rsa_generated":                      resourceKeycloakRealmKeystoreRsaGenerated(),
			"keycloak_realm_user_profile":                                resourceKeycloakRealmUserProfile(),
			"keycloak_required_action":                                   resourceKeycloakRequiredAction(),
			"keycloak_group":                                             resourceKeycloakGroup(),
			"keycloak_group_memberships":                                 resourceKeycloakGroupMemberships(),
			"keycloak_default_groups":                                    resourceKeycloakDefaultGroups(),
			"keycloak_default_roles":                                     resourceKeycloakDefaultRoles(),
			"keycloak_group_roles":                                       resourceKeycloakGroupRoles(),
			"keycloak_user":                                              resourceKeycloakUser(),
			"keycloak_user_roles":                                        resourceKeycloakUserRoles(),
			"keycloak_openid_client":                                     resourceKeycloakOpenidClient(),
			"keycloak_openid_client_scope":                               resourceKeycloakOpenidClientScope(),
			"keycloak_openid_default_client_scope":                       resourceKeycloakOpenidDefaultClientScope(),
			"keycloak_openid_optional_client_scope":                      resourceKeycloakOpenidOptionalClientScope(),
			"keycloak_ldap_user_federation":                              resourceKeycloakLdapUserFederation(),
			"keycloak_ldap_user_attribute_mapper":                        resourceKeycloakLdapUserAttributeMapper(),
			"keycloak_ldap_group_mapper":                                 resourceKeycloakLdapGroupMapper(),
			"keycloak_ldap_role_mapper":                                  resourceKeycloakLdapRoleMapper(),
			"keycloak_ldap_hardcoded_role_mapper":                        resourceKeycloakLdapHardcodedRoleMapper(),
			"keycloak_ldap_hardcoded_attribute_mapper":                   resourceKeycloakLdapHardcodedAttributeMapper(),
			"keycloak_ldap_hardcoded_group_mapper":                       resourceKeycloakLdapHardcodedGroupMapper(),
			"keycloak_ldap_msad_user_account_control_mapper":             resourceKeycloakLdapMsadUserAccountControlMapper(),
			"keycloak_ldap_msad_lds_user_account_control_mapper":         resourceKeycloakLdapMsadLdsUserAccountControlMapper(),
			"keycloak_ldap_full_name_mapper":                             resourceKeycloakLdapFullNameMapper(),
			"keycloak_ldap_custom_mapper":                                resourceKeycloakLdapCustomMapper(),
			"keycloak_custom_user_federation":                            resourceKeycloakCustomUserFederation(),
			"keycloak_openid_user_attribute_protocol_mapper":             resourceKeycloakOpenIdUserAttributeProtocolMapper(),
			"keycloak_openid_user_property_protocol_mapper":              resourceKeycloakOpenIdUserPropertyProtocolMapper(),
			"keycloak_openid_group_membership_protocol_mapper":           resourceKeycloakOpenIdGroupMembershipProtocolMapper(),
			"keycloak_openid_full_name_protocol_mapper":                  resourceKeycloakOpenIdFullNameProtocolMapper(),
			"keycloak_openid_hardcoded_claim_protocol_mapper":            resourceKeycloakOpenIdHardcodedClaimProtocolMapper(),
			"keycloak_openid_audience_protocol_mapper":                   resourceKeycloakOpenIdAudienceProtocolMapper(),
			"keycloak_openid_audience_resolve_protocol_mapper":           resourceKeycloakOpenIdAudienceResolveProtocolMapper(),
			"keycloak_openid_hardcoded_role_protocol_mapper":             resourceKeycloakOpenIdHardcodedRoleProtocolMapper(),
			"keycloak_openid_user_realm_role_protocol_mapper":            resourceKeycloakOpenIdUserRealmRoleProtocolMapper(),
			"keycloak_openid_user_client_role_protocol_mapper":           resourceKeycloakOpenIdUserClientRoleProtocolMapper(),
			"keycloak_openid_user_session_note_protocol_mapper":          resourceKeycloakOpenIdUserSessionNoteProtocolMapper(),
			"keycloak_openid_script_protocol_mapper":                     resourceKeycloakOpenIdScriptProtocolMapper(),
			"keycloak_openid_client_default_scopes":                      resourceKeycloakOpenidClientDefaultScopes(),
			"keycloak_openid_client_optional_scopes":                     resourceKeycloakOpenidClientOptionalScopes(),
			"keycloak_saml_client":                                       resourceKeycloakSamlClient(),
			"keycloak_saml_client_scope":                                 resourceKeycloakSamlClientScope(),
			"keycloak_saml_client_default_scopes":                        resourceKeycloakSamlClientDefaultScopes(),
			"keycloak_generic_client_protocol_mapper":                    resourceKeycloakGenericClientProtocolMapper(),
			"keycloak_generic_client_role_mapper":                        resourceKeycloakGenericClientRoleMapper(),
			"keycloak_generic_protocol_mapper":                           resourceKeycloakGenericProtocolMapper(),
			"keycloak_generic_role_mapper":                               resourceKeycloakGenericRoleMapper(),
			"keycloak_saml_user_attribute_protocol_mapper":               resourceKeycloakSamlUserAttributeProtocolMapper(),
			"keycloak_saml_user_property_protocol_mapper":                resourceKeycloakSamlUserPropertyProtocolMapper(),
			"keycloak_saml_script_protocol_mapper":                       resourceKeycloakSamlScriptProtocolMapper(),
			"keycloak_hardcoded_attribute_identity_provider_mapper":      resourceKeycloakHardcodedAttributeIdentityProviderMapper(),
			"keycloak_hardcoded_role_identity_provider_mapper":           resourceKeycloakHardcodedRoleIdentityProviderMapper(),
			"keycloak_attribute_importer_identity_provider_mapper":       resourceKeycloakAttributeImporterIdentityProviderMapper(),
			"keycloak_attribute_to_role_identity_provider_mapper":        resourceKeycloakAttributeToRoleIdentityProviderMapper(),
			"keycloak_user_template_importer_identity_provider_mapper":   resourceKeycloakUserTemplateImporterIdentityProviderMapper(),
			"keycloak_custom_identity_provider_mapper":                   resourceKeycloakCustomIdentityProviderMapper(),
			"keycloak_saml_identity_provider":                            resourceKeycloakSamlIdentityProvider(),
			"keycloak_oidc_google_identity_provider":                     resourceKeycloakOidcGoogleIdentityProvider(),
			"keycloak_oidc_identity_provider":                            resourceKeycloakOidcIdentityProvider(),
			"keycloak_openid_client_authorization_resource":              resourceKeycloakOpenidClientAuthorizationResource(),
			"keycloak_openid_client_group_policy":                        resourceKeycloakOpenidClientAuthorizationGroupPolicy(),
			"keycloak_openid_client_role_policy":                         resourceKeycloakOpenidClientAuthorizationRolePolicy(),
			"keycloak_openid_client_aggregate_policy":                    resourceKeycloakOpenidClientAuthorizationAggregatePolicy(),
			"keycloak_openid_client_js_policy":                           resourceKeycloakOpenidClientAuthorizationJSPolicy(),
			"keycloak_openid_client_time_policy":                         resourceKeycloakOpenidClientAuthorizationTimePolicy(),
			"keycloak_openid_client_user_policy":                         resourceKeycloakOpenidClientAuthorizationUserPolicy(),
			"keycloak_openid_client_client_policy":                       resourceKeycloakOpenidClientAuthorizationClientPolicy(),
			"keycloak_openid_client_authorization_scope":                 resourceKeycloakOpenidClientAuthorizationScope(),
			"keycloak_openid_client_authorization_permission":            resourceKeycloakOpenidClientAuthorizationPermission(),
			"keycloak_openid_client_service_account_role":                resourceKeycloakOpenidClientServiceAccountRole(),
			"keycloak_openid_client_service_account_realm_role":          resourceKeycloakOpenidClientServiceAccountRealmRole(),
			"keycloak_role":                                              resourceKeycloakRole(),
			"keycloak_authentication_flow":                               resourceKeycloakAuthenticationFlow(),
			"keycloak_authentication_subflow":                            resourceKeycloakAuthenticationSubFlow(),
			"keycloak_authentication_execution":                          resourceKeycloakAuthenticationExecution(),
			"keycloak_authentication_execution_config":                   resourceKeycloakAuthenticationExecutionConfig(),
			"keycloak_identity_provider_token_exchange_scope_permission": resourceKeycloakIdentityProviderTokenExchangeScopePermission(),
			"keycloak_openid_client_permissions":                         resourceKeycloakOpenidClientPermissions(),
			"keycloak_users_permissions":                                 resourceKeycloakUsersPermissions(),
			"keycloak_user_groups":                                       resourceKeycloakUserGroups(),
			"keycloak_group_permissions":                                 resourceKeycloakGroupPermissions(),
			"keycloak_authentication_bindings":                           resourceKeycloakAuthenticationBindings(),
		},
		Schema: map[string]*schema.Schema{
			"client_id": {
				Required:    true,
				Type:        schema.TypeString,
				DefaultFunc: schema.EnvDefaultFunc("KEYCLOAK_CLIENT_ID", nil),
			},
			"client_secret": {
				Optional:    true,
				Type:        schema.TypeString,
				DefaultFunc: schema.EnvDefaultFunc("KEYCLOAK_CLIENT_SECRET", nil),
			},
			"username": {
				Optional:    true,
				Type:        schema.TypeString,
				DefaultFunc: schema.EnvDefaultFunc("KEYCLOAK_USER", nil),
			},
			"password": {
				Optional:    true,
				Type:        schema.TypeString,
				DefaultFunc: schema.EnvDefaultFunc("KEYCLOAK_PASSWORD", nil),
			},
			"realm": {
				Optional:    true,
				Type:        schema.TypeString,
				DefaultFunc: schema.EnvDefaultFunc("KEYCLOAK_REALM", "master"),
			},
			"url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The base URL of the Keycloak instance, before `/auth`",
				DefaultFunc: schema.EnvDefaultFunc("KEYCLOAK_URL", nil),
			},
			"initial_login": {
				Optional:    true,
				Type:        schema.TypeBool,
				Description: "Whether or not to login to Keycloak instance on provider initialization",
				Default:     true,
			},
			"client_timeout": {
				Optional:    true,
				Type:        schema.TypeInt,
				Description: "Timeout (in seconds) of the Keycloak client",
				DefaultFunc: schema.EnvDefaultFunc("KEYCLOAK_CLIENT_TIMEOUT", 15),
			},
			"root_ca_certificate": {
				Optional:    true,
				Type:        schema.TypeString,
				Description: "Allows x509 calls using an unknown CA certificate (for development purposes)",
				Default:     "",
			},
			"tls_insecure_skip_verify": {
				Optional:    true,
				Type:        schema.TypeBool,
				Description: "Allows ignoring insecure certificates when set to true. Defaults to false. Disabling security check is dangerous and should be avoided.",
				Default:     false,
			},
			"red_hat_sso": {
				Optional:    true,
				Type:        schema.TypeBool,
				Description: "When true, the provider will treat the Keycloak instance as a Red Hat SSO server, specifically when parsing the version returned from the /serverinfo API endpoint.",
				Default:     false,
			},
			"base_path": {
				Optional:    true,
				Type:        schema.TypeString,
				DefaultFunc: schema.EnvDefaultFunc("KEYCLOAK_BASE_PATH", ""),
			},
			"additional_headers": {
				Optional: true,
				Type:     schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}

	provider.ConfigureContextFunc = func(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
		if client != nil {
			return client, nil
		}

		url := data.Get("url").(string)
		basePath := data.Get("base_path").(string)
		clientId := data.Get("client_id").(string)
		clientSecret := data.Get("client_secret").(string)
		username := data.Get("username").(string)
		password := data.Get("password").(string)
		realm := data.Get("realm").(string)
		initialLogin := data.Get("initial_login").(bool)
		clientTimeout := data.Get("client_timeout").(int)
		tlsInsecureSkipVerify := data.Get("tls_insecure_skip_verify").(bool)
		rootCaCertificate := data.Get("root_ca_certificate").(string)
		redHatSSO := data.Get("red_hat_sso").(bool)
		additionalHeaders := make(map[string]string)
		for k, v := range data.Get("additional_headers").(map[string]interface{}) {
			additionalHeaders[k] = v.(string)
		}

		var diags diag.Diagnostics

		userAgent := fmt.Sprintf("HashiCorp Terraform/%s (+https://www.terraform.io) Terraform Plugin SDK/%s", provider.TerraformVersion, meta.SDKVersionString())

		keycloakClient, err := keycloak.NewKeycloakClient(ctx, url, basePath, clientId, clientSecret, realm, username, password, initialLogin, clientTimeout, rootCaCertificate, tlsInsecureSkipVerify, userAgent, redHatSSO, additionalHeaders)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "error initializing keycloak provider",
				Detail:   err.Error(),
			})
		}

		return keycloakClient, diags
	}

	return provider
}
