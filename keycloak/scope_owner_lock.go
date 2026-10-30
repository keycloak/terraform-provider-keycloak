package keycloak

import "fmt"

// Scope owners are the entities that hold default/optional client scope assignments. The default and optional scope
// resources of the same owner modify overlapping server-side state, so their reconciliation must be serialized.

func RealmScopeOwnerKey(realmId string) string {
	return fmt.Sprintf("scopes/realm/%s", realmId)
}

func OpenidClientScopeOwnerKey(realmId, clientId string) string {
	return fmt.Sprintf("scopes/openid-client/%s/%s", realmId, clientId)
}

func SamlClientScopeOwnerKey(realmId, clientId string) string {
	return fmt.Sprintf("scopes/saml-client/%s/%s", realmId, clientId)
}

// LockScopeOwner blocks until the lock for the given scope owner is acquired and returns the function releasing it.
func (keycloakClient *KeycloakClient) LockScopeOwner(ownerKey string) func() {
	keycloakClient.Mutex.Lock(ownerKey)

	return func() {
		keycloakClient.Mutex.Unlock(ownerKey)
	}
}
