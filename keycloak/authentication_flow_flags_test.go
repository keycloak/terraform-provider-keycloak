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

func TestAuthenticationFlowFlagPreservation(t *testing.T) {
	for _, tc := range []struct {
		name                                    string
		subflow, builtIn, topLevel, readFailure bool
	}{
		{name: "built-in top-level", builtIn: true, topLevel: true},
		{name: "built-in nested", builtIn: true},
		{name: "custom top-level", topLevel: true},
		{name: "custom nested"},
		{name: "built-in subflow", subflow: true, builtIn: true},
		{name: "custom subflow", subflow: true},
		{name: "flow read failure stops before PUT", readFailure: true},
		{name: "subflow read failure stops before PUT", subflow: true, readFailure: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			metadataWrites, executionWrites := 0, 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch {
				case r.Method == "GET" && r.URL.Path == "/admin/realms/realm/authentication/flows/flow":
					if tc.readFailure {
						w.WriteHeader(http.StatusNotFound)
						fmt.Fprint(w, `{"error":"missing"}`)
						return
					}
					json.NewEncoder(w).Encode(&AuthenticationFlow{Id: "flow", Alias: "before", ProviderId: "basic-flow", BuiltIn: tc.builtIn, TopLevel: tc.topLevel})
				case r.Method == "PUT" && r.URL.Path == "/admin/realms/realm/authentication/flows/flow":
					metadataWrites++
					var body AuthenticationFlow
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Error(err)
					}
					if body.Id != "flow" || body.Alias != "desired" || body.Description != "captured description" || body.ProviderId != "basic-flow" || body.BuiltIn != tc.builtIn || body.TopLevel != tc.topLevel {
						t.Errorf("unexpected metadata payload: %#v", body)
					}
					w.WriteHeader(http.StatusNoContent)
				case r.Method == "GET" && r.URL.Path == "/admin/realms/realm/authentication/flows/parent/executions":
					fmt.Fprint(w, `[{"id":"execution","flowId":"flow","level":0}]`)
				case r.Method == "PUT" && r.URL.Path == "/admin/realms/realm/authentication/flows/parent/executions":
					executionWrites++
					var body authenticationExecutionRequirementUpdate
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Error(err)
					}
					if body.Id != "execution" || body.Requirement != "REQUIRED" || body.Priority != 4 {
						t.Errorf("execution settings changed: %#v", body)
					}
					w.WriteHeader(http.StatusNoContent)
				default:
					t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()
			client := &KeycloakClient{baseUrl: server.URL, httpClient: server.Client(), initialLogin: true, clientCredentials: &ClientCredentials{AccessToken: "fixture", TokenType: "Bearer"}}
			var err error
			if tc.subflow {
				desired := &AuthenticationSubFlow{Id: "flow", RealmId: "realm", ParentFlowAlias: "parent", Alias: "desired", Description: "captured description", ProviderId: "basic-flow", BuiltIn: !tc.builtIn, TopLevel: !tc.topLevel, Requirement: "REQUIRED", Priority: 4}
				before := *desired
				err = client.UpdateAuthenticationSubFlow(context.Background(), desired)
				if !reflect.DeepEqual(*desired, before) {
					t.Fatal("caller subflow was mutated")
				}
			} else {
				desired := &AuthenticationFlow{Id: "flow", RealmId: "realm", Alias: "desired", Description: "captured description", ProviderId: "basic-flow", BuiltIn: !tc.builtIn, TopLevel: !tc.topLevel}
				before := *desired
				err = client.UpdateAuthenticationFlow(context.Background(), desired)
				if !reflect.DeepEqual(*desired, before) {
					t.Fatal("caller flow was mutated")
				}
			}
			if tc.readFailure {
				if err == nil || metadataWrites != 0 || executionWrites != 0 {
					t.Fatalf("failed read must stop all writes: %v, %d/%d", err, metadataWrites, executionWrites)
				}
			} else if err != nil || metadataWrites != 1 || executionWrites != map[bool]int{true: 1, false: 0}[tc.subflow] {
				t.Fatalf("update = %v, writes = %d/%d", err, metadataWrites, executionWrites)
			}
		})
	}
}
