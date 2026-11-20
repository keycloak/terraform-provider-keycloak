package keycloak

import (
	"context"
	"fmt"
)

type OpenidClientAuthorizationPolicy struct {
	Id               string   `json:"id,omitempty"`
	RealmId          string   `json:"-"`
	ResourceServerId string   `json:"-"`
	Name             string   `json:"name"`
	Owner            string   `json:"owner"`
	DecisionStrategy string   `json:"decisionStrategy"`
	Logic            string   `json:"logic"`
	Policies         []string `json:"policies"`
	Resources        []string `json:"resources"`
	Scopes           []string `json:"scopes"`
	Type             string   `json:"type"`
}

func (keycloakClient *KeycloakClient) GetClientAuthorizationPolicyByName(ctx context.Context, realmId, resourceServerId, name string) (*OpenidClientAuthorizationPolicy, error) {
	var policies []OpenidClientAuthorizationPolicy
	params := map[string]string{"name": name}
	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/clients/%s/authz/resource-server/policy", realmId, resourceServerId), &policies, params)
	if err != nil {
		return nil, err
	}
	if len(policies) == 0 {
		return nil, fmt.Errorf("unable to find client authorization policy with name %s", name)
	}
	policy := policies[0]
	policy.RealmId = realmId
	policy.ResourceServerId = resourceServerId
	policy.Name = name
	return &policy, nil
}
