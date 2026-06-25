package keycloak

import (
	"context"
	"fmt"
	"strconv"
)

type OpenIdUserSessionNoteProtocolMapper struct {
	Id            string
	Name          string
	RealmId       string
	ClientId      string
	ClientScopeId string

	AddToIdToken            bool
	AddToAccessToken        bool
	AddToUserInfo           bool
	AddToTokenIntrospection bool

	ClaimName       string
	ClaimValueType  string
	UserSessionNote string
}

func (mapper *OpenIdUserSessionNoteProtocolMapper) convertToGenericProtocolMapper() *protocolMapper {
	return &protocolMapper{
		Id:             mapper.Id,
		Name:           mapper.Name,
		Protocol:       "openid-connect",
		ProtocolMapper: "oidc-usersessionmodel-note-mapper",
		Config: map[string]string{
			addToIdTokenField:            strconv.FormatBool(mapper.AddToIdToken),
			addToAccessTokenField:        strconv.FormatBool(mapper.AddToAccessToken),
			addToUserInfoField:           strconv.FormatBool(mapper.AddToUserInfo),
			addToTokenIntrospectionField: strconv.FormatBool(mapper.AddToTokenIntrospection),
			claimNameField:               mapper.ClaimName,
			claimValueTypeField:          mapper.ClaimValueType,
			userSessionNoteField:         mapper.UserSessionNote,
		},
	}
}

func (protocolMapper *protocolMapper) convertToOpenIdUserSessionNoteProtocolMapper(realmId, clientId, clientScopeId string) (*OpenIdUserSessionNoteProtocolMapper, error) {
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

	addToTokenIntrospection, err := parseBoolAndTreatEmptyStringAsFalse(protocolMapper.Config[addToTokenIntrospectionField])
	if err != nil {
		return nil, err
	}

	return &OpenIdUserSessionNoteProtocolMapper{
		Id:            protocolMapper.Id,
		Name:          protocolMapper.Name,
		RealmId:       realmId,
		ClientId:      clientId,
		ClientScopeId: clientScopeId,

		AddToIdToken:            addToIdToken,
		AddToAccessToken:        addToAccessToken,
		AddToUserInfo:           addToUserInfo,
		AddToTokenIntrospection: addToTokenIntrospection,

		ClaimName:       protocolMapper.Config[claimNameField],
		ClaimValueType:  protocolMapper.Config[claimValueTypeField],
		UserSessionNote: protocolMapper.Config[userSessionNoteField],
	}, nil
}

func (keycloakClient *KeycloakClient) GetOpenIdUserSessionNoteProtocolMapper(ctx context.Context, realmId, clientId, clientScopeId, mapperId string) (*OpenIdUserSessionNoteProtocolMapper, error) {
	var protocolMapper *protocolMapper

	err := keycloakClient.get(ctx, individualProtocolMapperPath(realmId, clientId, clientScopeId, mapperId), &protocolMapper, nil)
	if err != nil {
		return nil, err
	}

	return protocolMapper.convertToOpenIdUserSessionNoteProtocolMapper(realmId, clientId, clientScopeId)
}

func (keycloakClient *KeycloakClient) DeleteOpenIdUserSessionNoteProtocolMapper(ctx context.Context, realmId, clientId, clientScopeId, mapperId string) error {
	return keycloakClient.delete(ctx, individualProtocolMapperPath(realmId, clientId, clientScopeId, mapperId), nil)
}

func (keycloakClient *KeycloakClient) NewOpenIdUserSessionNoteProtocolMapper(ctx context.Context, mapper *OpenIdUserSessionNoteProtocolMapper) error {
	path := protocolMapperPath(mapper.RealmId, mapper.ClientId, mapper.ClientScopeId)

	_, location, err := keycloakClient.post(ctx, path, mapper.convertToGenericProtocolMapper())
	if err != nil {
		return err
	}

	mapper.Id = getIdFromLocationHeader(location)

	return nil
}

func (keycloakClient *KeycloakClient) UpdateOpenIdUserSessionNoteProtocolMapper(ctx context.Context, mapper *OpenIdUserSessionNoteProtocolMapper) error {
	path := individualProtocolMapperPath(mapper.RealmId, mapper.ClientId, mapper.ClientScopeId, mapper.Id)

	return keycloakClient.put(ctx, path, mapper.convertToGenericProtocolMapper())
}

func (keycloakClient *KeycloakClient) ValidateOpenIdUserSessionNoteProtocolMapper(ctx context.Context, mapper *OpenIdUserSessionNoteProtocolMapper) error {
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
