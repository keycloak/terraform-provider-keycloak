package keycloak

import (
	"context"
	"fmt"
	"strconv"
)

type OpenIdUserRealmRoleProtocolMapper struct {
	Id            string
	Name          string
	RealmId       string
	ClientId      string
	ClientScopeId string

	AddToIdToken            bool
	AddToAccessToken        bool
	AddToUserInfo           bool
	AddToTokenIntrospection bool

	RealmRolePrefix string
	Multivalued     bool
	ClaimName       string
	ClaimValueType  string
}

func (mapper *OpenIdUserRealmRoleProtocolMapper) convertToGenericProtocolMapper() *protocolMapper {
	return &protocolMapper{
		Id:             mapper.Id,
		Name:           mapper.Name,
		Protocol:       "openid-connect",
		ProtocolMapper: "oidc-usermodel-realm-role-mapper",
		Config: map[string]string{
			addToIdTokenField:                   strconv.FormatBool(mapper.AddToIdToken),
			addToAccessTokenField:               strconv.FormatBool(mapper.AddToAccessToken),
			addToUserInfoField:                  strconv.FormatBool(mapper.AddToUserInfo),
			addToTokenIntrospectionField:        strconv.FormatBool(mapper.AddToTokenIntrospection),
			claimNameField:                      mapper.ClaimName,
			claimValueTypeField:                 mapper.ClaimValueType,
			multivaluedField:                    strconv.FormatBool(mapper.Multivalued),
			userRealmRoleMappingRolePrefixField: mapper.RealmRolePrefix,
		},
	}
}

func (protocolMapper *protocolMapper) convertToOpenIdUserRealmRoleProtocolMapper(realmId, clientId, clientScopeId string) (*OpenIdUserRealmRoleProtocolMapper, error) {
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

	multivalued, err := parseBoolAndTreatEmptyStringAsFalse(protocolMapper.Config[multivaluedField])
	if err != nil {
		return nil, err
	}

	return &OpenIdUserRealmRoleProtocolMapper{
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
		Multivalued:     multivalued,
		RealmRolePrefix: protocolMapper.Config[userRealmRoleMappingRolePrefixField],
	}, nil
}

func (keycloakClient *KeycloakClient) GetOpenIdUserRealmRoleProtocolMapper(ctx context.Context, realmId, clientId, clientScopeId, mapperId string) (*OpenIdUserRealmRoleProtocolMapper, error) {
	var protocolMapper *protocolMapper

	err := keycloakClient.get(ctx, individualProtocolMapperPath(realmId, clientId, clientScopeId, mapperId), &protocolMapper, nil)
	if err != nil {
		return nil, err
	}

	return protocolMapper.convertToOpenIdUserRealmRoleProtocolMapper(realmId, clientId, clientScopeId)
}

func (keycloakClient *KeycloakClient) DeleteOpenIdUserRealmRoleProtocolMapper(ctx context.Context, realmId, clientId, clientScopeId, mapperId string) error {
	return keycloakClient.delete(ctx, individualProtocolMapperPath(realmId, clientId, clientScopeId, mapperId), nil)
}

func (keycloakClient *KeycloakClient) NewOpenIdUserRealmRoleProtocolMapper(ctx context.Context, mapper *OpenIdUserRealmRoleProtocolMapper) error {
	path := protocolMapperPath(mapper.RealmId, mapper.ClientId, mapper.ClientScopeId)

	_, location, err := keycloakClient.post(ctx, path, mapper.convertToGenericProtocolMapper())
	if err != nil {
		return err
	}

	mapper.Id = getIdFromLocationHeader(location)

	return nil
}

func (keycloakClient *KeycloakClient) UpdateOpenIdUserRealmRoleProtocolMapper(ctx context.Context, mapper *OpenIdUserRealmRoleProtocolMapper) error {
	path := individualProtocolMapperPath(mapper.RealmId, mapper.ClientId, mapper.ClientScopeId, mapper.Id)

	return keycloakClient.put(ctx, path, mapper.convertToGenericProtocolMapper())
}

func (keycloakClient *KeycloakClient) ValidateOpenIdUserRealmRoleProtocolMapper(ctx context.Context, mapper *OpenIdUserRealmRoleProtocolMapper) error {
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
