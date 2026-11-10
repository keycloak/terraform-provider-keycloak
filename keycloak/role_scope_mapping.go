package keycloak

import (
	"context"
	"fmt"
)

type RealmRoleRepresentation struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Composite   bool   `json:"composite"`
	ClientRole  bool   `json:"clientRole"`
	ContainerId string `json:"containerId"`
}

func roleScopeMappingUrl(realmId, clientId string, clientScopeId string, role *Role) string {
	if clientId != "" {
		if role.ClientRole {
			return fmt.Sprintf("/realms/%s/clients/%s/scope-mappings/clients/%s", realmId, clientId, role.ClientId)
		} else {
			return fmt.Sprintf("/realms/%s/clients/%s/scope-mappings/realm", realmId, clientId)
		}
	}

	if role.ClientRole {
		return fmt.Sprintf("/realms/%s/client-scopes/%s/scope-mappings/clients/%s", realmId, clientScopeId, role.ClientId)
	} else {
		return fmt.Sprintf("/realms/%s/client-scopes/%s/scope-mappings/realm", realmId, clientScopeId)
	}
}

func (keycloakClient *KeycloakClient) CreateRoleScopeMapping(ctx context.Context, realmId string, clientId string, clientScopeId string, role *Role) error {
	roleUrl := roleScopeMappingUrl(realmId, clientId, clientScopeId, role)

	_, _, err := keycloakClient.post(ctx, roleUrl, []Role{*role})
	if err != nil {
		return err
	}

	return nil
}

func (keycloakClient *KeycloakClient) GetRoleScopeMapping(ctx context.Context, realmId string, clientId string, clientScopeId string, role *Role) (*Role, error) {
	roleUrl := roleScopeMappingUrl(realmId, clientId, clientScopeId, role)
	var roles []Role

	err := keycloakClient.get(ctx, roleUrl, &roles, nil)
	if err != nil {
		return nil, err
	}

	for _, mappedRole := range roles {
		if mappedRole.Id == role.Id {
			return role, nil
		}
	}

	return nil, nil
}

func (keycloakClient *KeycloakClient) DeleteRoleScopeMapping(ctx context.Context, realmId string, clientId string, clientScopeId string, role *Role) error {
	roleUrl := roleScopeMappingUrl(realmId, clientId, clientScopeId, role)

	body := [1]RealmRoleRepresentation{
		{
			Id:          role.Id,
			Name:        role.Name,
			Description: role.Description,
			Composite:   role.Composite,
			ClientRole:  role.ClientRole,
			ContainerId: role.ContainerId,
		},
	}
	return keycloakClient.delete(ctx, roleUrl, body)
}
