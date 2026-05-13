package keycloak

import (
	"context"
	"fmt"
)

const clientRegistrationPolicyProviderType = "org.keycloak.services.clientregistration.policy.ClientRegistrationPolicy"

type ClientRegistrationPolicy struct {
	Id         string
	RealmId    string
	Name       string
	ProviderId string
	SubType    string
	Config     map[string][]string
}

func (p *ClientRegistrationPolicy) toComponent() *component {
	config := p.Config
	if config == nil {
		config = map[string][]string{}
	}
	return &component{
		Id:           p.Id,
		Name:         p.Name,
		ProviderId:   p.ProviderId,
		ProviderType: clientRegistrationPolicyProviderType,
		ParentId:     p.RealmId,
		SubType:      p.SubType,
		Config:       config,
	}
}

func componentToClientRegistrationPolicy(c *component) *ClientRegistrationPolicy {
	config := c.Config
	if config == nil {
		config = map[string][]string{}
	}
	return &ClientRegistrationPolicy{
		Id:         c.Id,
		RealmId:    c.ParentId,
		Name:       c.Name,
		ProviderId: c.ProviderId,
		SubType:    c.SubType,
		Config:     config,
	}
}

func (keycloakClient *KeycloakClient) NewClientRegistrationPolicy(ctx context.Context, policy *ClientRegistrationPolicy) error {
	_, location, err := keycloakClient.post(ctx, fmt.Sprintf("/realms/%s/components", policy.RealmId), policy.toComponent())
	if err != nil {
		return err
	}
	policy.Id = getIdFromLocationHeader(location)
	return nil
}

func (keycloakClient *KeycloakClient) GetClientRegistrationPolicy(ctx context.Context, realmId, id string) (*ClientRegistrationPolicy, error) {
	var c *component
	if err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/components/%s", realmId, id), &c, nil); err != nil {
		return nil, err
	}
	if c.ProviderType != clientRegistrationPolicyProviderType {
		return nil, fmt.Errorf("component %s in realm %s is not a client registration policy (providerType=%q)", id, realmId, c.ProviderType)
	}
	return componentToClientRegistrationPolicy(c), nil
}

func (keycloakClient *KeycloakClient) UpdateClientRegistrationPolicy(ctx context.Context, policy *ClientRegistrationPolicy) error {
	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/components/%s", policy.RealmId, policy.Id), policy.toComponent())
}

func (keycloakClient *KeycloakClient) DeleteClientRegistrationPolicy(ctx context.Context, realmId, id string) error {
	return keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/components/%s", realmId, id), nil)
}
