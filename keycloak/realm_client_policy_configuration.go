package keycloak

import (
	"context"
	"encoding/json"
	"fmt"
)

// Preserve JSON types and fields (including policy mode) without the older
// per-entry resources' lossy configuration string-map conversion.
type RealmClientPolicyConfiguration struct {
	Profiles       []json.RawMessage
	Policies       []json.RawMessage
	GlobalProfiles []json.RawMessage
	GlobalPolicies []json.RawMessage
}

func (c *KeycloakClient) GetRealmClientPolicyConfiguration(ctx context.Context, realm string) (*RealmClientPolicyConfiguration, error) {
	var profiles struct {
		Profiles       []json.RawMessage `json:"profiles"`
		GlobalProfiles []json.RawMessage `json:"globalProfiles"`
	}
	var policies struct {
		Policies       []json.RawMessage `json:"policies"`
		GlobalPolicies []json.RawMessage `json:"globalPolicies"`
	}
	if err := c.get(ctx, fmt.Sprintf("/realms/%s/client-policies/profiles", realm), &profiles, map[string]string{"include-global-profiles": "true"}); err != nil {
		return nil, err
	}
	if err := c.get(ctx, fmt.Sprintf("/realms/%s/client-policies/policies", realm), &policies, map[string]string{"include-global-policies": "true"}); err != nil {
		return nil, err
	}
	return &RealmClientPolicyConfiguration{profiles.Profiles, policies.Policies, profiles.GlobalProfiles, policies.GlobalPolicies}, nil
}

func (c *KeycloakClient) PutRealmClientPolicyCollection(ctx context.Context, realm, collection string, entries []json.RawMessage) error {
	if collection != "profiles" && collection != "policies" {
		return fmt.Errorf("invalid client-policy collection")
	}
	if entries == nil {
		entries = []json.RawMessage{}
	}
	// Global collections are server-owned and must never be included in a PUT.
	return c.put(ctx, fmt.Sprintf("/realms/%s/client-policies/%s", realm, collection), map[string]interface{}{collection: entries})
}
