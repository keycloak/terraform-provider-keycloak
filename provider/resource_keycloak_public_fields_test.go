package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestDisabledBruteForceOwnership(t *testing.T) {
	for _, tc := range []struct {
		name                    string
		headers, block, enabled bool
	}{
		{name: "legacy no block"},
		{name: "legacy headers", headers: true},
		{name: "legacy enabled block", block: true, enabled: true},
		{name: "disabled owned block", block: true},
		{name: "disabled block and headers", headers: true, block: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			config := map[string]interface{}{"realm": "fixture"}
			settings := map[string]interface{}{}
			if tc.headers {
				settings["headers"] = []interface{}{map[string]interface{}{}}
			}
			if tc.block {
				settings["brute_force_detection"] = []interface{}{map[string]interface{}{"enabled": tc.enabled, "max_login_failures": 43}}
			}
			if tc.headers || tc.block {
				config["security_defenses"] = []interface{}{settings}
			}
			data := schema.TestResourceDataRaw(t, resourceKeycloakRealm().Schema, config)
			v := version.Must(version.NewVersion("26.8.1"))
			realm, err := getRealmFromData(data, v)
			if err != nil {
				t.Fatal(err)
			}
			if realm.BruteForceProtected != tc.enabled {
				t.Fatalf("protection = %t", realm.BruteForceProtected)
			}
			if tc.block && (realm.MaxTemporaryLockouts == nil || *realm.MaxTemporaryLockouts != 0) {
				t.Fatal("configured zero max temporary lockouts was omitted")
			}
			if !tc.block && realm.MaxTemporaryLockouts != nil {
				t.Fatal("legacy omitted-block payload changed")
			}
			if tc.block && realm.FailureFactor != 43 {
				t.Fatal("configured disabled threshold was discarded")
			}
			// Refresh must retain opted-in disabled settings and observe threshold drift,
			// without introducing a disabled block for legacy users.
			realm.FailureFactor = 57
			setRealmData(data, realm, v, true)
			current := data.Get("security_defenses").([]interface{})
			gotBlock := false
			if len(current) > 0 {
				blocks := current[0].(map[string]interface{})["brute_force_detection"].([]interface{})
				gotBlock = len(blocks) == 1
				if gotBlock {
					got := blocks[0].(map[string]interface{})
					if got["enabled"].(bool) != tc.enabled || got["max_login_failures"].(int) != 57 {
						t.Fatalf("incorrect refreshed disabled settings: %v", got)
					}
				}
			}
			if gotBlock != tc.block {
				t.Fatalf("block presence = %t, want %t", gotBlock, tc.block)
			}
		})
	}
}

// Actual Terraform/SDK configuration is needed to distinguish explicitly empty
// sets and false values from observations of omitted optional/computed fields.
func TestRealmPublicConfigurationLifecycle(t *testing.T) {
	data := schema.TestResourceDataRaw(t, resourceKeycloakRealm().Schema, map[string]interface{}{"realm": "public-fields"})
	current, err := getRealmFromData(data, version.Must(version.NewVersion("26.8.1")))
	if err != nil {
		t.Fatal(err)
	}
	credentials, disabled := []string{"password"}, false
	current.Id = current.Realm
	current.RequiredCredentials = &credentials
	current.ScimApiEnabled = &disabled
	current.VerifiableCredentialsEnabled = &disabled
	var mu sync.Mutex
	writes := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/admin/serverinfo":
			fmt.Fprint(w, `{"systemInfo":{"version":"26.8.1"}}`)
		case r.URL.Path == "/admin/realms/public-fields" && r.Method == http.MethodGet:
			response := *current
			if response.RequiredCredentials != nil && len(*response.RequiredCredentials) == 0 {
				response.RequiredCredentials = nil
			}
			json.NewEncoder(w).Encode(&response)
		case r.URL.Path == "/admin/realms/public-fields" && r.Method == http.MethodPut:
			if err := json.NewDecoder(r.Body).Decode(current); err != nil {
				t.Error(err)
				w.WriteHeader(400)
				return
			}
			writes++
			w.WriteHeader(204)
		case r.URL.Path == "/admin/realms/public-fields" && r.Method == http.MethodDelete:
			w.WriteHeader(204)
		default:
			t.Errorf("unexpected HTTP request %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	client, err := keycloak.NewKeycloakClient(context.Background(), server.URL, "", "", "", "", "master", "", "", "fixture-token", "", "", "", "", false, 10, "", false, "", "", "test", false, nil, "26.8.1")
	if err != nil {
		t.Fatal(err)
	}
	provider := KeycloakProvider(client)
	config := func(required string) string {
		return fmt.Sprintf(`provider "keycloak" { url = %q }
resource "keycloak_realm" "test" {
 realm = "public-fields"
 required_credentials = %s
 scim_api_enabled = false
 verifiable_credentials_enabled = false
}`, server.URL, required)
	}
	check := func(expected int, count string) resource.TestCheckFunc {
		return func(state *terraform.State) error {
			mu.Lock()
			defer mu.Unlock()
			if writes != expected {
				return fmt.Errorf("writes %d, want %d", writes, expected)
			}
			if current.RequiredCredentials == nil || current.ScimApiEnabled == nil || *current.ScimApiEnabled || current.VerifiableCredentialsEnabled == nil || *current.VerifiableCredentialsEnabled {
				return fmt.Errorf("false values or required credentials were omitted")
			}
			return resource.TestCheckResourceAttr("keycloak_realm.test", "required_credentials.#", count)(state)
		}
	}
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: map[string]func() (tfprotov5.ProviderServer, error){"keycloak": func() (tfprotov5.ProviderServer, error) { return provider.GRPCProvider(), nil }},
		Steps: []resource.TestStep{
			{Config: config(`["password"]`) + "\nimport {\n to = keycloak_realm.test\n id = \"public-fields\"\n}", Check: check(0, "1")},
			{Config: config(`["password"]`), PreConfig: func() {
				mu.Lock()
				empty, enabled := []string{}, true
				current.RequiredCredentials = &empty
				current.ScimApiEnabled = &enabled
				current.VerifiableCredentialsEnabled = &enabled
				mu.Unlock()
			}, PlanOnly: true, ExpectNonEmptyPlan: true},
			{Config: config(`["password"]`), Check: check(1, "1")},
			{Config: config(`[]`), Check: check(2, "0")},
			{Config: config(`[]`), PlanOnly: true, ExpectNonEmptyPlan: false},
		},
	})
}

func TestRealmDataSourceDoesNotRetainDisabledBruteForceBlock(t *testing.T) {
	// A prior enabled observation must still disappear on a disabled data-source
	// refresh; resource ownership opt-in must not change the data source contract.
	data := schema.TestResourceDataRaw(t, dataSourceKeycloakRealm().Schema, map[string]interface{}{"realm": "fixture", "security_defenses": []interface{}{map[string]interface{}{"brute_force_detection": []interface{}{map[string]interface{}{"enabled": true}}}}})
	realm := &keycloak.Realm{Realm: "fixture", BruteForceProtected: false}
	setRealmData(data, realm, version.Must(version.NewVersion("26.8.1")), false)
	if value := data.Get("security_defenses").([]interface{}); len(value) != 0 {
		t.Fatal("disabled data-source block was retained")
	}
}
