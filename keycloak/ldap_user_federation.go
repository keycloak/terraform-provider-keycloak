package keycloak

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

type LdapUserFederation struct {
	Id      string
	Name    string
	RealmId string

	Enabled  bool
	Priority int

	ImportEnabled     bool
	EditMode          string // can be "READ_ONLY", "WRITABLE", or "UNSYNCED"
	SyncRegistrations bool   // I think this field controls whether BatchSizeForSync, FullSyncPeriod, and ChangedSyncPeriod are needed

	Vendor                 string // can be "other", "edirectory", "ad", "rhds", or "tivoli". honestly I don't think this field actually does anything
	UsernameLDAPAttribute  string
	RdnLDAPAttribute       string
	UuidLDAPAttribute      string
	UserObjectClasses      []string // api expects comma + space separated for some reason
	ConnectionUrl          string
	UsersDn                string
	BindDn                 string
	BindCredential         string
	CustomUserSearchFilter string // must start with '(' and end with ')'
	SearchScope            string // api expects "1" or "2", but that means "One Level" or "Subtree"

	StartTls                    bool
	UsePasswordModifyExtendedOp bool
	TrustEmail                  bool
	ValidatePasswordPolicy      bool
	UseTruststoreSpi            string // can be "ldapsOnly", "always", or "never"
	ConnectionTimeout           string // duration string (ex: 1h30m)
	ReadTimeout                 string // duration string (ex: 1h30m)
	Pagination                  bool

	ServerPrincipal                      string
	UseKerberosForPasswordAuthentication bool
	AllowKerberosAuthentication          bool
	KeyTab                               string
	KerberosRealm                        string

	BatchSizeForSync  int
	FullSyncPeriod    int // either a number, in milliseconds, or -1 if full sync is disabled
	ChangedSyncPeriod int // either a number, in milliseconds, or -1 if changed sync is disabled

	CachePolicy    string
	MaxLifespan    string // duration string (ex: 1h30m)
	EvictionDay    *int
	EvictionHour   *int
	EvictionMinute *int
}

