package provider

import (
	"dario.cat/mergo"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"github.com/keycloak/terraform-provider-keycloak/keycloak/types"
)

func resourceKeycloakOidcGithubIdentityProvider() *schema.Resource {
	oidcGithubSchema := map[string]*schema.Schema{
		"alias": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "The alias uniquely identifies an identity provider and it is also used to build the redirect uri. In case of github this is computed and always github",
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
			Default:     "github",
			Description: "provider id, is always github, unless you have a extended custom implementation",
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
		"base_url": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "https://github.com",
			Description: "Base URL for the GitHub instance, defaults to https://github.com",
		},
		"api_url": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "https://api.github.com",
			Description: "API URL for the GitHub instance, defaults to https://api.github.com",
		},
		"github_json_format": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Whether GitHub API shoulds accept JSON explicitly during token authentication requests, defaults to false",
		},
		"default_scopes": { //defaultScope
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "user:email",
			Description: "The scopes to be sent when asking for authorization. See the documentation for possible values, separator and default value'. Default to 'user:email'",
		},
		"hide_on_login_page": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Hide On Login Page.",
		},
	}
	oidcResource := resourceKeycloakIdentityProvider()
	oidcResource.Schema = mergeSchemas(oidcResource.Schema, oidcGithubSchema)
	oidcResource.CreateContext = resourceKeycloakIdentityProviderCreate(getOidcGithubIdentityProviderFromData, setOidcGithubIdentityProviderData)
	oidcResource.ReadContext = resourceKeycloakIdentityProviderRead(setOidcGithubIdentityProviderData)
	oidcResource.UpdateContext = resourceKeycloakIdentityProviderUpdate(getOidcGithubIdentityProviderFromData, setOidcGithubIdentityProviderData)
	return oidcResource
}

func getOidcGithubIdentityProviderFromData(data *schema.ResourceData, keycloakVersion *version.Version) (*keycloak.IdentityProvider, error) {
	rec, defaultConfig := getIdentityProviderFromData(data, keycloakVersion)
	rec.ProviderId = data.Get("provider_id").(string)

	aliasRaw, ok := data.GetOk("alias")
	if ok {
		rec.Alias = aliasRaw.(string)
	} else {
		rec.Alias = "github"
	}

	githubOidcIdentityProviderConfig := &keycloak.IdentityProviderConfig{
		ClientId:         data.Get("client_id").(string),
		ClientSecret:     data.Get("client_secret").(string),
		DefaultScope:     data.Get("default_scopes").(string),
		GithubJsonFormat: types.KeycloakBoolQuoted(data.Get("github_json_format").(bool)),
		BaseUrl:          data.Get("base_url").(string),
		ApiUrl:           data.Get("api_url").(string),

		//since keycloak v26 moved to IdentityProvider - still here fore backward compatibility
		HideOnLoginPage: types.KeycloakBoolQuoted(data.Get("hide_on_login_page").(bool)),
	}

	if err := mergo.Merge(githubOidcIdentityProviderConfig, defaultConfig); err != nil {
		return nil, err
	}

	rec.Config = githubOidcIdentityProviderConfig

	return rec, nil
}

func setOidcGithubIdentityProviderData(data *schema.ResourceData, identityProvider *keycloak.IdentityProvider, keycloakVersion *version.Version) error {
	setIdentityProviderData(data, identityProvider, keycloakVersion)
	data.Set("provider_id", identityProvider.ProviderId)
	data.Set("client_id", identityProvider.Config.ClientId)
	data.Set("github_json_format", identityProvider.Config.GithubJsonFormat)
	data.Set("base_url", identityProvider.Config.BaseUrl)
	data.Set("api_url", identityProvider.Config.ApiUrl)
	data.Set("default_scopes", identityProvider.Config.DefaultScope)

	if keycloakVersion.LessThan(keycloak.Version_26.AsVersion()) {
		// Since keycloak v26 the attribute "hideOnLoginPage" is not part of the identity provider config anymore!
		data.Set("hide_on_login_page", identityProvider.Config.HideOnLoginPage)
		return nil
	}

	return nil
}
