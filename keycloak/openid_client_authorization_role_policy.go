package keycloak

import (
	"context"
	"encoding/json"
	"fmt"
)

type OpenidClientAuthorizationRolePolicy struct {
	Id               string                          `json:"id,omitempty"`
	RealmId          string                          `json:"-"`
	ResourceServerId string                          `json:"-"`
	Name             string                          `json:"name"`
	DecisionStrategy string                          `json:"decisionStrategy"`
	Logic            string                          `json:"logic"`
	Type             string                          `json:"type"`
	Roles            []OpenidClientAuthorizationRole `json:"roles,omitempty"`
	Description      string                          `json:"description"`
	FetchRoles       bool                            `json:"fetchRoles,omitempty"`
}

type OpenidClientAuthorizationRole struct {
	Id       string `json:"id,omitempty"`
	Required bool   `json:"required"`
}

func (keycloakClient *KeycloakClient) NewOpenidClientAuthorizationRolePolicy(ctx context.Context, policy *OpenidClientAuthorizationRolePolicy) error {
	body, _, err := keycloakClient.post(ctx, fmt.Sprintf("/realms/%s/clients/%s/authz/resource-server/policy/role", policy.RealmId, policy.ResourceServerId), policy)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &policy)
	if err != nil {
		return err
	}
	return nil
}

func (keycloakClient *KeycloakClient) UpdateOpenidClientAuthorizationRolePolicy(ctx context.Context, policy *OpenidClientAuthorizationRolePolicy) error {
	err := keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/clients/%s/authz/resource-server/policy/role/%s", policy.RealmId, policy.ResourceServerId, policy.Id), policy)
	if err != nil {
		return err
	}
	return nil
}

func (keycloakClient *KeycloakClient) DeleteOpenidClientAuthorizationRolePolicy(ctx context.Context, realmId, resourceServerId, policyId string) error {
	return keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/clients/%s/authz/resource-server/policy/role/%s", realmId, resourceServerId, policyId), nil)
}

func (keycloakClient *KeycloakClient) GetOpenidClientAuthorizationRolePolicy(ctx context.Context, realmId, resourceServerId, policyId string) (*OpenidClientAuthorizationRolePolicy, error) {

	policy := OpenidClientAuthorizationRolePolicy{
		Id:               policyId,
		ResourceServerId: resourceServerId,
		RealmId:          realmId,
	}
	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/clients/%s/authz/resource-server/policy/role/%s", realmId, resourceServerId, policyId), &policy, nil)
	if err != nil {
		return nil, err
	}

	return &policy, nil
}
