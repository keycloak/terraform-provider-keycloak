package provider

import (
	"dario.cat/mergo"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"github.com/keycloak/terraform-provider-keycloak/keycloak/types"
)

func resourceKeycloakOidcFacebookIdentityProvider() *schema.Resource {
	oidcFacebookSchema := map[string]*schema.Schema{
		"alias": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "The alias uniquely identifies an identity provider and it is also used to build the redirect uri.",
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
			Default:     "facebook",
			Description: "provider id, is always facebook, unless you have a extended custom implementation",
		},
		"client_id": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The client identifier registered with the Facebook identity provider.",
		},
		"client_secret": {
			Type:        schema.TypeString,
			Required:    true,
			Sensitive:   true,
			Description: "The client secret registered with the Facebook identity provider.",
		},
		"fetched_fields": { //fetchedFields
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Provide additional fields which would be fetched using the profile request. This will be appended to the default set of 'id,name,email,first_name,last_name'.",
		},
		"default_scopes": { //defaultScope
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "openid profile email",
			Description: "The scopes to be sent when asking for authorization. See the documentation for possible values, separator and default value'. Default: 'openid profile email'",
		},
		"accepts_prompt_none_forward_from_client": { // acceptsPromptNoneForwardFromClient
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "This is just used together with Identity Provider Authenticator or when kc_idp_hint points to this identity provider. In case that client sends a request with prompt=none and user is not yet authenticated, the error will not be directly returned to client, but the request with prompt=none will be forwarded to this identity provider.",
		},
		"disable_user_info": { //disableUserInfo
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
	oidcResource := resourceKeycloakIdentityProvider()
	oidcResource.Schema = mergeSchemas(oidcResource.Schema, oidcFacebookSchema)
	oidcResource.CreateContext = resourceKeycloakIdentityProviderCreate(getOidcFacebookIdentityProviderFromData, setOidcFacebookIdentityProviderData)
	oidcResource.ReadContext = resourceKeycloakIdentityProviderRead(setOidcFacebookIdentityProviderData)
	oidcResource.UpdateContext = resourceKeycloakIdentityProviderUpdate(getOidcFacebookIdentityProviderFromData, setOidcFacebookIdentityProviderData)
	return oidcResource
}

func getOidcFacebookIdentityProviderFromData(data *schema.ResourceData, keycloakVersion *version.Version) (*keycloak.IdentityProvider, error) {
	rec, defaultConfig := getIdentityProviderFromData(data, keycloakVersion)
	rec.ProviderId = data.Get("provider_id").(string)

	aliasRaw, ok := data.GetOk("alias")
	if ok {
		rec.Alias = aliasRaw.(string)
	} else {
		rec.Alias = "facebook"
	}

	facebookOidcIdentityProviderConfig := &keycloak.IdentityProviderConfig{
		ClientId:                    data.Get("client_id").(string),
		ClientSecret:                data.Get("client_secret").(string),
		FetchedFields:               data.Get("fetched_fields").(string),
		DefaultScope:                data.Get("default_scopes").(string),
		AcceptsPromptNoneForwFrmClt: types.KeycloakBoolQuoted(data.Get("accepts_prompt_none_forward_from_client").(bool)),
		UseJwksUrl:                  true,
		DisableUserInfo:             types.KeycloakBoolQuoted(data.Get("disable_user_info").(bool)),

		//since keycloak v26 moved to IdentityProvider - still here fore backward compatibility
		HideOnLoginPage: types.KeycloakBoolQuoted(data.Get("hide_on_login_page").(bool)),
	}

	if err := mergo.Merge(facebookOidcIdentityProviderConfig, defaultConfig); err != nil {
		return nil, err
	}

	rec.Config = facebookOidcIdentityProviderConfig

	return rec, nil
}

func setOidcFacebookIdentityProviderData(data *schema.ResourceData, identityProvider *keycloak.IdentityProvider, keycloakVersion *version.Version) error {
	setIdentityProviderData(data, identityProvider, keycloakVersion)
	data.Set("provider_id", identityProvider.ProviderId)
	data.Set("client_id", identityProvider.Config.ClientId)
	data.Set("fetched_fields", identityProvider.Config.FetchedFields)
	data.Set("default_scopes", identityProvider.Config.DefaultScope)
	data.Set("accepts_prompt_none_forward_from_client", identityProvider.Config.AcceptsPromptNoneForwFrmClt)
	data.Set("disable_user_info", identityProvider.Config.DisableUserInfo)

	if keycloakVersion.LessThan(keycloak.Version_26.AsVersion()) {
		// Since keycloak v26 the attribute "hideOnLoginPage" is not part of the identity provider config anymore!
		data.Set("hide_on_login_page", identityProvider.Config.HideOnLoginPage)
		return nil
	}

	return nil
}
