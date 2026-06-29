package keycloak

import (
	"context"
	"encoding/json"
	"fmt"
)

type GenericClientAuthorizationPolicy struct {
	Id               string `json:"id,omitempty"`
	RealmId          string `json:"-"`
	ResourceServerId string `json:"-"`
	Name             string `json:"name"`
	DecisionStrategy string `json:"decisionStrategy"`
	Logic            string `json:"logic"`
	Type             string `json:"type"`
	Description      string `json:"description"`
}

func (keycloakClient *KeycloakClient) NewGenericClientAuthorizationPolicy(ctx context.Context, policy *GenericClientAuthorizationPolicy) error {
	// The generic policy endpoint resolves the provider from the "type" field in the body. This is
	// required for JAR-deployed policy providers (e.g. deployed JavaScript policies), whose type is
	// "script-<fileName>" and may contain characters such as "/" that cannot be put in the URL path.
	body, _, err := keycloakClient.post(ctx, fmt.Sprintf("/realms/%s/clients/%s/authz/resource-server/policy", policy.RealmId, policy.ResourceServerId), policy)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &policy)
	if err != nil {
		return err
	}
	return nil
}

func (keycloakClient *KeycloakClient) UpdateGenericClientAuthorizationPolicy(ctx context.Context, policy *GenericClientAuthorizationPolicy) error {
	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/clients/%s/authz/resource-server/policy/%s", policy.RealmId, policy.ResourceServerId, policy.Id), policy)
}

func (keycloakClient *KeycloakClient) DeleteGenericClientAuthorizationPolicy(ctx context.Context, realmId, resourceServerId, policyId string) error {
	return keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/clients/%s/authz/resource-server/policy/%s", realmId, resourceServerId, policyId), nil)
}

func (keycloakClient *KeycloakClient) GetGenericClientAuthorizationPolicy(ctx context.Context, realmId, resourceServerId, policyId string) (*GenericClientAuthorizationPolicy, error) {
	policy := GenericClientAuthorizationPolicy{
		Id:               policyId,
		ResourceServerId: resourceServerId,
		RealmId:          realmId,
	}
	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/clients/%s/authz/resource-server/policy/%s", realmId, resourceServerId, policyId), &policy, nil)
	if err != nil {
		return nil, err
	}

	return &policy, nil
}
