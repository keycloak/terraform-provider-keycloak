package keycloak

import (
	"context"
	"fmt"
)

type GroupPermissionsInput struct {
	Enabled bool `json:"enabled"`
}

type GroupPermissions struct {
	RealmId          string            `json:"-"`
	GroupId          string            `json:"-"`
	Enabled          bool              `json:"enabled"`
	Resource         string            `json:"resource"`
	ScopePermissions map[string]string `json:"scopePermissions"`
}

func (keycloakClient *KeycloakClient) EnableGroupPermissions(ctx context.Context, realmId, groupId string) error {
	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/groups/%s/management/permissions", realmId, groupId), GroupPermissionsInput{Enabled: true})
}

func (keycloakClient *KeycloakClient) DisableGroupPermissions(ctx context.Context, realmId, groupId string) error {
	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/groups/%s/management/permissions", realmId, groupId), GroupPermissionsInput{Enabled: false})
}

func (keycloakClient *KeycloakClient) GetGroupPermissions(ctx context.Context, realmId, groupId string) (*GroupPermissions, error) {
	var groupPermissions GroupPermissions

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/groups/%s/management/permissions", realmId, groupId), &groupPermissions, nil)
	if err != nil {
		return nil, err
	}

	groupPermissions.RealmId = realmId
	groupPermissions.GroupId = groupId

	return &groupPermissions, nil
}
