package keycloak

import (
	"context"
	"fmt"
)

// AuthenticationExecutionConfig https://www.keycloak.org/docs-api/latest/rest-api/index.html#AuthenticatorConfigRepresentation
type AuthenticationExecutionConfig struct {
	RealmId     string            `json:"-"`
	ExecutionId string            `json:"-"`
	Id          string            `json:"id,omitempty"`
	Alias       string            `json:"alias"`
	Config      map[string]string `json:"config"`
}

// NewAuthenticationExecutionConfig creates a new AuthenticationExecutionConfig
func (keycloakClient *KeycloakClient) NewAuthenticationExecutionConfig(ctx context.Context, config *AuthenticationExecutionConfig) (string, error) {
	_, location, err := keycloakClient.post(ctx, fmt.Sprintf("/realms/%s/authentication/executions/%s/config", config.RealmId, config.ExecutionId), config)
	if err != nil {
		return "", err
	}
	return getIdFromLocationHeader(location), nil
}

// GetAuthenticationExecutionConfig https://www.keycloak.org/docs-api/latest/rest-api/index.html#_get_adminrealmsrealmauthenticationconfigid
func (keycloakClient *KeycloakClient) GetAuthenticationExecutionConfig(ctx context.Context, config *AuthenticationExecutionConfig) error {
	return keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/authentication/config/%s", config.RealmId, config.Id), config, nil)
}

// UpdateAuthenticationExecutionConfig https://www.keycloak.org/docs-api/latest/rest-api/index.html#_put_adminrealmsrealmauthenticationconfigid
func (keycloakClient *KeycloakClient) UpdateAuthenticationExecutionConfig(ctx context.Context, config *AuthenticationExecutionConfig) error {
	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/authentication/config/%s", config.RealmId, config.Id), config)
}

// DeleteAuthenticationExecutionConfig https://www.keycloak.org/docs-api/latest/rest-api/index.html#_delete_adminrealmsrealmauthenticationconfigid
func (keycloakClient *KeycloakClient) DeleteAuthenticationExecutionConfig(ctx context.Context, config *AuthenticationExecutionConfig) error {
	return keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/authentication/config/%s", config.RealmId, config.Id), nil)
}
