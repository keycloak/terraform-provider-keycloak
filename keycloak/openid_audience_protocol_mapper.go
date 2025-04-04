package keycloak

import (
	"context"
	"fmt"
	"strconv"
)

type OpenIdAudienceProtocolMapper struct {
	Id            string
	Name          string
	RealmId       string
	ClientId      string
	ClientScopeId string

	AddToIdToken     bool
	AddToAccessToken bool

	IncludedClientAudience string
	IncludedCustomAudience string
}

func (mapper *OpenIdAudienceProtocolMapper) convertToGenericProtocolMapper() *protocolMapper {
	return &protocolMapper{
		Id:             mapper.Id,
		Name:           mapper.Name,
		Protocol:       "openid-connect",
		ProtocolMapper: "oidc-audience-mapper",
		Config: map[string]string{
			addToIdTokenField:           strconv.FormatBool(mapper.AddToIdToken),
			addToAccessTokenField:       strconv.FormatBool(mapper.AddToAccessToken),
			includedClientAudienceField: mapper.IncludedClientAudience,
			includedCustomAudienceField: mapper.IncludedCustomAudience,
		},
	}
}

func (protocolMapper *protocolMapper) convertToOpenIdAudienceProtocolMapper(realmId, clientId, clientScopeId string) (*OpenIdAudienceProtocolMapper, error) {
	addToIdToken, err := parseBoolAndTreatEmptyStringAsFalse(protocolMapper.Config[addToIdTokenField])
	if err != nil {
		return nil, err
	}

	addToAccessToken, err := parseBoolAndTreatEmptyStringAsFalse(protocolMapper.Config[addToAccessTokenField])
	if err != nil {
		return nil, err
	}

	return &OpenIdAudienceProtocolMapper{
		Id:            protocolMapper.Id,
		Name:          protocolMapper.Name,
		RealmId:       realmId,
		ClientId:      clientId,
		ClientScopeId: clientScopeId,

		AddToIdToken:     addToIdToken,
		AddToAccessToken: addToAccessToken,

		IncludedClientAudience: protocolMapper.Config[includedClientAudienceField],
		IncludedCustomAudience: protocolMapper.Config[includedCustomAudienceField],
	}, nil
}

func (keycloakClient *KeycloakClient) GetOpenIdAudienceProtocolMapper(ctx context.Context, realmId, clientId, clientScopeId, mapperId string) (*OpenIdAudienceProtocolMapper, error) {
	var protocolMapper *protocolMapper

	err := keycloakClient.get(ctx, individualProtocolMapperPath(realmId, clientId, clientScopeId, mapperId), &protocolMapper, nil)
	if err != nil {
		return nil, err
	}

	return protocolMapper.convertToOpenIdAudienceProtocolMapper(realmId, clientId, clientScopeId)
}

func (keycloakClient *KeycloakClient) DeleteOpenIdAudienceProtocolMapper(ctx context.Context, realmId, clientId, clientScopeId, mapperId string) error {
	return keycloakClient.delete(ctx, individualProtocolMapperPath(realmId, clientId, clientScopeId, mapperId), nil)
}

func (keycloakClient *KeycloakClient) NewOpenIdAudienceProtocolMapper(ctx context.Context, mapper *OpenIdAudienceProtocolMapper) error {
	path := protocolMapperPath(mapper.RealmId, mapper.ClientId, mapper.ClientScopeId)

	_, location, err := keycloakClient.post(ctx, path, mapper.convertToGenericProtocolMapper())
	if err != nil {
		return err
	}

	mapper.Id = getIdFromLocationHeader(location)

	return nil
}

func (keycloakClient *KeycloakClient) UpdateOpenIdAudienceProtocolMapper(ctx context.Context, mapper *OpenIdAudienceProtocolMapper) error {
	path := individualProtocolMapperPath(mapper.RealmId, mapper.ClientId, mapper.ClientScopeId, mapper.Id)

	return keycloakClient.put(ctx, path, mapper.convertToGenericProtocolMapper())
}

func (keycloakClient *KeycloakClient) ValidateOpenIdAudienceProtocolMapper(ctx context.Context, mapper *OpenIdAudienceProtocolMapper) error {
	if mapper.ClientId == "" && mapper.ClientScopeId == "" {
		return fmt.Errorf("validation error: one of ClientId or ClientScopeId must be set")
	}

	if mapper.ClientId != "" && mapper.ClientScopeId != "" {
		return fmt.Errorf("validation error: ClientId and ClientScopeId cannot both be set")
	}

	if mapper.IncludedClientAudience == "" && mapper.IncludedCustomAudience == "" {
		return fmt.Errorf("validation error: one of IncludedClientAudience or IncludedCustomAudience must be set")
	}

	if mapper.IncludedClientAudience != "" && mapper.IncludedCustomAudience != "" {
		return fmt.Errorf("validation error: IncludedClientAudience and IncludedCustomAudience cannot both be set")
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

	if mapper.IncludedClientAudience != "" {
		_, err = keycloakClient.GetGenericClientByClientId(ctx, mapper.RealmId, mapper.IncludedClientAudience)
		if err != nil {
			return fmt.Errorf("validation error: %w", err)
		}
	}

	return nil
}
