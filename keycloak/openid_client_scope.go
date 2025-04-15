package keycloak

import (
	"context"
	"fmt"
	"github.com/keycloak/terraform-provider-keycloak/keycloak/types"
)

type OpenidClientScope struct {
	Id          string `json:"id,omitempty"`
	RealmId     string `json:"-"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Protocol    string `json:"protocol"`
	Attributes  struct {
		DisplayOnConsentScreen types.KeycloakBoolQuoted `json:"display.on.consent.screen"` // boolean in string form
		ConsentScreenText      string                   `json:"consent.screen.text"`
		GuiOrder               string                   `json:"gui.order"`
		IncludeInTokenScope    types.KeycloakBoolQuoted `json:"include.in.token.scope"` // boolean in string form
	} `json:"attributes"`
}

type OpenidClientScopeFilterFunc func(*OpenidClientScope) bool

func (keycloakClient *KeycloakClient) NewOpenidClientScope(ctx context.Context, clientScope *OpenidClientScope) error {
	clientScope.Protocol = "openid-connect"

	_, location, err := keycloakClient.post(ctx, fmt.Sprintf("/realms/%s/client-scopes", clientScope.RealmId), clientScope)
	if err != nil {
		return err
	}

	clientScope.Id = getIdFromLocationHeader(location)

	return nil
}

func (keycloakClient *KeycloakClient) GetOpenidClientScope(ctx context.Context, realmId, id string) (*OpenidClientScope, error) {
	var clientScope OpenidClientScope

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/client-scopes/%s", realmId, id), &clientScope, nil)
	if err != nil {
		return nil, err
	}

	clientScope.RealmId = realmId

	return &clientScope, nil
}

func (keycloakClient *KeycloakClient) GetOpenidDefaultClientScopes(ctx context.Context, realmId, clientId string) (*[]OpenidClientScope, error) {
	var clientScopes []OpenidClientScope

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/clients/%s/default-client-scopes", realmId, clientId), &clientScopes, nil)
	if err != nil {
		return nil, err
	}

	for _, clientScope := range clientScopes {
		clientScope.RealmId = realmId
	}

	return &clientScopes, nil
}

func (keycloakClient *KeycloakClient) GetOpenidOptionalClientScopes(ctx context.Context, realmId, clientId string) (*[]OpenidClientScope, error) {
	var clientScopes []OpenidClientScope

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/clients/%s/optional-client-scopes", realmId, clientId), &clientScopes, nil)
	if err != nil {
		return nil, err
	}

	for _, clientScope := range clientScopes {
		clientScope.RealmId = realmId
	}

	return &clientScopes, nil
}

func (keycloakClient *KeycloakClient) UpdateOpenidClientScope(ctx context.Context, clientScope *OpenidClientScope) error {
	clientScope.Protocol = "openid-connect"

	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/client-scopes/%s", clientScope.RealmId, clientScope.Id), clientScope)
}

func (keycloakClient *KeycloakClient) DeleteOpenidClientScope(ctx context.Context, realmId, id string) error {
	return keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/client-scopes/%s", realmId, id), nil)
}

func (keycloakClient *KeycloakClient) ListOpenidClientScopesWithFilter(ctx context.Context, realmId string, filter OpenidClientScopeFilterFunc) ([]*OpenidClientScope, error) {
	var clientScopes []OpenidClientScope
	var openidClientScopes []*OpenidClientScope

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/client-scopes", realmId), &clientScopes, nil)
	if err != nil {
		return nil, err
	}

	for _, clientScope := range clientScopes {
		if clientScope.Protocol == "openid-connect" && filter(&clientScope) {
			scope := new(OpenidClientScope)
			*scope = clientScope

			scope.RealmId = realmId

			openidClientScopes = append(openidClientScopes, scope)
		}
	}

	return openidClientScopes, nil
}

func IncludeOpenidClientScopesMatchingNames(scopeNames []string) OpenidClientScopeFilterFunc {
	return func(scope *OpenidClientScope) bool {
		for _, scopeName := range scopeNames {
			if scopeName == scope.Name {
				return true
			}
		}

		return false
	}
}
