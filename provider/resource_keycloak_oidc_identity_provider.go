package provider

import (
	"dario.cat/mergo"
	"errors"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"github.com/keycloak/terraform-provider-keycloak/keycloak/types"
)

func resourceKeycloakOidcIdentityProvider() *schema.Resource {
	oidcSchema := map[string]*schema.Schema{
		"provider_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "oidc",
			Description: "provider id, is always oidc, unless you have a custom implementation",
		},
		"display_name": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "The human-friendly name of the identity provider, used in the log in form.",
		},
		"backchannel_supported": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "Does the external IDP support backchannel logout?",
		},
		"validate_signature": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Enable/disable signature validation of external IDP signatures.",
		},
		"authorization_url": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "OIDC authorization URL.",
		},
		"client_id": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Client ID.",
		},
		"client_secret": {
			Type:          schema.TypeString,
			Optional:      true,
			Sensitive:     true,
			Description:   "Client Secret.",
			ConflictsWith: []string{"client_secret_wo", "client_secret_wo_version"},
		},
		"client_secret_wo": {
			Type:          schema.TypeString,
			Optional:      true,
			Sensitive:     true,
			WriteOnly:     true,
			ConflictsWith: []string{"client_secret"},
			RequiredWith:  []string{"client_secret_wo_version"},
			Description:   "Client Secret as write-only argument",
		},
		"client_secret_wo_version": {
			Type:          schema.TypeInt,
			Optional:      true,
			ConflictsWith: []string{"client_secret"},
			RequiredWith:  []string{"client_secret_wo"},
			Description:   "Version of the Client secret write-only argument",
		},
		"user_info_url": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "User Info URL",
		},
		"jwks_url": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "JSON Web Key Set URL",
		},
		"hide_on_login_page": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Hide On Login Page.",
		},
		"token_url": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Token URL.",
		},
		"logout_url": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Logout URL",
		},
		"login_hint": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Login Hint.",
		},
		"ui_locales": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Pass current locale to identity provider",
		},
		"default_scopes": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "openid",
			Description: "The scopes to be sent when asking for authorization. It can be a space-separated list of scopes. Defaults to 'openid'.",
		},
		"accepts_prompt_none_forward_from_client": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "This is just used together with Identity Provider Authenticator or when kc_idp_hint points to this identity provider. In case that client sends a request with prompt=none and user is not yet authenticated, the error will not be directly returned to client, but the request with prompt=none will be forwarded to this identity provider.",
		},
		"disable_user_info": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Disable usage of User Info service to obtain additional user information?  Default is to use this OIDC service.",
		},
		"issuer": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The issuer identifier for the issuer of the response. If not provided, no validation will be performed.",
		},
		"disable_type_claim_check": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Disables the validation of the `typ` claim of tokens received from the Identity Provider. If this is `off` the type claim is validated (default).",
		},
	}
	oidcResource := resourceKeycloakIdentityProvider()
	oidcResource.Schema = mergeSchemas(oidcResource.Schema, oidcSchema)
	oidcResource.CreateContext = resourceKeycloakIdentityProviderCreate(getOidcIdentityProviderFromData, setOidcIdentityProviderData)
	oidcResource.ReadContext = resourceKeycloakIdentityProviderRead(setOidcIdentityProviderData)
	oidcResource.UpdateContext = resourceKeycloakIdentityProviderUpdate(getOidcIdentityProviderFromData, setOidcIdentityProviderData)
	oidcResource.ValidateRawResourceConfigFuncs = []schema.ValidateRawResourceConfigFunc{
		// validate that argument is required if none of the checkExists attributes exist
		requiredWithoutAll(cty.GetAttrPath("client_secret"), []cty.Path{cty.GetAttrPath("client_secret_wo"), cty.GetAttrPath("client_secret_wo_version")}),
	}
	return oidcResource
}

