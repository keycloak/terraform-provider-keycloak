package keycloak

import (
	"context"
	"fmt"
	"strconv"
)

type RealmKeystoreJavaKeystore struct {
	Id              string
	Name            string
	RealmId         string
	InternalRealmId string

	Active    bool
	Enabled   bool
	Priority  int
	Algorithm string

	Keystore         string
	KeystorePassword string
	KeyAlias         string
	KeyPassword      string
}

func convertFromRealmKeystoreJavaKeystoreToComponent(realmKey *RealmKeystoreJavaKeystore) *component {
	componentConfig := map[string][]string{
		"active": {
			strconv.FormatBool(realmKey.Active),
		},
		"enabled": {
			strconv.FormatBool(realmKey.Enabled),
		},
		"priority": {
			strconv.Itoa(realmKey.Priority),
		},
		"algorithm": {
			realmKey.Algorithm,
		},
		"keystore": {
			realmKey.Keystore,
		},
		"keystorePassword": {
			realmKey.KeystorePassword,
		},
		"keyAlias": {
			realmKey.KeyAlias,
		},
		"keyPassword": {
			realmKey.KeyPassword,
		},
	}

	var parentId string
	if realmKey.InternalRealmId != "" {
		parentId = realmKey.InternalRealmId
	} else {
		parentId = realmKey.RealmId
	}

	return &component{
		Id:           realmKey.Id,
		Name:         realmKey.Name,
		ParentId:     parentId,
		ProviderId:   "java-keystore",
		ProviderType: "org.keycloak.keys.KeyProvider",
		Config:       componentConfig,
	}
}

func convertFromComponentToRealmKeystoreJavaKeystore(component *component, realmId string) (*RealmKeystoreJavaKeystore, error) {
	active, err := parseBoolAndTreatEmptyStringAsFalse(component.getConfig("active"))
	if err != nil {
		return nil, err
	}

	enabled, err := parseBoolAndTreatEmptyStringAsFalse(component.getConfig("enabled"))
	if err != nil {
		return nil, err
	}

	priority := 0 // Default priority
	if component.getConfig("priority") != "" {
		priority, err = strconv.Atoi(component.getConfig("priority"))
		if err != nil {
			return nil, err
		}
	}

	realmKey := &RealmKeystoreJavaKeystore{
		Id:      component.Id,
		Name:    component.Name,
		RealmId: realmId,

		Active:           active,
		Enabled:          enabled,
		Priority:         priority,
		Algorithm:        component.getConfig("algorithm"),
		Keystore:         component.getConfig("keystore"),
		KeystorePassword: component.getConfig("keystorePassword"),
		KeyAlias:         component.getConfig("keyAlias"),
		KeyPassword:      component.getConfig("keyPassword"),
	}

	return realmKey, nil
}

func (keycloakClient *KeycloakClient) NewRealmKeystoreJavaKeystore(ctx context.Context, realmKey *RealmKeystoreJavaKeystore) error {
	_, location, err := keycloakClient.post(ctx, fmt.Sprintf("/realms/%s/components", realmKey.RealmId), convertFromRealmKeystoreJavaKeystoreToComponent(realmKey))
	if err != nil {
		return err
	}

	realmKey.Id = getIdFromLocationHeader(location)

	return nil
}

func (keycloakClient *KeycloakClient) GetRealmKeystoreJavaKeystore(ctx context.Context, realmId, id string) (*RealmKeystoreJavaKeystore, error) {
	var component *component

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/components/%s", realmId, id), &component, nil)
	if err != nil {
		return nil, err
	}

	return convertFromComponentToRealmKeystoreJavaKeystore(component, realmId)
}

func (keycloakClient *KeycloakClient) UpdateRealmKeystoreJavaKeystore(ctx context.Context, realmKey *RealmKeystoreJavaKeystore) error {
	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/components/%s", realmKey.RealmId, realmKey.Id), convertFromRealmKeystoreJavaKeystoreToComponent(realmKey))
}

func (keycloakClient *KeycloakClient) DeleteRealmKeystoreJavaKeystore(ctx context.Context, realmId, id string) error {
	return keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/components/%s", realmId, id), nil)
}
