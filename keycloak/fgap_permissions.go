package keycloak

import (
	"context"
	"fmt"
)

// FindFGAPv2PermissionByName finds a permission on the admin-permissions resource
// server with exactly the given name. Returns nil if not found.
// The API does a prefix match on name, so we filter client-side for exact equality.
func (keycloakClient *KeycloakClient) FindFGAPv2PermissionByName(ctx context.Context, realmId, apClientId, name string) (*OpenidClientAuthorizationPermission, error) {
	var results []OpenidClientAuthorizationPermission
	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/clients/%s/authz/resource-server/permission", realmId, apClientId), &results, map[string]string{
		"name": name,
	})
	if err != nil {
		return nil, err
	}
	for _, p := range results {
		if p.Name == name && p.Id != "" {
			return keycloakClient.GetOpenidClientAuthorizationPermission(ctx, realmId, apClientId, p.Id)
		}
	}
	return nil, nil
}

// FindFGAPv2PermissionByResourceAndScope returns the first permission on the
// admin-permissions resource server that covers the given resource and scope.
// Returns nil if not found. Used during import when permission names are unknown.
func (keycloakClient *KeycloakClient) FindFGAPv2PermissionByResourceAndScope(ctx context.Context, realmId, apClientId, resourceId, scope string) (*OpenidClientAuthorizationPermission, error) {
	var results []OpenidClientAuthorizationPermission
	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/clients/%s/authz/resource-server/permission", realmId, apClientId), &results, map[string]string{
		"resource": resourceId,
		"scope":    scope,
	})
	if err != nil {
		return nil, err
	}
	for _, p := range results {
		if p.Id != "" {
			return keycloakClient.GetOpenidClientAuthorizationPermission(ctx, realmId, apClientId, p.Id)
		}
	}
	return nil, nil
}

// GetFGAPv2Permission reads a single FGAPv2 scope permission and returns it
// with scope NAMES (not internal scope UUIDs) and policies populated.
// Resources are intentionally not populated: Keycloak's /resources sub-endpoint
// returns internal authorization resource IDs which differ from the entity UUIDs
// (role/group/client IDs) used when creating the permission. Entity IDs are
// therefore kept in Terraform state as configured rather than re-read from the API.
func (keycloakClient *KeycloakClient) GetFGAPv2Permission(ctx context.Context, realmId, apClientId, permId string) (*OpenidClientAuthorizationPermission, error) {
	perm := &OpenidClientAuthorizationPermission{
		RealmId:          realmId,
		ResourceServerId: apClientId,
	}
	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/clients/%s/authz/resource-server/permission/%s", realmId, apClientId, permId), perm, nil)
	if err != nil {
		return nil, err
	}
	perm.RealmId = realmId
	perm.ResourceServerId = apClientId

	var policies []OpenidClientAuthorizationPolicy
	err = keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/clients/%s/authz/resource-server/policy/%s/associatedPolicies", realmId, apClientId, permId), &policies, nil)
	if err != nil {
		return nil, err
	}
	perm.Policies = nil
	for _, p := range policies {
		perm.Policies = append(perm.Policies, p.Id)
	}

	var scopes []OpenidClientAuthorizationScope
	err = keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/clients/%s/authz/resource-server/permission/%s/scopes", realmId, apClientId, permId), &scopes, nil)
	if err != nil {
		return nil, err
	}
	perm.Scopes = nil
	for _, s := range scopes {
		perm.Scopes = append(perm.Scopes, s.Name)
	}

	// Entity UUIDs cannot be recovered from the /resources endpoint.
	perm.Resources = nil

	return perm, nil
}

// GetFGAPv2TypeResourceId returns the ID of the type-level resource for a given
// resource type (e.g. "Users") on the admin-permissions client. These resources
// are pre-created by Keycloak when admin permissions are enabled and are used for
// realm-wide permission management.
func (keycloakClient *KeycloakClient) GetFGAPv2TypeResourceId(ctx context.Context, realmId, apClientId, resourceType string) (string, error) {
	var resources []OpenidClientAuthorizationResource
	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/clients/%s/authz/resource-server/resource", realmId, apClientId), &resources, map[string]string{
		"type": resourceType,
	})
	if err != nil {
		return "", err
	}
	for _, r := range resources {
		if r.Name == resourceType {
			return r.Id, nil
		}
	}
	return "", fmt.Errorf("no type-level resource found for type %q on admin-permissions client %s", resourceType, apClientId)
}
