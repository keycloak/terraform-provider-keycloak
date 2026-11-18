package keycloak

import (
	"context"
	"fmt"
)

type ClientScope struct {
	Id          string            `json:"id,omitempty"`
	RealmId     string            `json:"-"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Protocol    string            `json:"protocol"`
	Attributes  map[string]string `json:"attributes"`
}

func (keycloakClient *KeycloakClient) ListClientScopes(ctx context.Context, realmId string) ([]ClientScope, error) {
	var scopes []ClientScope

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/client-scopes", realmId), &scopes, nil)
	if err != nil {
		return nil, err
	}

	return scopes, nil
}

func clientScopeAssignmentPath(realmId, clientId, assignmentType, scopeId string) (string, error) {
	if assignmentType != "default" && assignmentType != "optional" {
		return "", fmt.Errorf("unsupported client scope assignment type %q", assignmentType)
	}

	if clientId == "" {
		return fmt.Sprintf("/realms/%s/default-%s-client-scopes/%s", realmId, assignmentType, scopeId), nil
	}

	return fmt.Sprintf("/realms/%s/clients/%s/%s-client-scopes/%s", realmId, clientId, assignmentType, scopeId), nil
}

func (keycloakClient *KeycloakClient) AddClientScopeAssignment(ctx context.Context, realmId, clientId, assignmentType, scopeId string) error {
	path, err := clientScopeAssignmentPath(realmId, clientId, assignmentType, scopeId)
	if err != nil {
		return err
	}

	return keycloakClient.put(ctx, path, nil)
}

func (keycloakClient *KeycloakClient) RemoveClientScopeAssignment(ctx context.Context, realmId, clientId, assignmentType, scopeId string) error {
	path, err := clientScopeAssignmentPath(realmId, clientId, assignmentType, scopeId)
	if err != nil {
		return err
	}

	return keycloakClient.delete(ctx, path, nil)
}
