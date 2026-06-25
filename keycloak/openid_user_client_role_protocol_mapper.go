package keycloak

import (
	"context"
	"fmt"
	"strconv"
)

type OpenIdUserClientRoleProtocolMapper struct {
	Id            string
	Name          string
	RealmId       string
	ClientId      string
	ClientScopeId string

	AddToIdToken     bool
	AddToAccessToken bool
	AddToUserInfo    bool

	ClaimName               string
	ClaimValueType          string
	Multivalued             bool
	ClientIdForRoleMappings string
	ClientRolePrefix        string
}

func (mapper *OpenIdUserClientRoleProtocolMapper) convertToGenericProtocolMapper() *protocolMapper {
	return &protocolMapper{
		Id:             mapper.Id,
		Name:           mapper.Name,
		Protocol:       "openid-connect",
		ProtocolMapper: "oidc-usermodel-client-role-mapper",
		Config: map[string]string{
			addToIdTokenField:     strconv.FormatBool(mapper.AddToIdToken),
			addToAccessTokenField: strconv.FormatBool(mapper.AddToAccessToken),
			addToUserInfoField:    strconv.FormatBool(mapper.AddToUserInfo),

			claimNameField:                       mapper.ClaimName,
			claimValueTypeField:                  mapper.ClaimValueType,
			multivaluedField:                     strconv.FormatBool(mapper.Multivalued),
			userClientRoleMappingClientIdField:   mapper.ClientIdForRoleMappings,
			userClientRoleMappingRolePrefixField: mapper.ClientRolePrefix,
		},
	}
}

func (protocolMapper *protocolMapper) convertToOpenIdUserClientRoleProtocolMapper(realmId, clientId, clientScopeId string) (*OpenIdUserClientRoleProtocolMapper, error) {
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

	multivalued, err := parseBoolAndTreatEmptyStringAsFalse(protocolMapper.Config[multivaluedField])
	if err != nil {
		return nil, err
	}

	return &OpenIdUserClientRoleProtocolMapper{
		Id:            protocolMapper.Id,
		Name:          protocolMapper.Name,
		RealmId:       realmId,
		ClientId:      clientId,
		ClientScopeId: clientScopeId,

		AddToIdToken:     addToIdToken,
		AddToAccessToken: addToAccessToken,
		AddToUserInfo:    addToUserInfo,

		ClaimName:               protocolMapper.Config[claimNameField],
		ClaimValueType:          protocolMapper.Config[claimValueTypeField],
		Multivalued:             multivalued,
		ClientIdForRoleMappings: protocolMapper.Config[userClientRoleMappingClientIdField],
		ClientRolePrefix:        protocolMapper.Config[userClientRoleMappingRolePrefixField],
	}, nil
}

func (keycloakClient *KeycloakClient) GetOpenIdUserClientRoleProtocolMapper(ctx context.Context, realmId, clientId, clientScopeId, mapperId string) (*OpenIdUserClientRoleProtocolMapper, error) {
	var protocolMapper *protocolMapper

	err := keycloakClient.get(ctx, individualProtocolMapperPath(realmId, clientId, clientScopeId, mapperId), &protocolMapper, nil)
	if err != nil {
		return nil, err
	}

	return protocolMapper.convertToOpenIdUserClientRoleProtocolMapper(realmId, clientId, clientScopeId)
}

func (keycloakClient *KeycloakClient) DeleteOpenIdUserClientRoleProtocolMapper(ctx context.Context, realmId, clientId, clientScopeId, mapperId string) error {
	return keycloakClient.delete(ctx, individualProtocolMapperPath(realmId, clientId, clientScopeId, mapperId), nil)
}

func (keycloakClient *KeycloakClient) NewOpenIdUserClientRoleProtocolMapper(ctx context.Context, mapper *OpenIdUserClientRoleProtocolMapper) error {
	path := protocolMapperPath(mapper.RealmId, mapper.ClientId, mapper.ClientScopeId)

	_, location, err := keycloakClient.post(ctx, path, mapper.convertToGenericProtocolMapper())
	if err != nil {
		return err
	}

	mapper.Id = getIdFromLocationHeader(location)

	return nil
}

func (keycloakClient *KeycloakClient) UpdateOpenIdUserClientRoleProtocolMapper(ctx context.Context, mapper *OpenIdUserClientRoleProtocolMapper) error {
	path := individualProtocolMapperPath(mapper.RealmId, mapper.ClientId, mapper.ClientScopeId, mapper.Id)

	return keycloakClient.put(ctx, path, mapper.convertToGenericProtocolMapper())
}

func (keycloakClient *KeycloakClient) ValidateOpenIdUserClientRoleProtocolMapper(ctx context.Context, mapper *OpenIdUserClientRoleProtocolMapper) error {
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
