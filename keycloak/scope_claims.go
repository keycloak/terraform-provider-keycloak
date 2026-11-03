package keycloak

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// scopeClaims records, for the lifetime of a provider instance, which scopes the default and optional scope
// resources of a scope owner want to reconcile. This lets the resources detect that they were configured with the same
// scope, which they cannot see from their own configuration.
type scopeClaims struct {
	lock   sync.Mutex
	claims map[string]map[string]map[string]bool // owner key -> scope type -> scope names
}

func otherScopeType(t string) string {
	if t == "default" {
		return "optional"
	}

	return "default"
}

// ClaimScopes records the scopes of the given type ("default" or "optional") that the caller reconciles for the scope
// owner, replacing any previous claim of that type. It returns an error, without recording the claim, if the other
// type has claimed any of the same scopes.
func (keycloakClient *KeycloakClient) ClaimScopes(ownerKey, t string, scopeNames []string) error {
	c := keycloakClient.scopeClaims
	c.lock.Lock()
	defer c.lock.Unlock()

	var conflicts []string
	for _, name := range scopeNames {
		if c.claims[ownerKey][otherScopeType(t)][name] {
			conflicts = append(conflicts, name)
		}
	}

	if len(conflicts) > 0 {
		sort.Strings(conflicts)

		return fmt.Errorf("validation error: scopes [%s] are configured as both default and optional scopes for %s", strings.Join(conflicts, ", "), strings.TrimPrefix(ownerKey, "scopes/"))
	}

	if c.claims == nil {
		c.claims = make(map[string]map[string]map[string]bool)
	}
	if c.claims[ownerKey] == nil {
		c.claims[ownerKey] = make(map[string]map[string]bool)
	}

	names := make(map[string]bool, len(scopeNames))
	for _, name := range scopeNames {
		names[name] = true
	}
	c.claims[ownerKey][t] = names

	return nil
}

// ForNewRun returns a copy of keycloakClient that shares everything with it, including the connection,
// authentication, locks and server caches, except for the scope claims, which are empty. Scope claims must only ever reflect a single Terraform run, which a provider that reuses a client
// across runs gets by configuring each run with ForNewRun.
func (keycloakClient *KeycloakClient) ForNewRun() *KeycloakClient {
	c := *keycloakClient
	c.scopeClaims = &scopeClaims{}

	return &c
}

// ReleaseScopes removes the claim of the given type for the scope owner.
func (keycloakClient *KeycloakClient) ReleaseScopes(ownerKey, t string) {
	c := keycloakClient.scopeClaims
	c.lock.Lock()
	defer c.lock.Unlock()

	delete(c.claims[ownerKey], t)
}
