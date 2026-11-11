package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkterraform "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestKeycloakClientScopeAssignmentsImportRejectsEmptyComponents(t *testing.T) {
	for _, id := range []string{"realm/", "client/my-realm/", "client//my-client", "client/my-realm/my-client/extra"} {
		data := resourceKeycloakClientScopeAssignments().TestResourceData()
		data.SetId(id)
		if _, err := resourceKeycloakClientScopeAssignmentsImport(context.Background(), data, nil); err == nil {
			t.Errorf("%q: expected import error", id)
		}
	}
}

func TestKeycloakClientScopeAssignmentsDiff(t *testing.T) {
	tests := []struct {
		name           string
		defaultScopes  cty.Value
		optionalScopes cty.Value
		wantError      bool
	}{
		{
			name: "known overlap", wantError: true,
			defaultScopes:  cty.SetVal([]cty.Value{cty.StringVal("email")}),
			optionalScopes: cty.SetVal([]cty.Value{cty.StringVal("email")}),
		},
		{
			name:           "unknown sets",
			defaultScopes:  cty.UnknownVal(cty.Set(cty.String)),
			optionalScopes: cty.UnknownVal(cty.Set(cty.String)),
		},
		{
			name:           "unknown elements",
			defaultScopes:  cty.SetVal([]cty.Value{cty.UnknownVal(cty.String)}),
			optionalScopes: cty.SetVal([]cty.Value{cty.UnknownVal(cty.String)}),
		},
		{
			name:           "partially unknown sets",
			defaultScopes:  cty.SetVal([]cty.Value{cty.StringVal("profile"), cty.UnknownVal(cty.String)}),
			optionalScopes: cty.SetVal([]cty.Value{cty.StringVal("email"), cty.UnknownVal(cty.String)}),
		},
		{
			name:           "disjoint known sets",
			defaultScopes:  cty.SetVal([]cty.Value{cty.StringVal("profile")}),
			optionalScopes: cty.SetVal([]cty.Value{cty.StringVal("email")}),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := resourceKeycloakClientScopeAssignments()
			// This SDK helper preserves typed unknown cty values for the legacy Resource.Diff API.
			config := sdkterraform.NewResourceConfigShimmed(cty.ObjectVal(map[string]cty.Value{
				"id": cty.NullVal(cty.String), "realm_id": cty.StringVal("realm"), "client_id": cty.NullVal(cty.String),
				"default_scopes": test.defaultScopes, "optional_scopes": test.optionalScopes,
			}), r.CoreConfigSchema())
			_, err := r.Diff(context.Background(), nil, config, nil)
			if (err != nil) != test.wantError {
				t.Fatalf("got error %v, wantError %t", err, test.wantError)
			}
		})
	}
}

func TestPlanClientScopeAssignmentChanges(t *testing.T) {
	tests := []struct {
		name    string
		current clientScopeAssignments
		desired clientScopeAssignments
		want    clientScopeAssignmentChanges
	}{
		{
			name:    "moves in both directions preserve custom assignments",
			current: clientScopeAssignments{defaultScopes: []string{"to-optional", "custom-default"}, optionalScopes: []string{"to-default", "custom-optional"}},
			desired: clientScopeAssignments{defaultScopes: []string{"to-default", "custom-default"}, optionalScopes: []string{"to-optional", "custom-optional"}},
			want: clientScopeAssignmentChanges{
				removeDefaults: []string{"to-optional"}, removeOptionals: []string{"to-default"},
				addDefaults: []string{"to-default"}, addOptionals: []string{"to-optional"},
			},
		},
		{
			name:    "unchanged custom assignments",
			current: clientScopeAssignments{defaultScopes: []string{"custom-default"}, optionalScopes: []string{"custom-optional"}},
			desired: clientScopeAssignments{defaultScopes: []string{"custom-default"}, optionalScopes: []string{"custom-optional"}},
		},
		{
			name:    "reattach defaults after type-agnostic delete",
			current: clientScopeAssignments{defaultScopes: []string{"scope"}, optionalScopes: []string{"scope"}},
			desired: clientScopeAssignments{defaultScopes: []string{"scope"}},
			want:    clientScopeAssignmentChanges{removeOptionals: []string{"scope"}, addDefaults: []string{"scope"}},
		},
		{
			name:    "reattach optionals after type-agnostic delete",
			current: clientScopeAssignments{defaultScopes: []string{"scope"}, optionalScopes: []string{"scope"}},
			desired: clientScopeAssignments{optionalScopes: []string{"scope"}},
			want:    clientScopeAssignmentChanges{removeDefaults: []string{"scope"}, addOptionals: []string{"scope"}},
		},
		{
			name:    "restore out-of-band move",
			current: clientScopeAssignments{optionalScopes: []string{"custom"}},
			desired: clientScopeAssignments{defaultScopes: []string{"custom"}},
			want:    clientScopeAssignmentChanges{removeOptionals: []string{"custom"}, addDefaults: []string{"custom"}},
		},
		{
			name:    "explicit empty sets do not inject defaults",
			current: clientScopeAssignments{defaultScopes: []string{"roles", "profile"}},
			want:    clientScopeAssignmentChanges{removeDefaults: []string{"profile", "roles"}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := planClientScopeAssignmentChanges(&test.current, &test.desired)
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("got changes %+v, want %+v", got, test.want)
			}
		})
	}
}

