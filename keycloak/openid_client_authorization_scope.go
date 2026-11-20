package keycloak

import (
	"context"
	"encoding/json"
	"fmt"
)

type OpenidClientAuthorizationScope struct {
	Id               string `json:"id,omitempty"`
	RealmId          string `json:"-"`
	ResourceServerId string `json:"-"`
	Name             string `json:"name"`
	DisplayName      string `json:"displayName"`
	IconUri          string `json:"iconUri"`
}

func (keycloakClient *KeycloakClient) NewOpenidClientAuthorizationScope(ctx context.Context, scope *OpenidClientAuthorizationScope) error {
	body, _, err := keycloakClient.post(ctx, fmt.Sprintf("/realms/%s/clients/%s/authz/resource-server/scope", scope.RealmId, scope.ResourceServerId), scope)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &scope)
	if err != nil {
		return err
	}
	return nil
}

func (keycloakClient *KeycloakClient) GetOpenidClientAuthorizationScope(ctx context.Context, realm, resourceServerId, scopeId string) (*OpenidClientAuthorizationScope, error) {
	scope := OpenidClientAuthorizationScope{
		RealmId:          realm,
		ResourceServerId: resourceServerId,
	}
	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/clients/%s/authz/resource-server/scope/%s", realm, resourceServerId, scopeId), &scope, nil)
	if err != nil {
		return nil, err
	}
	return &scope, nil
}

func (keycloakClient *KeycloakClient) UpdateOpenidClientAuthorizationScope(ctx context.Context, scope *OpenidClientAuthorizationScope) error {
	err := keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/clients/%s/authz/resource-server/scope/%s", scope.RealmId, scope.ResourceServerId, scope.Id), scope)
	if err != nil {
		return err
	}
	return nil
}

func (keycloakClient *KeycloakClient) GetOpenidClientAuthorizationScopeByName(ctx context.Context, realm, resourceServerId, name string) (*OpenidClientAuthorizationScope, error) {
	var scopes []OpenidClientAuthorizationScope
	params := map[string]string{"name": name}
	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/clients/%s/authz/resource-server/scope", realm, resourceServerId), &scopes, params)
	if err != nil {
		return nil, err
	}
	for _, scope := range scopes {
		if scope.Name == name {
			scope.RealmId = realm
			scope.ResourceServerId = resourceServerId
			return &scope, nil
		}
	}
	return nil, fmt.Errorf("no authorization scope with name %s found", name)
}

func (keycloakClient *KeycloakClient) DeleteOpenidClientAuthorizationScope(ctx context.Context, realmId, resourceServerId, scopeId string) error {
	return keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/clients/%s/authz/resource-server/scope/%s", realmId, resourceServerId, scopeId), nil)
}
