package provider

import (
	"dario.cat/mergo"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"github.com/keycloak/terraform-provider-keycloak/keycloak/types"
)

func resourceKeycloakOidcMicrosoftIdentityProvider() *schema.Resource {
	microsoftSchema := map[string]*schema.Schema{
		"alias": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "The alias uniquely identifies an identity provider and it is also used to build the redirect uri. In case of microsoft this is computed and always microsoft",
		},
		"display_name": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "The human-friendly name of the identity provider, used in the log in form.",
		},
		"provider_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "microsoft",
			Description: "Provider ID, is always microsoft.",
		},
		"client_id": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Client ID.",
		},
		"client_secret": {
			Type:        schema.TypeString,
			Required:    true,
			Sensitive:   true,
			Description: "Client Secret.",
		},
		"tenant_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "Tenant ID.",
		},
		"prompt": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "Prompt parameter for Microsoft identity provider. Available options are 'login', 'none', 'consent' and 'select_account'.",
		},
		"default_scopes": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "openid profile email",
			Description: "The scopes to be sent when asking for authorization. See the documentation for possible values, separator and default value'. Default: 'openid profile email'",
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
		"hide_on_login_page": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Hide On Login Page.",
		},
	}
	microsoftResource := resourceKeycloakIdentityProvider()
	microsoftResource.Schema = mergeSchemas(microsoftResource.Schema, microsoftSchema)
	microsoftResource.CreateContext = resourceKeycloakIdentityProviderCreate(getOidcMicrosoftIdentityProviderFromData, setOidcMicrosoftIdentityProviderData)
	microsoftResource.ReadContext = resourceKeycloakIdentityProviderRead(setOidcMicrosoftIdentityProviderData)
	microsoftResource.UpdateContext = resourceKeycloakIdentityProviderUpdate(getOidcMicrosoftIdentityProviderFromData, setOidcMicrosoftIdentityProviderData)
	return microsoftResource
}

func getOidcMicrosoftIdentityProviderFromData(data *schema.ResourceData, keycloakVersion *version.Version) (*keycloak.IdentityProvider, error) {
	idp, defaultConfig := getIdentityProviderFromData(data, keycloakVersion)
	idp.ProviderId = data.Get("provider_id").(string)

	aliasRaw, ok := data.GetOk("alias")
	if ok {
		idp.Alias = aliasRaw.(string)
	} else {
		idp.Alias = "microsoft"
	}

	microsoftIdentityProviderConfig := &keycloak.IdentityProviderConfig{
		ClientId:     data.Get("client_id").(string),
		ClientSecret: data.Get("client_secret").(string),
		TenantId:     data.Get("tenant_id").(string),
		Prompt:       data.Get("prompt").(string),
		DefaultScope: data.Get("default_scopes").(string),
		AcceptsPromptNoneForwFrmClt: types.KeycloakBoolQuoted(data.Get("accepts_prompt_none_forward_from_client").(bool)),
		DisableUserInfo:             types.KeycloakBoolQuoted(data.Get("disable_user_info").(bool)),
	}

	if err := mergo.Merge(microsoftIdentityProviderConfig, defaultConfig); err != nil {
		return nil, err
	}

	idp.Config = microsoftIdentityProviderConfig

	return idp, nil
}

func setOidcMicrosoftIdentityProviderData(data *schema.ResourceData, identityProvider *keycloak.IdentityProvider, keycloakVersion *version.Version) error {
	setIdentityProviderData(data, identityProvider, keycloakVersion)

	data.Set("provider_id", identityProvider.ProviderId)
	data.Set("client_id", identityProvider.Config.ClientId)
	data.Set("tenant_id", identityProvider.Config.TenantId)
	data.Set("prompt", identityProvider.Config.Prompt)
	data.Set("default_scopes", identityProvider.Config.DefaultScope)
	data.Set("accepts_prompt_none_forward_from_client", identityProvider.Config.AcceptsPromptNoneForwFrmClt)
	data.Set("disable_user_info", identityProvider.Config.DisableUserInfo)

	return nil
}
