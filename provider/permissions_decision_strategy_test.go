package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

// A scope block without decision_strategy must send Keycloak's default
// strategy, not an empty string: Keycloak 26 rejects an empty
// decisionStrategy with "Cannot parse the JSON". Reading the default back
// must not plan a change either.
func TestKeycloakOpenidClientPermissionsWithoutDecisionStrategy(t *testing.T) {
	const prefix = "/admin/realms/test/clients/rm/authz/resource-server"

	var mu sync.Mutex
	permissions := map[string]*keycloak.OpenidClientAuthorizationPermission{}
	permission := func(id string) *keycloak.OpenidClientAuthorizationPermission {
		if p, ok := permissions[id]; ok {
			return p
		}
		p := &keycloak.OpenidClientAuthorizationPermission{Id: id, Name: id, DecisionStrategy: "UNANIMOUS", Type: "scope"}
		permissions[id] = p
		return p
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		path := r.URL.Path
		switch {
		case path == "/admin/serverinfo":
			fmt.Fprint(w, `{"systemInfo":{"version":"26.8.0"},"features":[{"name":"ADMIN_FINE_GRAINED_AUTHZ","enabled":true}]}`)
		case path == "/admin/realms/test/clients/app/management/permissions":
			if r.Method == http.MethodPut {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			fmt.Fprint(w, `{"enabled":true,"resource":"res","scopePermissions":{"token-exchange":"te"}}`)
		case path == "/admin/realms/test/clients" && r.URL.Query().Get("clientId") == "realm-management":
			fmt.Fprint(w, `[{"id":"rm","clientId":"realm-management"}]`)
		case r.Method == http.MethodPut && strings.HasPrefix(path, prefix+"/permission/scope/"):
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if body["decisionStrategy"] == "" {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, `{"error":"unknown_error","error_description":"Cannot parse the JSON"}`)
				return
			}
			p := permission(strings.TrimPrefix(path, prefix+"/permission/scope/"))
			p.DecisionStrategy, _ = body["decisionStrategy"].(string)
			p.Description, _ = body["description"].(string)
			p.Policies = nil
			for _, policy := range body["policies"].([]interface{}) {
				p.Policies = append(p.Policies, policy.(string))
			}
			w.WriteHeader(http.StatusCreated)
		case r.Method == http.MethodGet && strings.HasPrefix(path, prefix+"/policy/") && strings.HasSuffix(path, "/associatedPolicies"):
			p := permission(strings.TrimSuffix(strings.TrimPrefix(path, prefix+"/policy/"), "/associatedPolicies"))
			policies := []map[string]string{}
			for _, id := range p.Policies {
				policies = append(policies, map[string]string{"id": id})
			}
			json.NewEncoder(w).Encode(policies)
		case r.Method == http.MethodGet && (strings.HasSuffix(path, "/resources") || strings.HasSuffix(path, "/scopes")):
			fmt.Fprint(w, `[]`)
		case r.Method == http.MethodGet && strings.HasPrefix(path, prefix+"/permission/"):
			p := *permission(strings.TrimPrefix(path, prefix+"/permission/"))
			p.Policies = nil
			json.NewEncoder(w).Encode(p)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := keycloak.NewKeycloakClient(context.Background(), server.URL, "", "", "", "", "master", "", "", "test-token", "", "", "", "", false, 10, "", false, "", "", "", false, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	provider := &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{"keycloak_openid_client_permissions": resourceKeycloakOpenidClientPermissions()},
		ConfigureContextFunc: func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
			return client, nil
		},
	}
	factories := map[string]func() (tfprotov5.ProviderServer, error){
		"keycloak": func() (tfprotov5.ProviderServer, error) { return provider.GRPCProvider(), nil },
	}

	config := func(scope string) string {
		return fmt.Sprintf(`resource "keycloak_openid_client_permissions" "perms" {
	realm_id  = "test"
	client_id = "app"
	token_exchange_scope {
		%s
	}
}`, scope)
	}
	strategy := func(want string) resource.TestCheckFunc {
		return func(*terraform.State) error {
			mu.Lock()
			defer mu.Unlock()
			if got := permission("te").DecisionStrategy; got != want {
				return fmt.Errorf("decisionStrategy in Keycloak is %q, want %q", got, want)
			}
			return nil
		}
	}

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: factories,
		Steps: []resource.TestStep{
			{
				Config: config(`policies = ["policy-a"]`),
				Check:  strategy("UNANIMOUS"),
			},
			{
				Config: config(`policies = ["policy-a"]
		decision_strategy = "AFFIRMATIVE"`),
				Check: strategy("AFFIRMATIVE"),
			},
			{
				Config: config(`policies = ["policy-a"]`),
				Check:  strategy("UNANIMOUS"),
			},
		},
	})
}
