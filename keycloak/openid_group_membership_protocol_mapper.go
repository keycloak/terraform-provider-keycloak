package keycloak

import (
	"context"
	"fmt"
	"strconv"
)

type OpenIdGroupMembershipProtocolMapper struct {
	Id            string
	Name          string
	RealmId       string
	ClientId      string
	ClientScopeId string

	AddToIdToken            bool
	AddToAccessToken        bool
	AddToUserinfo           bool
	AddToTokenIntrospection bool

	ClaimName string
	FullPath  bool
}

func (mapper *OpenIdGroupMembershipProtocolMapper) convertToGenericProtocolMapper() *protocolMapper {
	return &protocolMapper{
		Id:             mapper.Id,
		Name:           mapper.Name,
		Protocol:       "openid-connect",
		ProtocolMapper: "oidc-group-membership-mapper",
		Config: map[string]string{
			fullPathField:                strconv.FormatBool(mapper.FullPath),
			addToIdTokenField:            strconv.FormatBool(mapper.AddToIdToken),
			addToAccessTokenField:        strconv.FormatBool(mapper.AddToAccessToken),
			addToUserInfoField:           strconv.FormatBool(mapper.AddToUserinfo),
			addToTokenIntrospectionField: strconv.FormatBool(mapper.AddToTokenIntrospection),
			claimNameField:               mapper.ClaimName,
		},
	}
}

func (protocolMapper *protocolMapper) convertToOpenIdGroupMembershipProtocolMapper(realmId, clientId, clientScopeId string) (*OpenIdGroupMembershipProtocolMapper, error) {
	fullPath, err := parseBoolAndTreatEmptyStringAsFalse(protocolMapper.Config[fullPathField])
	if err != nil {
		return nil, err
	}

	idTokenClaim, err := parseBoolAndTreatEmptyStringAsFalse(protocolMapper.Config[addToIdTokenField])
	if err != nil {
		return nil, err
	}

	accessTokenClaim, err := parseBoolAndTreatEmptyStringAsFalse(protocolMapper.Config[addToAccessTokenField])
	if err != nil {
		return nil, err
	}

	userinfoTokenClaim, err := parseBoolAndTreatEmptyStringAsFalse(protocolMapper.Config[addToUserInfoField])
	if err != nil {
		return nil, err
	}

	addToTokenIntrospectionClaim, err := parseBoolAndTreatEmptyStringAsFalse(protocolMapper.Config[addToTokenIntrospectionField])
	if err != nil {
		return nil, err
	}

	return &OpenIdGroupMembershipProtocolMapper{
		Id:            protocolMapper.Id,
		Name:          protocolMapper.Name,
		RealmId:       realmId,
		ClientId:      clientId,
		ClientScopeId: clientScopeId,

		ClaimName:               protocolMapper.Config[claimNameField],
		FullPath:                fullPath,
		AddToIdToken:            idTokenClaim,
		AddToAccessToken:        accessTokenClaim,
		AddToUserinfo:           userinfoTokenClaim,
		AddToTokenIntrospection: addToTokenIntrospectionClaim,
	}, nil
}

func (keycloakClient *KeycloakClient) GetOpenIdGroupMembershipProtocolMapper(ctx context.Context, realmId, clientId, clientScopeId, mapperId string) (*OpenIdGroupMembershipProtocolMapper, error) {
	var protocolMapper *protocolMapper

	err := keycloakClient.get(ctx, individualProtocolMapperPath(realmId, clientId, clientScopeId, mapperId), &protocolMapper, nil)
	if err != nil {
		return nil, err
	}

	return protocolMapper.convertToOpenIdGroupMembershipProtocolMapper(realmId, clientId, clientScopeId)
}

func (keycloakClient *KeycloakClient) DeleteOpenIdGroupMembershipProtocolMapper(ctx context.Context, realmId, clientId, clientScopeId, mapperId string) error {
	return keycloakClient.delete(ctx, individualProtocolMapperPath(realmId, clientId, clientScopeId, mapperId), nil)
}

func (keycloakClient *KeycloakClient) NewOpenIdGroupMembershipProtocolMapper(ctx context.Context, mapper *OpenIdGroupMembershipProtocolMapper) error {
	path := protocolMapperPath(mapper.RealmId, mapper.ClientId, mapper.ClientScopeId)

	_, location, err := keycloakClient.post(ctx, path, mapper.convertToGenericProtocolMapper())
	if err != nil {
		return err
	}

	mapper.Id = getIdFromLocationHeader(location)

	return nil
}

func (keycloakClient *KeycloakClient) UpdateOpenIdGroupMembershipProtocolMapper(ctx context.Context, mapper *OpenIdGroupMembershipProtocolMapper) error {
	path := individualProtocolMapperPath(mapper.RealmId, mapper.ClientId, mapper.ClientScopeId, mapper.Id)

	return keycloakClient.put(ctx, path, mapper.convertToGenericProtocolMapper())
}

func (keycloakClient *KeycloakClient) ValidateOpenIdGroupMembershipProtocolMapper(ctx context.Context, mapper *OpenIdGroupMembershipProtocolMapper) error {
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
