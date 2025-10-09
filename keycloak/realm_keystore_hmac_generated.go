package keycloak

import (
	"context"
	"fmt"
	"strconv"
)

type RealmKeystoreHmacGenerated struct {
	Id              string
	Name            string
	RealmId         string
	InternalRealmId string

	Active     bool
	Enabled    bool
	Priority   int
	SecretSize int
	Algorithm  string
}

func convertFromRealmKeystoreHmacGeneratedToComponent(realmKey *RealmKeystoreHmacGenerated) *component {
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
		"secretSize": {
			strconv.Itoa(realmKey.SecretSize),
		},
		"algorithm": {
			realmKey.Algorithm,
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
		ProviderId:   "hmac-generated",
		ProviderType: "org.keycloak.keys.KeyProvider",
		Config:       componentConfig,
	}
}

func convertFromComponentToRealmKeystoreHmacGenerated(component *component, realmId string) (*RealmKeystoreHmacGenerated, error) {
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

	secretSize := 64 // Default key size for hmac key
	if component.getConfig("secretSize") != "" {
		secretSize, err = strconv.Atoi(component.getConfig("secretSize"))
		if err != nil {
			return nil, err
		}
	}

	realmKey := &RealmKeystoreHmacGenerated{
		Id:      component.Id,
		Name:    component.Name,
		RealmId: realmId,

		Active:     active,
		Enabled:    enabled,
		Priority:   priority,
		Algorithm:  component.getConfig("algorithm"),
		SecretSize: secretSize,
	}

	return realmKey, nil
}

func (keycloakClient *KeycloakClient) NewRealmKeystoreHmacGenerated(ctx context.Context, realmKey *RealmKeystoreHmacGenerated) error {
	_, location, err := keycloakClient.post(ctx, fmt.Sprintf("/realms/%s/components", realmKey.RealmId), convertFromRealmKeystoreHmacGeneratedToComponent(realmKey))
	if err != nil {
		return err
	}

	realmKey.Id = getIdFromLocationHeader(location)

	return nil
}

func (keycloakClient *KeycloakClient) GetRealmKeystoreHmacGenerated(ctx context.Context, realmId, id string) (*RealmKeystoreHmacGenerated, error) {
	var component *component

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/components/%s", realmId, id), &component, nil)
	if err != nil {
		return nil, err
	}

	return convertFromComponentToRealmKeystoreHmacGenerated(component, realmId)
}

func (keycloakClient *KeycloakClient) UpdateRealmKeystoreHmacGenerated(ctx context.Context, realmKey *RealmKeystoreHmacGenerated) error {
	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/components/%s", realmKey.RealmId, realmKey.Id), convertFromRealmKeystoreHmacGeneratedToComponent(realmKey))
}

func (keycloakClient *KeycloakClient) DeleteRealmKeystoreHmacGenerated(ctx context.Context, realmId, id string) error {
	return keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/components/%s", realmId, id), nil)
}
