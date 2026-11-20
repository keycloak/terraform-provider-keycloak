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

func (keycloakClient *KeycloakClient) resolveAndHandleClientScopes(ctx context.Context, realmId string, scopeNames []string, handler func(context.Context, string, string) error) error {
	scopeIds, err := keycloakClient.resolveClientScopeNamesIntoIds(ctx, realmId, scopeNames)
	if err != nil {
		return err
	}

	for _, scopeId := range scopeIds {
		if err := handler(ctx, realmId, scopeId); err != nil {
			return err
		}
	}

	return nil
}

func (keycloakClient *KeycloakClient) markClientScopeAs(ctx context.Context, realmId, scopeId, t string) error {
	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/default-%s-client-scopes/%s", realmId, t, scopeId), nil)
}

func (keycloakClient *KeycloakClient) MarkClientScopesAsRealmDefault(ctx context.Context, realmId string, scopeNames []string) error {
	return keycloakClient.resolveAndHandleClientScopes(ctx, realmId, scopeNames, func(ctx context.Context, realmId, scopeId string) error {
		return keycloakClient.markClientScopeAs(ctx, realmId, scopeId, "default")
	})
}

func (keycloakClient *KeycloakClient) MarkClientScopesAsRealmOptional(ctx context.Context, realmId string, scopeNames []string) error {
	return keycloakClient.resolveAndHandleClientScopes(ctx, realmId, scopeNames, func(ctx context.Context, realmId, scopeId string) error {
		return keycloakClient.markClientScopeAs(ctx, realmId, scopeId, "optional")
	})
}

func (keycloakClient *KeycloakClient) unmarkClientScopeAs(ctx context.Context, realmId, scopeId, t string) error {
	return keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/default-%s-client-scopes/%s", realmId, t, scopeId), nil)
}

func (keycloakClient *KeycloakClient) UnmarkClientScopesAsRealmDefault(ctx context.Context, realmId string, scopeNames []string) error {
	return keycloakClient.resolveAndHandleClientScopes(ctx, realmId, scopeNames, func(ctx context.Context, realmId, scopeId string) error {
		return keycloakClient.unmarkClientScopeAs(ctx, realmId, scopeId, "default")
	})
}

func (keycloakClient *KeycloakClient) UnmarkClientScopesAsRealmOptional(ctx context.Context, realmId string, scopeNames []string) error {
	return keycloakClient.resolveAndHandleClientScopes(ctx, realmId, scopeNames, func(ctx context.Context, realmId, scopeId string) error {
		return keycloakClient.unmarkClientScopeAs(ctx, realmId, scopeId, "optional")
	})
}
