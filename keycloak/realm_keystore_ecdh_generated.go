package keycloak

import (
	"context"
	"fmt"
	"strconv"
)

type RealmKeystoreEcdhGenerated struct {
	Id       string
	Name     string
	RealmId  string
	ParentId string

	Active        bool
	Enabled       bool
	Priority      int
	EllipticCurve string
	Algorithm     string
}

func convertFromRealmKeystoreEcdhGeneratedToComponent(realmKey *RealmKeystoreEcdhGenerated) *component {
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
		"ecdhEllipticCurveKey": {
			realmKey.EllipticCurve,
		},
		"ecdhAlgorithm": {
			realmKey.Algorithm,
		},
	}

	// ParentId is deliberately not set, Keycloak defaults to the realm's internal id
	return &component{
		Id:           realmKey.Id,
		Name:         realmKey.Name,
		ProviderId:   "ecdh-generated",
		ProviderType: "org.keycloak.keys.KeyProvider",
		Config:       componentConfig,
	}
}

func convertFromComponentToRealmKeystoreEcdhGenerated(component *component, realmId string) (*RealmKeystoreEcdhGenerated, error) {
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

	realmKey := &RealmKeystoreEcdhGenerated{
		Id:       component.Id,
		Name:     component.Name,
		RealmId:  realmId,
		ParentId: component.ParentId,

		Active:        active,
		Enabled:       enabled,
		Priority:      priority,
		EllipticCurve: component.getConfig("ecdhEllipticCurveKey"),
		Algorithm:     component.getConfig("ecdhAlgorithm"),
	}

	return realmKey, nil
}

func (keycloakClient *KeycloakClient) NewRealmKeystoreEcdhGenerated(ctx context.Context, realmKey *RealmKeystoreEcdhGenerated) error {
	_, location, err := keycloakClient.post(ctx, fmt.Sprintf("/realms/%s/components", realmKey.RealmId), convertFromRealmKeystoreEcdhGeneratedToComponent(realmKey))
	if err != nil {
		return err
	}

	realmKey.Id = getIdFromLocationHeader(location)

	return nil
}

func (keycloakClient *KeycloakClient) GetRealmKeystoreEcdhGenerated(ctx context.Context, realmId, id string) (*RealmKeystoreEcdhGenerated, error) {
	var component *component

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/components/%s", realmId, id), &component, nil)
	if err != nil {
		return nil, err
	}

	return convertFromComponentToRealmKeystoreEcdhGenerated(component, realmId)
}

func (keycloakClient *KeycloakClient) UpdateRealmKeystoreEcdhGenerated(ctx context.Context, realmKey *RealmKeystoreEcdhGenerated) error {
	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/components/%s", realmKey.RealmId, realmKey.Id), convertFromRealmKeystoreEcdhGeneratedToComponent(realmKey))
}

func (keycloakClient *KeycloakClient) DeleteRealmKeystoreEcdhGenerated(ctx context.Context, realmId, id string) error {
	return keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/components/%s", realmId, id), nil)
}
