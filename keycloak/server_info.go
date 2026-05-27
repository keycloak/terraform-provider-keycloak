package keycloak

import "context"

type SystemInfo struct {
	ServerVersion string `json:"version"`
}

type ComponentType struct {
	Id string `json:"id"`
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
	Name      string `json:"name"`
	IsEnabled bool   `json:"isEnabled"`
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

func (keycloakClient *KeycloakClient) FGAPv2IsEnabled(ctx context.Context) (bool, error) {
	serverInfo, err := keycloakClient.GetServerInfo(ctx)
	if err != nil {
		return false, err
	}
	for _, f := range serverInfo.Features {
		if f.Name == "ADMIN_FINE_GRAINED_AUTHZ_V2" {
			return f.IsEnabled, nil
		}
	}
	return false, nil
}

func (keycloakClient *KeycloakClient) GetServerInfo(ctx context.Context) (*ServerInfo, error) {
	var serverInfo ServerInfo

	err := keycloakClient.get(ctx, "/serverinfo", &serverInfo, nil)
	if err != nil {
		return nil, err
	}

	return &serverInfo, nil
}