func getOidcIdentityProviderFromData(data *schema.ResourceData, keycloakVersion *version.Version) (*keycloak.IdentityProvider, error) {
	rec, defaultConfig := getIdentityProviderFromData(data, keycloakVersion)
	rec.ProviderId = data.Get("provider_id").(string)
	_, useJwksUrl := data.GetOk("jwks_url")

	oidcIdentityProviderConfig := &keycloak.IdentityProviderConfig{
		BackchannelSupported:        types.KeycloakBoolQuoted(data.Get("backchannel_supported").(bool)),
		ValidateSignature:           types.KeycloakBoolQuoted(data.Get("validate_signature").(bool)),
		AuthorizationUrl:            data.Get("authorization_url").(string),
		ClientId:                    data.Get("client_id").(string),
		ClientSecret:                data.Get("client_secret").(string),
		TokenUrl:                    data.Get("token_url").(string),
		LogoutUrl:                   data.Get("logout_url").(string),
		UILocales:                   types.KeycloakBoolQuoted(data.Get("ui_locales").(bool)),
		LoginHint:                   data.Get("login_hint").(string),
		JwksUrl:                     data.Get("jwks_url").(string),
		UserInfoUrl:                 data.Get("user_info_url").(string),
		UseJwksUrl:                  types.KeycloakBoolQuoted(useJwksUrl),
		DisableUserInfo:             types.KeycloakBoolQuoted(data.Get("disable_user_info").(bool)),
		DefaultScope:                data.Get("default_scopes").(string),
		AcceptsPromptNoneForwFrmClt: types.KeycloakBoolQuoted(data.Get("accepts_prompt_none_forward_from_client").(bool)),
		Issuer:                      data.Get("issuer").(string),
		DisableTypeClaimCheck:       types.KeycloakBoolQuoted(data.Get("disable_type_claim_check").(bool)),

		//since keycloak v26 moved to IdentityProvider - still here fore backward compatibility
		HideOnLoginPage: types.KeycloakBoolQuoted(data.Get("hide_on_login_page").(bool)),
	}

	if data.Get("client_secret_wo_version").(int) != 0 && data.HasChange("client_secret_wo_version") {
		clientSecretWriteOnly, clientSecretWriteOnlyDiags := data.GetRawConfigAt(cty.GetAttrPath("client_secret_wo"))
		if clientSecretWriteOnlyDiags.HasError() {
			return nil, errors.New("error reading 'client_secret_wo' argument")
		}

		oidcIdentityProviderConfig.ClientSecret = clientSecretWriteOnly.AsString()
	}

	if err := mergo.Merge(oidcIdentityProviderConfig, defaultConfig); err != nil {
		return nil, err
	}

	rec.Config = oidcIdentityProviderConfig

	return rec, nil
}

func setOidcIdentityProviderData(data *schema.ResourceData, identityProvider *keycloak.IdentityProvider, keycloakVersion *version.Version) error {
	setIdentityProviderData(data, identityProvider, keycloakVersion)
	data.Set("backchannel_supported", identityProvider.Config.BackchannelSupported)
	data.Set("jwks_url", identityProvider.Config.JwksUrl)
	data.Set("logout_url", identityProvider.Config.LogoutUrl)
	data.Set("validate_signature", identityProvider.Config.ValidateSignature)
	data.Set("authorization_url", identityProvider.Config.AuthorizationUrl)
	data.Set("client_id", identityProvider.Config.ClientId)
	data.Set("disable_user_info", identityProvider.Config.DisableUserInfo)
	data.Set("user_info_url", identityProvider.Config.UserInfoUrl)
	data.Set("token_url", identityProvider.Config.TokenUrl)
	data.Set("login_hint", identityProvider.Config.LoginHint)
	data.Set("ui_locales", identityProvider.Config.UILocales)
	data.Set("issuer", identityProvider.Config.Issuer)
	data.Set("disable_type_claim_check", identityProvider.Config.DisableTypeClaimCheck)

	if v, ok := data.GetOk("client_secret_wo_version"); ok && v != nil {
		data.Set("client_secret_wo_version", v.(int))
	}
	if keycloakVersion.LessThan(keycloak.Version_26.AsVersion()) {
		// Since keycloak v26 the attribute "hideOnLoginPage" is not part of the identity provider config anymore!
		data.Set("hide_on_login_page", identityProvider.Config.HideOnLoginPage)
		return nil
	}
	return nil
}
