package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestRealmClientRegistrationPolicyConfigLifecycle(t *testing.T) {
	var mu sync.Mutex
	config := map[string][]string{"first": {"one"}, "second": {"two"}}
	writes := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/admin/serverinfo":
			fmt.Fprint(w, `{"componentTypes":{}}`)
		case r.URL.Path == "/admin/realms/fixture/components/policy" && r.Method == http.MethodGet:
			json.NewEncoder(w).Encode(map[string]interface{}{"id": "policy", "name": "Policy", "providerId": "fixture", "subType": "anonymous", "config": config})
		case r.URL.Path == "/admin/realms/fixture/components/policy" && r.Method == http.MethodPut:
			var body struct {
				Config map[string][]string `json:"config"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			// Model Keycloak's merge semantics: only empty arrays delete keys.
			for k, v := range body.Config {
				if len(v) == 0 {
					delete(config, k)
				} else {
					config[k] = v
				}
			}
			writes++
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	client, err := keycloak.NewKeycloakClient(context.Background(), server.URL, "", "", "", "", "master", "", "", "fixture-token", "", "", "", "", false, 10, "", false, "", "", "test", false, nil, "26.8.2")
	if err != nil {
		t.Fatal(err)
	}
	p := KeycloakProvider(client)
	hcl := func(settings string) string {
		return fmt.Sprintf(`provider "keycloak" { url = %q }
resource "keycloak_realm_client_registration_policy" "test" {
 realm_id = "fixture"
 name = "Policy"
 provider_id = "fixture"
 sub_type = "anonymous"
 %s
}`, server.URL, settings)
	}
	check := func(expectedWrites, expectedKeys int) resource.TestCheckFunc {
		return func(_ *terraform.State) error {
			mu.Lock()
			defer mu.Unlock()
			if writes != expectedWrites || len(config) != expectedKeys {
				return fmt.Errorf("writes/config = %d/%v, want %d writes and %d keys", writes, config, expectedWrites, expectedKeys)
			}
			return nil
		}
	}
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: map[string]func() (tfprotov5.ProviderServer, error){"keycloak": func() (tfprotov5.ProviderServer, error) { return p.GRPCProvider(), nil }},
		Steps: []resource.TestStep{
			{Config: hcl("") + "\nimport {\n to = keycloak_realm_client_registration_policy.test\n id = \"fixture/policy\"\n}", Check: check(0, 2)},
			{Config: hcl(""), Check: check(0, 2)},
			{Config: hcl(`config = { first = "one" }`), Check: check(1, 1)},
			{Config: hcl(`config = {}`), Check: check(2, 0)},
			{Config: hcl(`config = {}`), PlanOnly: true, ExpectNonEmptyPlan: false},
		},
	})
}
