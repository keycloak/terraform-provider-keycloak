package keycloak

import (
	"context"
	"fmt"
	"strconv"
)

type OpenIdUserAttributeProtocolMapper struct {
	Id            string
	Name          string
	RealmId       string
	ClientId      string
	ClientScopeId string

	AddToIdToken            bool
	AddToAccessToken        bool
	AddToUserInfo           bool
	AddToTokenIntrospection bool

	UserAttribute  string
	ClaimName      string
	ClaimValueType string

	Multivalued              bool // indicates whether is this an array of attributes or a single attribute
	AggregateAttributeValues bool
}

func (mapper *OpenIdUserAttributeProtocolMapper) convertToGenericProtocolMapper() *protocolMapper {
	return &protocolMapper{
		Id:             mapper.Id,
		Name:           mapper.Name,
		Protocol:       "openid-connect",
		ProtocolMapper: "oidc-usermodel-attribute-mapper",
		Config: map[string]string{
			addToIdTokenField:             strconv.FormatBool(mapper.AddToIdToken),
			addToAccessTokenField:         strconv.FormatBool(mapper.AddToAccessToken),
			addToUserInfoField:            strconv.FormatBool(mapper.AddToUserInfo),
			addToTokenIntrospectionField:  strconv.FormatBool(mapper.AddToTokenIntrospection),
			userAttributeField:            mapper.UserAttribute,
			claimNameField:                mapper.ClaimName,
			claimValueTypeField:           mapper.ClaimValueType,
			multivaluedField:              strconv.FormatBool(mapper.Multivalued),
			aggregateAttributeValuesField: strconv.FormatBool(mapper.AggregateAttributeValues),
		},
	}
}

func (protocolMapper *protocolMapper) convertToOpenIdUserAttributeProtocolMapper(realmId, clientId, clientScopeId string) (*OpenIdUserAttributeProtocolMapper, error) {
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

	// multivalued's default is "", this is an issue when importing an existing mapper
	multivalued, err := parseBoolAndTreatEmptyStringAsFalse(protocolMapper.Config[multivaluedField])
	if err != nil {
		return nil, err
	}

	aggregateAttributeValues, err := parseBoolAndTreatEmptyStringAsFalse(protocolMapper.Config[aggregateAttributeValuesField])
	if err != nil {
		return nil, err
	}

	return &OpenIdUserAttributeProtocolMapper{
		Id:            protocolMapper.Id,
		Name:          protocolMapper.Name,
		RealmId:       realmId,
		ClientId:      clientId,
		ClientScopeId: clientScopeId,

		AddToIdToken:            addToIdToken,
		AddToAccessToken:        addToAccessToken,
		AddToUserInfo:           addToUserInfo,
		AddToTokenIntrospection: addToTokenIntrospection,

		UserAttribute:            protocolMapper.Config[userAttributeField],
		ClaimName:                protocolMapper.Config[claimNameField],
		ClaimValueType:           protocolMapper.Config[claimValueTypeField],
		Multivalued:              multivalued,
		AggregateAttributeValues: aggregateAttributeValues,
	}, nil
}

func (keycloakClient *KeycloakClient) GetOpenIdUserAttributeProtocolMapper(ctx context.Context, realmId, clientId, clientScopeId, mapperId string) (*OpenIdUserAttributeProtocolMapper, error) {
	var protocolMapper *protocolMapper

	err := keycloakClient.get(ctx, individualProtocolMapperPath(realmId, clientId, clientScopeId, mapperId), &protocolMapper, nil)
	if err != nil {
		return nil, err
	}

	return protocolMapper.convertToOpenIdUserAttributeProtocolMapper(realmId, clientId, clientScopeId)
}

func (keycloakClient *KeycloakClient) DeleteOpenIdUserAttributeProtocolMapper(ctx context.Context, realmId, clientId, clientScopeId, mapperId string) error {
	return keycloakClient.delete(ctx, individualProtocolMapperPath(realmId, clientId, clientScopeId, mapperId), nil)
}

func (keycloakClient *KeycloakClient) NewOpenIdUserAttributeProtocolMapper(ctx context.Context, mapper *OpenIdUserAttributeProtocolMapper) error {
	path := protocolMapperPath(mapper.RealmId, mapper.ClientId, mapper.ClientScopeId)

	_, location, err := keycloakClient.post(ctx, path, mapper.convertToGenericProtocolMapper())
	if err != nil {
		return err
	}

	mapper.Id = getIdFromLocationHeader(location)

	return nil
}

func (keycloakClient *KeycloakClient) UpdateOpenIdUserAttributeProtocolMapper(ctx context.Context, mapper *OpenIdUserAttributeProtocolMapper) error {
	path := individualProtocolMapperPath(mapper.RealmId, mapper.ClientId, mapper.ClientScopeId, mapper.Id)

	return keycloakClient.put(ctx, path, mapper.convertToGenericProtocolMapper())
}

func (keycloakClient *KeycloakClient) ValidateOpenIdUserAttributeProtocolMapper(ctx context.Context, mapper *OpenIdUserAttributeProtocolMapper) error {
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
