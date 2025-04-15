package keycloak

import (
	"context"
	"encoding/json"
	"fmt"
)

type OpenidClientAuthorizationClientScopePolicy struct {
	Id               string                                 `json:"id,omitempty"`
	RealmId          string                                 `json:"-"`
	ResourceServerId string                                 `json:"-"`
	Name             string                                 `json:"name"`
	DecisionStrategy string                                 `json:"decisionStrategy"`
	Logic            string                                 `json:"logic"`
	Type             string                                 `json:"type"`
	Scope            []OpenidClientAuthorizationClientScope `json:"clientScopes"`
	Description      string                                 `json:"description"`
}

type OpenidClientAuthorizationClientScope struct {
	Id       string `json:"id,omitempty"`
	Required bool   `json:"required,omitempty"`
}

func (keycloakClient *KeycloakClient) NewOpenidClientAuthorizationClientScopePolicy(ctx context.Context, policy *OpenidClientAuthorizationClientScopePolicy) error {
	body, _, err := keycloakClient.post(ctx, fmt.Sprintf("/realms/%s/clients/%s/authz/resource-server/policy/client-scope", policy.RealmId, policy.ResourceServerId), policy)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &policy)
	if err != nil {
		return err
	}
	return nil
}

func (keycloakClient *KeycloakClient) UpdateOpenidClientAuthorizationClientScopePolicy(ctx context.Context, policy *OpenidClientAuthorizationClientScopePolicy) error {
	err := keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/clients/%s/authz/resource-server/policy/client-scope/%s", policy.RealmId, policy.ResourceServerId, policy.Id), policy)
	if err != nil {
		return err
	}
	return nil
}

func (keycloakClient *KeycloakClient) DeleteOpenidClientAuthorizationClientScopePolicy(ctx context.Context, realmId, resourceServerId, policyId string) error {
	return keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/clients/%s/authz/resource-server/policy/client-scope/%s", realmId, resourceServerId, policyId), nil)
}

func (keycloakClient *KeycloakClient) GetOpenidClientAuthorizationClientScopePolicy(ctx context.Context, realmId, resourceServerId, policyId string) (*OpenidClientAuthorizationClientScopePolicy, error) {

	policy := OpenidClientAuthorizationClientScopePolicy{
		Id:               policyId,
		ResourceServerId: resourceServerId,
		RealmId:          realmId,
	}
	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/clients/%s/authz/resource-server/policy/client-scope/%s", realmId, resourceServerId, policyId), &policy, nil)
	if err != nil {
		return nil, err
	}

	return &policy, nil
}