func TestKeycloakClientScopeAssignmentsDeleteSkipsStaleAssignments(t *testing.T) {
	for _, clientId := range []string{"", "client"} {
		t.Run(clientScopeAssignmentsId("realm", clientId), func(t *testing.T) {
			defaultPath := "/admin/realms/realm/default-default-client-scopes"
			optionalPath := "/admin/realms/realm/default-optional-client-scopes"
			if clientId != "" {
				defaultPath = "/admin/realms/realm/clients/client/default-client-scopes"
				optionalPath = "/admin/realms/realm/clients/client/optional-client-scopes"
			}
			currentDefaults := []map[string]string{
				{"id": "concurrently-removed", "name": "concurrently-removed", "protocol": "openid-connect"},
				{"id": "owned-default", "name": "owned-default", "protocol": "openid-connect"},
				{"id": "moved-optional", "name": "moved-optional", "protocol": "openid-connect"},
				{"id": "foreign-default", "name": "foreign-default", "protocol": "openid-connect"},
			}
			currentOptionals := []map[string]string{
				{"id": "owned-optional", "name": "owned-optional", "protocol": "openid-connect"},
				{"id": "moved-default", "name": "moved-default", "protocol": "openid-connect"},
				{"id": "foreign-optional", "name": "foreign-optional", "protocol": "openid-connect"},
			}
			var deletes []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch r.Method + " " + r.URL.Path {
				case "GET /admin/serverinfo":
					fmt.Fprint(w, `{"systemInfo":{"version":"26.8.0"}}`)
				case "GET /admin/realms/realm/clients/client":
					fmt.Fprint(w, `{"id":"client","protocol":"openid-connect"}`)
				case "GET " + defaultPath:
					if err := json.NewEncoder(w).Encode(currentDefaults); err != nil {
						t.Error(err)
					}
				case "GET " + optionalPath:
					if err := json.NewEncoder(w).Encode(currentOptionals); err != nil {
						t.Error(err)
					}
				case "GET /admin/realms/realm/client-scopes":
					if err := json.NewEncoder(w).Encode(append(currentDefaults, currentOptionals...)); err != nil {
						t.Error(err)
					}
				case "DELETE " + defaultPath + "/concurrently-removed":
					http.NotFound(w, r)
				case "DELETE " + defaultPath + "/owned-default", "DELETE " + optionalPath + "/owned-optional":
					deletes = append(deletes, r.URL.Path)
					w.WriteHeader(http.StatusNoContent)
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
					http.NotFound(w, r)
				}
			}))
			defer server.Close()
			client, err := keycloak.NewKeycloakClient(context.Background(), server.URL, "", "", "", "", "master", "", "", "test-token", "", "", "", "", false, 10, "", false, "", "", "", false, nil, "")
			if err != nil {
				t.Fatal(err)
			}
			data := schema.TestResourceDataRaw(t, resourceKeycloakClientScopeAssignments().Schema, map[string]interface{}{
				"realm_id": "realm", "client_id": clientId,
				"default_scopes":  []interface{}{"concurrently-removed", "owned-default", "moved-default", "deleted-default"},
				"optional_scopes": []interface{}{"owned-optional", "moved-optional", "deleted-optional"},
			})
			if diags := resourceKeycloakClientScopeAssignmentsDelete(context.Background(), data, client); diags.HasError() {
				t.Fatal(diags)
			}
			want := []string{defaultPath + "/owned-default", optionalPath + "/owned-optional"}
			if !reflect.DeepEqual(deletes, want) {
				t.Fatalf("deleted %v, want %v", deletes, want)
			}
		})
	}
}