func convertFromLdapUserFederationToComponent(ldap *LdapUserFederation) (*component, error) {
	componentConfig := map[string][]string{
		"cachePolicy": {
			ldap.CachePolicy,
		},
		"enabled": {
			strconv.FormatBool(ldap.Enabled),
		},
		"priority": {
			strconv.Itoa(ldap.Priority),
		},
		"importEnabled": {
			strconv.FormatBool(ldap.ImportEnabled),
		},
		"editMode": {
			ldap.EditMode,
		},
		"syncRegistrations": {
			strconv.FormatBool(ldap.SyncRegistrations),
		},
		"vendor": {
			strings.ToLower(ldap.Vendor),
		},
		"usernameLDAPAttribute": {
			ldap.UsernameLDAPAttribute,
		},
		"rdnLDAPAttribute": {
			ldap.RdnLDAPAttribute,
		},
		"uuidLDAPAttribute": {
			ldap.UuidLDAPAttribute,
		},
		"userObjectClasses": {
			strings.Join(ldap.UserObjectClasses, ", "),
		},
		"connectionUrl": {
			ldap.ConnectionUrl,
		},
		"usersDn": {
			ldap.UsersDn,
		},
		"searchScope": {
			ldap.SearchScope,
		},
		"startTls": {
			strconv.FormatBool(ldap.StartTls),
		},
		"usePasswordModifyExtendedOp": {
			strconv.FormatBool(ldap.UsePasswordModifyExtendedOp),
		},
		"validatePasswordPolicy": {
			strconv.FormatBool(ldap.ValidatePasswordPolicy),
		},
		"trustEmail": {
			strconv.FormatBool(ldap.TrustEmail),
		},
		"pagination": {
			strconv.FormatBool(ldap.Pagination),
		},
		"batchSizeForSync": {
			strconv.Itoa(ldap.BatchSizeForSync),
		},
		"fullSyncPeriod": {
			strconv.Itoa(ldap.FullSyncPeriod),
		},
		"changedSyncPeriod": {
			strconv.Itoa(ldap.ChangedSyncPeriod),
		},

		"serverPrincipal": {
			ldap.ServerPrincipal,
		},
		"useKerberosForPasswordAuthentication": {
			strconv.FormatBool(ldap.UseKerberosForPasswordAuthentication),
		},
		"allowKerberosAuthentication": {
			strconv.FormatBool(ldap.AllowKerberosAuthentication),
		},
		"keyTab": {
			ldap.KeyTab,
		},
		"kerberosRealm": {
			ldap.KerberosRealm,
		},
	}

	if ldap.BindDn != "" && ldap.BindCredential != "" {
		componentConfig["bindDn"] = []string{ldap.BindDn}
		componentConfig["bindCredential"] = []string{ldap.BindCredential}

		componentConfig["authType"] = []string{"simple"}
	} else {
		componentConfig["authType"] = []string{"none"}
	}

	if ldap.SearchScope == "ONE_LEVEL" {
		componentConfig["searchScope"] = []string{"1"}
	} else {
		componentConfig["searchScope"] = []string{"2"}
	}

	if ldap.CustomUserSearchFilter != "" {
		componentConfig["customUserSearchFilter"] = []string{ldap.CustomUserSearchFilter}
	}

	if ldap.UseTruststoreSpi == "ONLY_FOR_LDAPS" {
		componentConfig["useTruststoreSpi"] = []string{"ldapsOnly"}
	} else {
		componentConfig["useTruststoreSpi"] = []string{strings.ToLower(ldap.UseTruststoreSpi)}
	}

	if ldap.ConnectionTimeout != "" {
		connectionTimeoutMs, err := getMillisecondsFromDurationString(ldap.ConnectionTimeout)
		if err != nil {
			return nil, err
		}

		componentConfig["connectionTimeout"] = []string{connectionTimeoutMs}
	} else {
		componentConfig["connectionTimeout"] = []string{} // the keycloak API will not unset this unless the config is present with an empty array
	}

	if ldap.ReadTimeout != "" {
		readTimeoutMs, err := getMillisecondsFromDurationString(ldap.ReadTimeout)
		if err != nil {
			return nil, err
		}

		componentConfig["readTimeout"] = []string{readTimeoutMs}
	} else {
		componentConfig["readTimeout"] = []string{} // the keycloak API will not unset this unless the config is present with an empty array
	}

	componentConfig["evictionHour"] = []string{}
	componentConfig["evictionMinute"] = []string{}
	componentConfig["evictionDay"] = []string{}
	componentConfig["maxLifespan"] = []string{}

	if ldap.CachePolicy != "" {
		if ldap.EvictionHour != nil {
			componentConfig["evictionHour"] = []string{strconv.Itoa(*ldap.EvictionHour)}
		}
		if ldap.EvictionMinute != nil {
			componentConfig["evictionMinute"] = []string{strconv.Itoa(*ldap.EvictionMinute)}
		}
		if ldap.EvictionDay != nil {
			componentConfig["evictionDay"] = []string{strconv.Itoa(*ldap.EvictionDay)}
		}

		if ldap.MaxLifespan != "" {
			maxLifespanMs, err := getMillisecondsFromDurationString(ldap.MaxLifespan)
			if err != nil {
				return nil, err
			}
			componentConfig["maxLifespan"] = []string{maxLifespanMs}
		}
	}

	return &component{
		Id:           ldap.Id,
		Name:         ldap.Name,
		ProviderId:   "ldap",
		ProviderType: userStorageProviderType,
		ParentId:     ldap.RealmId,
		Config:       componentConfig,
	}, nil
}

