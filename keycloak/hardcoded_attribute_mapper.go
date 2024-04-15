package keycloak

import (
	"context"
	"fmt"
)

type HardcodedAttributeMapper struct {
	Id                   string
	Name                 string
	RealmId              string
	LdapUserFederationId string
	AttributeName        string
	AttributeValue       string
}

func convertFromHardcodedAttributeMapperToComponent(hardcodedMapper *HardcodedAttributeMapper) *component {
	return &component{
		Id:           hardcodedMapper.Id,
		Name:         hardcodedMapper.Name,
		ProviderId:   "hardcoded-attribute-mapper",
		ProviderType: "org.keycloak.storage.ldap.mappers.LDAPStorageMapper",
		ParentId:     hardcodedMapper.LdapUserFederationId,

		Config: map[string][]string{
			"user.model.attribute": {
				hardcodedMapper.AttributeName,
			},
			"attribute.value": {
				hardcodedMapper.AttributeValue,
			},
		},
	}
}

func convertFromComponentToHardcodedAttributeMapper(component *component, realmId string) *HardcodedAttributeMapper {
	return &HardcodedAttributeMapper{
		Id:                   component.Id,
		Name:                 component.Name,
		RealmId:              realmId,
		LdapUserFederationId: component.ParentId,

		AttributeName:  component.getConfig("user.model.attribute"),
		AttributeValue: component.getConfig("attribute.value"),
	}
}

func (keycloakClient *KeycloakClient) NewHardcodedAttributeMapper(ctx context.Context, hardcodedMapper *HardcodedAttributeMapper) error {
	_, location, err := keycloakClient.post(ctx, fmt.Sprintf("/realms/%s/components", hardcodedMapper.RealmId), convertFromHardcodedAttributeMapperToComponent(hardcodedMapper))
	if err != nil {
		return err
	}

	hardcodedMapper.Id = getIdFromLocationHeader(location)

	return nil
}

func (keycloakClient *KeycloakClient) GetHardcodedAttributeMapper(ctx context.Context, realmId, id string) (*HardcodedAttributeMapper, error) {
	var component *component

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/components/%s", realmId, id), &component, nil)
	if err != nil {
		return nil, err
	}

	return convertFromComponentToHardcodedAttributeMapper(component, realmId), nil
}

func (keycloakClient *KeycloakClient) UpdateHardcodedAttributeMapper(ctx context.Context, hardcodedMapper *HardcodedAttributeMapper) error {
	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/components/%s", hardcodedMapper.RealmId, hardcodedMapper.Id), convertFromHardcodedAttributeMapperToComponent(hardcodedMapper))
}

func (keycloakClient *KeycloakClient) DeleteHardcodedAttributeMapper(ctx context.Context, realmId, id string) error {
	return keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/components/%s", realmId, id), nil)
}
