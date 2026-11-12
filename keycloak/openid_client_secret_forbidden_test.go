package keycloak

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/go-version"
)

// Keycloak 26.8 answers 403 on /client-secret to a caller with view-clients
// only. The client is still read, with its secret marked unreadable; any other
// failure of the same call is still an error.
func openidClientSecretServer(t *testing.T, secretStatus int) *KeycloakClient {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/client-secret"):
			if secretStatus != http.StatusOK {
				w.WriteHeader(secretStatus)
				w.Write([]byte(`{"error":"refused"}`))
				return
			}
			w.Write([]byte(`{"type":"secret","value":"s3cret"}`))
		case strings.HasSuffix(r.URL.Path, "/clients/abc"):
			w.Write([]byte(`{"id":"abc","clientId":"app"}`))
		case strings.HasSuffix(r.URL.Path, "/clients"):
			w.Write([]byte(`[{"id":"abc","clientId":"app"}]`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)
	// accessTokenProvided: a 403 makes the client try a token refresh first,
	// which a provided token skips, so the 403 is what the read sees.
	return &KeycloakClient{
		baseUrl: server.URL, httpClient: server.Client(), initialLogin: true, accessTokenProvided: true,
		version:           version.Must(version.NewVersion("26.8.1")),
		clientCredentials: &ClientCredentials{AccessToken: "test", TokenType: "Bearer"},
	}
}

func TestOpenidClientSecretForbiddenIsUnreadable(t *testing.T) {
	client := openidClientSecretServer(t, http.StatusForbidden)
	ctx := context.Background()

	byId, err := client.GetOpenidClient(ctx, "realm", "abc")
	if err != nil {
		t.Fatalf("GetOpenidClient: %v", err)
	}
	byClientId, err := client.GetOpenidClientByClientId(ctx, "realm", "app")
	if err != nil {
		t.Fatalf("GetOpenidClientByClientId: %v", err)
	}
	all, err := client.GetOpenidClients(ctx, "realm", true)
	if err != nil {
		t.Fatalf("GetOpenidClients: %v", err)
	}
	for _, c := range []*OpenidClient{byId, byClientId, all[0]} {
		if !c.ClientSecretUnreadable || c.ClientSecret != "" || c.ClientId != "app" {
			t.Fatalf("want the client with an unreadable secret, got %+v", c)
		}
	}
}

func TestOpenidClientSecretReadable(t *testing.T) {
	c, err := openidClientSecretServer(t, http.StatusOK).GetOpenidClient(context.Background(), "realm", "abc")
	if err != nil {
		t.Fatal(err)
	}
	if c.ClientSecretUnreadable || c.ClientSecret != "s3cret" {
		t.Fatalf("want the secret, got %+v", c)
	}
}

func TestOpenidClientSecretOtherErrorsStillFail(t *testing.T) {
	if _, err := openidClientSecretServer(t, http.StatusInternalServerError).GetOpenidClient(context.Background(), "realm", "abc"); err == nil {
		t.Fatal("a 500 on the secret must still fail the read")
	}
}
