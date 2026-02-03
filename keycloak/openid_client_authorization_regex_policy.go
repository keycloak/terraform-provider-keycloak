package keycloak

import (
	"context"
	"encoding/json"
	"fmt"
)

type OpenidClientAuthorizationRegexPolicy struct {
	Id                      string `json:"id,omitempty"`
	RealmId                 string `json:"-"`
	ResourceServerId        string `json:"-"`
	Name                    string `json:"name"`
	DecisionStrategy        string `json:"decisionStrategy"`
	Logic                   string `json:"logic"`
	Type                    string `json:"type"`
	Pattern                 string `json:"pattern"`
	Description             string `json:"description"`
	TargetClaim             string `json:"targetClaim"`
	TargetContextAttributes bool   `json:"targetContextAttributes"`
}

func (keycloakClient *KeycloakClient) NewOpenidClientAuthorizationRegexPolicy(ctx context.Context, policy *OpenidClientAuthorizationRegexPolicy) error {
	body, _, err := keycloakClient.post(ctx, fmt.Sprintf("/realms/%s/clients/%s/authz/resource-server/policy/regex", policy.RealmId, policy.ResourceServerId), policy)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &policy)
	if err != nil {
		return err
	}
	return nil
}

func (keycloakClient *KeycloakClient) UpdateOpenidClientAuthorizationRegexPolicy(ctx context.Context, policy *OpenidClientAuthorizationRegexPolicy) error {
	err := keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/clients/%s/authz/resource-server/policy/regex/%s", policy.RealmId, policy.ResourceServerId, policy.Id), policy)
	if err != nil {
		return err
	}
	return nil
}

func (keycloakClient *KeycloakClient) DeleteOpenidClientAuthorizationRegexPolicy(ctx context.Context, realmId, resourceServerId, policyId string) error {
	return keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/clients/%s/authz/resource-server/policy/regex/%s", realmId, resourceServerId, policyId), nil)
}

func (keycloakClient *KeycloakClient) GetOpenidClientAuthorizationRegexPolicy(ctx context.Context, realmId, resourceServerId, policyId string) (*OpenidClientAuthorizationRegexPolicy, error) {

	policy := OpenidClientAuthorizationRegexPolicy{
		Id:               policyId,
		ResourceServerId: resourceServerId,
		RealmId:          realmId,
	}
	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/clients/%s/authz/resource-server/policy/regex/%s", realmId, resourceServerId, policyId), &policy, nil)
	if err != nil {
		return nil, err
	}

	return &policy, nil
}
