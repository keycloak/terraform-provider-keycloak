package keycloak

import (
	"context"
	"fmt"
)

type LdapHardcodedRoleMapper struct {
	Id                   string
	Name                 string
	RealmId              string
	LdapUserFederationId string
	Role                 string
}

func convertFromLdapHardcodedRoleMapperToComponent(ldapMapper *LdapHardcodedRoleMapper) *component {
	return &component{
		Id:           ldapMapper.Id,
		Name:         ldapMapper.Name,
		ProviderId:   "hardcoded-ldap-role-mapper",
		ProviderType: "org.keycloak.storage.ldap.mappers.LDAPStorageMapper",
		ParentId:     ldapMapper.LdapUserFederationId,

		Config: map[string][]string{
			"role": {
				ldapMapper.Role,
			},
		},
	}
}

func convertFromComponentToLdapHardcodedRoleMapper(component *component, realmId string) *LdapHardcodedRoleMapper {
	return &LdapHardcodedRoleMapper{
		Id:                   component.Id,
		Name:                 component.Name,
		RealmId:              realmId,
		LdapUserFederationId: component.ParentId,

		Role: component.getConfig("role"),
	}
}

func (keycloakClient *KeycloakClient) ValidateLdapHardcodedRoleMapper(ctx context.Context, ldapMapper *LdapHardcodedRoleMapper) error {
	if len(ldapMapper.Role) == 0 {
		return fmt.Errorf("validation error: hardcoded role name must not be empty")
	}
	return nil
}

func (keycloakClient *KeycloakClient) NewLdapHardcodedRoleMapper(ctx context.Context, ldapMapper *LdapHardcodedRoleMapper) error {
	_, location, err := keycloakClient.post(ctx, fmt.Sprintf("/realms/%s/components", ldapMapper.RealmId), convertFromLdapHardcodedRoleMapperToComponent(ldapMapper))
	if err != nil {
		return err
	}

	ldapMapper.Id = getIdFromLocationHeader(location)

	return nil
}

func (keycloakClient *KeycloakClient) GetLdapHardcodedRoleMapper(ctx context.Context, realmId, id string) (*LdapHardcodedRoleMapper, error) {
	var component *component

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/components/%s", realmId, id), &component, nil)
	if err != nil {
		return nil, err
	}

	return convertFromComponentToLdapHardcodedRoleMapper(component, realmId), nil
}

func (keycloakClient *KeycloakClient) UpdateLdapHardcodedRoleMapper(ctx context.Context, ldapMapper *LdapHardcodedRoleMapper) error {
	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/components/%s", ldapMapper.RealmId, ldapMapper.Id), convertFromLdapHardcodedRoleMapperToComponent(ldapMapper))
}

func (keycloakClient *KeycloakClient) DeleteLdapHardcodedRoleMapper(ctx context.Context, realmId, id string) error {
	return keycloakClient.DeleteComponent(ctx, realmId, id)
}
