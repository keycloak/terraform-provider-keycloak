package keycloak

import (
	"context"
	"fmt"
	"strings"
)

type Group struct {
	Id          string              `json:"id,omitempty"`
	RealmId     string              `json:"-"`
	ParentId    string              `json:"-"`
	Name        string              `json:"name"`
	Path        string              `json:"path,omitempty"`
	SubGroups   []*Group            `json:"subGroups,omitempty"`
	RealmRoles  []string            `json:"realmRoles,omitempty"`
	ClientRoles map[string][]string `json:"clientRoles,omitempty"`
	Attributes  map[string][]string `json:"attributes"`
}

/*
 * There is no way to get a subgroup's parent ID using the Keycloak API (that I know of, PRs are welcome)
 * The best we can do is check subGroup's path with the group's path to figure out what sub-path to follow
 * until we find it.
 */
func (keycloakClient *KeycloakClient) groupParentId(ctx context.Context, group *Group) (string, error) {
	// Check the path of the group being passed in.
	// If there is only one group in the path, then this is a top-level group with no parentId
	if group.Path == "/"+group.Name {
		return "", nil
	}

	groups, err := keycloakClient.ListGroupsWithName(ctx, group.RealmId, group.Name)
	if err != nil {
		return "", err
	}

	var parentGroup Group
	if parentGroupId, found := findParentGroup(*group, groups, parentGroup); found {
		return parentGroupId, nil
	}

	// maybe panic here?  this should never happen
	return "", fmt.Errorf("unable to determine parent ID for group with path %s", group.Path)
}

func findParentGroup(group Group, ingroups []*Group, parentGroup Group) (string, bool) {
	for _, grp := range ingroups {
		if grp.Id == group.Id {
			return parentGroup.Id, true
		}
		if strings.HasPrefix(group.Path, grp.Path+"/") {

			if parentGroupId, found := findParentGroup(group, grp.SubGroups, *grp); found {
				return parentGroupId, found
			}
		}
	}
	return "", false
}

func (keycloakClient *KeycloakClient) ValidateGroupMembers(usernames []interface{}) error {
	for _, username := range usernames {
		if username.(string) != strings.ToLower(username.(string)) {
			return fmt.Errorf("expected all usernames within group membership to be lowercase")
		}
	}

	return nil
}

// NewGroup creates a new group based on the following rules:
// Top level groups are created via POST /realms/${realm_id}/groups
// Child groups are created via POST /realms/${realm_id}/groups/${parent_id}/children
func (keycloakClient *KeycloakClient) NewGroup(ctx context.Context, group *Group) error {
	var createGroupUrl string

	if group.ParentId == "" {
		createGroupUrl = fmt.Sprintf("/realms/%s/groups", group.RealmId)
	} else {
		createGroupUrl = fmt.Sprintf("/realms/%s/groups/%s/children", group.RealmId, group.ParentId)
	}

	_, location, err := keycloakClient.post(ctx, createGroupUrl, group)
	if err != nil {
		return err
	}

	group.Id = getIdFromLocationHeader(location)

	return nil
}

func (keycloakClient *KeycloakClient) GetGroups(ctx context.Context, realmId string) ([]*Group, error) {
	var groups []*Group

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/groups", realmId), &groups, nil)
	if err != nil {
		return nil, err
	}

	for _, group := range groups {
		group.RealmId = realmId
	}

	return groups, nil
}

func (keycloakClient *KeycloakClient) GetGroup(ctx context.Context, realmId, id string) (*Group, error) {
	var group Group

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/groups/%s", realmId, id), &group, nil)
	if err != nil {
		return nil, err
	}

	group.RealmId = realmId // it's important to set RealmId here because fetching the ParentId depends on it

	parentId, err := keycloakClient.groupParentId(ctx, &group)
	if err != nil {
		return nil, err
	}

	group.ParentId = parentId

	return &group, nil
}

