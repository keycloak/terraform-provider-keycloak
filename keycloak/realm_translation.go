package keycloak

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (keycloakClient *KeycloakClient) UpdateRealmLocalizationTexts(ctx context.Context, realmId string, locale string, texts map[string]string) error {
	var existingtexts map[string]string

	data, _ := keycloakClient.getRaw(ctx, fmt.Sprintf("/realms/%s/localization/%s", realmId, locale), nil)
	err := json.Unmarshal(data, &existingtexts)
	if err != nil {
		return nil
	}
	textsToDelete := make([]string, 0)
	for key := range existingtexts {
		if _, exists := texts[key]; !exists {
			textsToDelete = append(textsToDelete, key)
		}
	}
	for _, key := range textsToDelete {
		err := keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/localization/%s/%s", realmId, locale, key), nil)
		if err != nil {
			return err
		}
	}
	for key, value := range texts {
		err := keycloakClient.putPlain(ctx, fmt.Sprintf("/realms/%s/localization/%s/%s", realmId, locale, key), value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (keycloakClient *KeycloakClient) putPlain(ctx context.Context, path string, requestBody string) error {
	resourceUrl := keycloakClient.baseUrl + apiUrl + path
	request, err := http.NewRequestWithContext(ctx, http.MethodPut, resourceUrl, bytes.NewReader([]byte(requestBody)))
	if err != nil {
		return err
	}
	request.Header.Set("Content-type", "text/plain")
	_, _, err = keycloakClient.sendRequest(ctx, request, []byte(requestBody))
	return err
}

func (keycloakClient *KeycloakClient) GetRealmLocalizationTexts(ctx context.Context, realmId string, locale string) (*map[string]string, error) {
	keyValues := make(map[string]string)
	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/localization/%s", realmId, locale), &keyValues, nil)
	if err != nil {
		return nil, err
	}
	return &keyValues, nil
}

func (keycloakClient *KeycloakClient) DeleteRealmLocalizationTexts(ctx context.Context, realmId string, locale string, texts map[string]string) error {
	for key := range texts {
		err := keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/localization/%s/%s", realmId, locale, key), nil)
		if err != nil {
			return err
		}
	}
	return nil
}
