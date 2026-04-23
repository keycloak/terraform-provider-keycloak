package provider

import (
	"dario.cat/mergo"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakSpiffeIdentityProvider() *schema.Resource {
	spiffeSchema := map[string]*schema.Schema{
		"provider_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "spiffe",
			Description: "Provider ID, is always spiffe.",
		},
		"trust_domain": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The SPIFFE trust domain. This must use the spiffe:// scheme.",
		},
		"bundle_endpoint": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The SPIFFE bundle endpoint or OpenID Connect JWKS endpoint exposing SPIFFE public keys. Depending your Keycloak Realm \"ssl_required\" setting, this may need to be an HTTPS URL.",
		},
		"hide_on_login_page": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "This is always set to true for SPIFFE identity provider.",
		},
	}

	spiffeResource := resourceKeycloakIdentityProvider()
	spiffeResource.Schema = mergeSchemas(spiffeResource.Schema, spiffeSchema)
	spiffeResource.CreateContext = resourceKeycloakIdentityProviderCreate(getSpiffeIdentityProviderFromData, setSpiffeIdentityProviderData)
	spiffeResource.ReadContext = resourceKeycloakIdentityProviderRead(setSpiffeIdentityProviderData)
	spiffeResource.UpdateContext = resourceKeycloakIdentityProviderUpdate(getSpiffeIdentityProviderFromData, setSpiffeIdentityProviderData)
	return spiffeResource
}

func getSpiffeIdentityProviderFromData(data *schema.ResourceData, keycloakVersion *version.Version) (*keycloak.IdentityProvider, error) {
	idp, defaultConfig := getIdentityProviderFromData(data, keycloakVersion)
	idp.ProviderId = data.Get("provider_id").(string)
	// The SPIFFE Identity Provider is only used for client authentication, so we always hide it on the login page.
	idp.HideOnLogin = true

	spiffeIdentityProviderConfig := &keycloak.IdentityProviderConfig{
		TrustDomain:    data.Get("trust_domain").(string),
		BundleEndpoint: data.Get("bundle_endpoint").(string),
	}

	if err := mergo.Merge(spiffeIdentityProviderConfig, defaultConfig); err != nil {
		return nil, err
	}

	idp.Config = spiffeIdentityProviderConfig

	return idp, nil
}

func setSpiffeIdentityProviderData(data *schema.ResourceData, identityProvider *keycloak.IdentityProvider, keycloakVersion *version.Version) error {
	setIdentityProviderData(data, identityProvider, keycloakVersion)

	data.Set("trust_domain", identityProvider.Config.TrustDomain)
	data.Set("bundle_endpoint", identityProvider.Config.BundleEndpoint)

	return nil
}