func TestKeycloakClientScopeAssignmentsOpenidClientAcceptsOid4vcScope(t *testing.T) {
	var attached bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method + " " + r.URL.Path {
		case "GET /admin/serverinfo":
			fmt.Fprint(w, `{"systemInfo":{"version":"26.8.0"}}`)
		case "GET /admin/realms/realm/clients/client":
			fmt.Fprint(w, `{"id":"client","protocol":"openid-connect"}`)
		case "GET /admin/realms/realm/clients/client/default-client-scopes":
			if attached {
				fmt.Fprint(w, `[{"id":"vc-id","name":"vc","protocol":"oid4vc"}]`)
			} else {
				fmt.Fprint(w, `[]`)
			}
		case "GET /admin/realms/realm/clients/client/optional-client-scopes":
			fmt.Fprint(w, `[]`)
		case "GET /admin/realms/realm/client-scopes":
			fmt.Fprint(w, `[{"id":"vc-id","name":"vc","protocol":"oid4vc"}]`)
		case "PUT /admin/realms/realm/clients/client/default-client-scopes/vc-id":
			attached = true
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := keycloak.NewKeycloakClient(context.Background(), server.URL, "", "", "", "", "master", "", "", "test-token", "", "", "", "", false, 10, "", false, "", "", "", false, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	data := schema.TestResourceDataRaw(t, resourceKeycloakClientScopeAssignments().Schema, map[string]interface{}{
		"realm_id": "realm", "client_id": "client",
		"default_scopes":  []interface{}{"vc"},
		"optional_scopes": []interface{}{},
	})
	if diags := resourceKeycloakClientScopeAssignmentsReconcile(context.Background(), data, client); diags.HasError() {
		t.Fatal(diags)
	}
	if !attached {
		t.Fatal("oid4vc scope was not attached to the openid-connect client")
	}
}

func TestKeycloakClientScopeAssignmentsMoveReusesAssignedScopeIDs(t *testing.T) {
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		request := r.Method + " " + r.URL.Path
		if request == "GET /admin/serverinfo" {
			fmt.Fprint(w, `{"systemInfo":{"version":"26.8.0"}}`)
			return
		}
		requests = append(requests, request)
		switch request {
		case "GET /admin/realms/realm/clients/client":
			fmt.Fprint(w, `{"id":"client","protocol":"openid-connect"}`)
		case "GET /admin/realms/realm/clients/client/default-client-scopes":
			fmt.Fprint(w, `[{"id":"a-id","name":"a","protocol":"openid-connect"}]`)
		case "GET /admin/realms/realm/clients/client/optional-client-scopes":
			fmt.Fprint(w, `[{"id":"b-id","name":"b","protocol":"openid-connect"}]`)
		case "DELETE /admin/realms/realm/clients/client/default-client-scopes/a-id",
			"DELETE /admin/realms/realm/clients/client/optional-client-scopes/b-id",
			"PUT /admin/realms/realm/clients/client/default-client-scopes/b-id",
			"PUT /admin/realms/realm/clients/client/optional-client-scopes/a-id":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected request: %s", request)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := keycloak.NewKeycloakClient(context.Background(), server.URL, "", "", "", "", "master", "", "", "test-token", "", "", "", "", false, 10, "", false, "", "", "", false, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	data := schema.TestResourceDataRaw(t, resourceKeycloakClientScopeAssignments().Schema, map[string]interface{}{
		"realm_id": "realm", "client_id": "client",
		"default_scopes":  []interface{}{"b"},
		"optional_scopes": []interface{}{"a"},
	})
	if diags := resourceKeycloakClientScopeAssignmentsReconcile(context.Background(), data, client); diags.HasError() {
		t.Fatal(diags)
	}

	read := []string{
		"GET /admin/realms/realm/clients/client",
		"GET /admin/realms/realm/clients/client/default-client-scopes",
		"GET /admin/realms/realm/clients/client/optional-client-scopes",
	}
	want := append(append([]string{}, read...),
		"DELETE /admin/realms/realm/clients/client/default-client-scopes/a-id",
		"DELETE /admin/realms/realm/clients/client/optional-client-scopes/b-id",
		"PUT /admin/realms/realm/clients/client/default-client-scopes/b-id",
		"PUT /admin/realms/realm/clients/client/optional-client-scopes/a-id",
	)
	want = append(want, read...)
	if !reflect.DeepEqual(requests, want) {
		t.Fatalf("unexpected requests:\n got: %v\nwant: %v", requests, want)
	}
}

// clientScopeAssignmentsMockServer serves one realm and one client whose
// assignments live in memory. failPut makes the first PUT for that scope fail.
type clientScopeAssignmentsMockServer struct {
	mu        sync.Mutex
	clientRep string
	defaults  map[string][]string
	failPut   string
	writes    []string
}

func (m *clientScopeAssignmentsMockServer) serve(t *testing.T) *httptest.Server {
	scopes := []string{"keep", "old", "first", "second"}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.mu.Lock()
		defer m.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		path := r.URL.Path
		target := ""
		switch {
		case strings.HasPrefix(path, "/admin/realms/realm/default-default-client-scopes"):
			target = "realm"
		case strings.HasPrefix(path, "/admin/realms/realm/clients/client/default-client-scopes"):
			target = "client"
		}
		scopeName := path[strings.LastIndex(path, "/")+1:]
		encodeScopes := func(names []string) {
			reps := []map[string]string{}
			for _, name := range names {
				reps = append(reps, map[string]string{"id": name, "name": name, "protocol": "openid-connect"})
			}
			if err := json.NewEncoder(w).Encode(reps); err != nil {
				t.Error(err)
			}
		}
		switch {
		case r.Method == http.MethodGet && path == "/admin/serverinfo":
			fmt.Fprint(w, `{"systemInfo":{"version":"26.8.0"}}`)
		case r.Method == http.MethodGet && path == "/admin/realms/realm/clients/client":
			fmt.Fprint(w, m.clientRep)
		case r.Method == http.MethodGet && (path == "/admin/realms/realm/default-optional-client-scopes" || path == "/admin/realms/realm/clients/client/optional-client-scopes"):
			fmt.Fprint(w, `[]`)
		case r.Method == http.MethodGet && path == "/admin/realms/realm/client-scopes":
			encodeScopes(scopes)
		case r.Method == http.MethodGet && target != "":
			encodeScopes(m.defaults[target])
		case r.Method == http.MethodPut && target != "":
			if scopeName == m.failPut {
				m.failPut = ""
				http.Error(w, `{"error":"simulated server failure"}`, http.StatusInternalServerError)
				return
			}
			m.writes = append(m.writes, r.Method+" "+target+"/"+scopeName)
			m.defaults[target] = append(m.defaults[target], scopeName)
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodDelete && target != "":
			m.writes = append(m.writes, r.Method+" "+target+"/"+scopeName)
			remaining := []string{}
			for _, name := range m.defaults[target] {
				if name != scopeName {
					remaining = append(remaining, name)
				}
			}
			m.defaults[target] = remaining
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, path)
			http.NotFound(w, r)
		}
	}))
}

