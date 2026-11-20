package keycloak

import (
	"context"
	"fmt"
)

type RealmClientPolicyProfilePolicyCondition struct {
	Name          string                 `json:"condition"`
	Configuration map[string]interface{} `json:"configuration"`
}

type RealmClientPolicyProfilePolicy struct {
	Name        string                                    `json:"name"`
	RealmId     string                                    `json:"-"`
	Description string                                    `json:"description"`
	Enabled     bool                                      `json:"enabled"`
	Conditions  []RealmClientPolicyProfilePolicyCondition `json:"conditions"`
	Profiles    []string                                  `json:"profiles"`
}

type RealmClientPolicyProfilePolicies struct {
	Policies []RealmClientPolicyProfilePolicy `json:"policies"`
}

func (keycloakClient *KeycloakClient) UpdateRealmClientPolicyProfilePolicies(ctx context.Context, realmId string, policies *RealmClientPolicyProfilePolicies) error {
	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/client-policies/policies", realmId), policies)
}

func (keycloakClient *KeycloakClient) GetAllRealmClientPolicyProfilePolices(ctx context.Context, realmId string) (*RealmClientPolicyProfilePolicies, error) {
	var realmClientPolicyProfilePolicies *RealmClientPolicyProfilePolicies

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/client-policies/policies", realmId), &realmClientPolicyProfilePolicies, nil)
	if err != nil {
		return nil, err
	}

	return realmClientPolicyProfilePolicies, nil
}

func (keycloakClient *KeycloakClient) GetRealmClientPolicyProfilePolicyByName(ctx context.Context, realmId string, name string) (*RealmClientPolicyProfilePolicy, error) {
	realmClientPolicyProfilePolicies, err := keycloakClient.GetAllRealmClientPolicyProfilePolices(ctx, realmId)
	if err != nil {
		return nil, err
	}

	for _, policy := range realmClientPolicyProfilePolicies.Policies {
		if policy.Name == name {
			return &policy, nil
		}
	}

	return nil, fmt.Errorf("policy with name: %s not found", name)
}
