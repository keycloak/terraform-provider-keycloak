package keycloak

import (
	"context"
	"fmt"
)

func (keycloakClient *KeycloakClient) GetOpenidRealmDefaultDefaultClientScopes(ctx context.Context, realmId string) ([]*OpenidClientScope, error) {
	var clientScopes []*OpenidClientScope

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/default-default-client-scopes", realmId), &clientScopes, nil)
	if err != nil {
		return nil, err
	}

	return clientScopes, nil
}

func (keycloakClient *KeycloakClient) GetOpenidRealmDefaultDefaultClientScope(ctx context.Context, realmId, clientScopeId string) (*OpenidClientScope, error) {
	var clientScopes []OpenidClientScope

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/default-default-client-scopes", realmId), &clientScopes, nil)
	if err != nil {
		return nil, err
	}

	for _, clientScope := range clientScopes {
		if clientScope.Id == clientScopeId {
			return &clientScope, nil
		}
	}

	return nil, err
}

func (keycloakClient *KeycloakClient) PutOpenidRealmDefaultDefaultClientScope(ctx context.Context, realmId, clientScopeId string) error {
	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/default-default-client-scopes/%s", realmId, clientScopeId), nil)
}

func (keycloakClient *KeycloakClient) DeleteOpenidRealmDefaultDefaultClientScope(ctx context.Context, realmId, clientScopeId string) error {
	return keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/default-default-client-scopes/%s", realmId, clientScopeId), nil)
}
