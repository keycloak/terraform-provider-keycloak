package keycloak

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestRealmClientRegistrationPolicyConfigRemoval(t *testing.T) {
	for _, tc := range []struct {
		name       string
		desired    map[string]string
		want       map[string][]string
		readStatus int
	}{
		{name: "remove scalar and list keys", desired: map[string]string{"host-sending-registration-request-must-match": "false"}, want: map[string][]string{"host-sending-registration-request-must-match": {"false"}, "client-uris-must-match": {}, "trusted-hosts": {}}},
		{name: "clear entire config", desired: map[string]string{}, want: map[string][]string{"host-sending-registration-request-must-match": {}, "client-uris-must-match": {}, "trusted-hosts": {}}},
		{name: "retain and replace list", desired: map[string]string{"host-sending-registration-request-must-match": "true", "client-uris-must-match": "true", "trusted-hosts": "new.example, other.example"}, want: map[string][]string{"host-sending-registration-request-must-match": {"true"}, "client-uris-must-match": {"true"}, "trusted-hosts": {"new.example", "other.example"}}},
		{name: "failed read prevents write", desired: map[string]string{}, readStatus: http.StatusNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			puts := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch {
				case r.URL.Path == "/admin/serverinfo":
					fmt.Fprint(w, `{"componentTypes":{"org.keycloak.services.clientregistration.policy.ClientRegistrationPolicy":[{"id":"trusted-hosts","properties":[{"name":"trusted-hosts","type":"MultivaluedString"}]}]}}`)
				case r.Method == "GET" && r.URL.Path == "/admin/realms/realm/components/policy":
					if tc.readStatus != 0 {
						w.WriteHeader(tc.readStatus)
						fmt.Fprint(w, `{"error":"missing"}`)
						return
					}
					fmt.Fprint(w, `{"id":"policy","name":"Trusted Hosts","providerId":"trusted-hosts","providerType":"org.keycloak.services.clientregistration.policy.ClientRegistrationPolicy","subType":"anonymous","config":{"host-sending-registration-request-must-match":["true"],"client-uris-must-match":["true"],"trusted-hosts":["old.example"]}}`)
				case r.Method == "PUT" && r.URL.Path == "/admin/realms/realm/components/policy":
					puts++
					var body component
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Error(err)
					}
					if !reflect.DeepEqual(body.Config, tc.want) {
						t.Errorf("config = %#v, want %#v", body.Config, tc.want)
					}
					if body.ParentId != "" || body.Id != "policy" || body.ProviderId != "trusted-hosts" || body.SubType != "anonymous" {
						t.Error("policy identity changed")
					}
					w.WriteHeader(http.StatusNoContent)
				default:
					t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()
			client := &KeycloakClient{baseUrl: server.URL, httpClient: server.Client(), initialLogin: true, clientCredentials: &ClientCredentials{AccessToken: "fixture", TokenType: "Bearer"}}
			before := map[string]string{}
			for k, v := range tc.desired {
				before[k] = v
			}
			err := client.UpdateRealmClientRegistrationPolicy(context.Background(), &RealmClientRegistrationPolicy{Id: "policy", RealmId: "realm", Name: "Trusted Hosts", ProviderId: "trusted-hosts", SubType: "anonymous", Config: tc.desired})
			if tc.readStatus != 0 {
				if err == nil || puts != 0 {
					t.Fatal("failed read must stop before PUT")
				}
			} else if err != nil || puts != 1 {
				t.Fatalf("update = %v, writes = %d", err, puts)
			}
			if !reflect.DeepEqual(tc.desired, before) {
				t.Fatal("caller config was mutated")
			}
		})
	}
}
