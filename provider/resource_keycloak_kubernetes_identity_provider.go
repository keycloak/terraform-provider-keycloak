package provider

import (
	"dario.cat/mergo"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"github.com/keycloak/terraform-provider-keycloak/keycloak/types"
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
			Description: "The issuer of the Kubernetes service account tokens.",
		},
		"hide_on_login_page": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "Hide On Login Page.",
		},
	}
	kubernetesResource := resourceKeycloakIdentityProvider()
	kubernetesResource.Schema = mergeSchemas(kubernetesResource.Schema, kubernetesSchema)
	kubernetesResource.CreateContext = resourceKeycloakIdentityProviderCreate(getKubernetesProviderFromData, setKubernetesIdentityProviderData)
	kubernetesResource.ReadContext = resourceKeycloakIdentityProviderRead(setKubernetesIdentityProviderData)
	kubernetesResource.UpdateContext = resourceKeycloakIdentityProviderUpdate(getKubernetesProviderFromData, setKubernetesIdentityProviderData)
	return kubernetesResource
}

func getKubernetesProviderFromData(data *schema.ResourceData, keycloakVersion *version.Version) (*keycloak.IdentityProvider, error) {
	idp, defaultConfig := getIdentityProviderFromData(data, keycloakVersion)
	idp.ProviderId = data.Get("provider_id").(string)

	kubernetesIdentityProviderConfig := &keycloak.IdentityProviderConfig{
		Issuer: data.Get("issuer").(string),

		//since keycloak v26 moved to IdentityProvider - still here fore backward compatibility
		HideOnLoginPage: types.KeycloakBoolQuoted(data.Get("hide_on_login_page").(bool)),
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

	if keycloakVersion.LessThan(keycloak.Version_26.AsVersion()) {
		// Since keycloak v26 the attribute "hideOnLoginPage" is not part of the identity provider config anymore!
		data.Set("hide_on_login_page", identityProvider.Config.HideOnLoginPage)
		return nil
	}

	return nil
}
