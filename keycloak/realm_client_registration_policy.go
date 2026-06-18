package keycloak

import (
	"context"
	"fmt"
	"strings"
)

const realmClientRegistrationPolicyProviderType = "org.keycloak.services.clientregistration.policy.ClientRegistrationPolicy"

// MultiValueClientRegistrationConfigKeys are config fields that Keycloak stores as
// an array of individual values (map[string][]string). In Terraform they are
// expressed as a single comma-separated string, split into the array on write and
// re-joined on read. Keeping the whole comma string as one element would make
// Keycloak treat e.g. "roles,organization" as a single scope name and reject every
// dynamic client registration.
var MultiValueClientRegistrationConfigKeys = map[string]bool{
	"trusted-hosts":         true,
	"allowed-client-scopes": true,
}

type RealmClientRegistrationPolicy struct {
	Id         string
	Name       string
	RealmId    string
	ProviderId string
	SubType    string
	Config     map[string]string
}

func convertFromRealmClientRegistrationPolicyToComponent(policy *RealmClientRegistrationPolicy) *component {
	config := map[string][]string{}
	for k, v := range policy.Config {
		if MultiValueClientRegistrationConfigKeys[k] && strings.Contains(v, ",") {
			parts := strings.Split(v, ",")
			for i := range parts {
				parts[i] = strings.TrimSpace(parts[i])
			}
			config[k] = parts
		} else {
			config[k] = []string{v}
		}
	}

	return &component{
		Id:           policy.Id,
		Name:         policy.Name,
		ProviderId:   policy.ProviderId,
		ProviderType: realmClientRegistrationPolicyProviderType,
		ParentId:     policy.RealmId,
		SubType:      policy.SubType,
		Config:       config,
	}
}

func convertFromComponentToRealmClientRegistrationPolicy(c *component, realmId string) *RealmClientRegistrationPolicy {
	config := map[string]string{}
	for k, vals := range c.Config {
		if len(vals) == 0 {
			continue
		}

		if MultiValueClientRegistrationConfigKeys[k] && len(vals) > 1 {
			config[k] = strings.Join(vals, ",")
		} else {
			config[k] = vals[0]
		}
	}

	return &RealmClientRegistrationPolicy{
		Id:         c.Id,
		Name:       c.Name,
		RealmId:    realmId,
		ProviderId: c.ProviderId,
		SubType:    c.SubType,
		Config:     config,
	}
}

func (keycloakClient *KeycloakClient) NewRealmClientRegistrationPolicy(ctx context.Context, policy *RealmClientRegistrationPolicy) error {
	_, location, err := keycloakClient.post(ctx, fmt.Sprintf("/realms/%s/components", policy.RealmId), convertFromRealmClientRegistrationPolicyToComponent(policy))
	if err != nil {
		return err
	}

	policy.Id = getIdFromLocationHeader(location)

	return nil
}

func (keycloakClient *KeycloakClient) GetRealmClientRegistrationPolicy(ctx context.Context, realmId, id string) (*RealmClientRegistrationPolicy, error) {
	var component *component

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/components/%s", realmId, id), &component, nil)
	if err != nil {
		return nil, err
	}

	return convertFromComponentToRealmClientRegistrationPolicy(component, realmId), nil
}

func (keycloakClient *KeycloakClient) GetRealmClientRegistrationPolicies(ctx context.Context, realmId string) ([]*RealmClientRegistrationPolicy, error) {
	var components []*component

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/components?parent=%s&type=%s", realmId, realmId, realmClientRegistrationPolicyProviderType), &components, nil)
	if err != nil {
		return nil, err
	}

	policies := make([]*RealmClientRegistrationPolicy, 0, len(components))
	for _, c := range components {
		policies = append(policies, convertFromComponentToRealmClientRegistrationPolicy(c, realmId))
	}

	return policies, nil
}

func (keycloakClient *KeycloakClient) UpdateRealmClientRegistrationPolicy(ctx context.Context, policy *RealmClientRegistrationPolicy) error {
	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/components/%s", policy.RealmId, policy.Id), convertFromRealmClientRegistrationPolicyToComponent(policy))
}

func (keycloakClient *KeycloakClient) DeleteRealmClientRegistrationPolicy(ctx context.Context, realmId, id string) error {
	return keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/components/%s", realmId, id), nil)
}
