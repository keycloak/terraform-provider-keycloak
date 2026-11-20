package keycloak

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestOrganizationGroupRoleMappingsURL(t *testing.T) {
	for _, tc := range []struct {
		name           string
		realmId        string
		organizationId string
		groupId        string
		clientId       string
		call           func(client *KeycloakClient, realmId, organizationId, groupId, clientId string) error
		expectedPath   string
	}{
		{
			name:           "get realm-level",
			realmId:        "my-realm",
			organizationId: "",
			groupId:        "group-id",
			call: func(client *KeycloakClient, realmId, organizationId, groupId, clientId string) error {
				_, err := client.GetOrganizationGroupRoleMappings(context.Background(), realmId, organizationId, groupId)
				return err
			},
			expectedPath: "/admin/realms/my-realm/groups/group-id/role-mappings",
		},
		{
			name:           "get organization-scoped",
			realmId:        "my-realm",
			organizationId: "org-id",
			groupId:        "group-id",
			call: func(client *KeycloakClient, realmId, organizationId, groupId, clientId string) error {
				_, err := client.GetOrganizationGroupRoleMappings(context.Background(), realmId, organizationId, groupId)
				return err
			},
			expectedPath: "/admin/realms/my-realm/organizations/org-id/groups/group-id/role-mappings",
		},
		{
			name:           "add realm roles organization-scoped",
			realmId:        "my-realm",
			organizationId: "org-id",
			groupId:        "group-id",
			call: func(client *KeycloakClient, realmId, organizationId, groupId, clientId string) error {
				return client.AddRealmRolesToOrganizationGroup(context.Background(), realmId, organizationId, groupId, []*Role{})
			},
			expectedPath: "/admin/realms/my-realm/organizations/org-id/groups/group-id/role-mappings/realm",
		},
		{
			name:           "add client roles organization-scoped",
			realmId:        "my-realm",
			organizationId: "org-id",
			groupId:        "group-id",
			clientId:       "client-id",
			call: func(client *KeycloakClient, realmId, organizationId, groupId, clientId string) error {
				return client.AddClientRolesToOrganizationGroup(context.Background(), realmId, organizationId, groupId, clientId, []*Role{})
			},
			expectedPath: "/admin/realms/my-realm/organizations/org-id/groups/group-id/role-mappings/clients/client-id",
		},
		{
			name:           "remove realm roles organization-scoped",
			realmId:        "my-realm",
			organizationId: "org-id",
			groupId:        "group-id",
			call: func(client *KeycloakClient, realmId, organizationId, groupId, clientId string) error {
				return client.RemoveRealmRolesFromOrganizationGroup(context.Background(), realmId, organizationId, groupId, []*Role{})
			},
			expectedPath: "/admin/realms/my-realm/organizations/org-id/groups/group-id/role-mappings/realm",
		},
		{
			name:           "remove client roles organization-scoped",
			realmId:        "my-realm",
			organizationId: "org-id",
			groupId:        "group-id",
			clientId:       "client-id",
			call: func(client *KeycloakClient, realmId, organizationId, groupId, clientId string) error {
				return client.RemoveClientRolesFromOrganizationGroup(context.Background(), realmId, organizationId, groupId, clientId, []*Role{})
			},
			expectedPath: "/admin/realms/my-realm/organizations/org-id/groups/group-id/role-mappings/clients/client-id",
		},
		{
			name:           "add realm roles realm-level",
			realmId:        "my-realm",
			organizationId: "",
			groupId:        "group-id",
			call: func(client *KeycloakClient, realmId, organizationId, groupId, clientId string) error {
				return client.AddRealmRolesToOrganizationGroup(context.Background(), realmId, organizationId, groupId, []*Role{})
			},
			expectedPath: "/admin/realms/my-realm/groups/group-id/role-mappings/realm",
		},
		{
			name:           "add client roles realm-level",
			realmId:        "my-realm",
			organizationId: "",
			groupId:        "group-id",
			clientId:       "client-id",
			call: func(client *KeycloakClient, realmId, organizationId, groupId, clientId string) error {
				return client.AddClientRolesToOrganizationGroup(context.Background(), realmId, organizationId, groupId, clientId, []*Role{})
			},
			expectedPath: "/admin/realms/my-realm/groups/group-id/role-mappings/clients/client-id",
		},
		{
			name:           "remove realm roles realm-level",
			realmId:        "my-realm",
			organizationId: "",
			groupId:        "group-id",
			call: func(client *KeycloakClient, realmId, organizationId, groupId, clientId string) error {
				return client.RemoveRealmRolesFromOrganizationGroup(context.Background(), realmId, organizationId, groupId, []*Role{})
			},
			expectedPath: "/admin/realms/my-realm/groups/group-id/role-mappings/realm",
		},
		{
			name:           "remove client roles realm-level",
			realmId:        "my-realm",
			organizationId: "",
			groupId:        "group-id",
			clientId:       "client-id",
			call: func(client *KeycloakClient, realmId, organizationId, groupId, clientId string) error {
				return client.RemoveClientRolesFromOrganizationGroup(context.Background(), realmId, organizationId, groupId, clientId, []*Role{})
			},
			expectedPath: "/admin/realms/my-realm/groups/group-id/role-mappings/clients/client-id",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client, getPaths := urlCapturingClient(t, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte("{}"))
			})

			if err := tc.call(client, tc.realmId, tc.organizationId, tc.groupId, tc.clientId); err != nil {
				t.Fatalf("unexpected error: %v", err)
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

func TestGroupRoleMappingsDelegation(t *testing.T) {
	roles := []*Role{{Id: "role-id", Name: "role"}}
	for _, tc := range []struct {
		name, method, suffix string
		call                 func(*KeycloakClient) error
	}{
		{"get", http.MethodGet, "", func(c *KeycloakClient) error {
			_, err := c.GetGroupRoleMappings(context.Background(), "realm", "group")
			return err
		}},
		{"add realm", http.MethodPost, "/realm", func(c *KeycloakClient) error {
			return c.AddRealmRolesToGroup(context.Background(), "realm", "group", roles)
		}},
		{"add client", http.MethodPost, "/clients/client", func(c *KeycloakClient) error {
			return c.AddClientRolesToGroup(context.Background(), "realm", "group", "client", roles)
		}},
		{"remove realm", http.MethodDelete, "/realm", func(c *KeycloakClient) error {
			return c.RemoveRealmRolesFromGroup(context.Background(), "realm", "group", roles)
		}},
		{"remove client", http.MethodDelete, "/clients/client", func(c *KeycloakClient) error {
			return c.RemoveClientRolesFromGroup(context.Background(), "realm", "group", "client", roles)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client, paths := urlCapturingClient(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != tc.method {
					t.Errorf("method = %s, want %s", r.Method, tc.method)
				}
				if tc.method != http.MethodGet {
					var body []*Role
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Fatal(err)
					}
					if len(body) != 1 || body[0].Id != roles[0].Id || body[0].Name != roles[0].Name {
						t.Errorf("unexpected role payload: %v", body)
					}
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte("{}"))
			})
			if err := tc.call(client); err != nil {
				t.Fatal(err)
			}
			got := paths()
			want := "/admin/realms/realm/groups/group/role-mappings" + tc.suffix
			if len(got) != 1 || got[0] != want {
				t.Fatalf("paths = %v, want [%s]", got, want)
			}
		})
	}
}
