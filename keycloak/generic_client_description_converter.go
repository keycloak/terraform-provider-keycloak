package keycloak

import (
	"context"
	"encoding/json"
	"fmt"
)

// GenericClientRepresentation https://www.keycloak.org/docs-api/latest/javadocs/org/keycloak/representations/idm/ClientRepresentation.html
type GenericClientRepresentation struct {
	Access                             map[string]string        `json:"access"`
	AdminUrl                           string                   `json:"adminUrl"`
	Attributes                         map[string]string        `json:"attributes"`
	AuthenticationFlowBindingOverrides map[string]string        `json:"authenticationFlowBindingOverrides"`
	AuthorizationServicesEnabled       bool                     `json:"authorizationServicesEnabled"`
	AuthorizationSettings              map[string]string        `json:"authorizationSettings"`
	BaseUrl                            string                   `json:"baseUrl"`
	BearerOnly                         bool                     `json:"bearerOnly"`
	ClientAuthenticatorType            string                   `json:"clientAuthenticatorType"`
	ClientId                           string                   `json:"clientId"`
	ConsentRequired                    string                   `json:"consentRequired"`
	DefaultClientScopes                []string                 `json:"defaultClientScopes"`
	DefaultRoles                       []string                 `json:"defaultRoles"`
	Description                        string                   `json:"description"`
	DirectAccessGrantsEnabled          bool                     `json:"directAccessGrantsEnabled"`
	Enabled                            bool                     `json:"enabled"`
	FrontchannelLogout                 bool                     `json:"frontchannelLogout"`
	FullScopeAllowed                   bool                     `json:"fullScopeAllowed"`
	Id                                 string                   `json:"id"`
	ImplicitFlowEnabled                bool                     `json:"implicitFlowEnabled"`
	Name                               string                   `json:"name"`
	NotBefore                          int                      `json:"notBefore"`
	OptionalClientScopes               []string                 `json:"optionalClientScopes"`
	Origin                             string                   `json:"origin"`
	Protocol                           string                   `json:"protocol"`
	ProtocolMappers                    []*GenericProtocolMapper `json:"protocolMappers"`
	PublicClient                       bool                     `json:"publicClient"`
	RedirectUris                       []string                 `json:"redirectUris"`
	RegisteredNodes                    map[string]string        `json:"registeredNodes"`
	RegistrationAccessToken            string                   `json:"registrationAccessToken"`
	RootUrl                            string                   `json:"rootUrl"`
	Secret                             string                   `json:"secret"`
	ServiceAccountsEnabled             bool                     `json:"serviceAccountsEnabled"`
	StandardFlowEnabled                bool                     `json:"standardFlowEnabled"`
	SurrogateAuthRequired              bool                     `json:"surrogateAuthRequired"`
	WebOrigins                         []string                 `json:"webOrigins"`
}

func (keycloakClient *KeycloakClient) NewGenericClientDescription(ctx context.Context, realmId string, body string) (*GenericClientRepresentation, error) {
	var genericClientRepresentation GenericClientRepresentation

	result, err := keycloakClient.sendRaw(ctx, fmt.Sprintf("/realms/%s/client-description-converter", realmId), []byte(body))

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(result, &genericClientRepresentation)

	if err != nil {
		return nil, err
	}

	return &genericClientRepresentation, nil
}
