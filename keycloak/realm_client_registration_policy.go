package keycloak

import (
	"context"
	"fmt"
	"strings"
)

const realmClientRegistrationPolicyProviderType = "org.keycloak.services.clientregistration.policy.ClientRegistrationPolicy"

type RealmClientRegistrationPolicy struct {
	Id         string
	Name       string
	RealmId    string
	ProviderId string
	SubType    string
	Config     map[string]string
}

// MultiValuedClientRegistrationConfigKeys returns the config keys that Keycloak models as
// multi-valued for the given client registration policy provider. Keycloak stores these as
// an array of individual values (map[string][]string), while Terraform expresses them as a
// single comma-separated string that is split into the array on write and re-joined on read.
//
// The set is derived at runtime from the server info component-type metadata (properties with
// a "type" of "MultivaluedString" or "MultivaluedList"), so built-in and custom SPI policies
// are handled the same way without a hardcoded list.
func (keycloakClient *KeycloakClient) MultiValuedClientRegistrationConfigKeys(ctx context.Context, providerId string) (map[string]bool, error) {
	serverInfo, err := keycloakClient.GetServerInfoCached(ctx)
	if err != nil {
		return nil, err
	}

	return serverInfo.MultiValuedConfigKeys(realmClientRegistrationPolicyProviderType, providerId), nil
}

func convertFromRealmClientRegistrationPolicyToComponent(policy *RealmClientRegistrationPolicy, multiValuedKeys map[string]bool) *component {
	config := map[string][]string{}
	for k, v := range policy.Config {
		if multiValuedKeys[k] && strings.Contains(v, ",") {
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

func convertFromComponentToRealmClientRegistrationPolicy(c *component, realmId string, multiValuedKeys map[string]bool) *RealmClientRegistrationPolicy {
	config := map[string]string{}
	for k, vals := range c.Config {
		if len(vals) == 0 {
			continue
		}

		if multiValuedKeys[k] && len(vals) > 1 {
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
	multiValuedKeys, err := keycloakClient.MultiValuedClientRegistrationConfigKeys(ctx, policy.ProviderId)
	if err != nil {
		return err
	}

	_, location, err := keycloakClient.post(ctx, fmt.Sprintf("/realms/%s/components", policy.RealmId), convertFromRealmClientRegistrationPolicyToComponent(policy, multiValuedKeys))
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

	multiValuedKeys, err := keycloakClient.MultiValuedClientRegistrationConfigKeys(ctx, component.ProviderId)
	if err != nil {
		return nil, err
	}

	return convertFromComponentToRealmClientRegistrationPolicy(component, realmId, multiValuedKeys), nil
}

func (keycloakClient *KeycloakClient) GetRealmClientRegistrationPolicies(ctx context.Context, realmId string) ([]*RealmClientRegistrationPolicy, error) {
	var components []*component

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/components?parent=%s&type=%s", realmId, realmId, realmClientRegistrationPolicyProviderType), &components, nil)
	if err != nil {
		return nil, err
	}

	policies := make([]*RealmClientRegistrationPolicy, 0, len(components))
	for _, c := range components {
		multiValuedKeys, err := keycloakClient.MultiValuedClientRegistrationConfigKeys(ctx, c.ProviderId)
		if err != nil {
			return nil, err
		}

		policies = append(policies, convertFromComponentToRealmClientRegistrationPolicy(c, realmId, multiValuedKeys))
	}

	return policies, nil
}

func (keycloakClient *KeycloakClient) UpdateRealmClientRegistrationPolicy(ctx context.Context, policy *RealmClientRegistrationPolicy) error {
	multiValuedKeys, err := keycloakClient.MultiValuedClientRegistrationConfigKeys(ctx, policy.ProviderId)
	if err != nil {
		return err
	}

	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/components/%s", policy.RealmId, policy.Id), convertFromRealmClientRegistrationPolicyToComponent(policy, multiValuedKeys))
}

func (keycloakClient *KeycloakClient) DeleteRealmClientRegistrationPolicy(ctx context.Context, realmId, id string) error {
	return keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/components/%s", realmId, id), nil)
}
