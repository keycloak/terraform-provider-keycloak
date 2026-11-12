package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func roleReadTestClient(t *testing.T, handler http.Handler) *keycloak.KeycloakClient {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/admin/serverinfo" && r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"systemInfo":{"version":"26.8.1"}}`)
			return
		}
		handler.ServeHTTP(w, r)
	}))
	t.Cleanup(server.Close)
	client, err := keycloak.NewKeycloakClient(context.Background(), server.URL, "", "", "", "", "master", "", "", "test-token", "", "", "", "", false, 10, "", false, "", "", "test", false, nil, "26.8.1")
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func TestRoleReadRefreshesComposites(t *testing.T) {
	for _, test := range []struct {
		name                        string
		composite                   bool
		children                    []string
		roleStatus, compositeStatus int
		wantError                   bool
		want                        []string
		wantCompositeRead           bool
	}{
		{name: "last child removed", want: []string{}},
		{name: "composite children changed", composite: true, children: []string{"new-child"}, want: []string{"new-child"}, wantCompositeRead: true},
		{name: "composite with empty children", composite: true, want: []string{}, wantCompositeRead: true},
		{name: "composite permission error", composite: true, compositeStatus: 403, wantError: true, want: []string{"old-child"}, wantCompositeRead: true},
		{name: "role permission error", roleStatus: 403, wantError: true, want: []string{"old-child"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			var compositeReads atomic.Int32
			client := roleReadTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("unexpected method %s", r.Method)
				}
				w.Header().Set("Content-Type", "application/json")
				switch r.URL.Path {
				case "/admin/realms/realm/roles-by-id/parent":
					if test.roleStatus != 0 {
						w.WriteHeader(test.roleStatus)
						fmt.Fprint(w, `{"error":"forbidden"}`)
						return
					}
					json.NewEncoder(w).Encode(keycloak.Role{Id: "parent", Name: "parent", Composite: test.composite})
				case "/admin/realms/realm/roles-by-id/parent/composites":
					compositeReads.Add(1)
					if test.compositeStatus != 0 {
						w.WriteHeader(test.compositeStatus)
						fmt.Fprint(w, `{"error":"forbidden"}`)
						return
					}
					roles := []*keycloak.Role{}
					for _, id := range test.children {
						roles = append(roles, &keycloak.Role{Id: id, Name: id})
					}
					json.NewEncoder(w).Encode(roles)
				default:
					t.Errorf("unexpected path %s", r.URL.Path)
					http.NotFound(w, r)
				}
			}))
			data := schema.TestResourceDataRaw(t, resourceKeycloakRole().Schema, map[string]interface{}{"realm_id": "realm", "name": "parent", "composite_roles": []interface{}{"old-child"}})
			data.SetId("parent")
			diags := resourceKeycloakRoleRead(context.Background(), data, client)
			if diags.HasError() != test.wantError {
				t.Fatalf("diagnostics %v, want error %t", diags, test.wantError)
			}
			actual := []string{}
			for _, id := range data.Get("composite_roles").(*schema.Set).List() {
				actual = append(actual, id.(string))
			}
			if !reflect.DeepEqual(actual, test.want) {
				t.Errorf("composites %v, want %v", actual, test.want)
			}
			if (compositeReads.Load() > 0) != test.wantCompositeRead {
				t.Errorf("composite reads %d, want read %t", compositeReads.Load(), test.wantCompositeRead)
			}
		})
	}
}

// Exercise refresh/plan/apply so a cleared child is actually repaired, while a
// null computed input remains unowned by this resource (for example default roles).
func TestRoleCompositeRemovalDrift(t *testing.T) {
	var mu sync.Mutex
	hasChild := true
	writes := 0
	client := roleReadTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		child := &keycloak.Role{Id: "child", Name: "child"}
		switch {
		case r.URL.Path == "/admin/realms/realm/roles-by-id/parent" && r.Method == http.MethodGet:
			json.NewEncoder(w).Encode(&keycloak.Role{Id: "parent", Name: "parent", Composite: hasChild, Attributes: map[string][]string{}})
		case r.URL.Path == "/admin/realms/realm/roles-by-id/child" && r.Method == http.MethodGet:
			json.NewEncoder(w).Encode(child)
		case r.URL.Path == "/admin/realms/realm/roles-by-id/parent/composites" && r.Method == http.MethodGet:
			children := []*keycloak.Role{}
			if hasChild {
				children = append(children, child)
			}
			json.NewEncoder(w).Encode(children)
		case r.URL.Path == "/admin/realms/realm/roles-by-id/parent/composites" && r.Method == http.MethodPost:
			var children []*keycloak.Role
			if err := json.NewDecoder(r.Body).Decode(&children); err != nil {
				t.Error(err)
				w.WriteHeader(400)
				return
			}
			if len(children) != 1 || children[0].Id != "child" {
				t.Errorf("unexpected composite addition: %v", children)
			}
			hasChild = true
			writes++
			w.WriteHeader(204)
		case r.URL.Path == "/admin/realms/realm/roles-by-id/parent" && r.Method == http.MethodPut:
			writes++
			w.WriteHeader(204)
		case r.URL.Path == "/admin/realms/realm/roles-by-id/parent" && r.Method == http.MethodDelete:
			w.WriteHeader(204)
		default:
			t.Errorf("unexpected HTTP request: %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	provider := KeycloakProvider(client)
	config := func(composites string) string {
		return fmt.Sprintf(`provider "keycloak" {
 url = "http://fixture.invalid"
}
resource "keycloak_role" "parent" {
 realm_id = "realm"
 name = "parent"
 composite_roles = %s
}`, composites)
	}
	owned := config(`["child"]`)
	unowned := config(`null`)
	removeChild := func() { mu.Lock(); hasChild = false; mu.Unlock() }
	check := func(wantChild bool, wantWrites int) resource.TestCheckFunc {
		return func(s *terraform.State) error {
			mu.Lock()
			defer mu.Unlock()
			wantCount := "0"
			if wantChild {
				wantCount = "1"
			}
			if got := s.RootModule().Resources["keycloak_role.parent"].Primary.Attributes["composite_roles.#"]; got != wantCount {
				return fmt.Errorf("composite count %q, want %q", got, wantCount)
			}
			if hasChild != wantChild || writes != wantWrites {
				return fmt.Errorf("child %t, writes %d; want %t, %d", hasChild, writes, wantChild, wantWrites)
			}
			return nil
		}
	}
	resource.UnitTest(t, resource.TestCase{ProtoV5ProviderFactories: map[string]func() (tfprotov5.ProviderServer, error){"keycloak": func() (tfprotov5.ProviderServer, error) { return provider.GRPCProvider(), nil }}, Steps: []resource.TestStep{
		{Config: owned + "\nimport {\n to = keycloak_role.parent\n id = \"realm/parent\"\n}", Check: check(true, 0)},
		{Config: owned, PreConfig: removeChild, PlanOnly: true, ExpectNonEmptyPlan: true},
		{Config: owned, Check: check(true, 2)},
		{Config: unowned, Check: check(true, 2)},
		{Config: unowned, PreConfig: removeChild, PlanOnly: true, ExpectNonEmptyPlan: false},
		{Config: unowned, Check: check(false, 2)},
	}})
}
