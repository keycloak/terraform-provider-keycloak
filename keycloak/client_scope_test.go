package keycloak

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/hashicorp/go-version"
)

func TestClientScopeAssignmentEndpointsUseResolvedIDs(t *testing.T) {
	var methods []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method+" "+r.URL.Path)
		switch r.Method + " " + r.URL.Path {
		case "GET /admin/realms/realm/client-scopes":
			_, _ = w.Write([]byte(`[{"id":"scope-id","name":"scope","protocol":"openid-connect"}]`))
		case "PUT /admin/realms/realm/clients/client/default-client-scopes/scope-id",
			"DELETE /admin/realms/realm/default-optional-client-scopes/scope-id":
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	client := &KeycloakClient{
		baseUrl: server.URL, httpClient: server.Client(), initialLogin: true, accessTokenProvided: true,
		version:           version.Must(version.NewVersion("26.8.2")),
		clientCredentials: &ClientCredentials{AccessToken: "test", TokenType: "Bearer"},
	}
	ctx := context.Background()

	scopes, err := client.ListClientScopes(ctx, "realm")
	if err != nil {
		t.Fatalf("ListClientScopes: %v", err)
	}
	if len(scopes) != 1 || scopes[0].Id != "scope-id" || scopes[0].Protocol != "openid-connect" {
		t.Fatalf("unexpected client scopes: %+v", scopes)
	}
	if err := client.AddClientScopeAssignment(ctx, "realm", "client", "default", "scope-id"); err != nil {
		t.Fatalf("AddClientScopeAssignment: %v", err)
	}
	if err := client.RemoveClientScopeAssignment(ctx, "realm", "", "optional", "scope-id"); err != nil {
		t.Fatalf("RemoveClientScopeAssignment: %v", err)
	}

	expectedMethods := []string{
		"GET /admin/realms/realm/client-scopes",
		"PUT /admin/realms/realm/clients/client/default-client-scopes/scope-id",
		"DELETE /admin/realms/realm/default-optional-client-scopes/scope-id",
	}
	if !reflect.DeepEqual(methods, expectedMethods) {
		t.Fatalf("unexpected Keycloak requests:\n got: %v\nwant: %v", methods, expectedMethods)
	}
}