func clientScopeAssignmentsMockProviderFactories(t *testing.T, serverURL string) map[string]func() (tfprotov5.ProviderServer, error) {
	client, err := keycloak.NewKeycloakClient(context.Background(), serverURL, "", "", "", "", "master", "", "", "test-token", "", "", "", "", false, 10, "", false, "", "", "", false, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	provider := &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{"keycloak_client_scope_assignments": resourceKeycloakClientScopeAssignments()},
		ConfigureContextFunc: func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
			return client, nil
		},
	}
	return protoV5ProviderFactories(provider)
}

func TestKeycloakClientScopeAssignments_failedCreateRetryKeepsUnchangedAssignments(t *testing.T) {
	for _, clientId := range []string{"", "client"} {
		t.Run(clientScopeAssignmentsId("realm", clientId), func(t *testing.T) {
			target := "realm"
			clientIdAttr := ""
			if clientId != "" {
				target = "client"
				clientIdAttr = fmt.Sprintf("client_id = %q", clientId)
			}
			// The client omits its protocol, which Keycloak treats as openid-connect.
			mock := &clientScopeAssignmentsMockServer{
				clientRep: `{"id":"client"}`,
				defaults:  map[string][]string{target: {"keep", "old"}},
				failPut:   "second",
			}
			server := mock.serve(t)
			defer server.Close()

			config := fmt.Sprintf(`
resource "keycloak_client_scope_assignments" "assignments" {
	realm_id        = "realm"
	%s
	default_scopes  = ["keep", "first", "second"]
	optional_scopes = []
}`, clientIdAttr)
			resource.UnitTest(t, resource.TestCase{
				ProtoV5ProviderFactories: clientScopeAssignmentsMockProviderFactories(t, server.URL),
				Steps: []resource.TestStep{
					{
						Config:      config,
						ExpectError: regexp.MustCompile(`simulated server failure`),
					},
					{
						// A tainted failed create would be destroyed first, detaching "keep".
						Config: config,
						Check: func(*terraform.State) error {
							mock.mu.Lock()
							defer mock.mu.Unlock()
							want := []string{"DELETE " + target + "/old", "PUT " + target + "/first", "PUT " + target + "/second"}
							if !reflect.DeepEqual(mock.writes, want) {
								return fmt.Errorf("got writes %v, want %v", mock.writes, want)
							}
							return nil
						},
					},
				},
			})
		})
	}
}

