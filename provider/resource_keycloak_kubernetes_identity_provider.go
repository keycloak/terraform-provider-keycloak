package provider

import (
	"dario.cat/mergo"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakKubernetesIdentityProvider() *schema.Resource {
	kubernetesSchema := map[string]*schema.Schema{
		"provider_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "kubernetes",
			Description: "Provider ID, is always kubernetes.",
		},
		"issuer": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The issuer of the Kubernetes service account tokens. Depending your Keycloak Realm \"ssl_required\" setting, this may need to be an HTTPS URL.",
		},
		"hide_on_login_page": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "This is always set to true for Kubernetes identity provider.",
		},
	}
	kubernetesResource := resourceKeycloakIdentityProvider()
	kubernetesResource.Schema = mergeSchemas(kubernetesResource.Schema, kubernetesSchema)
	kubernetesResource.CreateContext = resourceKeycloakIdentityProviderCreate(getKubernetesIdentityProviderFromData, setKubernetesIdentityProviderData)
	kubernetesResource.ReadContext = resourceKeycloakIdentityProviderRead(setKubernetesIdentityProviderData)
	kubernetesResource.UpdateContext = resourceKeycloakIdentityProviderUpdate(getKubernetesIdentityProviderFromData, setKubernetesIdentityProviderData)
	return kubernetesResource
}

func getKubernetesIdentityProviderFromData(data *schema.ResourceData, keycloakVersion *version.Version) (*keycloak.IdentityProvider, error) {
	idp, defaultConfig := getIdentityProviderFromData(data, keycloakVersion)
	idp.ProviderId = data.Get("provider_id").(string)
	// The Kubernetes Identity Provider is only used for service accounts, so we always hide it on the login page.
	idp.HideOnLogin = true

	kubernetesIdentityProviderConfig := &keycloak.IdentityProviderConfig{
		Issuer: data.Get("issuer").(string),
	}

	if err := mergo.Merge(kubernetesIdentityProviderConfig, defaultConfig); err != nil {
		return nil, err
	}

	idp.Config = kubernetesIdentityProviderConfig

	return idp, nil
}

func setKubernetesIdentityProviderData(data *schema.ResourceData, identityProvider *keycloak.IdentityProvider, keycloakVersion *version.Version) error {
	setIdentityProviderData(data, identityProvider, keycloakVersion)

	data.Set("issuer", identityProvider.Config.Issuer)

	return nil
}
