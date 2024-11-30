package keycloak

import (
	"context"
	"encoding/json"
	"fmt"
)

type OpenidClientInitialAccessToken struct {
	Id             string `json:"id,omitempty"`
	RealmId        string `json:"realm_id"`
	Count          int    `json:"count"`
	Expiration     int    `json:"expiration"`
	RemainingCount int    `json:"remaining_count"`
	Token          string `json:"token"`
}

type OpenidClientInitialAccessTokenCreateModel struct {
	Count      int `json:"count"`
	Expiration int `json:"expiration"`
}

func (keycloakClient *KeycloakClient) NewOpenidClientInitialAccessToken(ctx context.Context, initialAccessToken *OpenidClientInitialAccessToken) (*OpenidClientInitialAccessToken, error) {
	createModel := OpenidClientInitialAccessTokenCreateModel{
		Count:      initialAccessToken.Count,
		Expiration: initialAccessToken.Expiration,
	}

	body, _, err := keycloakClient.post(ctx, fmt.Sprintf("/realms/%s/clients-initial-access", initialAccessToken.RealmId), createModel)
	if err != nil {
		return nil, err
	}

	var createdToken OpenidClientInitialAccessToken
	conversionErr := json.Unmarshal(body, &createdToken)
	if conversionErr != nil {
		return nil, conversionErr
	}
	createdToken.RealmId = initialAccessToken.RealmId

	initialAccessToken.Id = createdToken.Id
	initialAccessToken.Token = createdToken.Token
	return &createdToken, nil
}

func (keycloakClient *KeycloakClient) GetClientInitialAccessTokens(ctx context.Context, realmId string) (*[]OpenidClientInitialAccessToken, error) {

	var initialAccessTokens []OpenidClientInitialAccessToken

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/clients-initial-access", realmId), &initialAccessTokens, nil)
	if err != nil {
		return nil, err
	}

	for k := range initialAccessTokens {
		initialAccessTokens[k].RealmId = realmId
	}

	return &initialAccessTokens, nil
}

func (keycloakClient *KeycloakClient) GetClientInitialAccessToken(ctx context.Context, realmId, id string) (*OpenidClientInitialAccessToken, error) {

	var initialAccessTokens []OpenidClientInitialAccessToken

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/clients-initial-access", realmId), &initialAccessTokens, nil)
	if err != nil {
		return nil, err
	}

	var token *OpenidClientInitialAccessToken
	for k := range initialAccessTokens {
		initialAccessTokens[k].RealmId = realmId
		if initialAccessTokens[k].Id == id {
			token = &initialAccessTokens[k]
		}
	}

	return token, nil
}

func (keycloakClient *KeycloakClient) DeleteClientInitialAccessToken(ctx context.Context, realmId, id string) error {
	return keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/clients-initial-access/%s", realmId, id), nil)
}
