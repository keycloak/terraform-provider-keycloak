package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakAttributeImporterIdentityProviderMapper() *schema.Resource {
	mapperSchema := map[string]*schema.Schema{
		"user_attribute": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "User Attribute",
		},
		"attribute_name": {
			Type:          schema.TypeString,
			Optional:      true,
			Description:   "Attribute Name",
			ConflictsWith: []string{"attribute_friendly_name"},
		},
		"attribute_friendly_name": {
			Type:          schema.TypeString,
			Optional:      true,
			Description:   "Attribute Friendly Name",
			ConflictsWith: []string{"attribute_name"},
		},
		"claim_name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Claim Name",
		},
	}
	genericMapperResource := resourceKeycloakIdentityProviderMapper()
	genericMapperResource.Schema = mergeSchemas(genericMapperResource.Schema, mapperSchema)
	genericMapperResource.CreateContext = resourceKeycloakIdentityProviderMapperCreate(getAttributeImporterIdentityProviderMapperFromData, setAttributeImporterIdentityProviderMapperData)
	genericMapperResource.ReadContext = resourceKeycloakIdentityProviderMapperRead(setAttributeImporterIdentityProviderMapperData)
	genericMapperResource.UpdateContext = resourceKeycloakIdentityProviderMapperUpdate(getAttributeImporterIdentityProviderMapperFromData, setAttributeImporterIdentityProviderMapperData)
	return genericMapperResource
}

func getAttributeImporterIdentityProviderMapperFromData(ctx context.Context, data *schema.ResourceData, meta interface{}) (*keycloak.IdentityProviderMapper, error) {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	rec, _ := getIdentityProviderMapperFromData(data)
	identityProvider, err := keycloakClient.GetIdentityProvider(ctx, rec.Realm, rec.IdentityProviderAlias)
	if err != nil {
		return nil, err
	}

	rec.IdentityProviderMapper = fmt.Sprintf("%s-user-attribute-idp-mapper", identityProvider.ProviderId)
	rec.Config.UserAttribute = data.Get("user_attribute").(string)

	if identityProvider.ProviderId == "saml" {
		if attr, ok := data.GetOk("attribute_friendly_name"); ok {
			rec.Config.AttributeFriendlyName = attr.(string)
		} else if attr, ok := data.GetOk("attribute_name"); ok {
			rec.Config.Attribute = attr.(string)
		} else {
			return nil, fmt.Errorf(`provider.keycloak: keycloak_attribute_importer_identity_provider_mapper: %s: either "attribute_name" or "attribute_friendly_name" should be set for %s identity provider`, data.Get("name").(string), identityProvider.ProviderId)
		}
	} else if identityProvider.ProviderId == "oidc" {
		if _, ok := data.GetOk("claim_name"); !ok {
			return nil, fmt.Errorf(`provider.keycloak: keycloak_attribute_importer_identity_provider_mapper: %s: "claim_name": should be set for %s identity provider`, data.Get("name").(string), identityProvider.ProviderId)
		}

		rec.Config.Claim = data.Get("claim_name").(string)
	} else if identityProvider.ProviderId == "facebook" || identityProvider.ProviderId == "google" {
		rec.IdentityProviderMapper = fmt.Sprintf("%s-user-attribute-mapper", identityProvider.ProviderId)
		rec.Config.JsonField = data.Get("claim_name").(string)
		rec.Config.UserAttributeName = data.Get("user_attribute").(string)
	} else {
		return nil, fmt.Errorf(`provider.keycloak: keycloak_attribute_importer_identity_provider_mapper: %s: "%s" identity provider is not supported yet`, data.Get("name").(string), identityProvider.ProviderId)
	}

	return rec, nil
}

func setAttributeImporterIdentityProviderMapperData(data *schema.ResourceData, identityProviderMapper *keycloak.IdentityProviderMapper) error {
	setIdentityProviderMapperData(data, identityProviderMapper)

	claimName := identityProviderMapper.Config.Claim
	if claimName == "" {
		claimName = identityProviderMapper.Config.JsonField
	}

	data.Set("attribute_name", identityProviderMapper.Config.Attribute)
	data.Set("user_attribute", identityProviderMapper.Config.UserAttribute)
	data.Set("attribute_friendly_name", identityProviderMapper.Config.AttributeFriendlyName)
	data.Set("claim_name", claimName)

	return nil
}
