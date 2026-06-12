package keycloak

import (
	"context"
	"fmt"
)

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
		config[k] = []string{v}
	}

	return &component{
		Id:           policy.Id,
		Name:         policy.Name,
		ProviderId:   policy.ProviderId,
		ProviderType: "org.keycloak.services.clientregistration.policy.ClientRegistrationPolicy",
		ParentId:     policy.RealmId,
		SubType:      policy.SubType,
		Config:       config,
	}
}

func convertFromComponentToRealmClientRegistrationPolicy(c *component, realmId string) *RealmClientRegistrationPolicy {
	config := map[string]string{}
	for k, vals := range c.Config {
		if len(vals) > 0 {
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

func (keycloakClient *KeycloakClient) UpdateRealmClientRegistrationPolicy(ctx context.Context, policy *RealmClientRegistrationPolicy) error {
	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/components/%s", policy.RealmId, policy.Id), convertFromRealmClientRegistrationPolicyToComponent(policy))
}

func (keycloakClient *KeycloakClient) DeleteRealmClientRegistrationPolicy(ctx context.Context, realmId, id string) error {
	return keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/components/%s", realmId, id), nil)
}
