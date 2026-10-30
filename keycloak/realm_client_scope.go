package keycloak

import (
	"context"
	"errors"
	"fmt"
)

func (keycloakClient *KeycloakClient) getRealmClientScopesOfType(ctx context.Context, realmId, t string) ([]*OpenidClientScope, error) {
	var scopes []*OpenidClientScope

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/default-%s-client-scopes", realmId, t), &scopes, nil)
	if err != nil {
		return nil, err
	}

	return scopes, nil
}

func (keycloakClient *KeycloakClient) GetRealmDefaultClientScopes(ctx context.Context, realmId string) ([]*OpenidClientScope, error) {
	return keycloakClient.getRealmClientScopesOfType(ctx, realmId, "default")
}

func (keycloakClient *KeycloakClient) GetRealmOptionalClientScopes(ctx context.Context, realmId string) ([]*OpenidClientScope, error) {
	return keycloakClient.getRealmClientScopesOfType(ctx, realmId, "optional")
}

func (keycloakClient *KeycloakClient) resolveClientScopeNamesIntoIds(ctx context.Context, realmId string, scopeNames []string) ([]string, error) {
	var scopeIds []string
	var clientScopes []OpenidClientScope

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/client-scopes", realmId), &clientScopes, nil)
	if err != nil {
		return nil, err
	}

ScopeNames:
	for _, scopeName := range scopeNames {
		for _, clientScope := range clientScopes {
			if clientScope.Name == scopeName {
				scopeIds = append(scopeIds, clientScope.Id)
				continue ScopeNames
			}
		}

		return nil, errors.New(fmt.Sprintf("Client scope with name %s not found in realm %s", scopeName, realmId))
	}

	return scopeIds, nil
}

func (keycloakClient *KeycloakClient) markClientScopesAs(ctx context.Context, realmId string, scopeNames []string, t string) error {
	scopeIds, err := keycloakClient.resolveClientScopeNamesIntoIds(ctx, realmId, scopeNames)
	if err != nil {
		return err
	}

	currentScopes, err := keycloakClient.getRealmClientScopesOfType(ctx, realmId, t)
	if err != nil {
		return err
	}

	// Keycloak does not allow a client scope to be a realm default and a realm optional scope at the same time, so
	// remove it from the conflicting type first, mirroring the behavior of the Keycloak admin console.
	conflictingType := otherScopeType(t)
	conflictingScopes, err := keycloakClient.getRealmClientScopesOfType(ctx, realmId, conflictingType)
	if err != nil {
		return err
	}

	marked := make(map[string]bool, len(currentScopes))
	for _, s := range currentScopes {
		marked[s.Id] = true
	}
	conflicting := make(map[string]bool, len(conflictingScopes))
	for _, s := range conflictingScopes {
		conflicting[s.Id] = true
	}

	for _, scopeId := range scopeIds {
		if marked[scopeId] {
			continue
		}

		// Keycloak ignores the type on DELETE, so only unmark scopes that are marked as the conflicting type.
		// Unmarking any other scope would unmark it whatever its type.
		if conflicting[scopeId] {
			if err := keycloakClient.unmarkClientScopeAs(ctx, realmId, scopeId, conflictingType); err != nil {
				return err
			}
		}

		if err := keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/default-%s-client-scopes/%s", realmId, t, scopeId), nil); err != nil {
			return err
		}
	}

	return nil
}

func (keycloakClient *KeycloakClient) MarkClientScopesAsRealmDefault(ctx context.Context, realmId string, scopeNames []string) error {
	return keycloakClient.markClientScopesAs(ctx, realmId, scopeNames, "default")
}

func (keycloakClient *KeycloakClient) MarkClientScopesAsRealmOptional(ctx context.Context, realmId string, scopeNames []string) error {
	return keycloakClient.markClientScopesAs(ctx, realmId, scopeNames, "optional")
}

func (keycloakClient *KeycloakClient) unmarkClientScopeAs(ctx context.Context, realmId, scopeId, t string) error {
	err := keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/default-%s-client-scopes/%s", realmId, t, scopeId), nil)
	if err != nil && !ErrorIs404(err) {
		return err
	}

	return nil
}

// unmarkClientScopesAs only unmarks scopes that are currently marked as the given type. Keycloak ignores the type on
// DELETE and removes the assignment whatever its type, so this avoids undoing a scope that has since moved to the other
// type.
func (keycloakClient *KeycloakClient) unmarkClientScopesAs(ctx context.Context, realmId string, scopeNames []string, t string) error {
	currentScopes, err := keycloakClient.getRealmClientScopesOfType(ctx, realmId, t)
	if ErrorIs404(err) {
		return nil
	} else if err != nil {
		return err
	}

	names := make(map[string]bool, len(scopeNames))
	for _, name := range scopeNames {
		names[name] = true
	}

	for _, s := range currentScopes {
		if names[s.Name] {
			if err := keycloakClient.unmarkClientScopeAs(ctx, realmId, s.Id, t); err != nil {
				return err
			}
		}
	}

	return nil
}

func (keycloakClient *KeycloakClient) UnmarkClientScopesAsRealmDefault(ctx context.Context, realmId string, scopeNames []string) error {
	return keycloakClient.unmarkClientScopesAs(ctx, realmId, scopeNames, "default")
}

func (keycloakClient *KeycloakClient) UnmarkClientScopesAsRealmOptional(ctx context.Context, realmId string, scopeNames []string) error {
	return keycloakClient.unmarkClientScopesAs(ctx, realmId, scopeNames, "optional")
}
