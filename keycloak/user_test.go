package keycloak

import (
	"context"
	"net/http"
	"testing"
)

func TestOrganizationGroupMembershipURL(t *testing.T) {
	for _, tc := range []struct {
		name           string
		realmId        string
		organizationId string
		groupId        string
		userId         string
		call           func(client *KeycloakClient, realmId, organizationId, groupId, userId string) error
		expectedPath   string
	}{
		{
			name:           "add user to realm-level group",
			realmId:        "my-realm",
			organizationId: "",
			groupId:        "group-id",
			userId:         "user-id",
			call: func(client *KeycloakClient, realmId, organizationId, groupId, userId string) error {
				return client.AddUserToOrganizationGroup(context.Background(), realmId, organizationId, groupId, userId)
			},
			expectedPath: "/admin/realms/my-realm/users/user-id/groups/group-id",
		},
		{
			name:           "add user to organization group",
			realmId:        "my-realm",
			organizationId: "org-id",
			groupId:        "group-id",
			userId:         "user-id",
			call: func(client *KeycloakClient, realmId, organizationId, groupId, userId string) error {
				return client.AddUserToOrganizationGroup(context.Background(), realmId, organizationId, groupId, userId)
			},
			expectedPath: "/admin/realms/my-realm/organizations/org-id/groups/group-id/members/user-id",
		},
		{
			name:           "remove user from realm-level group",
			realmId:        "my-realm",
			organizationId: "",
			groupId:        "group-id",
			userId:         "user-id",
			call: func(client *KeycloakClient, realmId, organizationId, groupId, userId string) error {
				return client.RemoveUserFromOrganizationGroup(context.Background(), realmId, organizationId, groupId, userId)
			},
			expectedPath: "/admin/realms/my-realm/users/user-id/groups/group-id",
		},
		{
			name:           "remove user from organization group",
			realmId:        "my-realm",
			organizationId: "org-id",
			groupId:        "group-id",
			userId:         "user-id",
			call: func(client *KeycloakClient, realmId, organizationId, groupId, userId string) error {
				return client.RemoveUserFromOrganizationGroup(context.Background(), realmId, organizationId, groupId, userId)
			},
			expectedPath: "/admin/realms/my-realm/organizations/org-id/groups/group-id/members/user-id",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client, getPaths := urlCapturingClient(t, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte("{}"))
			})

			if err := tc.call(client, tc.realmId, tc.organizationId, tc.groupId, tc.userId); err != nil {
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
