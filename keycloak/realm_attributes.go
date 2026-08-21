package keycloak

import (
	"context"
	"fmt"
)

type RealmAttributes struct {
	RealmId    string                 `json:"-"`
	Attributes map[string]interface{} `json:"attributes"`
}

func (keycloakClient *KeycloakClient) GetRealmAttributes(ctx context.Context, realmId string) (*RealmAttributes, error) {
	realm, err := keycloakClient.GetRealm(ctx, realmId)
	if err != nil {
		return nil, err
	}

	return &RealmAttributes{
		RealmId:    realmId,
		Attributes: realm.Attributes,
	}, nil
}

func (keycloakClient *KeycloakClient) UpdateRealmAttributes(ctx context.Context, realmId string, attributes *RealmAttributes) error {
	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s", realmId), attributes)
}
