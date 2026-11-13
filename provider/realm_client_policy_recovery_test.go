package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

func TestClientPolicyFailedCreateRecovery(t *testing.T) {
	profiles, policies := json.RawMessage(`[]`), json.RawMessage(`[]`)
	fail := true
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/admin/serverinfo" {
			fmt.Fprint(w, `{}`)
			return
		}
		collection := strings.TrimPrefix(r.URL.Path, "/admin/realms/fixture/client-policies/")
		if collection != "profiles" && collection != "policies" {
			t.Errorf("unexpected %s", r.URL.Path)
			w.WriteHeader(404)
			return
		}
		if r.Method == http.MethodGet {
			if collection == "profiles" {
				fmt.Fprintf(w, `{"profiles":%s,"globalProfiles":[]}`, profiles)
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
		if collection == "policies" && fail {
			fail = false
			w.WriteHeader(400)
			fmt.Fprint(w, `{"error":"injected failure"}`)
			return
		}
		var body map[string]json.RawMessage
		json.NewDecoder(r.Body).Decode(&body)
		if collection == "profiles" {
			profiles = body[collection]
		} else {
			policies = body[collection]
		}
		w.WriteHeader(204)
	}))
	defer s.Close()
	c, e := keycloak.NewKeycloakClient(context.Background(), s.URL, "", "", "", "", "master", "", "", "fixture", "", "", "", "", false, 10, "", false, "", "", "test", false, nil, "26.8.2")
	if e != nil {
		t.Fatal(e)
	}
	provider := KeycloakProvider(c)
	config := fmt.Sprintf(`provider "keycloak" { url = %q }
resource "keycloak_realm_client_policy_configuration" "test" {
 realm_id = "fixture"
 profiles = jsonencode([{name="new"}])
 policies = jsonencode([{name="policy",profiles=["new"]}])
}`, s.URL)
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: map[string]func() (tfprotov5.ProviderServer, error){"keycloak": func() (tfprotov5.ProviderServer, error) { return provider.GRPCProvider(), nil }},
		Steps: []resource.TestStep{
			{Config: config, ExpectError: regexp.MustCompile("injected failure")},
			// Forget the tainted instance without deleting the remote singleton.
			{Config: fmt.Sprintf(`provider "keycloak" { url = %q }
removed {
 from = keycloak_realm_client_policy_configuration.test
 lifecycle { destroy = false }
}`, s.URL)},
			// Reimport the existing collections, then resume the incomplete write.
			{Config: config, ImportStateConfigExact: true, ImportState: true, ResourceName: "keycloak_realm_client_policy_configuration.test", ImportStateId: "fixture", ImportStatePersist: true},
			{Config: config, Check: resource.TestCheckResourceAttr("keycloak_realm_client_policy_configuration.test", "policies", `[{"name":"policy","profiles":["new"]}]`)},
			{Config: fmt.Sprintf(`provider "keycloak" { url = %q }
removed {
 from = keycloak_realm_client_policy_configuration.test
 lifecycle { destroy = false }
}`, s.URL)},
		},
	})
}
