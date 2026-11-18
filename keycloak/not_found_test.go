package keycloak

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/go-version"
)

// Regression tests for https://github.com/keycloak/terraform-provider-keycloak/issues/1816:
// a 404 removes a resource from the state, so only Keycloak's own JSON 404 may count as "not found".
func notFoundServer(t *testing.T, contentType, body string) *KeycloakClient {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if contentType != "" {
			w.Header().Set("Content-Type", contentType)
		}
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)

	return &KeycloakClient{
		baseUrl: server.URL, httpClient: server.Client(), initialLogin: true, accessTokenProvided: true,
		version:           version.Must(version.NewVersion("26.8.2")),
		clientCredentials: &ClientCredentials{AccessToken: "test", TokenType: "Bearer"},
	}
}

func TestNotFoundFromKeycloakIs404(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		body        string
	}{
		{"missing realm", "application/json", `{"error":"Realm not found."}`},
		{"with charset", "application/json;charset=UTF-8", `{"error":"Could not find client"}`},
		{"problem json", "application/problem+json", `{"status":404}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := notFoundServer(t, tt.contentType, tt.body)

			err := client.get(context.Background(), "/realms/gone", &Realm{}, nil)
			if !ErrorIs404(err) {
				t.Fatalf("want a 404 ApiError, got %v", err)
			}
		})
	}
}

func TestNotFoundFromProxyIsNot404(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		body        string
	}{
		{"traefik without routes", "text/plain; charset=utf-8", "404 page not found"},
		{"html error page", "text/html", "<html><body><h1>404 Not Found</h1></body></html>"},
		{"no content type", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := notFoundServer(t, tt.contentType, tt.body)

			err := client.get(context.Background(), "/realms/present", &Realm{}, nil)
			if err == nil {
				t.Fatal("a 404 that is not from Keycloak must fail the request")
			}
			if ErrorIs404(err) {
				t.Fatalf("a 404 that is not from Keycloak must not count as not found, got %v", err)
			}
			if !strings.Contains(err.Error(), "not treated as deleted") || !strings.Contains(err.Error(), tt.body) {
				t.Fatalf("want the reason and the response body in the error, got %v", err)
			}
		})
	}
}
