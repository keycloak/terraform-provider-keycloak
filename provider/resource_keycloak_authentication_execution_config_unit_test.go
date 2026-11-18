package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestMergeMaskedAuthenticationExecutionConfig(t *testing.T) {
	t.Parallel()

	remote := map[string]string{
		"secret.key":      keycloakMaskedSecret,
		"site.key":        "site",
		"imported.secret": keycloakMaskedSecret,
		"empty.secret":    keycloakMaskedSecret,
	}
	known := map[string]interface{}{
		"secret.key":   "my-secret",
		"site.key":     "outdated-site",
		"empty.secret": "",
		"removed.key":  "gone",
	}

	got := mergeMaskedAuthenticationExecutionConfig(remote, known)

	want := map[string]string{
		"secret.key":      "my-secret",          // masked -> known value is kept
		"site.key":        "site",               // not masked -> remote value wins
		"imported.secret": keycloakMaskedSecret, // masked and unknown -> stays masked
		"empty.secret":    "",                   // masked and known empty -> known empty value is kept
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected merge result\ngot:  %#v\nwant: %#v", got, want)
	}
}

// TestAuthenticationExecutionConfigUpdate_sendsRotatedSecret proves against a mocked Keycloak admin API that rotating
// a secret sends the new literal value to Keycloak, and that the masked value returned by the subsequent GET does
// not leak into state. This cannot be verified in an acceptance test because Keycloak never returns the real value.
func TestAuthenticationExecutionConfigUpdate_sendsRotatedSecret(t *testing.T) {
	t.Parallel()

	const (
		realmId     = "test-realm"
		executionId = "execution-id"
		configId    = "config-id"
		oldSecret   = "old-secret"
		newSecret   = "new-secret"
	)

	var putBodies []keycloak.AuthenticationExecutionConfig
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != fmt.Sprintf("/admin/realms/%s/authentication/config/%s", realmId, configId) {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		switch r.Method {
		case http.MethodPut:
			var body keycloak.AuthenticationExecutionConfig
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			putBodies = append(putBodies, body)
			w.WriteHeader(http.StatusNoContent)
		case http.MethodGet:
			// Keycloak >= 26.6.7 / 26.7.3 masks the secret, whatever was stored
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(keycloak.AuthenticationExecutionConfig{
				Id:     configId,
				Alias:  "alias",
				Config: map[string]string{"site.key": "site-key", "secret.key": keycloakMaskedSecret},
			})
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	t.Cleanup(server.Close)

	// an access token skips the login, so no live Keycloak is involved
	client, err := keycloak.NewKeycloakClient(testCtx, server.URL, "", "", "", "", "", "", "", "test-token", "", "", "", "", true, 10, "", false, "", "", "", false, nil, "")
	if err != nil {
		t.Fatal(err)
	}

	// the planned configuration carries the rotated secret
	data := schema.TestResourceDataRaw(t, resourceKeycloakAuthenticationExecutionConfig().Schema, map[string]interface{}{
		"realm_id":     realmId,
		"execution_id": executionId,
		"alias":        "alias",
		"config":       map[string]interface{}{"site.key": "site-key", "secret.key": newSecret},
	})
	data.SetId(configId)

	if diags := resourceKeycloakAuthenticationExecutionConfigUpdate(testCtx, data, client); diags.HasError() {
		t.Fatalf("update failed: %v", diags)
	}

	if len(putBodies) != 1 {
		t.Fatalf("expected exactly one PUT request, got %d", len(putBodies))
	}
	if got := putBodies[0].Config["secret.key"]; got != newSecret {
		t.Fatalf("expected PUT to send the rotated secret %q, got %q", newSecret, got)
	}
	if got := putBodies[0].Config["secret.key"]; got == oldSecret || got == keycloakMaskedSecret {
		t.Fatalf("PUT must not send the old or masked secret, got %q", got)
	}

	stateConfig := data.Get("config").(map[string]interface{})
	if got := stateConfig["secret.key"]; got != newSecret {
		t.Fatalf("expected state to keep the rotated secret %q after the masked GET, got %q", newSecret, got)
	}
	if got := stateConfig["site.key"]; got != "site-key" {
		t.Fatalf("expected non-secret value from Keycloak in state, got %q", got)
	}
}