func TestKeycloakClientScopeAssignments_rejectEmptyClientId(t *testing.T) {
	mock := &clientScopeAssignmentsMockServer{
		clientRep: `{"id":"client","protocol":"openid-connect"}`,
		defaults:  map[string][]string{"realm": {"keep"}},
	}
	server := mock.serve(t)
	defer server.Close()

	config := func(clientId string) string {
		return fmt.Sprintf(`
resource "keycloak_client_scope_assignments" "assignments" {
	realm_id        = "realm"
	client_id       = %s
	default_scopes  = []
	optional_scopes = []
}`, clientId)
	}
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: clientScopeAssignmentsMockProviderFactories(t, server.URL),
		Steps: []resource.TestStep{
			{
				Config:      config(`""`),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`expected "client_id" to not be an empty string`),
			},
			{
				// Unknown during planning; Terraform revalidates once known at apply.
				Config:      config(`substr(timestamp(), 0, 0)`),
				ExpectError: regexp.MustCompile(`expected "client_id" to not be an empty string`),
			},
		},
	})
	if len(mock.writes) != 0 {
		t.Fatalf("empty client_id mutated assignments: %v", mock.writes)
	}
}

func TestKeycloakClientScopeAssignmentsValidateBeforeMutation(t *testing.T) {
	tests := []struct {
		name         string
		desiredScope string
		scopeList    string
		bearerOnly   bool
		wantMessage  string
	}{
		{
			name:         "missing desired scope",
			desiredScope: "missing",
			scopeList:    `[]`,
			wantMessage:  `client scope "missing" does not exist`,
		},
		{
			name:         "wrong protocol",
			desiredScope: "wrong-protocol",
			scopeList:    `[{"id":"scope-id","name":"wrong-protocol","protocol":"saml"}]`,
			wantMessage:  `client scope "wrong-protocol" uses protocol "saml"`,
		},
		{
			name:         "bearer-only client",
			desiredScope: "scope",
			scopeList:    `[{"id":"scope-id","name":"scope","protocol":"openid-connect"}]`,
			bearerOnly:   true,
			wantMessage:  "BEARER-ONLY",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var mutations []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch r.Method + " " + r.URL.Path {
				case "GET /admin/serverinfo":
					fmt.Fprint(w, `{"systemInfo":{"version":"26.8.0"}}`)
				case "GET /admin/realms/realm/clients/client":
					fmt.Fprintf(w, `{"id":"client","protocol":"openid-connect","bearerOnly":%t}`, test.bearerOnly)
				case "GET /admin/realms/realm/clients/client/default-client-scopes":
					fmt.Fprint(w, `[{"id":"old-id","name":"old","protocol":"openid-connect"}]`)
				case "GET /admin/realms/realm/clients/client/optional-client-scopes":
					fmt.Fprint(w, `[]`)
				case "GET /admin/realms/realm/client-scopes":
					fmt.Fprint(w, test.scopeList)
				case "PUT /admin/realms/realm/clients/client/default-client-scopes/old-id",
					"DELETE /admin/realms/realm/clients/client/default-client-scopes/old-id":
					mutations = append(mutations, r.Method+" "+r.URL.Path)
					w.WriteHeader(http.StatusNoContent)
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
					http.NotFound(w, r)
				}
			}))
			defer server.Close()

			client, err := keycloak.NewKeycloakClient(context.Background(), server.URL, "", "", "", "", "master", "", "", "test-token", "", "", "", "", false, 10, "", false, "", "", "", false, nil, "")
			if err != nil {
				t.Fatal(err)
			}
			data := schema.TestResourceDataRaw(t, resourceKeycloakClientScopeAssignments().Schema, map[string]interface{}{
				"realm_id": "realm", "client_id": "client",
				"default_scopes":  []interface{}{test.desiredScope},
				"optional_scopes": []interface{}{},
			})

			diags := resourceKeycloakClientScopeAssignmentsReconcile(context.Background(), data, client)
			if !diags.HasError() {
				t.Fatal("expected assignment validation error")
			}
			if !strings.Contains(diags[0].Summary+" "+diags[0].Detail, test.wantMessage) {
				t.Fatalf("got diagnostic %v, want it to contain %q", diags[0], test.wantMessage)
			}
			if len(mutations) != 0 {
				t.Fatalf("mutated assignments before validation: %v", mutations)
			}
			if data.Id() != "" {
				t.Fatalf("validation failure retained an unmodified target: %q", data.Id())
			}
		})
	}
}

