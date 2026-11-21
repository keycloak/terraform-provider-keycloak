package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestRealmClientPolicyConfigurationUpdates(t *testing.T) {
	for _, tc := range []struct {
		name, desiredProfiles, desiredPolicies string
		failRead, failPolicy, wantError        bool
		wantWrites                             []string
	}{
		{name: "unchanged", desiredProfiles: `[{"name":"old"}]`, desiredPolicies: `[{"name":"policy","profiles":["old"]}]`},
		{name: "switch profile reference", desiredProfiles: `[{"name":"new"}]`, desiredPolicies: `[{"name":"policy","profiles":["new"]}]`, wantWrites: []string{"profiles", "policies", "profiles"}},
		{name: "clear", desiredProfiles: `[]`, desiredPolicies: `[]`, wantWrites: []string{"policies", "profiles"}},
		{name: "failed read", desiredProfiles: `[]`, desiredPolicies: `[]`, failRead: true, wantError: true},
		{name: "global collision", desiredProfiles: `[{"name":"global"}]`, desiredPolicies: `[]`, wantError: true},
		{name: "resume partial update", desiredProfiles: `[{"name":"new"}]`, desiredPolicies: `[{"name":"policy","profiles":["new"]}]`, failPolicy: true, wantError: true, wantWrites: []string{"profiles", "policies"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			profiles := json.RawMessage(`[{"name":"old"}]`)
			policies := json.RawMessage(`[{"name":"policy","profiles":["old"]}]`)
			var writes []string
			failPolicy := tc.failPolicy
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.URL.Path == "/admin/serverinfo" {
					fmt.Fprint(w, `{}`)
					return
				}
				collection := strings.TrimPrefix(r.URL.Path, "/admin/realms/fixture/client-policies/")
				if collection != "profiles" && collection != "policies" {
					t.Errorf("unexpected path %s", r.URL.Path)
					w.WriteHeader(404)
					return
				}
				if r.Method == http.MethodGet {
					if tc.failRead {
						w.WriteHeader(403)
						return
					}
					if collection == "profiles" {
						fmt.Fprintf(w, `{"profiles":%s,"globalProfiles":[{"name":"global"}]}`, profiles)
					} else {
						fmt.Fprintf(w, `{"policies":%s,"globalPolicies":[]}`, policies)
					}
					return
				}
				if r.Method != http.MethodPut {
					t.Errorf("unexpected method %s", r.Method)
					w.WriteHeader(500)
					return
				}
				writes = append(writes, collection)
				if collection == "policies" && failPolicy {
					w.WriteHeader(400)
					return
				}
				var body map[string]json.RawMessage
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
				}
				if len(body) != 1 || body[collection] == nil {
					t.Error("unexpected global fields or collection")
				}
				if collection == "profiles" {
					profiles = body[collection]
				} else {
					policies = body[collection]
				}
				// Model the server's reference validation at every stage.
				var ps []struct {
					Name string `json:"name"`
				}
				var rules []struct {
					Profiles []string `json:"profiles"`
				}
				json.Unmarshal(profiles, &ps)
				json.Unmarshal(policies, &rules)
				available := map[string]bool{}
				for _, p := range ps {
					available[p.Name] = true
				}
				for _, rule := range rules {
					for _, name := range rule.Profiles {
						if !available[name] {
							t.Errorf("removed referenced profile %s", name)
						}
					}
				}
				w.WriteHeader(204)
			}))
			defer server.Close()
			client, err := keycloak.NewKeycloakClient(context.Background(), server.URL, "", "", "", "", "master", "", "", "fixture-token", "", "", "", "", false, 10, "", false, "", "", "test", false, nil, "26.8.2")
			if err != nil {
				t.Fatal(err)
			}
			r := resourceKeycloakRealmClientPolicyConfiguration()
			data := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{"realm_id": "fixture", "profiles": tc.desiredProfiles, "policies": tc.desiredPolicies})
			ds := r.CreateContext(context.Background(), data, client)
			if ds.HasError() != tc.wantError {
				t.Fatalf("diagnostics: %v", ds)
			}
			if !reflect.DeepEqual(writes, tc.wantWrites) {
				t.Fatalf("writes %v, want %v", writes, tc.wantWrites)
			}
			if !tc.wantError || tc.failPolicy {
				if data.Id() != "fixture" {
					t.Fatal("singleton ID was not retained")
				}
			}
			if tc.failPolicy {
				failPolicy = false
				writes = nil
				ds = r.UpdateContext(context.Background(), data, client)
				if ds.HasError() || !reflect.DeepEqual(writes, []string{"policies", "profiles"}) {
					t.Fatalf("retry: %v, writes %v", ds, writes)
				}
			}
			if !tc.wantError || tc.failPolicy {
				if canonicalClientPolicyJSON(string(profiles)) != canonicalClientPolicyJSON(tc.desiredProfiles) || canonicalClientPolicyJSON(string(policies)) != canonicalClientPolicyJSON(tc.desiredPolicies) {
					t.Fatal("desired collections not installed")
				}
				writes = nil
				// Import must only observe, and deletion must not clear the singleton.
				imported := schema.TestResourceDataRaw(t, r.Schema, nil)
				imported.SetId("fixture")
				if _, err := r.Importer.StateContext(context.Background(), imported, client); err != nil {
					t.Fatal(err)
				}
				if !r.DeleteContext(context.Background(), imported, client).HasError() || len(writes) != 0 {
					t.Fatal("delete/import changed server collections")
				}
			}
		})
	}
}