func convertFromComponentToLdapUserFederation(component *component) (*LdapUserFederation, error) {
	enabled, err := parseBoolAndTreatEmptyStringAsFalse(component.getConfig("enabled"))
	if err != nil {
		return nil, err
	}

	priority, err := atoiAndTreatEmptyStringAsZero(component.getConfig("priority"))
	if err != nil {
		return nil, err
	}

	importEnabled, err := parseBoolAndTreatEmptyStringAsFalse(component.getConfig("importEnabled"))
	if err != nil {
		return nil, err
	}

	syncRegistrations, err := parseBoolAndTreatEmptyStringAsFalse(component.getConfig("syncRegistrations"))
	if err != nil {
		return nil, err
	}

	userObjectClasses := strings.Split(component.getConfig("userObjectClasses"), ", ")

	startTls, err := parseBoolAndTreatEmptyStringAsFalse(component.getConfig("startTls"))
	if err != nil {
		return nil, err
	}

	usePasswordModifyExtendedOp, err := parseBoolAndTreatEmptyStringAsFalse(component.getConfig("usePasswordModifyExtendedOp"))
	if err != nil {
		return nil, err
	}

	validatePasswordPolicy, err := parseBoolAndTreatEmptyStringAsFalse(component.getConfig("validatePasswordPolicy"))
	if err != nil {
		return nil, err
	}

	trustEmail, err := parseBoolAndTreatEmptyStringAsFalse(component.getConfig("trustEmail"))
	if err != nil {
		return nil, err
	}

	pagination, err := parseBoolAndTreatEmptyStringAsFalse(component.getConfig("pagination"))
	if err != nil {
		return nil, err
	}

	batchSizeForSync, err := atoiAndTreatEmptyStringAsZero(component.getConfig("batchSizeForSync"))
	if err != nil {
		return nil, err
	}

	fullSyncPeriod, err := atoiAndTreatEmptyStringAsZero(component.getConfig("fullSyncPeriod"))
	if err != nil {
		return nil, err
	}

	changedSyncPeriod, err := atoiAndTreatEmptyStringAsZero(component.getConfig("changedSyncPeriod"))
	if err != nil {
		return nil, err
	}

	useKerberosForPasswordAuthentication, err := parseBoolAndTreatEmptyStringAsFalse(component.getConfig("useKerberosForPasswordAuthentication"))
	if err != nil {
		return nil, err
	}

	allowKerberosAuthentication, err := parseBoolAndTreatEmptyStringAsFalse(component.getConfig("allowKerberosAuthentication"))
	if err != nil {
		return nil, err
	}

	ldap := &LdapUserFederation{
		Id:      component.Id,
		Name:    component.Name,
		RealmId: component.ParentId,

		Enabled:  enabled,
		Priority: priority,

		ImportEnabled:     importEnabled,
		EditMode:          component.getConfig("editMode"),
		SyncRegistrations: syncRegistrations,

		Vendor:                 strings.ToUpper(component.getConfig("vendor")),
		UsernameLDAPAttribute:  component.getConfig("usernameLDAPAttribute"),
		RdnLDAPAttribute:       component.getConfig("rdnLDAPAttribute"),
		UuidLDAPAttribute:      component.getConfig("uuidLDAPAttribute"),
		UserObjectClasses:      userObjectClasses,
		ConnectionUrl:          component.getConfig("connectionUrl"),
		UsersDn:                component.getConfig("usersDn"),
		BindDn:                 component.getConfig("bindDn"),
		BindCredential:         component.getConfig("bindCredential"),
		CustomUserSearchFilter: component.getConfig("customUserSearchFilter"),
		SearchScope:            component.getConfig("searchScope"),

		StartTls:                    startTls,
		UsePasswordModifyExtendedOp: usePasswordModifyExtendedOp,
		ValidatePasswordPolicy:      validatePasswordPolicy,
		TrustEmail:                  trustEmail,
		UseTruststoreSpi:            component.getConfig("useTruststoreSpi"),
		Pagination:                  pagination,

		ServerPrincipal:                      component.getConfig("serverPrincipal"),
		UseKerberosForPasswordAuthentication: useKerberosForPasswordAuthentication,
		AllowKerberosAuthentication:          allowKerberosAuthentication,
		KeyTab:                               component.getConfig("keyTab"),
		KerberosRealm:                        component.getConfig("kerberosRealm"),

		BatchSizeForSync:  batchSizeForSync,
		FullSyncPeriod:    fullSyncPeriod,
		ChangedSyncPeriod: changedSyncPeriod,

		CachePolicy: component.getConfig("cachePolicy"),
	}

	if bindDn := component.getConfig("bindDn"); bindDn != "" {
		ldap.BindDn = bindDn
	}

	if bindCredential := component.getConfig("bindCredential"); bindCredential != "" {
		ldap.BindCredential = bindCredential
	}

	if customUserSearchFilter := component.getConfig("customUserSearchFilter"); customUserSearchFilter != "" {
		ldap.CustomUserSearchFilter = customUserSearchFilter
	}

	if component.getConfig("searchScope") == "1" {
		ldap.SearchScope = "ONE_LEVEL"
	} else {
		ldap.SearchScope = "SUBTREE"
	}

	if useTruststoreSpi := component.getConfig("useTruststoreSpi"); useTruststoreSpi == "ldapsOnly" {
		ldap.UseTruststoreSpi = "ONLY_FOR_LDAPS"
	} else {
		ldap.UseTruststoreSpi = strings.ToUpper(useTruststoreSpi)
	}

	if connectionTimeout, ok := component.getConfigOk("connectionTimeout"); ok {
		connectionTimeoutDurationString, err := GetDurationStringFromMilliseconds(connectionTimeout)
		if err != nil {
			return nil, err
		}

		ldap.ConnectionTimeout = connectionTimeoutDurationString
	}

	if readTimeout, ok := component.getConfigOk("readTimeout"); ok {
		readTimeoutDurationString, err := GetDurationStringFromMilliseconds(readTimeout)
		if err != nil {
			return nil, err
		}

		ldap.ReadTimeout = readTimeoutDurationString
	}

	if maxLifespan, ok := component.getConfigOk("maxLifespan"); ok {
		maxLifespanString, err := GetDurationStringFromMilliseconds(maxLifespan)
		if err != nil {
			return nil, err
		}

		ldap.MaxLifespan = maxLifespanString
	}

	defaultEvictioValue := -1

	if evictionDay, ok := component.getConfigOk("evictionDay"); ok {
		evictionDayInt, err := atoiAndTreatEmptyStringAsZero(evictionDay)
		if err != nil {
			return nil, fmt.Errorf("unable to parse `evictionDay`: %w", err)
		}

		ldap.EvictionDay = &evictionDayInt
	} else {
		ldap.EvictionDay = &defaultEvictioValue
	}

	if evictionHour, ok := component.getConfigOk("evictionHour"); ok {
		evictionHourInt, err := atoiAndTreatEmptyStringAsZero(evictionHour)
		if err != nil {
			return nil, fmt.Errorf("unable to parse `evictionHour`: %w", err)
		}

		ldap.EvictionHour = &evictionHourInt
	} else {
		ldap.EvictionHour = &defaultEvictioValue
	}
	if evictionMinute, ok := component.getConfigOk("evictionMinute"); ok {
		evictionMinuteInt, err := atoiAndTreatEmptyStringAsZero(evictionMinute)
		if err != nil {
			return nil, fmt.Errorf("unable to parse `evictionMinute`: %w", err)
		}

		ldap.EvictionMinute = &evictionMinuteInt
	} else {
		ldap.EvictionMinute = &defaultEvictioValue
	}

	return ldap, nil
}

