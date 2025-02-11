package keycloak

import (
	"context"
	"fmt"
	"strconv"
)

type RealmKeystoreRsaGenerated struct {
	Id              string
	Name            string
	RealmId         string
	InternalRealmId string

	Active    bool
	Enabled   bool
	Priority  int
	Algorithm string
	KeySize   int

	PrivateKey  string
	Certificate string
}

func convertFromRealmKeystoreRsaGeneratedToComponent(realmKey *RealmKeystoreRsaGenerated) *component {
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
		"keySize": {
			strconv.Itoa(realmKey.KeySize),
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
		ProviderId:   "rsa-generated",
		ProviderType: "org.keycloak.keys.KeyProvider",
		Config:       componentConfig,
	}
}

func convertFromComponentToRealmKeystoreRsaGenerated(component *component, realmId string) (*RealmKeystoreRsaGenerated, error) {
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

	keySize := 2048 // Default key size for rsa key
	if component.getConfig("keySize") != "" {
		keySize, err = strconv.Atoi(component.getConfig("keySize"))
		if err != nil {
			return nil, err
		}
	}

	realmKey := &RealmKeystoreRsaGenerated{
		Id:      component.Id,
		Name:    component.Name,
		RealmId: realmId,

		Active:      active,
		Enabled:     enabled,
		Priority:    priority,
		Algorithm:   component.getConfig("algorithm"),
		KeySize:     keySize,
		PrivateKey:  component.getConfig("privateKey"),
		Certificate: component.getConfig("certificate"),
	}

	return realmKey, nil
}

func (keycloakClient *KeycloakClient) NewRealmKeystoreRsaGenerated(ctx context.Context, realmKey *RealmKeystoreRsaGenerated) error {
	_, location, err := keycloakClient.post(ctx, fmt.Sprintf("/realms/%s/components", realmKey.RealmId), convertFromRealmKeystoreRsaGeneratedToComponent(realmKey))
	if err != nil {
		return err
	}

	realmKey.Id = getIdFromLocationHeader(location)

	return nil
}

func (keycloakClient *KeycloakClient) GetRealmKeystoreRsaGenerated(ctx context.Context, realmId, id string) (*RealmKeystoreRsaGenerated, error) {
	var component *component

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/components/%s", realmId, id), &component, nil)
	if err != nil {
		return nil, err
	}

	return convertFromComponentToRealmKeystoreRsaGenerated(component, realmId)
}

func (keycloakClient *KeycloakClient) UpdateRealmKeystoreRsaGenerated(ctx context.Context, realmKey *RealmKeystoreRsaGenerated) error {
	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/components/%s", realmKey.RealmId, realmKey.Id), convertFromRealmKeystoreRsaGeneratedToComponent(realmKey))
}

func (keycloakClient *KeycloakClient) DeleteRealmKeystoreRsaGenerated(ctx context.Context, realmId, id string) error {
	return keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/components/%s", realmId, id), nil)
}
