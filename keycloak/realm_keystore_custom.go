package keycloak

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

const realmKeystoreProviderType = "org.keycloak.keys.KeyProvider"

// RealmKeystoreSecretValue is the value Keycloak returns instead of the actual value of a config
// property that is flagged as secret.
const RealmKeystoreSecretValue = "**********"

type RealmKeystoreCustom struct {
	Id         string
	Name       string
	RealmId    string
	ParentId   string
	ProviderId string
	Config     map[string]string
}

// RealmKeystoreCustomConfigProperties returns the config properties that the given key provider
// declares in the server info, keyed by name. Keycloak drops undeclared config keys when it returns
// or updates a component, so only these keys can be managed.
func (keycloakClient *KeycloakClient) RealmKeystoreCustomConfigProperties(ctx context.Context, providerId string) (map[string]ComponentTypeProperty, error) {
	serverInfo, err := keycloakClient.GetServerInfoCached(ctx)
	if err != nil {
		return nil, err
	}

	return serverInfo.ConfigProperties(realmKeystoreProviderType, providerId), nil
}

// ValidateRealmKeystoreCustom checks that the key provider is installed and that every config key is
// declared by it and has a value. Keycloak silently drops undeclared keys and empty values, which
// would otherwise show up as a permanent diff.
func (keycloakClient *KeycloakClient) ValidateRealmKeystoreCustom(ctx context.Context, keystore *RealmKeystoreCustom) error {
	serverInfo, err := keycloakClient.GetServerInfoCached(ctx)
	if err != nil {
		return err
	}

	if !serverInfo.ComponentTypeIsInstalled(realmKeystoreProviderType, keystore.ProviderId) {
		return fmt.Errorf("key provider with id %s is not installed on the server", keystore.ProviderId)
	}

	properties := serverInfo.ConfigProperties(realmKeystoreProviderType, keystore.ProviderId)

	for k, v := range keystore.Config {
		if _, ok := properties[k]; !ok {
			if len(properties) == 0 {
				return fmt.Errorf("config key %q is not declared by key provider %s, which declares no config keys", k, keystore.ProviderId)
			}

			declared := make([]string, 0, len(properties))
			for name := range properties {
				declared = append(declared, name)
			}
			sort.Strings(declared)

			return fmt.Errorf("config key %q is not declared by key provider %s, declared keys are: %s", k, keystore.ProviderId, strings.Join(declared, ", "))
		}

		if strings.TrimSpace(v) == "" {
			return fmt.Errorf("config key %q must not be empty, remove the key instead", k)
		}
	}

	return nil
}

func (keycloakClient *KeycloakClient) multiValuedRealmKeystoreConfigKeys(ctx context.Context, providerId string) (map[string]bool, error) {
	serverInfo, err := keycloakClient.GetServerInfoCached(ctx)
	if err != nil {
		return nil, err
	}

	return serverInfo.MultiValuedConfigKeys(realmKeystoreProviderType, providerId), nil
}

// convertFromRealmKeystoreCustomToComponent builds the component for a create or update. Keys in
// removedKeys are sent without a value, which makes Keycloak delete them; keys that are left out of
// an update are kept as they are.
func convertFromRealmKeystoreCustomToComponent(keystore *RealmKeystoreCustom, multiValuedKeys map[string]bool, removedKeys []string) *component {
	config := map[string][]string{}
	for k, v := range keystore.Config {
		if multiValuedKeys[k] && strings.Contains(v, ",") {
			parts := strings.Split(v, ",")
			for i := range parts {
				parts[i] = strings.TrimSpace(parts[i])
			}
			config[k] = parts
		} else {
			config[k] = []string{v}
		}
	}

	for _, k := range removedKeys {
		if _, ok := config[k]; !ok {
			config[k] = []string{}
		}
	}

	// ParentId is deliberately not set, Keycloak defaults to the realm's internal id
	return &component{
		Id:           keystore.Id,
		Name:         keystore.Name,
		ProviderId:   keystore.ProviderId,
		ProviderType: realmKeystoreProviderType,
		Config:       config,
	}
}

// convertFromComponentToRealmKeystoreCustom joins all values of a config key, not only those of declared
// multi-valued keys, so no value is lost if a key provider stores several values for another key.
func convertFromComponentToRealmKeystoreCustom(c *component, realmId string) *RealmKeystoreCustom {
	config := map[string]string{}
	for k, vals := range c.Config {
		if len(vals) == 0 {
			continue
		}

		config[k] = strings.Join(vals, ",")
	}

	return &RealmKeystoreCustom{
		Id:         c.Id,
		Name:       c.Name,
		RealmId:    realmId,
		ParentId:   c.ParentId,
		ProviderId: c.ProviderId,
		Config:     config,
	}
}

func (keycloakClient *KeycloakClient) NewRealmKeystoreCustom(ctx context.Context, keystore *RealmKeystoreCustom) error {
	multiValuedKeys, err := keycloakClient.multiValuedRealmKeystoreConfigKeys(ctx, keystore.ProviderId)
	if err != nil {
		return err
	}

	_, location, err := keycloakClient.post(ctx, fmt.Sprintf("/realms/%s/components", keystore.RealmId), convertFromRealmKeystoreCustomToComponent(keystore, multiValuedKeys, nil))
	if err != nil {
		return err
	}

	keystore.Id = getIdFromLocationHeader(location)

	return nil
}

func (keycloakClient *KeycloakClient) GetRealmKeystoreCustom(ctx context.Context, realmId, id string) (*RealmKeystoreCustom, error) {
	var component *component

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/components/%s", realmId, id), &component, nil)
	if err != nil {
		return nil, err
	}

	return convertFromComponentToRealmKeystoreCustom(component, realmId), nil
}

// GetRealmKeystoresCustom returns all key provider components of a realm with the given name and provider id.
func (keycloakClient *KeycloakClient) GetRealmKeystoresCustom(ctx context.Context, realmId, name, providerId string) ([]*RealmKeystoreCustom, error) {
	var components []*component

	params := map[string]string{
		"type": realmKeystoreProviderType,
		"name": name,
	}

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/components", realmId), &components, params)
	if err != nil {
		return nil, err
	}

	keystores := make([]*RealmKeystoreCustom, 0, len(components))
	for _, c := range components {
		// the name filter of the components endpoint is exact, but check it anyway to be safe
		if c.Name != name || c.ProviderId != providerId {
			continue
		}

		keystores = append(keystores, convertFromComponentToRealmKeystoreCustom(c, realmId))
	}

	return keystores, nil
}

func (keycloakClient *KeycloakClient) UpdateRealmKeystoreCustom(ctx context.Context, keystore *RealmKeystoreCustom, removedKeys []string) error {
	multiValuedKeys, err := keycloakClient.multiValuedRealmKeystoreConfigKeys(ctx, keystore.ProviderId)
	if err != nil {
		return err
	}

	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/components/%s", keystore.RealmId, keystore.Id), convertFromRealmKeystoreCustomToComponent(keystore, multiValuedKeys, removedKeys))
}

func (keycloakClient *KeycloakClient) DeleteRealmKeystoreCustom(ctx context.Context, realmId, id string) error {
	return keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/components/%s", realmId, id), nil)
}
