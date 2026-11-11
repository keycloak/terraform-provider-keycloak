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

func TestClientScopeLifecycleUsesClientScopeRepresentation(t *testing.T) {
	var methods []string
	var written ClientScope
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method+" "+r.URL.Path)
		switch r.Method + " " + r.URL.Path {
		case "POST /admin/realms/realm/client-scopes":
			if err := json.NewDecoder(r.Body).Decode(&written); err != nil {
				t.Errorf("decode create body: %v", err)
			}

			w.Header().Set("Location", "/admin/realms/realm/client-scopes/scope-id")
			w.WriteHeader(http.StatusCreated)
		case "GET /admin/realms/realm/client-scopes/scope-id":
			_, _ = w.Write([]byte(`{"id":"scope-id","name":"custom","description":"scope","protocol":"saml","attributes":{"saml.attr":"value"}}`))
		case "PUT /admin/realms/realm/client-scopes/scope-id",
			"DELETE /admin/realms/realm/client-scopes/scope-id":
			if r.Method == http.MethodPut {
				written = ClientScope{}
				if err := json.NewDecoder(r.Body).Decode(&written); err != nil {
					t.Errorf("decode update body: %v", err)
				}
			}
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

	scope := &ClientScope{
		RealmId:     "realm",
		Name:        "custom",
		Description: "scope",
		Protocol:    "saml",
		Attributes:  map[string]string{"saml.attr": "value"},
	}
	if err := client.NewClientScope(ctx, scope); err != nil {
		t.Fatalf("NewClientScope: %v", err)
	}
	if scope.Id != "scope-id" {
		t.Fatalf("expected created scope ID scope-id, got %q", scope.Id)
	}
	if written.Name != scope.Name || written.Protocol != scope.Protocol || !reflect.DeepEqual(written.Attributes, scope.Attributes) {
		t.Fatalf("unexpected scope representation sent to Keycloak: %+v", written)
	}

	fetched, err := client.GetClientScope(ctx, "realm", "scope-id")
	if err != nil {
		t.Fatalf("GetClientScope: %v", err)
	}
	if fetched.RealmId != "realm" || fetched.Protocol != "saml" || fetched.Attributes["saml.attr"] != "value" {
		t.Fatalf("unexpected scope returned from Keycloak: %+v", fetched)
	}

	fetched.Description = "updated"
	fetched.Attributes = map[string]string{}
	if err := client.UpdateClientScope(ctx, fetched); err != nil {
		t.Fatalf("UpdateClientScope: %v", err)
	}
	if written.Description != "updated" {
		t.Fatalf("expected updated description in request, got %q", written.Description)
	}
	if written.Attributes == nil || len(written.Attributes) != 0 {
		t.Fatalf("expected empty attributes map in update request, got %#v", written.Attributes)
	}
	if err := client.DeleteClientScope(ctx, "realm", "scope-id"); err != nil {
		t.Fatalf("DeleteClientScope: %v", err)
	}

	expectedMethods := []string{
		"POST /admin/realms/realm/client-scopes",
		"GET /admin/realms/realm/client-scopes/scope-id",
		"PUT /admin/realms/realm/client-scopes/scope-id",
		"DELETE /admin/realms/realm/client-scopes/scope-id",
	}
	if !reflect.DeepEqual(methods, expectedMethods) {
		t.Fatalf("unexpected Keycloak requests:\n got: %v\nwant: %v", methods, expectedMethods)
	}
}

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
