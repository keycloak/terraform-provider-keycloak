package keycloak

import (
	"context"
	"fmt"
	"strconv"
)

type RealmKeystoreRsaEncGenerated struct {
	Id       string
	Name     string
	RealmId  string
	ParentId string

	Active    bool
	Enabled   bool
	Priority  int
	Algorithm string
	KeySize   int

	PrivateKey  string
	Certificate string
}

func convertFromRealmKeystoreRsaEncGeneratedToComponent(realmKey *RealmKeystoreRsaEncGenerated) *component {
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

	// ParentId is deliberately not set, Keycloak defaults to the realm's internal id
	return &component{
		Id:           realmKey.Id,
		Name:         realmKey.Name,
		ProviderId:   "rsa-enc-generated",
		ProviderType: "org.keycloak.keys.KeyProvider",
		Config:       componentConfig,
	}
}

func convertFromComponentToRealmKeystoreRsaEncGenerated(component *component, realmId string) (*RealmKeystoreRsaEncGenerated, error) {
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

	realmKey := &RealmKeystoreRsaEncGenerated{
		Id:       component.Id,
		Name:     component.Name,
		RealmId:  realmId,
		ParentId: component.ParentId,

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

func (keycloakClient *KeycloakClient) NewRealmKeystoreRsaEncGenerated(ctx context.Context, realmKey *RealmKeystoreRsaEncGenerated) error {
	_, location, err := keycloakClient.post(ctx, fmt.Sprintf("/realms/%s/components", realmKey.RealmId), convertFromRealmKeystoreRsaEncGeneratedToComponent(realmKey))
	if err != nil {
		return err
	}

	realmKey.Id = getIdFromLocationHeader(location)

	return nil
}

func (keycloakClient *KeycloakClient) GetRealmKeystoreRsaEncGenerated(ctx context.Context, realmId, id string) (*RealmKeystoreRsaEncGenerated, error) {
	var component *component

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/components/%s", realmId, id), &component, nil)
	if err != nil {
		return nil, err
	}

	return convertFromComponentToRealmKeystoreRsaEncGenerated(component, realmId)
}

func (keycloakClient *KeycloakClient) UpdateRealmKeystoreRsaEncGenerated(ctx context.Context, realmKey *RealmKeystoreRsaEncGenerated) error {
	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/components/%s", realmKey.RealmId, realmKey.Id), convertFromRealmKeystoreRsaEncGeneratedToComponent(realmKey))
}

func (keycloakClient *KeycloakClient) DeleteRealmKeystoreRsaEncGenerated(ctx context.Context, realmId, id string) error {
	return keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/components/%s", realmId, id), nil)
}
