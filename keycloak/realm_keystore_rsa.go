package keycloak

import (
	"context"
	"fmt"
	"strconv"
)

type RealmKeystoreRsa struct {
	Id      string
	Name    string
	RealmId string

	Active    bool
	Enabled   bool
	Priority  int
	Algorithm string

	PrivateKey  string
	Certificate string
	ProviderId  string
	ExtraConfig map[string]interface{}
}

func convertFromRealmKeystoreRsaToComponent(realmKey *RealmKeystoreRsa) *component {
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
		"privateKey": {
			realmKey.PrivateKey,
		},
		"certificate": {
			realmKey.Certificate,
		},
	}

	// merge extra config
	if realmKey.ExtraConfig != nil {
		for k, v := range realmKey.ExtraConfig {
			componentConfig[k] = []string{fmt.Sprint(v)}
		}
	}

	return &component{
		Id:           realmKey.Id,
		Name:         realmKey.Name,
		ParentId:     realmKey.RealmId,
		ProviderId:   realmKey.ProviderId,
		ProviderType: "org.keycloak.keys.KeyProvider",
		Config:       componentConfig,
	}
}

func convertFromComponentToRealmKeystoreRsa(component *component, realmId string) (*RealmKeystoreRsa, error) {
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

	realmKey := &RealmKeystoreRsa{
		Id:      component.Id,
		Name:    component.Name,
		RealmId: realmId,

		Active:      active,
		Enabled:     enabled,
		Priority:    priority,
		Algorithm:   component.getConfig("algorithm"),
		PrivateKey:  component.getConfig("privateKey"),
		Certificate: component.getConfig("certificate"),
		ProviderId:  component.ProviderId,
	}

	// capture extra config (anything not in the known keys)
	known := map[string]struct{}{
		"active": {}, "enabled": {}, "priority": {}, "algorithm": {}, "privateKey": {}, "certificate": {},
	}
	extra := map[string]interface{}{}
	for k, vals := range component.Config {
		if _, ok := known[k]; ok {
			continue
		}
		if len(vals) > 0 {
			extra[k] = vals[0]
		}
	}
	realmKey.ExtraConfig = extra

	return realmKey, nil
}

func (keycloakClient *KeycloakClient) NewRealmKeystoreRsa(ctx context.Context, realmKey *RealmKeystoreRsa) error {
	_, location, err := keycloakClient.post(ctx, fmt.Sprintf("/realms/%s/components", realmKey.RealmId), convertFromRealmKeystoreRsaToComponent(realmKey))
	if err != nil {
		return err
	}

	realmKey.Id = getIdFromLocationHeader(location)

	return nil
}

func (keycloakClient *KeycloakClient) GetRealmKeystoreRsa(ctx context.Context, realmId, id string) (*RealmKeystoreRsa, error) {
	var component *component

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/components/%s", realmId, id), &component, nil)
	if err != nil {
		return nil, err
	}

	return convertFromComponentToRealmKeystoreRsa(component, realmId)
}

func (keycloakClient *KeycloakClient) UpdateRealmKeystoreRsa(ctx context.Context, realmKey *RealmKeystoreRsa) error {
	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/components/%s", realmKey.RealmId, realmKey.Id), convertFromRealmKeystoreRsaToComponent(realmKey))
}

func (keycloakClient *KeycloakClient) DeleteRealmKeystoreRsa(ctx context.Context, realmId, id string) error {
	return keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/components/%s", realmId, id), nil)
}