func TestValidateNoDuplicateClientScopeAssignments(t *testing.T) {
	tests := []struct {
		name           string
		defaultScopes  []string
		optionalScopes []string
		wantError      bool
	}{
		{name: "disjoint sets", defaultScopes: []string{"profile"}, optionalScopes: []string{"email"}},
		{name: "empty sets"},
		{name: "overlap", defaultScopes: []string{"profile", "email"}, optionalScopes: []string{"email"}, wantError: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateNoDuplicateClientScopeAssignments(test.defaultScopes, test.optionalScopes)
			if (err != nil) != test.wantError {
				t.Fatalf("got error %v, wantError %t", err, test.wantError)
			}
		})
	}
}

func TestAccKeycloakClientScopeAssignments_samlDefaultOptionalMove(t *testing.T) {
	t.Parallel()
	client := acctest.RandomWithPrefix("tf-acc")
	scope := acctest.RandomWithPrefix("tf-acc")
	signingPrivateKeyExpr := testKeycloakSamlClientDefaultScopes_signingPrivateKeyExpr(t)
	baseConfig := testKeycloakSamlClientDefaultScopes_base(client, signingPrivateKeyExpr) + fmt.Sprintf(`
resource "keycloak_client_scope" "scope" {
	realm_id = data.keycloak_realm.realm.id
	name     = %q

	saml {}
}
`, scope)

	config := func(defaultScopes, optionalScopes []string) string {
		return baseConfig + fmt.Sprintf(`
resource "keycloak_client_scope_assignments" "assignments" {
	realm_id        = data.keycloak_realm.realm.id
	client_id       = keycloak_saml_client.client.id
	default_scopes  = %s
	optional_scopes = %s
	depends_on      = [keycloak_client_scope.scope]
}
`, arrayOfStringsForTerraformResource(defaultScopes), arrayOfStringsForTerraformResource(optionalScopes))
	}

	resourceName := "keycloak_client_scope_assignments.assignments"
	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: config([]string{"role_list", scope}, nil),
				Check:  testAccCheckSamlClientScopeAssignments(resourceName, []string{"role_list", scope}, nil),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: config([]string{"role_list"}, []string{scope}),
				Check:  testAccCheckSamlClientScopeAssignments(resourceName, []string{"role_list"}, []string{scope}),
			},
			{
				Config: config([]string{"role_list", scope}, nil),
				Check:  testAccCheckSamlClientScopeAssignments(resourceName, []string{"role_list", scope}, nil),
			},
		},
	})
}

