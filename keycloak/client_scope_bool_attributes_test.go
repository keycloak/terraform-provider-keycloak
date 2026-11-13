package keycloak

import (
	"context"
	"fmt"
	"net/http"
	"testing"
)

// Keycloak's ClientScopeModel treats display.on.consent.screen,
// include.in.token.scope and include.in.openid.provider.metadata as true when
// they are not stored, and otherwise parses them with Boolean.parseBoolean.
func TestClientScopeBoolAttributesFollowKeycloakDefaults(t *testing.T) {
	for _, tc := range []struct {
		name       string
		attributes string
		want       bool
	}{
		{name: "not stored", attributes: `{}`, want: true},
		{name: "true", attributes: `{"%s": "true"}`, want: true},
		{name: "upper case true", attributes: `{"%s": "TRUE"}`, want: true},
		{name: "false", attributes: `{"%s": "false"}`, want: false},
		{name: "empty", attributes: `{"%s": ""}`, want: false},
		{name: "not a boolean", attributes: `{"%s": "yes"}`, want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, key := range []string{"display.on.consent.screen", "include.in.token.scope", "include.in.openid.provider.metadata"} {
				attributes := tc.attributes
				if attributes != `{}` {
					attributes = fmt.Sprintf(attributes, key)
				}
				client, _ := urlCapturingClient(t, func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					fmt.Fprintf(w, `{"id": "scope-id", "name": "offline_access", "protocol": "openid-connect", "attributes": %s}`, attributes)
				})

				scope, err := client.GetOpenidClientScope(context.Background(), "realm", "scope-id")
				if err != nil {
					t.Fatal(err)
				}
				got := map[string]bool{
					"display.on.consent.screen":           bool(scope.Attributes.DisplayOnConsentScreen),
					"include.in.token.scope":              bool(scope.Attributes.IncludeInTokenScope),
					"include.in.openid.provider.metadata": bool(scope.Attributes.IncludeInOpenidProviderMetadata),
				}
				for k, v := range got {
					want := true
					if k == key {
						want = tc.want
					}
					if v != want {
						t.Errorf("%s=%s: %s read as %t, want %t", key, attributes, k, v, want)
					}
				}
				if _, ok := scope.Attributes.ExtraConfig[key]; ok {
					t.Errorf("%s left in extra_config", key)
				}
			}
		})
	}
}

func TestSamlClientScopeDisplayOnConsentScreenNotStoredIsTrue(t *testing.T) {
	client, _ := urlCapturingClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id": "scope-id", "name": "role_list", "protocol": "saml", "attributes": {}}`)
	})

	scope, err := client.GetSamlClientScope(context.Background(), "realm", "scope-id")
	if err != nil {
		t.Fatal(err)
	}
	if !scope.Attributes.DisplayOnConsentScreen {
		t.Error("display.on.consent.screen not stored: read as false, want true")
	}
}
