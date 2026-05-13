package provider

import (
	"dario.cat/mergo"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakOidcOpenshiftV4IdentityProvider() *schema.Resource {
	oidcOpenshiftV4Schema := map[string]*schema.Schema{
		"alias": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "The alias uniquely identifies an identity provider and it is also used to build the redirect uri. Defaults to openshift-v4 if not set.",
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
			Default:     "openshift-v4",
			Description: "provider id, is always openshift-v4, unless you have an extended custom implementation",
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
			Required:    true,
			Description: "Base URL of the OpenShift 4 cluster, e.g. https://openshift.example.com:8443.",
		},
		"default_scopes": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "user:full",
			Description: "The scopes to be sent when asking for authorization. Defaults to 'user:full'.",
		},
		"hide_on_login_page": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Hide On Login Page.",
		},
	}
	oidcResource := resourceKeycloakIdentityProvider()
	oidcResource.Schema = mergeSchemas(oidcResource.Schema, oidcOpenshiftV4Schema)
	oidcResource.CreateContext = resourceKeycloakIdentityProviderCreate(getOidcOpenshiftV4IdentityProviderFromData, setOidcOpenshiftV4IdentityProviderData)
	oidcResource.ReadContext = resourceKeycloakIdentityProviderRead(setOidcOpenshiftV4IdentityProviderData)
	oidcResource.UpdateContext = resourceKeycloakIdentityProviderUpdate(getOidcOpenshiftV4IdentityProviderFromData, setOidcOpenshiftV4IdentityProviderData)
	return oidcResource
}

func getOidcOpenshiftV4IdentityProviderFromData(data *schema.ResourceData, keycloakVersion *version.Version) (*keycloak.IdentityProvider, error) {
	rec, defaultConfig := getIdentityProviderFromData(data, keycloakVersion)
	rec.ProviderId = data.Get("provider_id").(string)

	aliasRaw, ok := data.GetOk("alias")
	if ok {
		rec.Alias = aliasRaw.(string)
	} else {
		rec.Alias = "openshift-v4"
	}

	openshiftV4OidcIdentityProviderConfig := &keycloak.IdentityProviderConfig{
		ClientId:     data.Get("client_id").(string),
		ClientSecret: data.Get("client_secret").(string),
		DefaultScope: data.Get("default_scopes").(string),
		BaseUrl:      data.Get("base_url").(string),
	}

	if err := mergo.Merge(openshiftV4OidcIdentityProviderConfig, defaultConfig); err != nil {
		return nil, err
	}

	rec.Config = openshiftV4OidcIdentityProviderConfig

	return rec, nil
}

func setOidcOpenshiftV4IdentityProviderData(data *schema.ResourceData, identityProvider *keycloak.IdentityProvider, keycloakVersion *version.Version) error {
	setIdentityProviderData(data, identityProvider, keycloakVersion)
	if err := data.Set("provider_id", identityProvider.ProviderId); err != nil {
		return err
	}
	if err := data.Set("client_id", identityProvider.Config.ClientId); err != nil {
		return err
	}
	if err := data.Set("base_url", identityProvider.Config.BaseUrl); err != nil {
		return err
	}
	if err := data.Set("default_scopes", identityProvider.Config.DefaultScope); err != nil {
		return err
	}

	return nil
}
