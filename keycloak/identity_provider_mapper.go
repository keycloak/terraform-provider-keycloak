package keycloak

import (
	"context"
	"fmt"
	"reflect"
)

type IdentityProviderMapperConfig struct {
	UserAttribute         string                 `json:"user.attribute,omitempty"`
	UserAttributeName     string                 `json:"userAttribute,omitempty"`
	Claim                 string                 `json:"claim,omitempty"`
	ClaimValue            string                 `json:"claim.value,omitempty"`
	HardcodedAttribute    string                 `json:"attribute,omitempty"`
	Attribute             string                 `json:"attribute.name,omitempty"`
	AttributeValue        string                 `json:"attribute.value,omitempty"`
	AttributeFriendlyName string                 `json:"attribute.friendly.name,omitempty"`
	Template              string                 `json:"template,omitempty"`
	Role                  string                 `json:"role,omitempty"`
	Group                 string                 `json:"group,omitempty"`
	JsonField             string                 `json:"jsonField,omitEmpty"`
	ExtraConfig           map[string]interface{} `json:"-"`
}

type IdentityProviderMapper struct {
	Realm                  string                        `json:"-"`
	Provider               string                        `json:"-"`
	Id                     string                        `json:"id,omitempty"`
	Name                   string                        `json:"name,omitempty"`
	IdentityProviderAlias  string                        `json:"identityProviderAlias,omitempty"`
	IdentityProviderMapper string                        `json:"identityProviderMapper,omitempty"`
	Config                 *IdentityProviderMapperConfig `json:"config,omitempty"`
}

func (keycloakClient *KeycloakClient) NewIdentityProviderMapper(ctx context.Context, identityProviderMapper *IdentityProviderMapper) error {
	_, location, err := keycloakClient.post(ctx, fmt.Sprintf("/realms/%s/identity-provider/instances/%s/mappers", identityProviderMapper.Realm, identityProviderMapper.IdentityProviderAlias), identityProviderMapper)
	if err != nil {
		return err
	}

	identityProviderMapper.Id = getIdFromLocationHeader(location)

	return nil
}

func (keycloakClient *KeycloakClient) GetIdentityProviderMappers(ctx context.Context, realm, alias string) ([]*IdentityProviderMapper, error) {
	var identityProviderMapper []*IdentityProviderMapper

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/identity-provider/instances/%s/mappers", realm, alias), &identityProviderMapper, nil)
	if err != nil {
		return nil, err
	}

	return identityProviderMapper, nil
}

func (keycloakClient *KeycloakClient) GetIdentityProviderMapper(ctx context.Context, realm, alias, id string) (*IdentityProviderMapper, error) {
	var identityProviderMapper IdentityProviderMapper
	identityProviderMapper.Realm = realm
	identityProviderMapper.IdentityProviderAlias = alias

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/identity-provider/instances/%s/mappers/%s", realm, alias, id), &identityProviderMapper, nil)
	if err != nil {
		return nil, err
	}

	return &identityProviderMapper, nil
}

func (keycloakClient *KeycloakClient) UpdateIdentityProviderMapper(ctx context.Context, identityProviderMapper *IdentityProviderMapper) error {
	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/identity-provider/instances/%s/mappers/%s", identityProviderMapper.Realm, identityProviderMapper.IdentityProviderAlias, identityProviderMapper.Id), identityProviderMapper)
}

func (keycloakClient *KeycloakClient) DeleteIdentityProviderMapper(ctx context.Context, realm, alias, id string) error {
	return keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/identity-provider/instances/%s/mappers/%s", realm, alias, id), nil)
}

func (f *IdentityProviderMapperConfig) UnmarshalJSON(data []byte) error {
	return unmarshalExtraConfig(data, reflect.ValueOf(f).Elem(), &f.ExtraConfig)
}

func (f *IdentityProviderMapperConfig) MarshalJSON() ([]byte, error) {
	return marshalExtraConfig(reflect.ValueOf(f).Elem(), f.ExtraConfig)
}
