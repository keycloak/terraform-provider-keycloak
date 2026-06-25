package keycloak

import (
	"context"
	"fmt"
	"strings"
)

type OpenIdHardcodedRoleProtocolMapper struct {
	Id            string
	Name          string
	RealmId       string
	ClientId      string
	ClientScopeId string

	RoleId string
}

var roleField = "role"

func parseRoleClientIdAndName(roleProp string) (string, string) {
	parts := strings.Split(roleProp, ".")

	if len(parts) == 2 {
		return parts[0], parts[1]
	}

	return "", parts[0]
}

func (keycloakClient *KeycloakClient) getRolePropFromRole(ctx context.Context, role *Role) (string, error) {
	if role.ClientRole {
		client, err := keycloakClient.GetOpenidClient(ctx, role.RealmId, role.ContainerId)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("%s.%s", client.ClientId, role.Name), nil
	}

	return role.Name, nil
}

func (mapper *OpenIdHardcodedRoleProtocolMapper) convertToGenericProtocolMapper(roleProp string) *protocolMapper {
	return &protocolMapper{
		Id:             mapper.Id,
		Name:           mapper.Name,
		Protocol:       "openid-connect",
		ProtocolMapper: "oidc-hardcoded-role-mapper",
		Config: map[string]string{
			roleField: roleProp,
		},
	}
}

func (protocolMapper *protocolMapper) convertToOpenIdHardcodedRoleProtocolMapper(realmId, clientId, clientScopeId, roleId string) (*OpenIdHardcodedRoleProtocolMapper, error) {
	return &OpenIdHardcodedRoleProtocolMapper{
		Id:            protocolMapper.Id,
		Name:          protocolMapper.Name,
		RealmId:       realmId,
		ClientId:      clientId,
		ClientScopeId: clientScopeId,

		RoleId: roleId,
	}, nil
}

func (keycloakClient *KeycloakClient) GetOpenIdHardcodedRoleProtocolMapper(ctx context.Context, realmId, clientId, clientScopeId, mapperId string) (*OpenIdHardcodedRoleProtocolMapper, error) {
	var protocolMapper *protocolMapper

	err := keycloakClient.get(ctx, individualProtocolMapperPath(realmId, clientId, clientScopeId, mapperId), &protocolMapper, nil)
	if err != nil {
		return nil, err
	}

	roleClientId, roleName := parseRoleClientIdAndName(protocolMapper.Config[roleField])

	var roleClientUId = ""
	if roleClientId != "" {
		client, err := keycloakClient.GetOpenidClientByClientId(ctx, realmId, roleClientId)
		if err != nil {
			return nil, err
		}

		roleClientUId = client.Id
	}

	role, err := keycloakClient.GetRoleByName(ctx, realmId, roleClientUId, roleName)
	if err != nil {
		return nil, err
	}

	return protocolMapper.convertToOpenIdHardcodedRoleProtocolMapper(realmId, clientId, clientScopeId, role.Id)
}

func (keycloakClient *KeycloakClient) DeleteOpenIdHardcodedRoleProtocolMapper(ctx context.Context, realmId, clientId, clientScopeId, mapperId string) error {
	return keycloakClient.delete(ctx, individualProtocolMapperPath(realmId, clientId, clientScopeId, mapperId), nil)
}

func (keycloakClient *KeycloakClient) NewOpenIdHardcodedRoleProtocolMapper(ctx context.Context, mapper *OpenIdHardcodedRoleProtocolMapper) error {
	role, err := keycloakClient.GetRole(ctx, mapper.RealmId, mapper.RoleId)
	if err != nil {
		return err
	}

	roleProp, err := keycloakClient.getRolePropFromRole(ctx, role)
	if err != nil {
		return err
	}

	path := protocolMapperPath(mapper.RealmId, mapper.ClientId, mapper.ClientScopeId)

	_, location, err := keycloakClient.post(ctx, path, mapper.convertToGenericProtocolMapper(roleProp))
	if err != nil {
		return err
	}

	mapper.Id = getIdFromLocationHeader(location)

	return nil
}

func (keycloakClient *KeycloakClient) UpdateOpenIdHardcodedRoleProtocolMapper(ctx context.Context, mapper *OpenIdHardcodedRoleProtocolMapper) error {
	role, err := keycloakClient.GetRole(ctx, mapper.RealmId, mapper.RoleId)
	if err != nil {
		return err
	}

	roleProp, err := keycloakClient.getRolePropFromRole(ctx, role)
	if err != nil {
		return err
	}

	path := individualProtocolMapperPath(mapper.RealmId, mapper.ClientId, mapper.ClientScopeId, mapper.Id)

	return keycloakClient.put(ctx, path, mapper.convertToGenericProtocolMapper(roleProp))
}

func (keycloakClient *KeycloakClient) ValidateOpenIdHardcodedRoleProtocolMapper(ctx context.Context, mapper *OpenIdHardcodedRoleProtocolMapper) error {
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
