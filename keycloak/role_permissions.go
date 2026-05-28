package keycloak

import (
	"context"
	"fmt"
)

type RolePermissions struct {
	RealmId string
	RoleId  string
	Enabled bool
}

// GetAdminPermissionsClientId returns the ID of the realm's admin-permissions
// resource server, which Keycloak creates when adminPermissionsEnabled = true.
func (keycloakClient *KeycloakClient) GetAdminPermissionsClientId(ctx context.Context, realmId string) (string, error) {
	var realm Realm
	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s", realmId), &realm, nil)
	if err != nil {
		return "", err
	}
	if realm.AdminPermissionsClient == nil {
		return "", fmt.Errorf("realm %s does not have an admin-permissions client; ensure adminPermissionsEnabled = true on the realm", realmId)
	}
	return realm.AdminPermissionsClient.Id, nil
}

// GetRolePermissions returns the fine-grained permissions state for a role.
// In FGAPv2, permissions are enabled implicitly when the realm has
// adminPermissionsEnabled = true; this simply confirms that state.
func (keycloakClient *KeycloakClient) GetRolePermissions(ctx context.Context, realmId, roleId string) (*RolePermissions, error) {
	var realm Realm
	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s", realmId), &realm, nil)
	if err != nil {
		return nil, err
	}
	return &RolePermissions{
		RealmId: realmId,
		RoleId:  roleId,
		Enabled: realm.AdminPermissionsEnabled,
	}, nil
}
