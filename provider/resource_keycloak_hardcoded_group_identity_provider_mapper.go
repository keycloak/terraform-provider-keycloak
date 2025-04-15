package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakHardcodedGroupIdentityProviderMapper() *schema.Resource {
	mapperSchema := map[string]*schema.Schema{
		"group": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Group Name",
		},
	}
	genericMapperResource := resourceKeycloakIdentityProviderMapper()
	genericMapperResource.Schema = mergeSchemas(genericMapperResource.Schema, mapperSchema)
	genericMapperResource.CreateContext = resourceKeycloakIdentityProviderMapperCreate(getHardcodedGroupIdentityProviderMapperFromData, setHardcodedGroupIdentityProviderMapperData)
	genericMapperResource.ReadContext = resourceKeycloakIdentityProviderMapperRead(setHardcodedGroupIdentityProviderMapperData)
	genericMapperResource.UpdateContext = resourceKeycloakIdentityProviderMapperUpdate(getHardcodedGroupIdentityProviderMapperFromData, setHardcodedGroupIdentityProviderMapperData)
	return genericMapperResource
}

func getHardcodedGroupIdentityProviderMapperFromData(_ context.Context, data *schema.ResourceData, _ interface{}) (*keycloak.IdentityProviderMapper, error) {
	rec, _ := getIdentityProviderMapperFromData(data)

	rec.IdentityProviderMapper = "oidc-hardcoded-group-idp-mapper"
	rec.Config.Group = data.Get("group").(string)

	return rec, nil
}

func setHardcodedGroupIdentityProviderMapperData(data *schema.ResourceData, identityProviderMapper *keycloak.IdentityProviderMapper) error {
	setIdentityProviderMapperData(data, identityProviderMapper)
	data.Set("group", identityProviderMapper.Config.Group)

	return nil
}
