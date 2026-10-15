package keycloak

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/hashicorp/go-version"
)

func TestRealmZeroValueUpdates(t *testing.T) {
	for _, tc := range []struct {
		name          string
		serverVersion string
		create        bool
		idle, maximum int
		enabled       bool
	}{
		{name: "creation keeps omitted server defaults", create: true},
		{name: "update resets zero and false values"},
		{name: "update retains positive and true values", idle: 123, maximum: 456, enabled: true},
		{name: "26.0 update omits unsupported false admin permissions", serverVersion: "26.0.8"},
		{name: "26.0 preserves explicit true validation", serverVersion: "26.0.8", enabled: true},
		{name: "26.1 update resets admin permissions", serverVersion: "26.1.0"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			realm := &Realm{Realm: "fixture", Enabled: true, SslRequired: "all", AccessTokenLifespan: 300,
				SsoSessionIdleTimeoutRememberMe: tc.idle, SsoSessionMaxLifespanRememberMe: tc.maximum,
				OfflineSessionMaxLifespanEnabled: tc.enabled, Attributes: map[string]interface{}{"fixture": "preserve"}}
			theme := ""
			if tc.enabled {
				theme = "keycloak"
			}
			realm.LoginTheme, realm.AccountTheme, realm.AdminTheme, realm.EmailTheme = theme, theme, theme, theme
			realm.OrganizationsEnabled, realm.AdminPermissionsEnabled = tc.enabled, tc.enabled
			before, err := json.Marshal(realm)
			if err != nil {
				t.Fatal(err)
			}
			var want map[string]interface{}
			if err := json.Unmarshal(before, &want); err != nil {
				t.Fatal(err)
			}
			if tc.create {
				delete(want, "ssoSessionIdleTimeoutRememberMe")
				delete(want, "ssoSessionMaxLifespanRememberMe")
				delete(want, "offlineSessionMaxLifespanEnabled")
			} else {
				want["ssoSessionIdleTimeoutRememberMe"] = float64(tc.idle)
				want["ssoSessionMaxLifespanRememberMe"] = float64(tc.maximum)
				want["offlineSessionMaxLifespanEnabled"] = tc.enabled
			}
			if !tc.create {
				for k, v := range map[string]interface{}{
					"loginTheme": theme, "accountTheme": theme, "adminTheme": theme, "emailTheme": theme,
					"organizationsEnabled": tc.enabled, "adminPermissionsEnabled": tc.enabled,
				} {
					want[k] = v
				}
			}
			serverVersion := tc.serverVersion
			if serverVersion == "" {
				serverVersion = "26.8.1"
			}
			if serverVersion == "26.0.8" && !tc.enabled {
				delete(want, "adminPermissionsEnabled")
			}
			writes := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				writes++
				method, path, status := "PUT", "/admin/realms/fixture", http.StatusNoContent
				if tc.create {
					method, path, status = "POST", "/admin/realms", http.StatusCreated
				}
				if r.Method != method || r.URL.Path != path {
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
				}
				var body map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
				}
				if !reflect.DeepEqual(body, want) {
					t.Errorf("request = %#v, want %#v", body, want)
				}
				w.WriteHeader(status)
			}))
			defer server.Close()
			client := &KeycloakClient{version: version.Must(version.NewVersion(serverVersion)), baseUrl: server.URL, httpClient: server.Client(), initialLogin: true, clientCredentials: &ClientCredentials{AccessToken: "fixture", TokenType: "Bearer"}}
			if tc.create {
				err = client.NewRealm(context.Background(), realm)
			} else {
				err = client.UpdateRealm(context.Background(), realm)
			}
			if err != nil || writes != 1 {
				t.Fatalf("operation = %v, writes = %d", err, writes)
			}
			after, err := json.Marshal(realm)
			if err != nil || string(before) != string(after) {
				t.Fatal("caller realm was mutated")
			}
		})
	}
}
