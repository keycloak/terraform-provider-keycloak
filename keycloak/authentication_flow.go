package keycloak

import (
	"context"
	"fmt"
	"time"
)

type AuthenticationFlow struct {
	Id          string `json:"id,omitempty"`
	RealmId     string `json:"-"`
	Alias       string `json:"alias"`
	Description string `json:"description"`
	ProviderId  string `json:"providerId"` // "basic-flow" or "client-flow"
	TopLevel    bool   `json:"topLevel"`
	BuiltIn     bool   `json:"builtIn"`
}

func (keycloakClient *KeycloakClient) ListAuthenticationFlows(ctx context.Context, realmId string) ([]*AuthenticationFlow, error) {
	var authenticationFlows []*AuthenticationFlow

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/authentication/flows", realmId), &authenticationFlows, nil)
	if err != nil {
		return nil, err
	}

	for _, authenticationFlow := range authenticationFlows {
		authenticationFlow.RealmId = realmId
	}

	return authenticationFlows, nil
}

func (keycloakClient *KeycloakClient) NewAuthenticationFlow(ctx context.Context, authenticationFlow *AuthenticationFlow) error {
	authenticationFlow.TopLevel = true
	authenticationFlow.BuiltIn = false

	_, location, err := keycloakClient.post(ctx, fmt.Sprintf("/realms/%s/authentication/flows", authenticationFlow.RealmId), authenticationFlow)
	if err != nil {
		return err
	}
	authenticationFlow.Id = getIdFromLocationHeader(location)

	return nil
}

type authenticationFlowCopyRequest struct {
	NewName string `json:"newName"`
}

// CopyAuthenticationFlow duplicates an existing flow (including all of its executions and
// subflows) under a new alias. Unlike a built-in flow, the resulting copy is not builtIn and
// can be freely modified.
//
// The copy endpoint does not reliably return a Location header across Keycloak versions, so
// callers must look up the resulting flow by its new alias (e.g. via GetAuthenticationFlowFromAlias)
// rather than relying on a returned id.
func (keycloakClient *KeycloakClient) CopyAuthenticationFlow(ctx context.Context, realmId, flowAlias, newName string) error {
	_, _, err := keycloakClient.post(ctx, fmt.Sprintf("/realms/%s/authentication/flows/%s/copy", realmId, flowAlias), &authenticationFlowCopyRequest{NewName: newName})
	return err
}

func (keycloakClient *KeycloakClient) GetAuthenticationFlow(ctx context.Context, realmId, id string) (*AuthenticationFlow, error) {
	var authenticationFlow AuthenticationFlow
	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/authentication/flows/%s", realmId, id), &authenticationFlow, nil)
	if err != nil {
		return nil, err
	}

	authenticationFlow.RealmId = realmId
	return &authenticationFlow, nil
}

func (keycloakClient *KeycloakClient) GetAuthenticationFlowFromAlias(ctx context.Context, realmId, alias string) (*AuthenticationFlow, error) {
	var authenticationFlows []*AuthenticationFlow
	var authenticationFlow *AuthenticationFlow = nil

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/authentication/flows", realmId), &authenticationFlows, nil)
	if err != nil {
		return nil, err
	}

	// Retry 3 more times if not found, sometimes it took split milliseconds the Authentication to populate
	if len(authenticationFlows) == 0 {
		for i := 0; i < 3; i++ {
			err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/authentication/flows", realmId), &authenticationFlows, nil)

			if len(authenticationFlows) > 0 {
				break
			}

			if err != nil {
				return nil, err
			}

			time.Sleep(time.Millisecond * 50)
		}

		if len(authenticationFlows) == 0 {
			return nil, fmt.Errorf("no authentication flow found for alias %s", alias)
		}
	}

	for _, authFlow := range authenticationFlows {
		if authFlow.Alias == alias {
			authenticationFlow = authFlow
		}
	}

	if authenticationFlow == nil {
		return nil, fmt.Errorf("no authentication flow found for alias %s", alias)
	}
	authenticationFlow.RealmId = realmId

	return authenticationFlow, nil
}

func (keycloakClient *KeycloakClient) UpdateAuthenticationFlow(ctx context.Context, authenticationFlow *AuthenticationFlow) error {
	authenticationFlow.TopLevel = true
	authenticationFlow.BuiltIn = false

	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/authentication/flows/%s", authenticationFlow.RealmId, authenticationFlow.Id), authenticationFlow)
}

func (keycloakClient *KeycloakClient) DeleteAuthenticationFlow(ctx context.Context, realmId, id string) error {
	err := keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/authentication/flows/%s", realmId, id), nil)
	if err != nil {
		// For whatever reason, this fails sometimes with a 500 during acceptance tests. try again
		return keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/authentication/flows/%s", realmId, id), nil)
	}
	return nil
}
