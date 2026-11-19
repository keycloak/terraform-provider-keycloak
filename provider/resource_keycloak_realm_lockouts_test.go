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

// Model Keycloak's partial PUT: an omitted maxTemporaryLockouts keeps the old
// value, while an explicit zero resets it. Exercise expansion, serialization,
// refresh, and the final plan through the actual Terraform/SDK lifecycle.
func TestRealmMaxTemporaryLockoutsReset(t *testing.T) {
	for _, serverVersion := range []string{"26.0.8", "26.8.1"} {
		t.Run(serverVersion, func(t *testing.T) {
			data := schema.TestResourceDataRaw(t, resourceKeycloakRealm().Schema, map[string]interface{}{"realm": "lockouts-test", "security_defenses": []interface{}{map[string]interface{}{"brute_force_detection": []interface{}{map[string]interface{}{"permanent_lockout": true, "max_temporary_lockouts": 3}}}}})
			current, err := getRealmFromData(data, version.Must(version.NewVersion(serverVersion)))
			if err != nil {
				t.Fatal(err)
			}
			current.Id = current.Realm
			var mu sync.Mutex
			writes, initialWrites := 0, 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mu.Lock()
				defer mu.Unlock()
				w.Header().Set("Content-Type", "application/json")
				switch {
				case r.URL.Path == "/admin/serverinfo" && r.Method == http.MethodGet:
					json.NewEncoder(w).Encode(map[string]interface{}{"systemInfo": map[string]string{"version": serverVersion}})
				case r.URL.Path == "/admin/realms/lockouts-test" && r.Method == http.MethodGet:
					json.NewEncoder(w).Encode(current)
				case r.URL.Path == "/admin/realms/lockouts-test" && r.Method == http.MethodPut:
					if err := json.NewDecoder(r.Body).Decode(current); err != nil {
						t.Error(err)
						w.WriteHeader(400)
						return
					}
					writes++
					w.WriteHeader(204)
				case r.URL.Path == "/admin/realms/lockouts-test" && r.Method == http.MethodDelete:
					w.WriteHeader(204)
				default:
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
					http.NotFound(w, r)
				}
			}))
			defer server.Close()
			client, err := keycloak.NewKeycloakClient(context.Background(), server.URL, "", "", "", "", "master", "", "", "test-token", "", "", "", "", false, 10, "", false, "", "", "test", false, nil, serverVersion)
			if err != nil {
				t.Fatal(err)
			}
			provider := KeycloakProvider(client)
			config := func(maximum int) string {
				return fmt.Sprintf(`provider "keycloak" { url = "http://fixture.invalid" }
resource "keycloak_realm" "test" {
 realm = "lockouts-test"
 security_defenses {
  brute_force_detection {
   permanent_lockout = true
   max_temporary_lockouts = %d
  }
 }
}`, maximum)
			}
			check := func(maximum int) resource.TestCheckFunc {
				return func(s *terraform.State) error {
					if err := resource.TestCheckResourceAttr("keycloak_realm.test", "security_defenses.0.brute_force_detection.0.max_temporary_lockouts", fmt.Sprint(maximum))(s); err != nil {
						return err
					}
					mu.Lock()
					defer mu.Unlock()
					raw, err := json.Marshal(current)
					if err != nil {
						return err
					}
					var realm map[string]interface{}
					if err = json.Unmarshal(raw, &realm); err != nil {
						return err
					}
					if value, present := realm["maxTemporaryLockouts"]; !present || value != float64(maximum) {
						return fmt.Errorf("server maxTemporaryLockouts = %v (present %t), want %d", value, present, maximum)
					}
					if maximum == 3 {
						initialWrites = writes
					} else if writes != initialWrites+1 {
						return fmt.Errorf("reset PUT count %d, want 1", writes-initialWrites)
					}
					return nil
				}
			}
			resource.UnitTest(t, resource.TestCase{ProtoV5ProviderFactories: map[string]func() (tfprotov5.ProviderServer, error){"keycloak": func() (tfprotov5.ProviderServer, error) { return provider.GRPCProvider(), nil }}, Steps: []resource.TestStep{
				{Config: config(3) + "\nimport {\n to = keycloak_realm.test\n id = \"lockouts-test\"\n}", Check: check(3)},
				{Config: config(0), Check: check(0)},
				{Config: config(0), PlanOnly: true, ExpectNonEmptyPlan: false},
			}})
		})
	}
}

func TestRealmMaxTemporaryLockoutsOmittedBlock(t *testing.T) {
	for _, headers := range []bool{false, true} {
		config := map[string]interface{}{"realm": "lockouts-test"}
		if headers {
			config["security_defenses"] = []interface{}{map[string]interface{}{"headers": []interface{}{map[string]interface{}{}}}}
		}
		data := schema.TestResourceDataRaw(t, resourceKeycloakRealm().Schema, config)
		realm, err := getRealmFromData(data, version.Must(version.NewVersion("26.8.1")))
		if err != nil {
			t.Fatal(err)
		}
		raw, err := json.Marshal(realm)
		if err != nil {
			t.Fatal(err)
		}
		var payload map[string]interface{}
		if err = json.Unmarshal(raw, &payload); err != nil {
			t.Fatal(err)
		}
		if _, present := payload["maxTemporaryLockouts"]; present {
			t.Fatalf("headers=%t: omitted brute-force block unexpectedly owns maxTemporaryLockouts", headers)
		}
		// The data source shares this flatten helper: absent API values remain a
		// numeric zero in Terraform rather than a nil/pointer value.
		if got := getBruteForceDetectionSettings(realm, version.Must(version.NewVersion("26.8.1")))["max_temporary_lockouts"]; got != 0 {
			t.Fatalf("absent value flattened as %v, want 0", got)
		}
	}
}
