package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestKeycloakRealmUserProfileNumericValidatorRead(t *testing.T) {
	data := schema.TestResourceDataRaw(t, resourceKeycloakRealmUserProfile().Schema, nil)
	profile := &keycloak.RealmUserProfile{Attributes: []*keycloak.RealmUserProfileAttribute{{
		Name: "username", Validations: map[string]keycloak.RealmUserProfileValidationConfig{
			"length": {"min": float64(3), "max": float64(1000000), "ignore.empty.value": true},
		},
	}}}
	if err := setRealmUserProfileData(context.Background(), nil, data, profile); err != nil {
		t.Fatal(err)
	}
	attrs := data.Get("attribute").([]interface{})
	if len(attrs) != 1 {
		t.Fatalf("expected the default attribute to be retained, got %d", len(attrs))
	}
	validators := attrs[0].(map[string]interface{})["validator"].(*schema.Set).List()
	config := validators[0].(map[string]interface{})["config"].(map[string]interface{})
	for key, expected := range map[string]string{"min": "3", "max": "1000000", "ignore.empty.value": "true"} {
		if config[key] != expected {
			t.Errorf("validator %s = %#v, want %q", key, config[key], expected)
		}
	}
}

func TestRealmUserProfileImportReadOnly(t *testing.T) {
	for _, tc := range []struct {
		name, id string
		status   int
		valid    bool
	}{
		{"existing", "fixture", 200, true}, {"missing", "fixture", 404, false},
		{"forbidden", "fixture", 403, false}, {"empty", "", 200, false},
		{"malformed", "fixture/profile", 200, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reads := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/admin/serverinfo" {
					w.Header().Set("Content-Type", "application/json")
					fmt.Fprint(w, `{}`)
					return
				}
				if r.Method != http.MethodGet || r.URL.Path != "/admin/realms/fixture/users/profile" {
					t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
					w.WriteHeader(500)
					return
				}
				reads++
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				fmt.Fprint(w, `{"attributes":[{"name":"username","validations":{"length":{"min":3,"max":1000000}}}],"groups":[{"name":"personal"}],"unmanagedAttributePolicy":"ENABLED"}`)
			}))
			defer server.Close()
			client, err := keycloak.NewKeycloakClient(context.Background(), server.URL, "", "", "", "", "master", "", "", "fixture-token", "", "", "", "", false, 10, "", false, "", "", "test", false, nil, "26.8.2")
			if err != nil {
				t.Fatal(err)
			}
			r := resourceKeycloakRealmUserProfile()
			data := schema.TestResourceDataRaw(t, r.Schema, nil)
			data.SetId(tc.id)
			result, err := r.Importer.StateContext(context.Background(), data, client)
			if (err == nil) != tc.valid {
				t.Fatalf("import error: %v", err)
			}
			if tc.valid {
				if len(result) != 1 || data.Id() != "fixture" || data.Get("realm_id") != "fixture" || len(data.Get("attribute").([]interface{})) != 1 || data.Get("group").(*schema.Set).Len() != 1 || data.Get("unmanaged_attribute_policy") != "ENABLED" {
					t.Fatalf("incomplete import state: %#v", data.State())
				}
			}
			if tc.id == "" || tc.id == "fixture/profile" {
				if reads != 0 {
					t.Fatal("invalid ID made an HTTP request")
				}
			} else if reads == 0 || (tc.valid && reads != 1) {
				t.Fatalf("unexpected profile read count: %d", reads)
			}
		})
	}
}
