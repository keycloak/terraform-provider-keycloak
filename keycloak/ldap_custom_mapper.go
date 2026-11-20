package keycloak

import (
	"context"
	"fmt"
)

type LdapCustomMapper struct {
	Id                   string
	Name                 string
	RealmId              string
	LdapUserFederationId string
	ProviderId           string
	ProviderType         string
	Config               map[string]string
}

func convertFromLdapCustomMapperToComponent(ldapCustomMapper *LdapCustomMapper) *component {
	return &component{
		Id:           ldapCustomMapper.Id,
		Name:         ldapCustomMapper.Name,
		ProviderId:   ldapCustomMapper.ProviderId,
		ProviderType: ldapCustomMapper.ProviderType,
		ParentId:     ldapCustomMapper.LdapUserFederationId,
		Config:       convertToComponentConfig(ldapCustomMapper.Config),
	}
}

func convertFromComponentToLdapCustomMapper(component *component, realmId string) (*LdapCustomMapper, error) {
	return &LdapCustomMapper{
		Id:                   component.Id,
		Name:                 component.Name,
		RealmId:              realmId,
		LdapUserFederationId: component.ParentId,
		ProviderId:           component.ProviderId,
		ProviderType:         component.ProviderType,
		Config:               convertFromComponentConfig(component.Config),
	}, nil
}

func convertFromComponentConfig(originalMap map[string][]string) map[string]string {
	convertedMap := make(map[string]string)

	for key, values := range originalMap {
		convertedMap[key] = values[0]
	}

	return convertedMap
}

func convertToComponentConfig(originalMap map[string]string) map[string][]string {
	convertedMap := make(map[string][]string)

	for key, value := range originalMap {
		convertedMap[key] = []string{value}
	}

	return convertedMap
}

func (keycloakClient *KeycloakClient) NewLdapCustomMapper(ctx context.Context, ldapCustomMapper *LdapCustomMapper) error {
	_, location, err := keycloakClient.post(ctx, fmt.Sprintf("/realms/%s/components", ldapCustomMapper.RealmId), convertFromLdapCustomMapperToComponent(ldapCustomMapper))
	if err != nil {
		return err
	}

	ldapCustomMapper.Id = getIdFromLocationHeader(location)

	return nil
}

func (keycloakClient *KeycloakClient) GetLdapCustomMapper(ctx context.Context, realmId, id string) (*LdapCustomMapper, error) {
	var component *component

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/components/%s", realmId, id), &component, nil)
	if err != nil {
		return nil, err
	}

	return convertFromComponentToLdapCustomMapper(component, realmId)
}

func (keycloakClient *KeycloakClient) UpdateLdapCustomMapper(ctx context.Context, ldapCustomMapper *LdapCustomMapper) error {
	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/components/%s", ldapCustomMapper.RealmId, ldapCustomMapper.Id), convertFromLdapCustomMapperToComponent(ldapCustomMapper))
}

func (keycloakClient *KeycloakClient) DeleteLdapCustomMapper(ctx context.Context, realmId, id string) error {
	return keycloakClient.DeleteComponent(ctx, realmId, id)
}
