package keycloak

import (
	"context"
	"fmt"
	"strconv"
)

type RealmKeystoreCustom struct {
	Id      string
	Name    string
	RealmId string

	Active   bool
	Enabled  bool
	Priority int

	ProviderId   string
	ProviderType string

	ExtraConfig map[string]interface{} `json:"-"`
}

func convertFromRealmKeystoreCustomToComponent(realmKey *RealmKeystoreCustom) *component {
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
		"providerId": {
			realmKey.ProviderId,
		},
		"providerType": {
			realmKey.ProviderType,
		},
	}

	for key, value := range realmKey.ExtraConfig {
		if strVal, ok := value.(string); ok {
			componentConfig[key] = []string{strVal}
		} else if boolVal, ok := value.(bool); ok {
			componentConfig[key] = []string{strconv.FormatBool(boolVal)}
		} else if intVal, ok := value.(int); ok {
			componentConfig[key] = []string{strconv.Itoa(intVal)}
		}
	}

	return &component{
		Id:           realmKey.Id,
		Name:         realmKey.Name,
		ParentId:     realmKey.RealmId,
		ProviderId:   realmKey.ProviderId,
		ProviderType: realmKey.ProviderType,
		Config:       componentConfig,
	}
}

func convertFromComponentToRealmKeystoreCustom(component *component, realmId string) (*RealmKeystoreCustom, error) {
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

	providerId := component.ProviderId
	providerType := component.ProviderType

	realmKey := &RealmKeystoreCustom{
		Id:      component.Id,
		Name:    component.Name,
		RealmId: realmId,

		Active:       active,
		Enabled:      enabled,
		Priority:     priority,
		ProviderId:   providerId,
		ProviderType: providerType,
		ExtraConfig:  make(map[string]interface{}),
	}

	for key, value := range component.Config {
		if len(value) > 0 {
			switch key {
			case "active", "enabled", "priority":
				continue
			default:
				realmKey.ExtraConfig[key] = value[0]
			}
		}
	}

	return realmKey, nil
}

func (keycloakClient *KeycloakClient) NewRealmKeystoreCustom(ctx context.Context, realmKey *RealmKeystoreCustom) error {
	_, location, err := keycloakClient.post(ctx, fmt.Sprintf("/realms/%s/components", realmKey.RealmId), convertFromRealmKeystoreCustomToComponent(realmKey))
	if err != nil {
		return err
	}

	realmKey.Id = getIdFromLocationHeader(location)

	return nil
}

func (keycloakClient *KeycloakClient) GetRealmKeystoreCustom(ctx context.Context, realmId, id string) (*RealmKeystoreCustom, error) {
	var component *component

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/components/%s", realmId, id), &component, nil)
	if err != nil {
		return nil, err
	}

	return convertFromComponentToRealmKeystoreCustom(component, realmId)
}

func (keycloakClient *KeycloakClient) UpdateRealmKeystoreCustom(ctx context.Context, realmKey *RealmKeystoreCustom) error {
	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/components/%s", realmKey.RealmId, realmKey.Id), convertFromRealmKeystoreCustomToComponent(realmKey))
}

func (keycloakClient *KeycloakClient) DeleteRealmKeystoreCustom(ctx context.Context, realmId, id string) error {
	return keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/components/%s", realmId, id), nil)
}
