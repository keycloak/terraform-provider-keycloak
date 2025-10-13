package keycloak

import (
	"context"
	"fmt"
	"strconv"
)

type OpenIdSubProtocolMapper struct {
	Id            string
	Name          string
	RealmId       string
	ClientId      string
	ClientScopeId string

	AddToAccessToken        bool
	AddToTokenIntrospection bool
}

func (mapper *OpenIdSubProtocolMapper) convertToGenericProtocolMapper() *protocolMapper {
	return &protocolMapper{
		Id:             mapper.Id,
		Name:           mapper.Name,
		Protocol:       "openid-connect",
		ProtocolMapper: "oidc-sub-mapper",
		Config: map[string]string{
			addToAccessTokenField:        strconv.FormatBool(mapper.AddToAccessToken),
			addToTokenIntrospectionField: strconv.FormatBool(mapper.AddToTokenIntrospection),
		},
	}
}

func (protocolMapper *protocolMapper) convertToOpenIdSubProtocolMapper(realmId, clientId, clientScopeId string) (*OpenIdSubProtocolMapper, error) {
	addToAccessToken, err := parseBoolAndTreatEmptyStringAsFalse(protocolMapper.Config[addToAccessTokenField])
	if err != nil {
		return nil, err
	}

	addToTokenIntrospection, err := parseBoolAndTreatEmptyStringAsFalse(protocolMapper.Config[addToTokenIntrospectionField])
	if err != nil {
		return nil, err
	}

	return &OpenIdSubProtocolMapper{
		Id:            protocolMapper.Id,
		Name:          protocolMapper.Name,
		RealmId:       realmId,
		ClientId:      clientId,
		ClientScopeId: clientScopeId,

		AddToAccessToken:        addToAccessToken,
		AddToTokenIntrospection: addToTokenIntrospection,
	}, nil
}

func (keycloakClient *KeycloakClient) GetOpenIdSubProtocolMapper(ctx context.Context, realmId, clientId, clientScopeId, mapperId string) (*OpenIdSubProtocolMapper, error) {
	var protoMapper *protocolMapper

	err := keycloakClient.get(ctx, individualProtocolMapperPath(realmId, clientId, clientScopeId, mapperId), &protoMapper, nil)
	if err != nil {
		return nil, err
	}

	return protoMapper.convertToOpenIdSubProtocolMapper(realmId, clientId, clientScopeId)
}

func (keycloakClient *KeycloakClient) DeleteOpenIdSubProtocolMapper(ctx context.Context, realmId, clientId, clientScopeId, mapperId string) error {
	return keycloakClient.delete(ctx, individualProtocolMapperPath(realmId, clientId, clientScopeId, mapperId), nil)
}

func (keycloakClient *KeycloakClient) NewOpenIdSubProtocolMapper(ctx context.Context, mapper *OpenIdSubProtocolMapper) error {
	path := protocolMapperPath(mapper.RealmId, mapper.ClientId, mapper.ClientScopeId)

	_, location, err := keycloakClient.post(ctx, path, mapper.convertToGenericProtocolMapper())
	if err != nil {
		return err
	}

	mapper.Id = getIdFromLocationHeader(location)

	return nil
}

func (keycloakClient *KeycloakClient) UpdateOpenIdSubProtocolMapper(ctx context.Context, mapper *OpenIdSubProtocolMapper) error {
	path := individualProtocolMapperPath(mapper.RealmId, mapper.ClientId, mapper.ClientScopeId, mapper.Id)

	return keycloakClient.put(ctx, path, mapper.convertToGenericProtocolMapper())
}

func (keycloakClient *KeycloakClient) ValidateOpenIdSubProtocolMapper(ctx context.Context, mapper *OpenIdSubProtocolMapper) error {
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
