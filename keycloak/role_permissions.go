package keycloak

import (
	"context"
	"fmt"
)

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