func TestAccKeycloakClientScopeAssignments_openidDefaultOptionalMove(t *testing.T) {
	t.Parallel()
	client := acctest.RandomWithPrefix("tf-acc")
	customDefault := acctest.RandomWithPrefix("tf-acc")
	customOptional := acctest.RandomWithPrefix("tf-acc")
	defaults := []string{"profile", "email", "roles", "web-origins", "acr", "basic", customDefault}
	resourceName := "keycloak_client_scope_assignments.assignments"
	config := func(defaultScopes, optionalScopes []string) string {
		return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = %q
}

resource "keycloak_openid_client" "client" {
	realm_id    = data.keycloak_realm.realm.id
	client_id   = %q
	access_type = "PUBLIC"
}

resource "keycloak_client_scope" "custom_default" {
	realm_id = data.keycloak_realm.realm.id
	name     = %q

	openid_connect {}
}

resource "keycloak_client_scope" "custom_optional" {
	realm_id = data.keycloak_realm.realm.id
	name     = %q

	openid_connect {}
}

resource "keycloak_client_scope_assignments" "assignments" {
	realm_id        = data.keycloak_realm.realm.id
	client_id       = keycloak_openid_client.client.id
	default_scopes  = %s
	optional_scopes = %s
	depends_on      = [keycloak_client_scope.custom_default, keycloak_client_scope.custom_optional]
}
`, testAccRealm.Realm, client, customDefault, customOptional, arrayOfStringsForTerraformResource(defaultScopes), arrayOfStringsForTerraformResource(optionalScopes))
	}

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: config(defaults, []string{customOptional}),
				Check:  testAccCheckOpenidClientScopeAssignments(resourceName, defaults, []string{customOptional}),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: config([]string{"profile", "roles", "web-origins", "acr", "basic", customDefault}, []string{"email", customOptional}),
				Check:  testAccCheckOpenidClientScopeAssignments(resourceName, []string{"profile", "roles", "web-origins", "acr", "basic", customDefault}, []string{"email", customOptional}),
			},
			{
				Config: config(defaults, []string{customOptional}),
				Check:  testAccCheckOpenidClientScopeAssignments(resourceName, defaults, []string{customOptional}),
			},
		},
	})
}

func TestAccKeycloakClientScopeAssignments_realmDefaultOptionalMove(t *testing.T) {
	t.Parallel()
	realm := acctest.RandomWithPrefix("tf-acc")
	scope := acctest.RandomWithPrefix("tf-acc")
	config := func(defaultScopes, optionalScopes []string) string {
		return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm = %q
}

resource "keycloak_client_scope" "scope" {
	realm_id = keycloak_realm.realm.id
	name     = %q

	openid_connect {}
}

resource "keycloak_client_scope_assignments" "assignments" {
	realm_id        = keycloak_realm.realm.id
	default_scopes  = %s
	optional_scopes = %s
	depends_on      = [keycloak_client_scope.scope]
}
`, realm, scope, arrayOfStringsForTerraformResource(defaultScopes), arrayOfStringsForTerraformResource(optionalScopes))
	}
	resourceName := "keycloak_client_scope_assignments.assignments"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: config([]string{"profile", "email", scope}, []string{"offline_access"}),
				Check:  testAccCheckRealmClientScopeAssignments(resourceName, []string{"profile", "email", scope}, []string{"offline_access"}),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					resourceState, ok := state.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("resource not found: %s", resourceName)
					}
					return "realm/" + resourceState.Primary.Attributes["realm_id"], nil
				},
				ImportStateVerify: true,
			},
			{
				Config: config([]string{"profile", "email"}, []string{"offline_access", scope}),
				Check:  testAccCheckRealmClientScopeAssignments(resourceName, []string{"profile", "email"}, []string{"offline_access", scope}),
			},
			{
				Config: config([]string{"profile", "email", scope}, []string{"offline_access"}),
				Check:  testAccCheckRealmClientScopeAssignments(resourceName, []string{"profile", "email", scope}, []string{"offline_access"}),
			},
		},
	})
}

