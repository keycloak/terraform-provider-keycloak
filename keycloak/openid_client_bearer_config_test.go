package keycloak

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenidClientBearerConfiguration(t *testing.T) {
	writes := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/admin/serverinfo":
			fmt.Fprint(w, `{"systemInfo":{"version":"26.8.1"},"themes":{}}`)
		case strings.HasSuffix(r.URL.Path, "/client-secret"):
			t.Error("scope attachment must not request credentials")
			w.WriteHeader(403)
		case r.URL.Path == "/admin/realms/fixture/clients/id":
			fmt.Fprint(w, `{"id":"id","clientId":"realm-management","bearerOnly":true,"standardFlowEnabled":true}`)
		case r.URL.Path == "/admin/realms/fixture/client-scopes":
			fmt.Fprint(w, `[{"id":"scope","name":"profile","protocol":"openid-connect"}]`)
		case r.Method == "GET" && strings.Contains(r.URL.Path, "-client-scopes"):
			fmt.Fprint(w, `[]`)
		case r.Method == "PUT" && strings.HasSuffix(r.URL.Path, "-client-scopes/scope"):
			writes++
			w.WriteHeader(204)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer server.Close()
	api := &KeycloakClient{baseUrl: server.URL, httpClient: server.Client(), clientCredentials: &ClientCredentials{AccessToken: "fixture-token", TokenType: "Bearer"}, accessTokenProvided: true}
	ctx := context.Background()
	for _, test := range []struct {
		name   string
		client OpenidClient
		valid  bool
	}{
		{"factory_bearer", OpenidClient{Id: "existing", BearerOnly: true, StandardFlowEnabled: true}, true},
		{"new_bearer_standard_rejected", OpenidClient{BearerOnly: true, StandardFlowEnabled: true}, false},
		{"public_browser_requires_redirect", OpenidClient{PublicClient: true, StandardFlowEnabled: true}, false},
		{"bearer_implicit_rejected", OpenidClient{BearerOnly: true, ImplicitFlowEnabled: true}, false},
		{"bearer_password_rejected", OpenidClient{BearerOnly: true, DirectAccessGrantsEnabled: true}, false},
		{"bearer_service_accounts_rejected", OpenidClient{BearerOnly: true, ServiceAccountsEnabled: true}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := api.ValidateOpenidClient(ctx, &test.client)
			if (err == nil) != test.valid {
				t.Fatalf("validation error=%v, valid=%v", err, test.valid)
			}
		})
	}
	for _, kind := range []string{"default", "optional"} {
		if err := api.attachOpenidClientScopes(ctx, "fixture", "id", kind, []string{"profile"}); err != nil {
			t.Fatal(err)
		}
	}
	if writes != 2 {
		t.Fatalf("scope writes=%d, want 2", writes)
	}
}
