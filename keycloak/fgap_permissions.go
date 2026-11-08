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

// GetFGAPv2Permission reads a single FGAPv2 scope permission and returns it
// with scope NAMES (not internal scope UUIDs), policies and resources populated.
// Resources come back as the authorization resources' NAMES: Keycloak names a
// permission's per-entity resource after the entity's id
// (AdminPermissionsSchema.getOrCreateResource), and a type-wide permission's
// resource after the resource type ("Users", "Groups", ...), so a caller gets
// the entity ids it configured plus, for a type-wide permission, the type's
// name; the internal resource UUIDs are not returned.
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

	var resources []OpenidClientAuthorizationResource
	err = keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/clients/%s/authz/resource-server/permission/%s/resources", realmId, apClientId, permId), &resources, nil)
	if err != nil {
		return nil, err
	}
	perm.Resources = nil
	for _, r := range resources {
		perm.Resources = append(perm.Resources, r.Name)
	}

	return perm, nil
}