func TestAccKeycloakClientScopeAssignments_rejectDuplicateScopeTypes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: `
resource "keycloak_client_scope_assignments" "assignments" {
	realm_id        = "unused"
	default_scopes  = ["duplicate"]
	optional_scopes = ["duplicate"]
}
`,
				ExpectError: regexp.MustCompile(`client scope "duplicate" cannot be assigned as both default and optional`),
			},
		},
	})
}

func testAccCheckOpenidClientScopeAssignments(resourceName string, expectedDefaults, expectedOptionals []string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		realmId := resourceState.Primary.Attributes["realm_id"]
		clientId := resourceState.Primary.Attributes["client_id"]
		defaultScopes, err := keycloakClient.GetOpenidClientDefaultScopes(testCtx, realmId, clientId)
		if err != nil {
			return err
		}
		optionalScopes, err := keycloakClient.GetOpenidClientOptionalScopes(testCtx, realmId, clientId)
		if err != nil {
			return err
		}
		return compareClientScopeNames(scopeNamesFromOpenidScopePointers(defaultScopes), scopeNamesFromOpenidScopePointers(optionalScopes), expectedDefaults, expectedOptionals)
	}
}

func testAccCheckRealmClientScopeAssignments(resourceName string, expectedDefaults, expectedOptionals []string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		realmId := resourceState.Primary.Attributes["realm_id"]
		defaultScopes, err := keycloakClient.GetRealmDefaultClientScopes(testCtx, realmId)
		if err != nil {
			return err
		}
		optionalScopes, err := keycloakClient.GetRealmOptionalClientScopes(testCtx, realmId)
		if err != nil {
			return err
		}
		return compareClientScopeNames(scopeNamesFromOpenidScopePointers(defaultScopes), scopeNamesFromOpenidScopePointers(optionalScopes), expectedDefaults, expectedOptionals)
	}
}

func testAccCheckSamlClientScopeAssignments(resourceName string, expectedDefaults, expectedOptionals []string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		realmId := resourceState.Primary.Attributes["realm_id"]
		clientId := resourceState.Primary.Attributes["client_id"]
		defaultScopes, err := keycloakClient.GetSamlClientDefaultScopes(testCtx, realmId, clientId)
		if err != nil {
			return err
		}
		optionalScopes, err := keycloakClient.GetSamlClientOptionalScopes(testCtx, realmId, clientId)
		if err != nil {
			return err
		}

		var actualDefaults, actualOptionals []string
		for _, scope := range defaultScopes {
			actualDefaults = append(actualDefaults, scope.Name)
		}
		for _, scope := range optionalScopes {
			actualOptionals = append(actualOptionals, scope.Name)
		}
		return compareClientScopeNames(actualDefaults, actualOptionals, expectedDefaults, expectedOptionals)
	}
}

func scopeNamesFromOpenidScopes(scopes []keycloak.OpenidClientScope) []string {
	names := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		names = append(names, scope.Name)
	}
	return names
}

func scopeNamesFromOpenidScopePointers(scopes []*keycloak.OpenidClientScope) []string {
	names := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		names = append(names, scope.Name)
	}
	return names
}

func compareClientScopeNames(actualDefaults, actualOptionals, expectedDefaults, expectedOptionals []string) error {
	actualDefaults = append([]string(nil), actualDefaults...)
	actualOptionals = append([]string(nil), actualOptionals...)
	expectedDefaults = append([]string(nil), expectedDefaults...)
	expectedOptionals = append([]string(nil), expectedOptionals...)
	sort.Strings(actualDefaults)
	sort.Strings(actualOptionals)
	sort.Strings(expectedDefaults)
	sort.Strings(expectedOptionals)
	if !reflect.DeepEqual(actualDefaults, expectedDefaults) {
		return fmt.Errorf("default scopes mismatch: got %v, want %v", actualDefaults, expectedDefaults)
	}
	if !reflect.DeepEqual(actualOptionals, expectedOptionals) {
		return fmt.Errorf("optional scopes mismatch: got %v, want %v", actualOptionals, expectedOptionals)
	}
	return nil
}
