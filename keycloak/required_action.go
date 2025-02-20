package keycloak

import (
	"context"
	"fmt"
)

type RequiredAction struct {
	Id            string            `json:"-"`
	RealmId       string            `json:"-"`
	Alias         string            `json:"alias"`
	Name          string            `json:"name"`
	ProviderId    string            `json:"providerId"`
	Enabled       bool              `json:"enabled"`
	DefaultAction bool              `json:"defaultAction"`
	Priority      int               `json:"priority"`
	Config        map[string]string `json:"config"`
}

func (requiredActions *RequiredAction) getConfig(val string) string {
	if len(requiredActions.Config[val]) == 0 {
		return ""
	}
	return requiredActions.Config[val]
}

func (requiredActions *RequiredAction) getConfigOk(val string) (string, bool) {
	if v, ok := requiredActions.Config[val]; ok {
		return v, true
	}
	return "", false
}
func (keycloakClient *KeycloakClient) GetRequiredActions(ctx context.Context, realmId string) ([]*RequiredAction, error) {
	var requiredActions []*RequiredAction

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/authentication/required-actions", realmId), &requiredActions, nil)
	if err != nil {
		return nil, err
	}

	for _, requiredAction := range requiredActions {
		requiredAction.RealmId = realmId
	}

	return requiredActions, nil
}

func (keycloakClient *KeycloakClient) GetUnregisteredRequiredActions(ctx context.Context, realmId string) ([]*RequiredAction, error) {
	var unregisteredRequiredActions []*RequiredAction

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/authentication/unregistered-required-actions", realmId), &unregisteredRequiredActions, nil)
	if err != nil {
		return nil, err
	}

	for _, unregisteredRequiredAction := range unregisteredRequiredActions {
		unregisteredRequiredAction.RealmId = realmId
	}

	return unregisteredRequiredActions, nil
}

func (keycloakClient *KeycloakClient) GetRequiredAction(ctx context.Context, realmId string, alias string) (*RequiredAction, error) {
	var requiredAction RequiredAction

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/authentication/required-actions/%s", realmId, alias), &requiredAction, nil)
	if err != nil {
		return nil, err
	}
	requiredAction.RealmId = realmId
	return &requiredAction, nil
}

func (keycloakClient *KeycloakClient) RegisterRequiredAction(ctx context.Context, requiredAction *RequiredAction) error {
	_, _, err := keycloakClient.post(ctx, fmt.Sprintf("/realms/%s/authentication/register-required-action", requiredAction.RealmId), requiredAction)
	return err
}

func (keycloakClient *KeycloakClient) CreateRequiredAction(ctx context.Context, requiredAction *RequiredAction) error {
	requiredAction.Id = fmt.Sprintf("%s/%s", requiredAction.RealmId, requiredAction.Alias)
	return keycloakClient.UpdateRequiredAction(ctx, requiredAction)
}

func (keycloakClient *KeycloakClient) UpdateRequiredAction(ctx context.Context, requiredAction *RequiredAction) error {

	err := keycloakClient.ValidateRequiredAction(ctx, requiredAction)
	if err != nil {
		return err
	}

	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/authentication/required-actions/%s", requiredAction.RealmId, requiredAction.Alias), requiredAction)
}

func (keycloakClient *KeycloakClient) DeleteRequiredAction(ctx context.Context, realmName string, alias string) error {
	err := keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/authentication/required-actions/%s", realmName, alias), nil)
	if err != nil {
		// For whatever reason, this fails sometimes with a 500 during acceptance tests. try again
		return keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/authentication/required-actions/%s", realmName, alias), nil)
	}

	return nil
}

func (keycloakClient *KeycloakClient) ValidateRequiredAction(ctx context.Context, requiredAction *RequiredAction) error {
	serverInfo, err := keycloakClient.GetServerInfo(ctx)
	if err != nil {
		return err
	}

	if requiredAction.DefaultAction && !requiredAction.Enabled {
		return fmt.Errorf("validation error: a 'default' required action should be enabled, set 'defaultAction' to 'false' or set 'enabled' to 'true'")
	}

	if !serverInfo.providerInstalled("required-action", requiredAction.Alias) {
		return fmt.Errorf("validation error: required action \"%s\" does not exist on the server, installed providers: %s", requiredAction.Alias, serverInfo.getInstalledProvidersNames("required-action"))
	}

	return nil
}
