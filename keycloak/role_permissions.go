package keycloak

import (
	"context"
	"fmt"
)

type RolePermissionsInput struct {
	Enabled bool `json:"enabled"`
}

type RolePermissions struct {
	RealmId          string            `json:"-"`
	RoleId           string            `json:"-"`
	Enabled          bool              `json:"enabled"`
	Resource         string            `json:"resource"`
	ScopePermissions map[string]string `json:"scopePermissions"`
}

func (keycloakClient *KeycloakClient) EnableRolePermissions(ctx context.Context, realmId, roleId string) error {
	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/roles-by-id/%s/management/permissions", realmId, roleId), RolePermissionsInput{Enabled: true})
}

func (keycloakClient *KeycloakClient) DisableRolePermissions(ctx context.Context, realmId, roleId string) error {
	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/roles-by-id/%s/management/permissions", realmId, roleId), RolePermissionsInput{Enabled: false})
}

func (keycloakClient *KeycloakClient) GetRolePermissions(ctx context.Context, realmId, roleId string) (*RolePermissions, error) {
	var rolePermissions RolePermissions

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/roles-by-id/%s/management/permissions", realmId, roleId), &rolePermissions, nil)
	if err != nil {
		return nil, err
	}

	rolePermissions.RealmId = realmId
	rolePermissions.RoleId = roleId

	return &rolePermissions, nil
}
