package keycloak

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/go-version"
)

// urlCapturingClient returns a KeycloakClient backed by a mock HTTP server that
// records the path of every request it receives. The provided handler is
// responsible for writing the response body. This is used to assert that the
// client methods build the correct (organization-scoped or realm-level) URLs
// without requiring a live Keycloak instance.
func urlCapturingClient(t *testing.T, handler http.HandlerFunc) (*KeycloakClient, func() []string) {
	t.Helper()

	var paths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		handler(w, r)
	}))
	t.Cleanup(server.Close)

	client := &KeycloakClient{
		baseUrl:           server.URL,
		httpClient:        server.Client(),
		initialLogin:      true,
		version:           version.Must(version.NewVersion("25.0.0")),
		clientCredentials: &ClientCredentials{AccessToken: "test", TokenType: "Bearer"},
	}

	return client, func() []string { return paths }
}

func TestGetOrganizationGroupMembersURL(t *testing.T) {
	for _, tc := range []struct {
		name           string
		realmId        string
		organizationId string
		groupId        string
		expectedPath   string
	}{
		{
			name:           "realm-level",
			realmId:        "my-realm",
			organizationId: "",
			groupId:        "group-id",
			expectedPath:   "/admin/realms/my-realm/groups/group-id/members",
		},
		{
			name:           "organization-scoped",
			realmId:        "my-realm",
			organizationId: "org-id",
			groupId:        "group-id",
			expectedPath:   "/admin/realms/my-realm/organizations/org-id/groups/group-id/members",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var offsets []string
			client, getPaths := urlCapturingClient(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.Query().Get("max") != "50" {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL)
				}
				first := r.URL.Query().Get("first")
				offsets = append(offsets, first)
				w.Header().Set("Content-Type", "application/json")
				if first == "0" {
					_, _ = w.Write([]byte(`[{"id":"user-1","username":"alice"}]`))
				} else if first == "50" {
					_, _ = w.Write([]byte(`[{"id":"user-2","username":"bob"}]`))
				} else {
					_, _ = w.Write([]byte("[]"))
				}
			})

			users, err := client.GetOrganizationGroupMembers(context.Background(), tc.realmId, tc.organizationId, tc.groupId)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if fmt.Sprint(offsets) != "[0 50 100]" {
				t.Fatalf("unexpected pagination offsets: %v", offsets)
			}
			if len(users) != 2 || users[0].Id != "user-1" || users[1].Id != "user-2" {
				t.Fatalf("unexpected members: %v", users)
			}
			for _, user := range users {
				if user.RealmId != tc.realmId {
					t.Errorf("unexpected realm: %s", user.RealmId)
				}
			}

			paths := getPaths()
			if len(paths) == 0 {
				t.Fatal("expected at least one request to be made")
			}
			if paths[len(paths)-1] != tc.expectedPath {
				t.Fatalf("expected request path %q, got %q", tc.expectedPath, paths[len(paths)-1])
			}
		})
	}
}
