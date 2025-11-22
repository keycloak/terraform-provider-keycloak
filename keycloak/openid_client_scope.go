package keycloak

import (
	"context"
	"fmt"
	"github.com/keycloak/terraform-provider-keycloak/keycloak/types"
	"regexp"
)

type OpenidClientScope struct {
	Id          string `json:"id,omitempty"`
	RealmId     string `json:"-"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Protocol    string `json:"protocol"`
	Attributes  struct {
		DisplayOnConsentScreen types.KeycloakBoolQuoted `json:"display.on.consent.screen"` // boolean in string form
		ConsentScreenText      string                   `json:"consent.screen.text"`
		GuiOrder               string                   `json:"gui.order"`
		IncludeInTokenScope    types.KeycloakBoolQuoted `json:"include.in.token.scope"` // boolean in string form
		IsDynamicScope         types.KeycloakBoolQuoted `json:"is.dynamic.scope"`       // boolean in string form
		DynamicScopeRegexp     string                   `json:"dynamic.scope.regexp,omitempty"`
	} `json:"attributes"`

	// Helper fields for easier access (not sent to API)
	Dynamic            bool   `json:"-"`
	DynamicScopeRegexp string `json:"-"`
}

type OpenidClientScopeFilterFunc func(*OpenidClientScope) bool

func (keycloakClient *KeycloakClient) NewOpenidClientScope(ctx context.Context, clientScope *OpenidClientScope) error {
	clientScope.Protocol = "openid-connect"
	clientScope.syncToAttributes()

	_, location, err := keycloakClient.post(ctx, fmt.Sprintf("/realms/%s/client-scopes", clientScope.RealmId), clientScope)
	if err != nil {
		return err
	}

	clientScope.Id = getIdFromLocationHeader(location)

	return nil
}

func (keycloakClient *KeycloakClient) GetOpenidClientScope(ctx context.Context, realmId, id string) (*OpenidClientScope, error) {
	var clientScope OpenidClientScope

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/client-scopes/%s", realmId, id), &clientScope, nil)
	if err != nil {
		return nil, err
	}

	clientScope.RealmId = realmId
	clientScope.syncFromAttributes()

	return &clientScope, nil
}

func (keycloakClient *KeycloakClient) GetOpenidDefaultClientScopes(ctx context.Context, realmId, clientId string) (*[]OpenidClientScope, error) {
	var clientScopes []OpenidClientScope

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/clients/%s/default-client-scopes", realmId, clientId), &clientScopes, nil)
	if err != nil {
		return nil, err
	}

	for _, clientScope := range clientScopes {
		clientScope.RealmId = realmId
	}

	return &clientScopes, nil
}

func (keycloakClient *KeycloakClient) GetOpenidOptionalClientScopes(ctx context.Context, realmId, clientId string) (*[]OpenidClientScope, error) {
	var clientScopes []OpenidClientScope

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/clients/%s/optional-client-scopes", realmId, clientId), &clientScopes, nil)
	if err != nil {
		return nil, err
	}

	for _, clientScope := range clientScopes {
		clientScope.RealmId = realmId
	}

	return &clientScopes, nil
}

func (keycloakClient *KeycloakClient) UpdateOpenidClientScope(ctx context.Context, clientScope *OpenidClientScope) error {
	clientScope.Protocol = "openid-connect"
	clientScope.syncToAttributes()

	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/client-scopes/%s", clientScope.RealmId, clientScope.Id), clientScope)
}

func (keycloakClient *KeycloakClient) DeleteOpenidClientScope(ctx context.Context, realmId, id string) error {
	return keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/client-scopes/%s", realmId, id), nil)
}

func (keycloakClient *KeycloakClient) ListOpenidClientScopesWithFilter(ctx context.Context, realmId string, filter OpenidClientScopeFilterFunc) ([]*OpenidClientScope, error) {
	var clientScopes []OpenidClientScope
	var openidClientScopes []*OpenidClientScope

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/client-scopes", realmId), &clientScopes, nil)
	if err != nil {
		return nil, err
	}

	for _, clientScope := range clientScopes {
		if clientScope.Protocol == "openid-connect" && filter(&clientScope) {
			scope := new(OpenidClientScope)
			*scope = clientScope

			scope.RealmId = realmId
			scope.syncFromAttributes()

			openidClientScopes = append(openidClientScopes, scope)
		}
	}

	return openidClientScopes, nil
}

func IncludeOpenidClientScopesMatchingNames(scopeNames []string) OpenidClientScopeFilterFunc {
	return func(scope *OpenidClientScope) bool {
		for _, scopeName := range scopeNames {
			if scopeName == scope.Name {
				return true
			}
		}

		return false
	}
}

// syncToAttributes copies helper fields to attributes before sending to API
func (scope *OpenidClientScope) syncToAttributes() {
	scope.Attributes.IsDynamicScope = types.KeycloakBoolQuoted(scope.Dynamic)
	scope.Attributes.DynamicScopeRegexp = scope.DynamicScopeRegexp
}

// syncFromAttributes copies attributes to helper fields after receiving from API
func (scope *OpenidClientScope) syncFromAttributes() {
	scope.Dynamic = bool(scope.Attributes.IsDynamicScope)
	scope.DynamicScopeRegexp = scope.Attributes.DynamicScopeRegexp
}

// IsDynamicScope Helper function to determine if a scope name is dynamic
func (scope *OpenidClientScope) IsDynamicScope() bool {
	return scope.Dynamic
}

// ValidateWildcardPattern validates that the wildcard pattern contains exactly one asterisk
// Keycloak validation: [^\s\*]*\*{1}[^\s\*]* (exactly one asterisk, no whitespace)
// Keycloak converts: "resource:*" -> "resource:(.*)" at runtime for regex matching
// We enforce stricter validation to prevent confusing patterns like "resource:.*" which won't work as expected
func (scope *OpenidClientScope) ValidateWildcardPattern() error {
	// If dynamic is true, regexp is required
	if scope.IsDynamicScope() && scope.DynamicScopeRegexp == "" {
		return fmt.Errorf("dynamic scope requires a wildcard pattern (e.g., 'resource:*', 'api:read:*')")
	}

	// Always validate the regexp pattern if provided (even if isDynamic is false)
	if scope.DynamicScopeRegexp != "" {
		// First check Keycloak's basic validation: exactly one asterisk, no whitespace
		matched, _ := regexp.MatchString(`^[^\s\*]*\*{1}[^\s\*]*$`, scope.DynamicScopeRegexp)
		if !matched {
			return fmt.Errorf("wildcard pattern must contain exactly one asterisk (*) and no whitespace")
		}

		// Additional validation: warn about regex metacharacters that won't work as expected
		// Keycloak replaces "*" with "(.*)" at runtime, so "resource:.*" would become "resource:.*" (no replacement)
		// which matches "resource:" + single char, not what users expect
		if regexp.MustCompile(`[.?+\[\](){}^$|\\]`).MatchString(scope.DynamicScopeRegexp) {
			return fmt.Errorf("wildcard pattern should not contain regex metacharacters (., ?, +, etc.). Use * as the only wildcard")
		}
	}

	return nil
}
