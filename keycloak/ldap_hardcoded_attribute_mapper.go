package keycloak

import (
	"context"
	"fmt"
)

type LdapHardcodedAttributeMapper struct {
	Id                   string
	Name                 string
	RealmId              string
	LdapUserFederationId string
	AttributeName        string
	AttributeValue       string
}

func convertFromLdapHardcodedAttributeMapperToComponent(ldapMapper *LdapHardcodedAttributeMapper) *component {
	return &component{
		Id:           ldapMapper.Id,
		Name:         ldapMapper.Name,
		ProviderId:   "hardcoded-ldap-attribute-mapper",
		ProviderType: "org.keycloak.storage.ldap.mappers.LDAPStorageMapper",
		ParentId:     ldapMapper.LdapUserFederationId,

		Config: map[string][]string{
			"ldap.attribute.name": {
				ldapMapper.AttributeName,
			},
			"ldap.attribute.value": {
				ldapMapper.AttributeValue,
			},
		},
	}
}

func convertFromComponentToLdapHardcodedAttributeMapper(component *component, realmId string) *LdapHardcodedAttributeMapper {
	return &LdapHardcodedAttributeMapper{
		Id:                   component.Id,
		Name:                 component.Name,
		RealmId:              realmId,
		LdapUserFederationId: component.ParentId,

		AttributeName:  component.getConfig("ldap.attribute.name"),
		AttributeValue: component.getConfig("ldap.attribute.value"),
	}
}

func (keycloakClient *KeycloakClient) NewLdapHardcodedAttributeMapper(ctx context.Context, ldapMapper *LdapHardcodedAttributeMapper) error {
	_, location, err := keycloakClient.post(ctx, fmt.Sprintf("/realms/%s/components", ldapMapper.RealmId), convertFromLdapHardcodedAttributeMapperToComponent(ldapMapper))
	if err != nil {
		return err
	}

	ldapMapper.Id = getIdFromLocationHeader(location)

	return nil
}

func (keycloakClient *KeycloakClient) GetLdapHardcodedAttributeMapper(ctx context.Context, realmId, id string) (*LdapHardcodedAttributeMapper, error) {
	var component *component

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/components/%s", realmId, id), &component, nil)
	if err != nil {
		return nil, err
	}

	return convertFromComponentToLdapHardcodedAttributeMapper(component, realmId), nil
}

func (keycloakClient *KeycloakClient) UpdateLdapHardcodedAttributeMapper(ctx context.Context, ldapMapper *LdapHardcodedAttributeMapper) error {
	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/components/%s", ldapMapper.RealmId, ldapMapper.Id), convertFromLdapHardcodedAttributeMapperToComponent(ldapMapper))
}

func (keycloakClient *KeycloakClient) DeleteLdapHardcodedAttributeMapper(ctx context.Context, realmId, id string) error {
	return keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/components/%s", realmId, id), nil)
}
