package keycloak

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type Role struct {
	Id          string `json:"id,omitempty"`
	RealmId     string `json:"-"`
	ClientId    string `json:"-"`
	RoleId      string `json:"-"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ClientRole  bool   `json:"clientRole"`
	ContainerId string `json:"containerId"`
	Composite   bool   `json:"composite"`
	//extra attributes of a role
	Attributes map[string][]string `json:"attributes"`
}

type UsersInRole struct {
	Role  *Role
	Users *[]User
}

/*
 * Realm roles: /realms/${realm_id}/roles
 * Client roles: /realms/${realm_id}/clients/${client_id}/roles
 */
func roleByNameUrl(realmId, clientId string) string {
	if clientId == "" {
		return fmt.Sprintf("/realms/%s/roles", realmId)
	}

	return fmt.Sprintf("/realms/%s/clients/%s/roles", realmId, clientId)
}

func (keycloakClient *KeycloakClient) CreateRole(ctx context.Context, role *Role) error {
	roleUrl := roleByNameUrl(role.RealmId, role.ClientId)

	if role.ClientId != "" {
		role.ContainerId = role.ClientId
		role.ClientRole = true
	}

	_, _, err := keycloakClient.post(ctx, roleUrl, role)
	if err != nil {
		return err
	}

	var createdRole Role
	var roleName = url.PathEscape(role.Name)

	err = keycloakClient.get(ctx, fmt.Sprintf("%s/%s", roleUrl, roleName), &createdRole, nil)
	if err != nil {
		return err
	}

	role.Id = createdRole.Id

	// seems like role attributes aren't respected on create, so a following update is needed
	return keycloakClient.UpdateRole(ctx, role)
}

func (keycloakClient *KeycloakClient) GetRealmRoles(ctx context.Context, realmId string) ([]*Role, error) {
	var roles []*Role

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/roles", realmId), &roles, nil)
	if err != nil {
		return nil, err
	}

	for _, role := range roles {
		role.RealmId = realmId
	}

	return roles, nil
}

func (keycloakClient *KeycloakClient) GetClientRoles(ctx context.Context, realmId string, clients []*OpenidClient) ([]*Role, error) {
	var roles []*Role

	for _, client := range clients {
		var rolesClient []*Role

		err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/clients/%s/roles", realmId, client.Id), &rolesClient, nil)
		if err != nil {
			return nil, err
		}

		for _, roleClient := range rolesClient {
			roleClient.RealmId = realmId
			roleClient.ClientId = client.Id
		}

		roles = append(roles, rolesClient...)
	}

	return roles, nil
}

func (keycloakClient *KeycloakClient) GetClientRoleUsers(ctx context.Context, realmId string, roles []*Role) (*[]UsersInRole, error) {
	var usersInRoles []UsersInRole

	for _, role := range roles {
		var usersInRole UsersInRole

		usersInRole.Role = role
		err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/clients/%s/roles/%s/users", realmId, role.ClientId, role.Name), &usersInRole.Users, nil)
		if usersInRole.Users == nil {
			continue
		}
		if err != nil {
			return nil, err
		}

		usersInRoles = append(usersInRoles, usersInRole)
	}

	return &usersInRoles, nil
}

func (keycloakClient *KeycloakClient) GetRole(ctx context.Context, realmId, id string) (*Role, error) {
	var role Role
	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/roles-by-id/%s", realmId, id), &role, nil)
	if err != nil {
		return nil, err
	}

	role.RealmId = realmId

	if role.ClientRole {
		role.ClientId = role.ContainerId
	}

	return &role, nil
}

func (keycloakClient *KeycloakClient) GetRoleByName(ctx context.Context, realmId, clientId, name string) (*Role, error) {
	var role Role
	var roleName = url.PathEscape(name)

	err := keycloakClient.get(ctx, fmt.Sprintf("%s/%s", roleByNameUrl(realmId, clientId), roleName), &role, nil)
	if err != nil {
		return nil, err
	}

	role.RealmId = realmId

	if role.ClientRole {
		role.ClientId = role.ContainerId
	}

	return &role, nil
}

func (keycloakClient *KeycloakClient) UpdateRole(ctx context.Context, role *Role) error {
	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/roles-by-id/%s", role.RealmId, role.Id), role)
}

func (keycloakClient *KeycloakClient) DeleteRole(ctx context.Context, realmId, id string) error {
	err := keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/roles-by-id/%s", realmId, id), nil)
	if err != nil {
		tflog.Debug(ctx, "Failed to delete role, trying again", map[string]interface{}{
			"roleId": id,
		})

		return keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/roles-by-id/%s", realmId, id), nil)
	}

	return nil
}

func (keycloakClient *KeycloakClient) AddCompositesToRole(ctx context.Context, role *Role, compositeRoles []*Role) error {
	_, _, err := keycloakClient.post(ctx, fmt.Sprintf("/realms/%s/roles-by-id/%s/composites", role.RealmId, role.Id), compositeRoles)
	if err != nil {
		return err
	}

	return nil
}

func (keycloakClient *KeycloakClient) RemoveCompositesFromRole(ctx context.Context, role *Role, compositeRoles []*Role) error {
	err := keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/roles-by-id/%s/composites", role.RealmId, role.Id), compositeRoles)
	if err != nil {
		return err
	}

	return nil
}

func (keycloakClient *KeycloakClient) GetRoleComposites(ctx context.Context, role *Role) ([]*Role, error) {
	var composites []*Role

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/roles-by-id/%s/composites", role.RealmId, role.Id), &composites, nil)
	if err != nil {
		return nil, err
	}

	return composites, nil
}
