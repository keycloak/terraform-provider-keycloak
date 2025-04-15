package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakHardcodedAttributeIdentityProviderMapper() *schema.Resource {
	mapperSchema := map[string]*schema.Schema{
		"attribute_name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "OIDC Claim",
		},
		"attribute_value": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "User Attribute",
		},
		"user_session": {
			Type:        schema.TypeBool,
			Required:    true,
			ForceNew:    true,
			Description: "Is Attribute Related To a User Session",
		},
	}
	genericMapperResource := resourceKeycloakIdentityProviderMapper()
	genericMapperResource.Schema = mergeSchemas(genericMapperResource.Schema, mapperSchema)
	genericMapperResource.CreateContext = resourceKeycloakIdentityProviderMapperCreate(getHardcodedAttributeIdentityProviderMapperFromData, setHardcodedAttributeIdentityProviderMapperData)
	genericMapperResource.ReadContext = resourceKeycloakIdentityProviderMapperRead(setHardcodedAttributeIdentityProviderMapperData)
	genericMapperResource.UpdateContext = resourceKeycloakIdentityProviderMapperUpdate(getHardcodedAttributeIdentityProviderMapperFromData, setHardcodedAttributeIdentityProviderMapperData)
	return genericMapperResource
}

func getHardcodedAttributeIdentityProviderMapperFromData(_ context.Context, data *schema.ResourceData, _ interface{}) (*keycloak.IdentityProviderMapper, error) {
	rec, _ := getIdentityProviderMapperFromData(data)

	rec.IdentityProviderMapper = getHardcodedAttributeIdentityProviderMapperType(data.Get("user_session").(bool))
	rec.Config.HardcodedAttribute = data.Get("attribute_name").(string)
	rec.Config.AttributeValue = data.Get("attribute_value").(string)

	return rec, nil
}

func setHardcodedAttributeIdentityProviderMapperData(data *schema.ResourceData, identityProviderMapper *keycloak.IdentityProviderMapper) error {
	setIdentityProviderMapperData(data, identityProviderMapper)
	data.Set("attribute_name", identityProviderMapper.Config.HardcodedAttribute)
	data.Set("attribute_value", identityProviderMapper.Config.AttributeValue)

	mapperType, err := getUserSessionFromHardcodedAttributeIdentityProviderMapperType(identityProviderMapper.IdentityProviderMapper)
	if err != nil {
		return err
	}

	data.Set("user_session", mapperType)

	return nil
}

func getHardcodedAttributeIdentityProviderMapperType(userSession bool) string {
	if userSession {
		return "hardcoded-user-session-attribute-idp-mapper"
	} else {
		return "hardcoded-attribute-idp-mapper"
	}
}

func getUserSessionFromHardcodedAttributeIdentityProviderMapperType(mapperType string) (bool, error) {
	if mapperType == "hardcoded-user-session-attribute-idp-mapper" {
		return true, nil
	} else if mapperType == "hardcoded-attribute-idp-mapper" {
		return false, nil
	} else {
		return false, fmt.Errorf(`provider.keycloak: keycloak_hardcoded_attribute_identity_provider_mapper: mapper type "%s" is not valid`, mapperType)
	}
}
