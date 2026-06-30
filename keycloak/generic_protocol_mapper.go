package keycloak

import (
	"context"
	"fmt"
)

type GenericProtocolMapper struct {
	ClientId       string            `json:"-"`
	ClientScopeId  string            `json:"-"`
	Config         map[string]string `json:"config"`
	Id             string            `json:"id,omitempty"`
	Name           string            `json:"name"`
	Protocol       string            `json:"protocol"`
	ProtocolMapper string            `json:"protocolMapper"`
	RealmId        string            `json:"-"`
}

type OpenidClientWithGenericProtocolMappers struct {
	OpenidClient
	ProtocolMappers []*GenericProtocolMapper
}

func (keycloakClient *KeycloakClient) NewGenericProtocolMapper(ctx context.Context, genericProtocolMapper *GenericProtocolMapper) error {
	path := protocolMapperPath(genericProtocolMapper.RealmId, genericProtocolMapper.ClientId, genericProtocolMapper.ClientScopeId)

	_, location, err := keycloakClient.post(ctx, path, genericProtocolMapper)
	if err != nil {
		return err
	}

	genericProtocolMapper.Id = getIdFromLocationHeader(location)

	return nil
}

func (keycloakClient *KeycloakClient) GetGenericProtocolMappers(ctx context.Context, realmId string, clientId string) (*OpenidClientWithGenericProtocolMappers, error) {
	var openidClientWithGenericProtocolMappers OpenidClientWithGenericProtocolMappers

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/clients/%s", realmId, clientId), &openidClientWithGenericProtocolMappers, nil)
	if err != nil {
		return nil, err
	}

	openidClientWithGenericProtocolMappers.RealmId = realmId
	openidClientWithGenericProtocolMappers.ClientId = clientId

	for _, protocolMapper := range openidClientWithGenericProtocolMappers.ProtocolMappers {
		protocolMapper.RealmId = realmId
		protocolMapper.ClientId = clientId
	}

	return &openidClientWithGenericProtocolMappers, nil

}

func (keycloakClient *KeycloakClient) GetGenericProtocolMapper(ctx context.Context, realmId string, clientId string, clientScopeId string, mapperId string) (*GenericProtocolMapper, error) {
	var genericProtocolMapper GenericProtocolMapper

	err := keycloakClient.get(ctx, individualProtocolMapperPath(realmId, clientId, clientScopeId, mapperId), &genericProtocolMapper, nil)
	if err != nil {
		return nil, err
	}

	// these values are not provided by the keycloak API
	genericProtocolMapper.ClientId = clientId
	genericProtocolMapper.ClientScopeId = clientScopeId
	genericProtocolMapper.RealmId = realmId

	return &genericProtocolMapper, nil
}

func (keycloakClient *KeycloakClient) GetGenericProtocolMapperByName(ctx context.Context, realmId string, clientId string, clientScopeId string, mapperName string) (*GenericProtocolMapper, error) {
	protocolMappers, err := keycloakClient.listGenericProtocolMappers(ctx, realmId, clientId, clientScopeId)
	if err != nil {
		return nil, err
	}

	for _, protocolMapper := range protocolMappers {
		if protocolMapper.Name == mapperName {
			return &GenericProtocolMapper{
				ClientId:       clientId,
				ClientScopeId:  clientScopeId,
				RealmId:        realmId,
				Id:             protocolMapper.Id,
				Name:           protocolMapper.Name,
				Protocol:       protocolMapper.Protocol,
				ProtocolMapper: protocolMapper.ProtocolMapper,
				Config:         protocolMapper.Config,
			}, nil
		}
	}

	return nil, fmt.Errorf("no protocol mapper found with name %s for this client", mapperName)
}

func (keycloakClient *KeycloakClient) UpdateGenericProtocolMapper(ctx context.Context, genericProtocolMapper *GenericProtocolMapper) error {
	path := individualProtocolMapperPath(genericProtocolMapper.RealmId, genericProtocolMapper.ClientId, genericProtocolMapper.ClientScopeId, genericProtocolMapper.Id)

	return keycloakClient.put(ctx, path, genericProtocolMapper)
}

func (keycloakClient *KeycloakClient) DeleteGenericProtocolMapper(ctx context.Context, realmId string, clientId string, clientScopeId string, mapperId string) error {
	return keycloakClient.delete(ctx, individualProtocolMapperPath(realmId, clientId, clientScopeId, mapperId), nil)
}

func (mapper *GenericProtocolMapper) Validate(ctx context.Context, keycloakClient *KeycloakClient) error {
	if mapper.ClientId == "" && mapper.ClientScopeId == "" {
		return fmt.Errorf("validation error: one of ClientId or ClientScopeId must be set")
	}
	if mapper.ClientId != "" && mapper.ClientScopeId != "" {
		return fmt.Errorf("validation error: only one of ClientId or ClientScopeId must be set")
	}

	protocolMappers, err := keycloakClient.listGenericProtocolMappers(ctx, mapper.RealmId, mapper.ClientId, mapper.ClientScopeId)
	if err != nil {
		return err
	}

	for _, protocolMapper := range protocolMappers {
		if protocolMapper.Name == mapper.Name {
			return fmt.Errorf("validation error: a protocol mapper with name %s already exists for this client", mapper.Name)
		}
	}

	return nil
}
