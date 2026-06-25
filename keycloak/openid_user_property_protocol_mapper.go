package keycloak

import (
	"context"
	"fmt"
	"strconv"
)

type OpenIdUserPropertyProtocolMapper struct {
	Id            string
	Name          string
	RealmId       string
	ClientId      string
	ClientScopeId string

	AddToIdToken     bool
	AddToAccessToken bool
	AddToUserInfo    bool

	UserProperty   string
	ClaimName      string
	ClaimValueType string
}

func (mapper *OpenIdUserPropertyProtocolMapper) convertToGenericProtocolMapper() *protocolMapper {
	return &protocolMapper{
		Id:             mapper.Id,
		Name:           mapper.Name,
		Protocol:       "openid-connect",
		ProtocolMapper: "oidc-usermodel-property-mapper",
		Config: map[string]string{
			addToIdTokenField:     strconv.FormatBool(mapper.AddToIdToken),
			addToAccessTokenField: strconv.FormatBool(mapper.AddToAccessToken),
			addToUserInfoField:    strconv.FormatBool(mapper.AddToUserInfo),
			userPropertyField:     mapper.UserProperty,
			claimNameField:        mapper.ClaimName,
			claimValueTypeField:   mapper.ClaimValueType,
		},
	}
}

func (protocolMapper *protocolMapper) convertToOpenIdUserPropertyProtocolMapper(realmId, clientId, clientScopeId string) (*OpenIdUserPropertyProtocolMapper, error) {
	addToIdToken, err := parseBoolAndTreatEmptyStringAsFalse(protocolMapper.Config[addToIdTokenField])
	if err != nil {
		return nil, err
	}

	addToAccessToken, err := parseBoolAndTreatEmptyStringAsFalse(protocolMapper.Config[addToAccessTokenField])
	if err != nil {
		return nil, err
	}

	addToUserInfo, err := parseBoolAndTreatEmptyStringAsFalse(protocolMapper.Config[addToUserInfoField])
	if err != nil {
		return nil, err
	}

	return &OpenIdUserPropertyProtocolMapper{
		Id:            protocolMapper.Id,
		Name:          protocolMapper.Name,
		RealmId:       realmId,
		ClientId:      clientId,
		ClientScopeId: clientScopeId,

		AddToIdToken:     addToIdToken,
		AddToAccessToken: addToAccessToken,
		AddToUserInfo:    addToUserInfo,

		UserProperty:   protocolMapper.Config[userPropertyField],
		ClaimName:      protocolMapper.Config[claimNameField],
		ClaimValueType: protocolMapper.Config[claimValueTypeField],
	}, nil
}

func (keycloakClient *KeycloakClient) GetOpenIdUserPropertyProtocolMapper(ctx context.Context, realmId, clientId, clientScopeId, mapperId string) (*OpenIdUserPropertyProtocolMapper, error) {
	var protocolMapper *protocolMapper

	err := keycloakClient.get(ctx, individualProtocolMapperPath(realmId, clientId, clientScopeId, mapperId), &protocolMapper, nil)
	if err != nil {
		return nil, err
	}

	return protocolMapper.convertToOpenIdUserPropertyProtocolMapper(realmId, clientId, clientScopeId)
}

func (keycloakClient *KeycloakClient) DeleteOpenIdUserPropertyProtocolMapper(ctx context.Context, realmId, clientId, clientScopeId, mapperId string) error {
	return keycloakClient.delete(ctx, individualProtocolMapperPath(realmId, clientId, clientScopeId, mapperId), nil)
}

func (keycloakClient *KeycloakClient) NewOpenIdUserPropertyProtocolMapper(ctx context.Context, mapper *OpenIdUserPropertyProtocolMapper) error {
	path := protocolMapperPath(mapper.RealmId, mapper.ClientId, mapper.ClientScopeId)

	_, location, err := keycloakClient.post(ctx, path, mapper.convertToGenericProtocolMapper())
	if err != nil {
		return err
	}

	mapper.Id = getIdFromLocationHeader(location)

	return nil
}

func (keycloakClient *KeycloakClient) UpdateOpenIdUserPropertyProtocolMapper(ctx context.Context, mapper *OpenIdUserPropertyProtocolMapper) error {
	path := individualProtocolMapperPath(mapper.RealmId, mapper.ClientId, mapper.ClientScopeId, mapper.Id)

	return keycloakClient.put(ctx, path, mapper.convertToGenericProtocolMapper())
}

func (keycloakClient *KeycloakClient) ValidateOpenIdUserPropertyProtocolMapper(ctx context.Context, mapper *OpenIdUserPropertyProtocolMapper) error {
	existingMapper, err := keycloakClient.getProtocolMapperByName(ctx, mapper.RealmId, mapper.ClientId, mapper.ClientScopeId, mapper.Name)
	if err != nil {
		return err
	}

	if existingMapper != nil && existingMapper.Id != mapper.Id {
		return fmt.Errorf("validation error: a protocol mapper with name %s already exists for this client or client scope", mapper.Name)
	}

	return nil
}