func (keycloakClient *KeycloakClient) ValidateLdapUserFederation(ctx context.Context, ldap *LdapUserFederation) error {
	if (ldap.BindDn == "" && ldap.BindCredential != "") || (ldap.BindDn != "" && ldap.BindCredential == "") {
		return fmt.Errorf("validation error: authentication requires both BindDN and BindCredential to be set")
	}

	return nil
}

func (keycloakClient *KeycloakClient) NewLdapUserFederation(ctx context.Context, realmId string, ldapUserFederation *LdapUserFederation) error {
	component, err := convertFromLdapUserFederationToComponent(ldapUserFederation)
	if err != nil {
		return err
	}

	_, location, err := keycloakClient.post(ctx, fmt.Sprintf("/realms/%s/components", realmId), component)
	if err != nil {
		return err
	}

	ldapUserFederation.Id = getIdFromLocationHeader(location)

	return nil
}

func (keycloakClient *KeycloakClient) GetLdapUserFederation(ctx context.Context, realmId, id string) (*LdapUserFederation, error) {
	var component *component

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/components/%s", realmId, id), &component, nil)
	if err != nil {
		return nil, err
	}

	return convertFromComponentToLdapUserFederation(component)
}

func (keycloakClient *KeycloakClient) GetLdapUserFederationMappers(ctx context.Context, realmId, id string) (*[]interface{}, error) {
	var components []*component
	var ldapUserFederationMappers []interface{}

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/components?parent=%s&type=org.keycloak.storage.ldap.mappers.LDAPStorageMapper", realmId, id), &components, nil)
	if err != nil {
		return nil, err
	}
	for _, component := range components {
		switch component.ProviderId {
		case "full-name-ldap-mapper":
			mapper, err := convertFromComponentToLdapFullNameMapper(component, realmId)
			if err != nil {
				return nil, err
			}
			ldapUserFederationMappers = append(ldapUserFederationMappers, mapper)
		case "group-ldap-mapper":
			mapper, err := convertFromComponentToLdapGroupMapper(component, realmId)
			if err != nil {
				return nil, err
			}
			ldapUserFederationMappers = append(ldapUserFederationMappers, mapper)
		case "hardcoded-ldap-group-mapper":
			mapper := convertFromComponentToLdapHardcodedGroupMapper(component, realmId)
			ldapUserFederationMappers = append(ldapUserFederationMappers, mapper)
		case "hardcoded-ldap-role-mapper":
			mapper := convertFromComponentToLdapHardcodedRoleMapper(component, realmId)
			ldapUserFederationMappers = append(ldapUserFederationMappers, mapper)
		case "hardcoded-ldap-attribute-mapper":
			mapper := convertFromComponentToLdapHardcodedAttributeMapper(component, realmId)
			ldapUserFederationMappers = append(ldapUserFederationMappers, mapper)
		case "msad-lds-user-account-control-mapper":
			mapper, err := convertFromComponentToLdapMsadLdsUserAccountControlMapper(component, realmId)
			if err != nil {
				return nil, err
			}
			ldapUserFederationMappers = append(ldapUserFederationMappers, mapper)
		case "msad-user-account-control-mapper":
			mapper, err := convertFromComponentToLdapMsadUserAccountControlMapper(component, realmId)
			if err != nil {
				return nil, err
			}
			ldapUserFederationMappers = append(ldapUserFederationMappers, mapper)
		case "user-attribute-ldap-mapper":
			mapper, err := convertFromComponentToLdapUserAttributeMapper(component, realmId)
			if err != nil {
				return nil, err
			}
			ldapUserFederationMappers = append(ldapUserFederationMappers, mapper)
		case "role-ldap-mapper":
			mapper, err := convertFromComponentToLdapRoleMapper(component, realmId)
			if err != nil {
				return nil, err
			}
			ldapUserFederationMappers = append(ldapUserFederationMappers, mapper)
		}
	}

	return &ldapUserFederationMappers, nil
}

func (keycloakClient *KeycloakClient) UpdateLdapUserFederation(ctx context.Context, realmId string, ldapUserFederation *LdapUserFederation) error {
	component, err := convertFromLdapUserFederationToComponent(ldapUserFederation)
	if err != nil {
		return err
	}

	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/components/%s", realmId, ldapUserFederation.Id), component)
}

func (keycloakClient *KeycloakClient) DeleteLdapUserFederation(ctx context.Context, realmId, id string) error {
	return keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/components/%s", realmId, id), nil)
}

func (keycloakClient *KeycloakClient) DeleteLdapUserFederationMappers(ctx context.Context, realmId, id string) error {
	var components []*component

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/components?parent=%s&type=org.keycloak.storage.ldap.mappers.LDAPStorageMapper", realmId, id), &components, nil)
	if err != nil {
		return err
	}

	for _, component := range components {
		err = keycloakClient.DeleteComponent(ctx, realmId, component.Id)
		if err != nil {
			return err
		}
	}

	return nil
}
