package keycloak

import (
	"fmt"
)

func (keycloakClient *KeycloakClient) GetOpenidRealmDefaultClientScopes(realmId string) (*[]OpenidClientScope, error) {
	var clientScopes []OpenidClientScope

	err := keycloakClient.get(fmt.Sprintf("/realms/%s/default-default-client-scopes", realmId), &clientScopes, nil)
	if err != nil {
		return nil, err
	}

	return &clientScopes, nil
}

func (keycloakClient *KeycloakClient) GetOpenidRealmDefaultClientScope(realmId, clientScopeId string) (*OpenidClientScope, error) {
	var clientScopes []OpenidClientScope

	err := keycloakClient.get(fmt.Sprintf("/realms/%s/default-default-client-scopes", realmId), &clientScopes, nil)
	if err != nil {
		return nil, err
	}

	for _, clientScope := range clientScopes {
		if clientScope.Id == clientScopeId {
			return &clientScope, nil
		}
	}

	return nil, err
}

func (keycloakClient *KeycloakClient) PutOpenidRealmDefaultClientScope(realmId, clientScopeId string) error {
	return keycloakClient.put(fmt.Sprintf("/realms/%s/default-default-client-scopes/%s", realmId, clientScopeId), nil)
}

func (keycloakClient *KeycloakClient) DeleteOpenidRealmDefaultClientScope(realmId, clientScopeId string) error {
	return keycloakClient.delete(fmt.Sprintf("/realms/%s/default-default-client-scopes/%s", realmId, clientScopeId), nil)
}
