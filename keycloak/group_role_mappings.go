package keycloak

import (
	"context"
	"fmt"
)

func groupRoleMappingsUrl(realmId, organizationId, groupId string) string {
	if organizationId != "" {
		return fmt.Sprintf("/realms/%s/organizations/%s/groups/%s/role-mappings", realmId, organizationId, groupId)
	}
	return fmt.Sprintf("/realms/%s/groups/%s/role-mappings", realmId, groupId)
}

func (keycloakClient *KeycloakClient) GetGroupRoleMappings(ctx context.Context, realmId string, groupId string) (*RoleMapping, error) {
	return keycloakClient.GetOrganizationGroupRoleMappings(ctx, realmId, "", groupId)
}

func (keycloakClient *KeycloakClient) GetOrganizationGroupRoleMappings(ctx context.Context, realmId, organizationId, groupId string) (*RoleMapping, error) {
	var roleMapping *RoleMapping
	err := keycloakClient.get(ctx, groupRoleMappingsUrl(realmId, organizationId, groupId), &roleMapping, nil)
	if err != nil {
		return nil, err
	}

	return roleMapping, nil
}

func (keycloakClient *KeycloakClient) AddRealmRolesToGroup(ctx context.Context, realmId, groupId string, roles []*Role) error {
	return keycloakClient.AddRealmRolesToOrganizationGroup(ctx, realmId, "", groupId, roles)
}

func (keycloakClient *KeycloakClient) AddRealmRolesToOrganizationGroup(ctx context.Context, realmId, organizationId, groupId string, roles []*Role) error {
	_, _, err := keycloakClient.post(ctx, fmt.Sprintf("%s/realm", groupRoleMappingsUrl(realmId, organizationId, groupId)), roles)
	return err
}

func (keycloakClient *KeycloakClient) AddClientRolesToGroup(ctx context.Context, realmId, groupId, clientId string, roles []*Role) error {
	return keycloakClient.AddClientRolesToOrganizationGroup(ctx, realmId, "", groupId, clientId, roles)
}

func (keycloakClient *KeycloakClient) AddClientRolesToOrganizationGroup(ctx context.Context, realmId, organizationId, groupId, clientId string, roles []*Role) error {
	_, _, err := keycloakClient.post(ctx, fmt.Sprintf("%s/clients/%s", groupRoleMappingsUrl(realmId, organizationId, groupId), clientId), roles)
	return err
}

func (keycloakClient *KeycloakClient) RemoveRealmRolesFromGroup(ctx context.Context, realmId, groupId string, roles []*Role) error {
	return keycloakClient.RemoveRealmRolesFromOrganizationGroup(ctx, realmId, "", groupId, roles)
}

func (keycloakClient *KeycloakClient) RemoveRealmRolesFromOrganizationGroup(ctx context.Context, realmId, organizationId, groupId string, roles []*Role) error {
	err := keycloakClient.delete(ctx, fmt.Sprintf("%s/realm", groupRoleMappingsUrl(realmId, organizationId, groupId)), roles)
	return err
}

func (keycloakClient *KeycloakClient) RemoveClientRolesFromGroup(ctx context.Context, realmId, groupId, clientId string, roles []*Role) error {
	return keycloakClient.RemoveClientRolesFromOrganizationGroup(ctx, realmId, "", groupId, clientId, roles)
}

func (keycloakClient *KeycloakClient) RemoveClientRolesFromOrganizationGroup(ctx context.Context, realmId, organizationId, groupId, clientId string, roles []*Role) error {
	err := keycloakClient.delete(ctx, fmt.Sprintf("%s/clients/%s", groupRoleMappingsUrl(realmId, organizationId, groupId), clientId), roles)
	return err
}
