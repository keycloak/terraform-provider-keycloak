package keycloak

import (
	"context"
	"fmt"
)

// https://www.keycloak.org/docs-api/4.2/rest-api/index.html#_component_resource

type component struct {
	Id           string              `json:"id,omitempty"`
	Name         string              `json:"name"`
	ProviderId   string              `json:"providerId"`
	ProviderType string              `json:"providerType"`
	ParentId     string              `json:"parentId"`
	SubType      string              `json:"subType,omitempty"`
	Config       map[string][]string `json:"config"`
}

func (component *component) getConfig(val string) string {
	if len(component.Config[val]) == 0 {
		return ""
	}

	return component.Config[val][0]
}

func (component *component) getConfigOk(val string) (string, bool) {
	if v, ok := component.Config[val]; ok {
		return v[0], true
	}

	return "", false
}

func (keycloakClient *KeycloakClient) DeleteComponent(ctx context.Context, realmId, id string) error {
	return keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/components/%s", realmId, id), nil)
}
