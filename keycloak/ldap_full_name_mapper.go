package keycloak

import (
	"context"
	"fmt"
	"strconv"
)

type LdapFullNameMapper struct {
	Id                   string
	Name                 string
	RealmId              string
	LdapUserFederationId string

	LdapFullNameAttribute string
	ReadOnly              bool
	WriteOnly             bool
}

func convertFromLdapFullNameMapperToComponent(ldapFullNameMapper *LdapFullNameMapper) *component {
	return &component{
		Id:           ldapFullNameMapper.Id,
		Name:         ldapFullNameMapper.Name,
		ProviderId:   "full-name-ldap-mapper",
		ProviderType: "org.keycloak.storage.ldap.mappers.LDAPStorageMapper",
		ParentId:     ldapFullNameMapper.LdapUserFederationId,
		Config: map[string][]string{
			"ldap.full.name.attribute": {
				ldapFullNameMapper.LdapFullNameAttribute,
			},
			"read.only": {
				strconv.FormatBool(ldapFullNameMapper.ReadOnly),
			},
			"write.only": {
				strconv.FormatBool(ldapFullNameMapper.WriteOnly),
			},
		},
	}
}

func convertFromComponentToLdapFullNameMapper(component *component, realmId string) (*LdapFullNameMapper, error) {
	readOnly, err := parseBoolAndTreatEmptyStringAsFalse(component.getConfig("read.only"))
	if err != nil {
		return nil, err
	}

	writeOnly, err := parseBoolAndTreatEmptyStringAsFalse(component.getConfig("write.only"))
	if err != nil {
		return nil, err
	}

	return &LdapFullNameMapper{
		Id:                   component.Id,
		Name:                 component.Name,
		RealmId:              realmId,
		LdapUserFederationId: component.ParentId,

		LdapFullNameAttribute: component.getConfig("ldap.full.name.attribute"),
		ReadOnly:              readOnly,
		WriteOnly:             writeOnly,
	}, nil
}

// ValidateLdapFullNameMapper the keycloak api client is passed in order to fetch the ldap provider for writable validation
func (keycloakClient *KeycloakClient) ValidateLdapFullNameMapper(ctx context.Context, mapper *LdapFullNameMapper) error {
	if mapper.ReadOnly && mapper.WriteOnly {
		return fmt.Errorf("validation error: ldap full name mapper cannot be both read only and write only")
	}

	// the mapper can't be write-only if the ldap provider is not writable
	if mapper.WriteOnly {
		ldapUserFederation, err := keycloakClient.GetLdapUserFederation(ctx, mapper.RealmId, mapper.LdapUserFederationId)
		if err != nil {
			return err
		}

		if ldapUserFederation.EditMode != "WRITABLE" {
			return fmt.Errorf("validation error: ldap full name mapper cannot be write only when ldap provider is not writable")
		}
	}

	return nil
}

func (keycloakClient *KeycloakClient) NewLdapFullNameMapper(ctx context.Context, ldapFullNameMapper *LdapFullNameMapper) error {
	_, location, err := keycloakClient.post(ctx, fmt.Sprintf("/realms/%s/components", ldapFullNameMapper.RealmId), convertFromLdapFullNameMapperToComponent(ldapFullNameMapper))
	if err != nil {
		return err
	}

	ldapFullNameMapper.Id = getIdFromLocationHeader(location)

	return nil
}

func (keycloakClient *KeycloakClient) GetLdapFullNameMapper(ctx context.Context, realmId, id string) (*LdapFullNameMapper, error) {
	var component *component

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/components/%s", realmId, id), &component, nil)
	if err != nil {
		return nil, err
	}

	return convertFromComponentToLdapFullNameMapper(component, realmId)
}

func (keycloakClient *KeycloakClient) UpdateLdapFullNameMapper(ctx context.Context, ldapFullNameMapper *LdapFullNameMapper) error {
	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/components/%s", ldapFullNameMapper.RealmId, ldapFullNameMapper.Id), convertFromLdapFullNameMapperToComponent(ldapFullNameMapper))
}

func (keycloakClient *KeycloakClient) DeleteLdapFullNameMapper(ctx context.Context, realmId, id string) error {
	return keycloakClient.DeleteComponent(ctx, realmId, id)
}
