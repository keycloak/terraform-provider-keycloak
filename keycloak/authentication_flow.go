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
	// A flow created moments ago (e.g. earlier in the same apply) can take a few milliseconds
	// to appear in the list, and the list is never empty in practice because realms ship with
	// built-in flows. Retry while the requested alias is absent before giving up.
	for attempt := 0; ; attempt++ {
		var authenticationFlows []*AuthenticationFlow

		err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/authentication/flows", realmId), &authenticationFlows, nil)
		if err != nil {
			return nil, err
		}

		for _, authFlow := range authenticationFlows {
			if authFlow.Alias == alias {
				authFlow.RealmId = realmId
				return authFlow, nil
			}
		}

		if attempt >= 3 {
			return nil, fmt.Errorf("no authentication flow found for alias %s", alias)
		}

		time.Sleep(time.Millisecond * 50)
	}
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
