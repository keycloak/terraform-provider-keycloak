package keycloak

import (
	"context"
	"fmt"
)

type RealmEventsConfig struct {
	AdminEventsDetailsEnabled bool     `json:"adminEventsDetailsEnabled"`
	AdminEventsEnabled        bool     `json:"adminEventsEnabled"`
	EnabledEventTypes         []string `json:"enabledEventTypes"`
	EventsEnabled             bool     `json:"eventsEnabled"`
	EventsExpiration          int      `json:"eventsExpiration"`
	EventsListeners           []string `json:"eventsListeners"`
}

func (keycloakClient *KeycloakClient) GetRealmEventsConfig(ctx context.Context, realmId string) (*RealmEventsConfig, error) {
	var realmEventsConfig RealmEventsConfig

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/events/config", realmId), &realmEventsConfig, nil)
	if err != nil {
		return nil, err
	}

	return &realmEventsConfig, nil
}

func (keycloakClient *KeycloakClient) UpdateRealmEventsConfig(ctx context.Context, realmId string, realmEventsConfig *RealmEventsConfig) error {
	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/events/config", realmId), realmEventsConfig)
}
