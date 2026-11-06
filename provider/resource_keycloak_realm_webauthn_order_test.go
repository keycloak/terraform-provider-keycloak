package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"strconv"
	"sync"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

// Use the real SDK/Terraform lifecycle against a local HTTP fixture: in-memory
// ResourceData alone does not carry raw configuration or computed prior state.
func TestRealmWebAuthnSignatureAlgorithmsOrdered(t *testing.T) {
	for _, block := range []string{"web_authn_policy", "web_authn_passwordless_policy"} {
		t.Run(block, func(t *testing.T) {
			data := schema.TestResourceDataRaw(t, resourceKeycloakRealm().Schema, map[string]interface{}{"realm": "ordered-test", block: []interface{}{map[string]interface{}{"signature_algorithms": []interface{}{"ES256", "RS256"}}}})
			current, err := getRealmFromData(data, version.Must(version.NewVersion("26.8.1")))
			if err != nil {
				t.Fatal(err)
			}
			current.Id = current.Realm
			var mu sync.Mutex
			writes := 0
			algorithms := func() *[]string {
				if block == "web_authn_policy" {
					return &current.WebAuthnPolicySignatureAlgorithms
				}
				return &current.WebAuthnPolicyPasswordlessSignatureAlgorithms
			}
			*algorithms() = []string{"RS256", "ES256"}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mu.Lock()
				defer mu.Unlock()
				w.Header().Set("Content-Type", "application/json")
				switch {
				case r.URL.Path == "/admin/serverinfo":
					fmt.Fprint(w, `{"systemInfo":{"version":"26.8.1"}}`)
				case r.URL.Path == "/admin/realms/ordered-test" && r.Method == http.MethodGet:
					json.NewEncoder(w).Encode(current)
				case r.URL.Path == "/admin/realms/ordered-test" && r.Method == http.MethodPut:
					if err := json.NewDecoder(r.Body).Decode(&current); err != nil {
						t.Error(err)
						w.WriteHeader(400)
						return
					}
					writes++
					w.WriteHeader(204)
				case r.URL.Path == "/admin/realms/ordered-test" && r.Method == http.MethodDelete:
					w.WriteHeader(204)
				default:
					t.Errorf("unexpected HTTP request: %s %s", r.Method, r.URL.Path)
					http.NotFound(w, r)
				}
			}))
			defer server.Close()
			client, err := keycloak.NewKeycloakClient(context.Background(), server.URL, "", "", "", "", "master", "", "", "test-token", "", "", "", "", false, 10, "", false, "", "", "test", false, nil, "26.8.1")
			if err != nil {
				t.Fatal(err)
			}
			provider := KeycloakProvider(client)
			legacyProvider := KeycloakProvider(client)
			for _, policy := range []string{"web_authn_policy", "web_authn_passwordless_policy"} {
				policySchema := legacyProvider.ResourcesMap["keycloak_realm"].Schema[policy].Elem.(*schema.Resource).Schema
				delete(policySchema, "signature_algorithms_ordered")
				policySchema["signature_algorithms"].ConflictsWith = nil
			}
			// Emulate the pre-change Read exactly, excluding the new computed key
			// when flattening into the prior schema.
			legacyResource := legacyProvider.ResourcesMap["keycloak_realm"]
			legacyResource.ReadContext = func(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
				updated := resourceKeycloakRealm().Data(data.State())
				if diags := resourceKeycloakRealmRead(ctx, updated, meta); diags.HasError() {
					return diags
				}
				for key := range legacyResource.Schema {
					value := updated.Get(key)
					if key == "web_authn_policy" || key == "web_authn_passwordless_policy" {
						for _, policy := range value.([]interface{}) {
							delete(policy.(map[string]interface{}), "signature_algorithms_ordered")
						}
					}
					if err := data.Set(key, value); err != nil {
						return diag.FromErr(err)
					}
				}
				return nil
			}
			factories := func(p *schema.Provider) map[string]func() (tfprotov5.ProviderServer, error) {
				return map[string]func() (tfprotov5.ProviderServer, error){"keycloak": func() (tfprotov5.ProviderServer, error) { return p.GRPCProvider(), nil }}
			}
			config := func(attribute, values string) string {
				return fmt.Sprintf(`provider "keycloak" {
 client_id = "fixture"
 url = %q
}
resource "keycloak_realm" "test" {
 realm = "ordered-test"
 %s { %s = %s }
}
data "keycloak_realm" "test" {
 realm = keycloak_realm.test.realm
 depends_on = [keycloak_realm.test]
}`, server.URL, block, attribute, values)
			}
			ordered := config("signature_algorithms_ordered", `["RS256", "ES256"]`)
			legacy := config("signature_algorithms", `["ES256", "RS256"]`)
			check := func(want []string, expectedWrites int, legacyState ...bool) resource.TestCheckFunc {
				return func(state *terraform.State) error {
					attributes := state.RootModule().Resources["keycloak_realm.test"].Primary.Attributes
					countKey := block + ".0.signature_algorithms_ordered.#"
					if len(legacyState) == 1 && legacyState[0] {
						if _, present := attributes[countKey]; present {
							return fmt.Errorf("prior-schema state unexpectedly contains %s", countKey)
						}
					} else {
						if got := attributes[countKey]; got != strconv.Itoa(len(want)) {
							return fmt.Errorf("state %s = %q, want %d", countKey, got, len(want))
						}
						for index, algorithm := range want {
							attribute := fmt.Sprintf("%s.0.signature_algorithms_ordered.%d", block, index)
							if got, present := attributes[attribute]; !present || got != algorithm {
								return fmt.Errorf("state %s = %q (present %t), want %q", attribute, got, present, algorithm)
							}
						}
					}
					if len(legacyState) == 0 {
						dataAttributes := state.RootModule().Resources["data.keycloak_realm.test"].Primary.Attributes
						if got := dataAttributes[countKey]; got != strconv.Itoa(len(want)) {
							return fmt.Errorf("data source %s = %q, want %d", countKey, got, len(want))
						}
						for index, algorithm := range want {
							attribute := fmt.Sprintf("%s.0.signature_algorithms_ordered.%d", block, index)
							if got, present := dataAttributes[attribute]; !present || got != algorithm {
								return fmt.Errorf("data source %s = %q (present %t), want %q", attribute, got, present, algorithm)
							}
						}
					}
					mu.Lock()
					defer mu.Unlock()
					if !reflect.DeepEqual(*algorithms(), want) {
						return fmt.Errorf("algorithms %v, want %v", *algorithms(), want)
					}
					if writes != expectedWrites {
						return fmt.Errorf("writes %d, want %d", writes, expectedWrites)
					}
					return nil
				}
			}
			testCase := resource.TestCase{
				Steps: []resource.TestStep{
					// Produce actual prior-schema state with only the legacy set, then
					// upgrade and adopt server preference without a realm update.
					{Config: legacy + "\nimport {\n to = keycloak_realm.test\n id = \"ordered-test\"\n}", ProtoV5ProviderFactories: factories(legacyProvider), Check: check([]string{"RS256", "ES256"}, 0, true)},
					{Config: legacy, Check: check([]string{"RS256", "ES256"}, 0)},
					{Config: ordered, Check: check([]string{"RS256", "ES256"}, 0)},
					{Config: ordered, PlanOnly: true, ExpectNonEmptyPlan: false},
					{Config: ordered, PreConfig: func() { mu.Lock(); *algorithms() = []string{"ES256", "RS256"}; mu.Unlock() }, PlanOnly: true, ExpectNonEmptyPlan: true},
					{Config: ordered, Check: check([]string{"RS256", "ES256"}, 1)},
					{Config: config("signature_algorithms_ordered", `["ES256", "RS256"]`), Check: check([]string{"ES256", "RS256"}, 2)},
					// Switching to the legacy set retains its historical ordering semantics.
					{Config: legacy, Check: check([]string{"ES256", "RS256"}, 2)},
					{Config: legacy, PreConfig: func() { mu.Lock(); *algorithms() = []string{"RS256", "ES256"}; mu.Unlock() }, Check: check([]string{"RS256", "ES256"}, 2)},
					{Config: ordered, Check: check([]string{"RS256", "ES256"}, 2)},
					{Config: config("signature_algorithms", `["ES256"]`) + "\n", PlanOnly: true, ExpectNonEmptyPlan: true},
					{Config: fmt.Sprintf(`provider "keycloak" { url = %q }
resource "keycloak_realm" "test" {
 realm = "ordered-test"
%s {
 signature_algorithms = ["ES256"]
signature_algorithms_ordered = ["ES256"]
 }
}`, server.URL, block), ExpectError: regexp.MustCompile("conflicts with")},
					{Config: ordered, Check: check([]string{"RS256", "ES256"}, 2)},
				},
			}
			for index := range testCase.Steps {
				if testCase.Steps[index].ProtoV5ProviderFactories == nil {
					testCase.Steps[index].ProtoV5ProviderFactories = factories(provider)
				}
			}
			resource.UnitTest(t, testCase)
		})
	}
}
