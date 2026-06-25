package keycloak

import (
	"context"
	"fmt"
)

type SamlUserPropertyProtocolMapper struct {
	Id            string
	Name          string
	RealmId       string
	ClientId      string
	ClientScopeId string

	UserProperty            string
	FriendlyName            string
	SamlAttributeName       string
	SamlAttributeNameFormat string
}

func (mapper *SamlUserPropertyProtocolMapper) convertToGenericProtocolMapper() *protocolMapper {
	return &protocolMapper{
		Id:             mapper.Id,
		Name:           mapper.Name,
		Protocol:       "saml",
		ProtocolMapper: "saml-user-property-mapper",
		Config: map[string]string{
			attributeNameField:       mapper.SamlAttributeName,
			attributeNameFormatField: mapper.SamlAttributeNameFormat,
			friendlyNameField:        mapper.FriendlyName,
			userAttributeField:       mapper.UserProperty,
		},
	}
}

func (protocolMapper *protocolMapper) convertToSamlUserPropertyProtocolMapper(realmId, clientId, clientScopeId string) *SamlUserPropertyProtocolMapper {
	return &SamlUserPropertyProtocolMapper{
		Id:            protocolMapper.Id,
		Name:          protocolMapper.Name,
		RealmId:       realmId,
		ClientId:      clientId,
		ClientScopeId: clientScopeId,

		UserProperty:            protocolMapper.Config[userAttributeField],
		FriendlyName:            protocolMapper.Config[friendlyNameField],
		SamlAttributeName:       protocolMapper.Config[attributeNameField],
		SamlAttributeNameFormat: protocolMapper.Config[attributeNameFormatField],
	}
}

func (keycloakClient *KeycloakClient) GetSamlUserPropertyProtocolMapper(ctx context.Context, realmId, clientId, clientScopeId, mapperId string) (*SamlUserPropertyProtocolMapper, error) {
	var protocolMapper *protocolMapper

	err := keycloakClient.get(ctx, individualProtocolMapperPath(realmId, clientId, clientScopeId, mapperId), &protocolMapper, nil)
	if err != nil {
		return nil, err
	}

	return protocolMapper.convertToSamlUserPropertyProtocolMapper(realmId, clientId, clientScopeId), nil
}

func (keycloakClient *KeycloakClient) DeleteSamlUserPropertyProtocolMapper(ctx context.Context, realmId, clientId, clientScopeId, mapperId string) error {
	return keycloakClient.delete(ctx, individualProtocolMapperPath(realmId, clientId, clientScopeId, mapperId), nil)
}

func (keycloakClient *KeycloakClient) NewSamlUserPropertyProtocolMapper(ctx context.Context, mapper *SamlUserPropertyProtocolMapper) error {
	path := protocolMapperPath(mapper.RealmId, mapper.ClientId, mapper.ClientScopeId)

	_, location, err := keycloakClient.post(ctx, path, mapper.convertToGenericProtocolMapper())
	if err != nil {
		return err
	}

	mapper.Id = getIdFromLocationHeader(location)

	return nil
}

func (keycloakClient *KeycloakClient) UpdateSamlUserPropertyProtocolMapper(ctx context.Context, mapper *SamlUserPropertyProtocolMapper) error {
	path := individualProtocolMapperPath(mapper.RealmId, mapper.ClientId, mapper.ClientScopeId, mapper.Id)

	return keycloakClient.put(ctx, path, mapper.convertToGenericProtocolMapper())
}

func (keycloakClient *KeycloakClient) ValidateSamlUserPropertyProtocolMapper(ctx context.Context, mapper *SamlUserPropertyProtocolMapper) error {
	if mapper.ClientId == "" && mapper.ClientScopeId == "" {
		return fmt.Errorf("validation error: one of ClientId or ClientScopeId must be set")
	}

	protocolMappers, err := keycloakClient.listProtocolMappers(ctx, mapper.RealmId, mapper.ClientId, mapper.ClientScopeId)
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
