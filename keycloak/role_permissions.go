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

// IsAdminPermissionsClient reports whether the client with the given internal
// id is the realm's admin-permissions resource server. A realm with admin
// permissions disabled has none, so the answer there is false, not an error.
func (keycloakClient *KeycloakClient) IsAdminPermissionsClient(ctx context.Context, realmId, id string) (bool, error) {
	var realm Realm
	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s", realmId), &realm, nil)
	if err != nil {
		return false, err
	}
	return realm.AdminPermissionsClient != nil && realm.AdminPermissionsClient.Id == id, nil
}