func (keycloakClient *KeycloakClient) GetGroupByName(ctx context.Context, realmId, name string) (*Group, error) {
	var groups []Group

	// We can't get a group by name, so we have to search for it
	params := map[string]string{
		"search": name,
	}

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/groups", realmId), &groups, params)
	if err != nil {
		return nil, err
	}

	if len(groups) == 0 {
		return nil, fmt.Errorf("no group with name " + name + " found")
	}

	// The search may return more than 1 result even if there is a group exactly matching the search string
	groupsPtr := make([]*Group, len(groups))
	for i := range groups {
		groupsPtr[i] = &groups[i]
	}
	group := getGroupByDFS(name, groupsPtr)
	if group != nil {
		group.RealmId = realmId // it's important to set RealmId here because fetching the ParentId depends on it

		parentId, err := keycloakClient.groupParentId(ctx, group)
		if err != nil {
			return nil, err
		}

		group.ParentId = parentId

		return group, nil
	}

	return nil, fmt.Errorf("no group with name " + name + " found")
}

/*
Find group by name in groups returned by /groups?search=${group_name}
If there are multiple groups match the name, it will return the first one it found, using DFS algorithm
*/
func getGroupByDFS(groupName string, groups []*Group) *Group {
	for _, group := range groups {
		if groupName == group.Name {
			return group
		}
		groupFound := getGroupByDFS(groupName, group.SubGroups)
		if groupFound != nil {
			return groupFound
		}
	}
	return nil
}

func (keycloakClient *KeycloakClient) UpdateGroup(ctx context.Context, group *Group) error {
	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/groups/%s", group.RealmId, group.Id), group)
}

func (keycloakClient *KeycloakClient) DeleteGroup(ctx context.Context, realmId, id string) error {
	return keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/groups/%s", realmId, id), nil)
}

func (keycloakClient *KeycloakClient) ListGroupsWithName(ctx context.Context, realmId, name string) ([]*Group, error) {
	var groups []*Group

	params := map[string]string{
		"search": name,
	}

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/groups", realmId), &groups, params)
	if err != nil {
		return nil, err
	}

	return groups, nil
}

func (keycloakClient *KeycloakClient) GetGroupMembers(ctx context.Context, realmId, groupId string) ([]*User, error) {
	var users []*User
	var first, pagination = 0, 50
	var iterationUsers []*User

	for ok := true; ok; ok = len(iterationUsers) > 0 {
		iterationUsers = nil
		err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/groups/%s/members?max=%d&first=%d", realmId, groupId, pagination, first), &iterationUsers, nil)
		if err != nil {
			return nil, err
		}
		users = append(users, iterationUsers...)
		first += pagination
	}

	for _, user := range users {
		user.RealmId = realmId
	}

	return users, nil
}

func defaultGroupURL(realmName, groupId string) string {
	return fmt.Sprintf("/realms/%s/default-groups/%s", realmName, groupId)
}

// PutDefaultGroup will PUT a new group ID to the realm default groups. This is effectively
// an "upsert".
func (keycloakClient *KeycloakClient) PutDefaultGroup(ctx context.Context, realmName, groupId string) error {
	url := defaultGroupURL(realmName, groupId)
	return keycloakClient.put(ctx, url, nil)
}

// DeleteDefaultGroup deletes a group ID from the realm default groups.
func (keycloakClient *KeycloakClient) DeleteDefaultGroup(ctx context.Context, realmName, groupId string) error {
	url := defaultGroupURL(realmName, groupId)
	return keycloakClient.delete(ctx, url, nil)
}

// GetDefaultGroups returns all the default groups for a realm.
func (keycloakClient *KeycloakClient) GetDefaultGroups(ctx context.Context, realmName string) ([]Group, error) {
	url := fmt.Sprintf("/realms/%s/default-groups", realmName)

	var defaultGroups []Group
	err := keycloakClient.get(ctx, url, &defaultGroups, nil)

	return defaultGroups, err
}
