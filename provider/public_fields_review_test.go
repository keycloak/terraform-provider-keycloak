package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkterraform "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"net/http"
	"net/http/httptest"
	"testing"
)

func reviewClient(t *testing.T, h http.HandlerFunc) *keycloak.KeycloakClient {
	t.Helper()
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/admin/serverinfo" {
			fmt.Fprint(w, `{}`)
			return
		}
		h(w, r)
	}))
	t.Cleanup(s.Close)
	c, e := keycloak.NewKeycloakClient(context.Background(), s.URL, "", "", "", "", "master", "", "", "fixture", "", "", "", "", false, 10, "", false, "", "", "test", false, nil, "26.8.2")
	if e != nil {
		t.Fatal(e)
	}
	return c
}

func TestBindingUpdatesOmitMaskedFeatures(t *testing.T) {
	for _, action := range []string{"create", "update", "delete"} {
		t.Run(action, func(t *testing.T) {
			writes := 0
			c := reviewClient(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodGet {
					fmt.Fprint(w, `{"realm":"fixture","sslRequired":"external","scimApiEnabled":false,"verifiableCredentialsEnabled":false}`)
					return
				}
				var body map[string]interface{}
				json.NewDecoder(r.Body).Decode(&body)
				writes++
				for _, k := range []string{"scimApiEnabled", "verifiableCredentialsEnabled"} {
					if _, ok := body[k]; ok {
						t.Errorf("binding writes unmanaged feature %s", k)
					}
				}
				w.WriteHeader(204)
			})
			r := resourceKeycloakAuthenticationBindings()
			d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{"realm_id": "fixture"})
			d.SetId("fixture")
			var f func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics = r.UpdateContext
			if action == "create" {
				f = r.CreateContext
			}
			if action == "delete" {
				f = r.DeleteContext
			}
			if ds := f(context.Background(), d, c); ds.HasError() {
				t.Fatal(ds)
			}
			if writes != 1 {
				t.Fatalf("writes %d", writes)
			}
		})
	}
}

func TestRealmUpdateReadsPersistedCredentials(t *testing.T) {
	r := resourceKeycloakRealm()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{"realm": "fixture"})
	d.SetId("fixture")
	observed, e := getRealmFromData(d, version.Must(version.NewVersion("26.8.2")))
	if e != nil {
		t.Fatal(e)
	}
	creds := []string{"password"}
	observed.RequiredCredentials = &creds
	c := reviewClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			json.NewEncoder(w).Encode(observed)
			return
		}
		w.WriteHeader(204)
	})
	if ds := r.UpdateContext(context.Background(), d, c); ds.HasError() {
		t.Fatal(ds)
	}
	if d.Get("required_credentials").(*schema.Set).Len() != 1 {
		t.Fatal("partial request replaced persisted credentials in state")
	}
}

func TestClientImportPublicFieldOwnership(t *testing.T) {
	for _, configured := range []bool{false, true} {
		t.Run(fmt.Sprint(configured), func(t *testing.T) {
			writes := 0
			c := reviewClient(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodGet {
					if r.URL.RawQuery != "" {
						fmt.Fprint(w, `[{"id":"client","clientId":"fixture","publicClient":true,"surrogateAuthRequired":true,"nodeReRegistrationTimeout":30}]`)
					} else {
						fmt.Fprint(w, `{"id":"client","clientId":"fixture","publicClient":true}`)
					}
					return
				}
				var body map[string]interface{}
				json.NewDecoder(r.Body).Decode(&body)
				writes++
				for k, want := range map[string]interface{}{"surrogateAuthRequired": false, "nodeReRegistrationTimeout": float64(0)} {
					v, ok := body[k]
					if ok != configured || (configured && v != want) {
						t.Errorf("%s = %v present %t, configured %t", k, v, ok, configured)
					}
				}
				w.WriteHeader(204)
			})
			r := resourceKeycloakOpenidClient()
			config := map[string]interface{}{"realm_id": "fixture", "client_id": "fixture", "access_type": "PUBLIC", "import": true}
			if configured {
				config["surrogate_auth_required"] = false
				config["node_re_registration_timeout"] = 0
			}
			raw := map[string]cty.Value{"surrogate_auth_required": cty.NullVal(cty.Bool), "node_re_registration_timeout": cty.NullVal(cty.Number)}
			if configured {
				raw["surrogate_auth_required"] = cty.False
				raw["node_re_registration_timeout"] = cty.NumberIntVal(0)
			}
			state := &sdkterraform.InstanceState{RawConfig: cty.ObjectVal(raw)}
			diff, e := r.Diff(context.Background(), state, sdkterraform.NewResourceConfigRaw(config), c)
			if e != nil {
				t.Fatal(e)
			}
			_, ds := r.Apply(context.Background(), nil, diff, c)
			if ds.HasError() {
				t.Fatal(ds)
			}
			if writes != 1 {
				t.Fatalf("writes %d", writes)
			}
		})
	}
}
