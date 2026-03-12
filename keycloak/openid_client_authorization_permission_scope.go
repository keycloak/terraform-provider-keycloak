package keycloak

import (
	"context"
	"encoding/json"
	"fmt"
)

type OpenidClientAuthorizationPermissionScope struct {
	Id               string   `json:"id,omitempty"`
	RealmId          string   `json:"-"`
	ResourceServerId string   `json:"-"`
	Name             string   `json:"name"`
	Description      string   `json:"description"`
	DecisionStrategy string   `json:"decisionStrategy"`
	Policies         []string `json:"policies"`
	Resources        []string `json:"resources"`
	Scopes           []string `json:"scopes"`
	ResourceType     string   `json:"resourceType"`
}

func (keycloakClient *KeycloakClient) GetOpenidClientAuthorizationPermissionScope(ctx context.Context, realm, resourceServerId, id string) (*OpenidClientAuthorizationPermissionScope, error) {
	permission := OpenidClientAuthorizationPermissionScope{
		RealmId:          realm,
		ResourceServerId: resourceServerId,
		Id:               id,
	}

	var policies []OpenidClientAuthorizationPolicy
	var resources []OpenidClientAuthorizationResource
	var scopes []OpenidClientAuthorizationScope

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/clients/%s/authz/resource-server/permission/%s", realm, resourceServerId, id), &permission, nil)
	if err != nil {
		return nil, err
	}

	err = keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/clients/%s/authz/resource-server/policy/%s/associatedPolicies", realm, resourceServerId, id), &policies, nil)
	if err != nil {
		return nil, err
	}

	err = keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/clients/%s/authz/resource-server/permission/%s/resources", realm, resourceServerId, id), &resources, nil)
	if err != nil {
		return nil, err
	}

	err = keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/clients/%s/authz/resource-server/permission/%s/scopes", realm, resourceServerId, id), &scopes, nil)
	if err != nil {
		return nil, err
	}

	for _, policy := range policies {
		permission.Policies = append(permission.Policies, policy.Id)
	}

	for _, resource := range resources {
		permission.Resources = append(permission.Resources, resource.Id)
	}

	for _, resource := range scopes {
		permission.Scopes = append(permission.Scopes, resource.Id)
	}

	return &permission, nil
}

func (keycloakClient *KeycloakClient) NewOpenidClientAuthorizationPermissionScope(ctx context.Context, permission *OpenidClientAuthorizationPermissionScope) error {
	body, _, err := keycloakClient.post(ctx, fmt.Sprintf("/realms/%s/clients/%s/authz/resource-server/permission/scope", permission.RealmId, permission.ResourceServerId), permission)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &permission)
	if err != nil {
		return err
	}
	return nil
}

func (keycloakClient *KeycloakClient) UpdateOpenidClientAuthorizationPermissionScope(ctx context.Context, permission *OpenidClientAuthorizationPermissionScope) error {
	err := keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/clients/%s/authz/resource-server/permission/scope/%s", permission.RealmId, permission.ResourceServerId, permission.Id), permission)
	if err != nil {
		return err
	}
	return nil
}

func (keycloakClient *KeycloakClient) DeleteOpenidClientAuthorizationPermissionScope(ctx context.Context, realmId, resourceServerId, permissionId string) error {
	return keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/clients/%s/authz/resource-server/permission/%s", realmId, resourceServerId, permissionId), nil)
}
