package keycloak

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPublicConfigurationOmissionAndResets(t *testing.T) {
	disabled, timeout := false, 0
	empty := []string{}
	cases := []struct {
		name     string
		value    interface{}
		expected map[string]interface{}
		absent   []string
	}{
		{"realm omission", &Realm{}, nil, []string{"requiredCredentials", "scimApiEnabled", "verifiableCredentialsEnabled"}},
		{"realm explicit resets", &Realm{RequiredCredentials: &empty, ScimApiEnabled: &disabled, VerifiableCredentialsEnabled: &disabled, MaxTemporaryLockouts: &timeout}, map[string]interface{}{"scimApiEnabled": false, "verifiableCredentialsEnabled": false, "maxTemporaryLockouts": float64(0)}, nil},
		{"client omission", &OpenidClient{}, nil, []string{"surrogateAuthRequired", "nodeReRegistrationTimeout"}},
		{"client explicit resets", &OpenidClient{SurrogateAuthRequired: &disabled, NodeReRegistrationTimeout: &timeout}, map[string]interface{}{"surrogateAuthRequired": false, "nodeReRegistrationTimeout": float64(0)}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := json.Marshal(tc.value)
			if err != nil {
				t.Fatal(err)
			}
			var got map[string]interface{}
			if err = json.Unmarshal(encoded, &got); err != nil {
				t.Fatal(err)
			}
			for _, key := range tc.absent {
				if _, present := got[key]; present {
					t.Fatalf("omitted %s unexpectedly serialized", key)
				}
			}
			for key, want := range tc.expected {
				if value, present := got[key]; !present || value != want {
					t.Fatalf("%s = %v (present %t), want %v", key, value, present, want)
				}
			}
			if tc.name == "realm explicit resets" {
				values, ok := got["requiredCredentials"].([]interface{})
				if !ok || len(values) != 0 {
					t.Fatal("empty required credentials not serialized as an empty array")
				}
			}
		})
	}
}

func TestRealmOptionalFeatureSwitchValidation(t *testing.T) {
	for _, tc := range []struct {
		name                           string
		configured, enabled, wantError bool
	}{
		{name: "explicit false with disabled feature"},
		{name: "true with disabled feature", configured: true, wantError: true},
		{name: "true with enabled feature", configured: true, enabled: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.Path != "/admin/serverinfo" {
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
					w.WriteHeader(400)
					return
				}
				fmt.Fprintf(w, `{"features":[{"name":"SCIM_API","enabled":%t},{"name":"OID4VC_VCI","enabled":%t}]}`, tc.enabled, tc.enabled)
			}))
			defer server.Close()
			client, err := NewKeycloakClient(context.Background(), server.URL, "", "", "", "", "master", "", "", "fixture-token", "", "", "", "", false, 10, "", false, "", "", "test", false, nil, "26.8.1")
			if err != nil {
				t.Fatal(err)
			}
			for _, field := range []string{"SCIM_API", "OID4VC_VCI"} {
				realm := &Realm{Realm: "fixture", SslRequired: "external"}
				if field == "SCIM_API" {
					realm.ScimApiEnabled = &tc.configured
				} else {
					realm.VerifiableCredentialsEnabled = &tc.configured
				}
				err = client.ValidateRealm(context.Background(), realm)
				if (err != nil) != tc.wantError {
					t.Fatalf("%s validation error %v, expected error %t", field, err, tc.wantError)
				}
			}
		})
	}
}
