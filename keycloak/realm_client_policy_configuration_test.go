package keycloak

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestRealmClientPolicyConfigurationTransport(t *testing.T) {
	payload := `[{"name":"fixture","description":"","executors":[{"executor":"fixture","configuration":{"enabled":false,"number":9007199254740993,"list":["a"],"object":{"x":true}}}]}]`
	var entries []json.RawMessage
	if err := json.Unmarshal([]byte(payload), &entries); err != nil {
		t.Fatal(err)
	}
	writes := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "GET" {
			switch r.URL.Path {
			case "/admin/realms/fixture/client-policies/profiles":
				if r.URL.Query().Get("include-global-profiles") != "true" {
					t.Error("global profile observation missing")
				}
				_, _ = w.Write([]byte(`{"profiles":` + payload + `,"globalProfiles":[{"name":"global"}]}`))
			case "/admin/realms/fixture/client-policies/policies":
				if r.URL.Query().Get("include-global-policies") != "true" {
					t.Error("global policy observation missing")
				}
				_, _ = w.Write([]byte(`{"policies":[],"globalPolicies":[]}`))
			default:
				t.Errorf("unexpected path %s", r.URL.Path)
			}
			return
		}
		writes++
		var body map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if r.Method != "PUT" || r.URL.Path != "/admin/realms/fixture/client-policies/profiles" || len(body) != 1 {
			t.Error("unexpected update or global fields in payload")
		}
		var got []json.RawMessage
		if err := json.Unmarshal(body["profiles"], &got); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, entries) {
			t.Error("configuration JSON types or large integer changed")
		}
		w.WriteHeader(204)
	}))
	defer server.Close()
	client := &KeycloakClient{baseUrl: server.URL, httpClient: server.Client(), initialLogin: true, clientCredentials: &ClientCredentials{AccessToken: "fixture", TokenType: "Bearer"}}
	result, err := client.GetRealmClientPolicyConfiguration(context.Background(), "fixture")
	if err != nil || !reflect.DeepEqual(result.Profiles, entries) || len(result.GlobalProfiles) != 1 {
		t.Fatalf("read failed: %v", err)
	}
	if err = client.PutRealmClientPolicyCollection(context.Background(), "fixture", "profiles", entries); err != nil || writes != 1 {
		t.Fatalf("write failed: %v", err)
	}
	if err = client.PutRealmClientPolicyCollection(context.Background(), "fixture", "globalProfiles", entries); err == nil || writes != 1 {
		t.Fatal("global collection write not rejected locally")
	}
}
