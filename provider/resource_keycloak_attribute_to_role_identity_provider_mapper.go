package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakAttributeToRoleIdentityProviderMapper() *schema.Resource {
	mapperSchema := map[string]*schema.Schema{
		"attribute_name": {
			Type:          schema.TypeString,
			Optional:      true,
			Description:   "Attribute Name",
			ConflictsWith: []string{"attribute_friendly_name"},
		},
		"attribute_value": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Attribute Value",
		},
		"claim_name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "OIDC Claim Name",
		},
		"claim_value": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "OIDC Claim Value",
		},
		"role": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Role Name",
		},
		"attribute_friendly_name": {
			Type:          schema.TypeString,
			Optional:      true,
			Description:   "Attribute Friendly Name",
			ConflictsWith: []string{"attribute_name"},
		},
	}
	genericMapperResource := resourceKeycloakIdentityProviderMapper()
	genericMapperResource.Schema = mergeSchemas(genericMapperResource.Schema, mapperSchema)
	genericMapperResource.CreateContext = resourceKeycloakIdentityProviderMapperCreate(getAttributeToRoleIdentityProviderMapperFromData, setAttributeToRoleIdentityProviderMapperData)
	genericMapperResource.ReadContext = resourceKeycloakIdentityProviderMapperRead(setAttributeToRoleIdentityProviderMapperData)
	genericMapperResource.UpdateContext = resourceKeycloakIdentityProviderMapperUpdate(getAttributeToRoleIdentityProviderMapperFromData, setAttributeToRoleIdentityProviderMapperData)
	return genericMapperResource
}

func getAttributeToRoleIdentityProviderMapperFromData(ctx context.Context, data *schema.ResourceData, meta interface{}) (*keycloak.IdentityProviderMapper, error) {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	rec, _ := getIdentityProviderMapperFromData(data)
	identityProvider, err := keycloakClient.GetIdentityProvider(ctx, rec.Realm, rec.IdentityProviderAlias)
	if err != nil {
		return nil, err
	}

	rec.IdentityProviderMapper = fmt.Sprintf("%s-role-idp-mapper", identityProvider.ProviderId)
	rec.Config.Role = data.Get("role").(string)

	if identityProvider.ProviderId == "saml" {
		if attr, ok := data.GetOk("attribute_friendly_name"); ok {
			rec.Config.AttributeFriendlyName = attr.(string)
		} else if attr, ok := data.GetOk("attribute_name"); ok {
			rec.Config.Attribute = attr.(string)
		} else {
			return nil, fmt.Errorf(`provider.keycloak: keycloak_attribute_to_role_identity_provider_mapper: %s: either "attribute_name" or "attribute_friendly_name" should be set for %s identity provider`, data.Get("name").(string), identityProvider.ProviderId)
		}
		if _, ok := data.GetOk("attribute_value"); !ok {
			return nil, fmt.Errorf(`provider.keycloak: keycloak_attribute_to_role_identity_provider_mapper: %s: "attribute_value": required field for %s identity provider`, data.Get("name").(string), identityProvider.ProviderId)
		}
		rec.Config.AttributeValue = data.Get("attribute_value").(string)
	} else if identityProvider.ProviderId == "oidc" {
		if _, ok := data.GetOk("claim_name"); !ok {
			return nil, fmt.Errorf(`provider.keycloak: keycloak_attribute_to_role_identity_provider_mapper: %s: "claim_name": required field for %s identity provider`, data.Get("name").(string), identityProvider.ProviderId)
		}
		if _, ok := data.GetOk("claim_value"); !ok {
			return nil, fmt.Errorf(`provider.keycloak: keycloak_attribute_to_role_identity_provider_mapper: %s: "claim_value": required field for %s identity provider`, data.Get("name").(string), identityProvider.ProviderId)
		}
		rec.Config.Claim = data.Get("claim_name").(string)
		rec.Config.ClaimValue = data.Get("claim_value").(string)
	} else {
		return nil, fmt.Errorf(`provider.keycloak: keycloak_attribute_to_role_identity_provider_mapper: %s: "%s" identity provider is not supported yet`, data.Get("name").(string), identityProvider.ProviderId)
	}

	return rec, nil
}

func setAttributeToRoleIdentityProviderMapperData(data *schema.ResourceData, identityProviderMapper *keycloak.IdentityProviderMapper) error {
	setIdentityProviderMapperData(data, identityProviderMapper)
	data.Set("role", identityProviderMapper.Config.Role)
	data.Set("attribute_name", identityProviderMapper.Config.Attribute)
	data.Set("attribute_value", identityProviderMapper.Config.AttributeValue)
	data.Set("claim_name", identityProviderMapper.Config.Claim)
	data.Set("claim_value", identityProviderMapper.Config.ClaimValue)
	data.Set("attribute_friendly_name", identityProviderMapper.Config.AttributeFriendlyName)

	return nil
}
