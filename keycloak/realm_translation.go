package keycloak

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type RealmLocaleTranslation struct {
	Locale       string            `json:"locale"`
	Translations map[string]string `json:"translations"`
}

func (keycloakClient *KeycloakClient) UpdateRealmTranslations(ctx context.Context, realmId string, locale string, translations map[string]string) error {
	var existingTranslations map[string]string

	data, _ := keycloakClient.getRaw(ctx, fmt.Sprintf("/realms/%s/localization/%s", realmId, locale), nil)
	err := json.Unmarshal(data, &existingTranslations)
	if err != nil {
		return nil
	}
	translationsToDelete := make([]string, 0)
	for key := range existingTranslations {
		if _, exists := translations[key]; !exists {
			translationsToDelete = append(translationsToDelete, key)
		}
	}
	for _, key := range translationsToDelete {
		err := keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/localization/%s/%s", realmId, locale, key), nil)
		if err != nil {
			return err
		}
	}
	for key, value := range translations {
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

func (keycloakClient *KeycloakClient) GetRealmTranslations(ctx context.Context, realmId string, locale string) (*map[string]string, error) {
	keyValues := make(map[string]string)
	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/localization/%s", realmId, locale), &keyValues, nil)
	if err != nil {
		return nil, err
	}
	return &keyValues, nil
}

func (keycloakClient *KeycloakClient) DeleteRealmTranslations(ctx context.Context, realmId string, locale string, translations map[string]string) error {
	for key := range translations {
		err := keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/localization/%s/%s", realmId, locale, key), nil)
		if err != nil {
			return err
		}
	}
	return nil
}
