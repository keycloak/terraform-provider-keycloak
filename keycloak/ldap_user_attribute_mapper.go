package keycloak

import (
	"context"
	"fmt"
	"strconv"
)

type LdapUserAttributeMapper struct {
	Id                   string
	Name                 string
	RealmId              string
	LdapUserFederationId string

	LdapAttribute           string
	IsMandatoryInLdap       bool
	ReadOnly                bool
	AlwaysReadValueFromLdap bool
	UserModelAttribute      string
	ForceDefaultValue       bool
	AttributeDefaultValue   string
	IsBinaryAttribute       bool
}

func convertFromLdapUserAttributeMapperToComponent(ldapUserAttributeMapper *LdapUserAttributeMapper) *component {
	return &component{
		Id:           ldapUserAttributeMapper.Id,
		Name:         ldapUserAttributeMapper.Name,
		ProviderId:   "user-attribute-ldap-mapper",
		ProviderType: "org.keycloak.storage.ldap.mappers.LDAPStorageMapper",
		ParentId:     ldapUserAttributeMapper.LdapUserFederationId,
		Config: map[string][]string{
			"ldap.attribute": {
				ldapUserAttributeMapper.LdapAttribute,
			},
			"is.mandatory.in.ldap": {
				strconv.FormatBool(ldapUserAttributeMapper.IsMandatoryInLdap),
			},
			"read.only": {
				strconv.FormatBool(ldapUserAttributeMapper.ReadOnly),
			},
			"always.read.value.from.ldap": {
				strconv.FormatBool(ldapUserAttributeMapper.AlwaysReadValueFromLdap),
			},
			"user.model.attribute": {
				ldapUserAttributeMapper.UserModelAttribute,
			},
			"attribute.force.default": {
				strconv.FormatBool(ldapUserAttributeMapper.ForceDefaultValue),
			},
			"attribute.default.value": {
				ldapUserAttributeMapper.AttributeDefaultValue,
			},
			"is.binary.attribute": {
				strconv.FormatBool(ldapUserAttributeMapper.IsBinaryAttribute),
			},
		},
	}
}

func convertFromComponentToLdapUserAttributeMapper(component *component, realmId string) (*LdapUserAttributeMapper, error) {
	isMandatoryInLdap, err := parseBoolAndTreatEmptyStringAsFalse(component.getConfig("is.mandatory.in.ldap"))
	if err != nil {
		return nil, err
	}

	readOnly, err := parseBoolAndTreatEmptyStringAsFalse(component.getConfig("read.only"))
	if err != nil {
		return nil, err
	}

	alwaysReadValueFromLdap, err := parseBoolAndTreatEmptyStringAsFalse(component.getConfig("always.read.value.from.ldap"))
	if err != nil {
		return nil, err
	}

	forceDefaultValue, err := parseBoolAndTreatEmptyStringAsFalse(component.getConfig("attribute.force.default"))
	if err != nil {
		return nil, err
	}

	isBinaryAttribute, err := parseBoolAndTreatEmptyStringAsFalse(component.getConfig("is.binary.attribute"))
	if err != nil {
		return nil, err
	}

	return &LdapUserAttributeMapper{
		Id:                   component.Id,
		Name:                 component.Name,
		RealmId:              realmId,
		LdapUserFederationId: component.ParentId,

		LdapAttribute:           component.getConfig("ldap.attribute"),
		IsMandatoryInLdap:       isMandatoryInLdap,
		ReadOnly:                readOnly,
		AlwaysReadValueFromLdap: alwaysReadValueFromLdap,
		UserModelAttribute:      component.getConfig("user.model.attribute"),
		ForceDefaultValue:       forceDefaultValue,
		AttributeDefaultValue:   component.getConfig("attribute.default.value"),
		IsBinaryAttribute:       isBinaryAttribute,
	}, nil
}

func (keycloakClient *KeycloakClient) NewLdapUserAttributeMapper(ctx context.Context, ldapUserAttributeMapper *LdapUserAttributeMapper) error {
	_, location, err := keycloakClient.post(ctx, fmt.Sprintf("/realms/%s/components", ldapUserAttributeMapper.RealmId), convertFromLdapUserAttributeMapperToComponent(ldapUserAttributeMapper))
	if err != nil {
		return err
	}

	ldapUserAttributeMapper.Id = getIdFromLocationHeader(location)

	return nil
}

func (keycloakClient *KeycloakClient) GetLdapUserAttributeMapper(ctx context.Context, realmId, id string) (*LdapUserAttributeMapper, error) {
	var component *component

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/components/%s", realmId, id), &component, nil)
	if err != nil {
		return nil, err
	}

	return convertFromComponentToLdapUserAttributeMapper(component, realmId)
}

func (keycloakClient *KeycloakClient) UpdateLdapUserAttributeMapper(ctx context.Context, ldapUserAttributeMapper *LdapUserAttributeMapper) error {
	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/components/%s", ldapUserAttributeMapper.RealmId, ldapUserAttributeMapper.Id), convertFromLdapUserAttributeMapperToComponent(ldapUserAttributeMapper))
}

func (keycloakClient *KeycloakClient) DeleteLdapUserAttributeMapper(ctx context.Context, realmId, id string) error {
	return keycloakClient.DeleteComponent(ctx, realmId, id)
}
