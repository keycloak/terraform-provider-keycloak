package keycloak

import (
	"context"
	"fmt"
	"strconv"
)

type SamlUserAttributeProtocolMapper struct {
	Id            string
	Name          string
	RealmId       string
	ClientId      string
	ClientScopeId string

	UserAttribute            string
	FriendlyName             string
	SamlAttributeName        string
	SamlAttributeNameFormat  string
	AggregateAttributeValues bool
}

func (mapper *SamlUserAttributeProtocolMapper) convertToGenericProtocolMapper() *protocolMapper {
	return &protocolMapper{
		Id:             mapper.Id,
		Name:           mapper.Name,
		Protocol:       "saml",
		ProtocolMapper: "saml-user-attribute-mapper",
		Config: map[string]string{
			attributeNameField:            mapper.SamlAttributeName,
			attributeNameFormatField:      mapper.SamlAttributeNameFormat,
			friendlyNameField:             mapper.FriendlyName,
			userAttributeField:            mapper.UserAttribute,
			aggregateAttributeValuesField: strconv.FormatBool(mapper.AggregateAttributeValues),
		},
	}
}

func (protocolMapper *protocolMapper) convertToSamlUserAttributeProtocolMapper(realmId, clientId, clientScopeId string) (*SamlUserAttributeProtocolMapper, error) {
	aggregateAttributeValues, err := parseBoolAndTreatEmptyStringAsFalse(protocolMapper.Config[aggregateAttributeValuesField])
	if err != nil {
		return nil, err
	}

	return &SamlUserAttributeProtocolMapper{
		Id:            protocolMapper.Id,
		Name:          protocolMapper.Name,
		RealmId:       realmId,
		ClientId:      clientId,
		ClientScopeId: clientScopeId,

		UserAttribute:            protocolMapper.Config[userAttributeField],
		FriendlyName:             protocolMapper.Config[friendlyNameField],
		SamlAttributeName:        protocolMapper.Config[attributeNameField],
		SamlAttributeNameFormat:  protocolMapper.Config[attributeNameFormatField],
		AggregateAttributeValues: aggregateAttributeValues,
	}, nil
}

func (keycloakClient *KeycloakClient) GetSamlUserAttributeProtocolMapper(ctx context.Context, realmId, clientId, clientScopeId, mapperId string) (*SamlUserAttributeProtocolMapper, error) {
	var protocolMapper *protocolMapper

	err := keycloakClient.get(ctx, individualProtocolMapperPath(realmId, clientId, clientScopeId, mapperId), &protocolMapper, nil)
	if err != nil {
		return nil, err
	}

	return protocolMapper.convertToSamlUserAttributeProtocolMapper(realmId, clientId, clientScopeId)
}

func (keycloakClient *KeycloakClient) DeleteSamlUserAttributeProtocolMapper(ctx context.Context, realmId, clientId, clientScopeId, mapperId string) error {
	return keycloakClient.delete(ctx, individualProtocolMapperPath(realmId, clientId, clientScopeId, mapperId), nil)
}

func (keycloakClient *KeycloakClient) NewSamlUserAttributeProtocolMapper(ctx context.Context, mapper *SamlUserAttributeProtocolMapper) error {
	path := protocolMapperPath(mapper.RealmId, mapper.ClientId, mapper.ClientScopeId)

	_, location, err := keycloakClient.post(ctx, path, mapper.convertToGenericProtocolMapper())
	if err != nil {
		return err
	}

	mapper.Id = getIdFromLocationHeader(location)

	return nil
}

func (keycloakClient *KeycloakClient) UpdateSamlUserAttributeProtocolMapper(ctx context.Context, mapper *SamlUserAttributeProtocolMapper) error {
	path := individualProtocolMapperPath(mapper.RealmId, mapper.ClientId, mapper.ClientScopeId, mapper.Id)

	return keycloakClient.put(ctx, path, mapper.convertToGenericProtocolMapper())
}

func (keycloakClient *KeycloakClient) ValidateSamlUserAttributeProtocolMapper(ctx context.Context, mapper *SamlUserAttributeProtocolMapper) error {
	if mapper.ClientId == "" && mapper.ClientScopeId == "" {
		return fmt.Errorf("validation error: one of ClientId or ClientScopeId must be set")
	}

	protocolMappers, err := keycloakClient.listGenericProtocolMappers(ctx, mapper.RealmId, mapper.ClientId, mapper.ClientScopeId)
	if err != nil {
		return err
	}

	for _, protocolMapper := range protocolMappers {
		if protocolMapper.Name == mapper.Name && protocolMapper.Id != mapper.Id {
			return fmt.Errorf("validation error: a protocol mapper with name %s already exists for this client", mapper.Name)
		}
	}

	return nil
}
