package keycloak

import "context"

type SystemInfo struct {
	ServerVersion string `json:"version"`
}

type ComponentType struct {
	Id         string                  `json:"id"`
	Properties []ComponentTypeProperty `json:"properties"`
}

type ComponentTypeProperty struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type ProviderType struct {
	Internal  bool                `json:"internal"`
	Providers map[string]Provider `json:"providers"`
}

type Provider struct {
}

type Theme struct {
	Name    string   `json:"name"`
	Locales []string `json:"locales,omitempty"`
}

type FeatureRepresentation struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

type ServerInfo struct {
	SystemInfo     SystemInfo                 `json:"systemInfo"`
	ComponentTypes map[string][]ComponentType `json:"componentTypes"`
	ProviderTypes  map[string]ProviderType    `json:"providers"`
	Themes         map[string][]Theme         `json:"themes"`
	Features       []FeatureRepresentation    `json:"features"`
}

func (serverInfo *ServerInfo) ThemeIsInstalled(t, themeName string) bool {
	if themes, ok := serverInfo.Themes[t]; ok {
		for _, theme := range themes {
			if theme.Name == themeName {
				return true
			}
		}
	}

	return false
}

// MultiValuedConfigKeys returns the set of config keys that Keycloak models as multi-valued
// (an array of individual values) for the given component type and provider id. Keycloak
// reports these as properties with a "type" of "MultivaluedString" or "MultivaluedList" in
// the server info component-type metadata.
func (serverInfo *ServerInfo) MultiValuedConfigKeys(componentType, providerId string) map[string]bool {
	keys := map[string]bool{}

	for _, ct := range serverInfo.ComponentTypes[componentType] {
		if ct.Id != providerId {
			continue
		}

		for _, p := range ct.Properties {
			if p.Type == "MultivaluedString" || p.Type == "MultivaluedList" {
				keys[p.Name] = true
			}
		}
	}

	return keys
}

func (serverInfo *ServerInfo) ComponentTypeIsInstalled(componentType, componentTypeId string) bool {
	if componentTypes, ok := serverInfo.ComponentTypes[componentType]; ok {
		for _, componentType := range componentTypes {
			if componentType.Id == componentTypeId {
				return true
			}
		}
	}

	return false
}

func (serverInfo *ServerInfo) getInstalledProvidersNames(providerType string) []string {
	providers := serverInfo.ProviderTypes[providerType].Providers
	keys := make([]string, 0, len(providers))
	for p := range providers {
		keys = append(keys, p)
	}
	return keys
}

func (serverInfo *ServerInfo) providerInstalled(providerType, providerName string) bool {
	providers := serverInfo.ProviderTypes[providerType].Providers
	for p := range providers {
		if p == providerName {
			return true
		}
	}
	return false
}

// FGAPv2IsEnabled reports whether the ADMIN_FINE_GRAINED_AUTHZ_V2 feature is
// enabled on the server. The result is memoised on the client because the
// feature flag is a server-level setting that does not change during a
// Terraform run; this avoids a /serverinfo round-trip on every guarded
// resource operation. Only successful lookups are cached, so a transient
// error does not poison the cache.
func (keycloakClient *KeycloakClient) FGAPv2IsEnabled(ctx context.Context) (bool, error) {
	keycloakClient.fgapv2Mu.Lock()
	defer keycloakClient.fgapv2Mu.Unlock()

	if keycloakClient.fgapv2Cached != nil {
		return *keycloakClient.fgapv2Cached, nil
	}

	serverInfo, err := keycloakClient.GetServerInfo(ctx)
	if err != nil {
		return false, err
	}

	enabled := false
	for _, f := range serverInfo.Features {
		if f.Name == "ADMIN_FINE_GRAINED_AUTHZ_V2" {
			enabled = f.Enabled
			break
		}
	}

	keycloakClient.fgapv2Cached = &enabled
	return enabled, nil
}

func (keycloakClient *KeycloakClient) GetServerInfo(ctx context.Context) (*ServerInfo, error) {
	var serverInfo ServerInfo

	err := keycloakClient.get(ctx, "/serverinfo", &serverInfo, nil)
	if err != nil {
		return nil, err
	}

	return &serverInfo, nil
}
