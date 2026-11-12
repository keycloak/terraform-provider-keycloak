package keycloak

import (
	"context"
	"fmt"
	"strconv"
)

type LdapMsadUserAccountControlMapper struct {
	Id                   string
	Name                 string
	RealmId              string
	LdapUserFederationId string

	LdapPasswordPolicyHintsEnabled bool
	AlwaysReadEnabledValueFromLdap bool
}

func convertFromLdapMsadUserAccountControlMapperToComponent(ldapMsadUserAccountControlMapper *LdapMsadUserAccountControlMapper) *component {
	return &component{
		Id:           ldapMsadUserAccountControlMapper.Id,
		Name:         ldapMsadUserAccountControlMapper.Name,
		ProviderId:   "msad-user-account-control-mapper",
		ProviderType: "org.keycloak.storage.ldap.mappers.LDAPStorageMapper",
		ParentId:     ldapMsadUserAccountControlMapper.LdapUserFederationId,
		Config: map[string][]string{
			"ldap.password.policy.hints.enabled": {
				strconv.FormatBool(ldapMsadUserAccountControlMapper.LdapPasswordPolicyHintsEnabled),
			},
			"always.read.enabled.value.from.ldap": {
				strconv.FormatBool(ldapMsadUserAccountControlMapper.AlwaysReadEnabledValueFromLdap),
			},
		},
	}
}

func convertFromComponentToLdapMsadUserAccountControlMapper(component *component, realmId string) (*LdapMsadUserAccountControlMapper, error) {
	ldapPasswordPolicyHintsEnabled, err := parseBoolAndTreatEmptyStringAsFalse(component.getConfig("ldap.password.policy.hints.enabled"))
	if err != nil {
		return nil, err
	}

	alwaysReadEnabledValueFromLdap, err := parseBoolAndTreatEmptyStringAsFalse(component.getConfig("always.read.enabled.value.from.ldap"))
	if err != nil {
		return nil, err
	}

	return &LdapMsadUserAccountControlMapper{
		Id:                   component.Id,
		Name:                 component.Name,
		RealmId:              realmId,
		LdapUserFederationId: component.ParentId,

		LdapPasswordPolicyHintsEnabled: ldapPasswordPolicyHintsEnabled,
		AlwaysReadEnabledValueFromLdap: alwaysReadEnabledValueFromLdap,
	}, nil
}

func (keycloakClient *KeycloakClient) NewLdapMsadUserAccountControlMapper(ctx context.Context, ldapMsadUserAccountControlMapper *LdapMsadUserAccountControlMapper) error {
	_, location, err := keycloakClient.post(ctx, fmt.Sprintf("/realms/%s/components", ldapMsadUserAccountControlMapper.RealmId), convertFromLdapMsadUserAccountControlMapperToComponent(ldapMsadUserAccountControlMapper))
	if err != nil {
		return err
	}

	ldapMsadUserAccountControlMapper.Id = getIdFromLocationHeader(location)

	return nil
}

func (keycloakClient *KeycloakClient) GetLdapMsadUserAccountControlMapper(ctx context.Context, realmId, id string) (*LdapMsadUserAccountControlMapper, error) {
	var component *component

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/components/%s", realmId, id), &component, nil)
	if err != nil {
		return nil, err
	}

	return convertFromComponentToLdapMsadUserAccountControlMapper(component, realmId)
}

func (keycloakClient *KeycloakClient) UpdateLdapMsadUserAccountControlMapper(ctx context.Context, ldapMsadUserAccountControlMapper *LdapMsadUserAccountControlMapper) error {
	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/components/%s", ldapMsadUserAccountControlMapper.RealmId, ldapMsadUserAccountControlMapper.Id), convertFromLdapMsadUserAccountControlMapperToComponent(ldapMsadUserAccountControlMapper))
}

func (keycloakClient *KeycloakClient) DeleteLdapMsadUserAccountControlMapper(ctx context.Context, realmId, id string) error {
	return keycloakClient.DeleteComponent(ctx, realmId, id)
}
