package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakUserTemplateImporterIdentityProviderMapper() *schema.Resource {
	mapperSchema := map[string]*schema.Schema{
		"template": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Username For Template Import",
		},
	}
	genericMapperResource := resourceKeycloakIdentityProviderMapper()
	genericMapperResource.Schema = mergeSchemas(genericMapperResource.Schema, mapperSchema)
	genericMapperResource.CreateContext = resourceKeycloakIdentityProviderMapperCreate(getUserTemplateImporterIdentityProviderMapperFromData, setUserTemplateImporterIdentityProviderMapperData)
	genericMapperResource.ReadContext = resourceKeycloakIdentityProviderMapperRead(setUserTemplateImporterIdentityProviderMapperData)
	genericMapperResource.UpdateContext = resourceKeycloakIdentityProviderMapperUpdate(getUserTemplateImporterIdentityProviderMapperFromData, setUserTemplateImporterIdentityProviderMapperData)
	return genericMapperResource
}

func getUserTemplateImporterIdentityProviderMapperFromData(ctx context.Context, data *schema.ResourceData, meta interface{}) (*keycloak.IdentityProviderMapper, error) {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	rec, _ := getIdentityProviderMapperFromData(data)
	identityProvider, err := keycloakClient.GetIdentityProvider(ctx, rec.Realm, rec.IdentityProviderAlias)
	if err != nil {
		return nil, err
	}

	if identityProvider.ProviderId == "facebook" || identityProvider.ProviderId == "google" || identityProvider.ProviderId == "keycloak-oidc" {
		rec.IdentityProviderMapper = "oidc-username-idp-mapper"
	} else {
		rec.IdentityProviderMapper = fmt.Sprintf("%s-username-idp-mapper", identityProvider.ProviderId)
	}

	rec.Config.Template = data.Get("template").(string)

	return rec, nil
}

func setUserTemplateImporterIdentityProviderMapperData(data *schema.ResourceData, identityProviderMapper *keycloak.IdentityProviderMapper) error {
	setIdentityProviderMapperData(data, identityProviderMapper)
	data.Set("template", identityProviderMapper.Config.Template)

	return nil
}
