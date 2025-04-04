package keycloak

import (
	"context"
	"encoding/json"
	"fmt"
)

type OpenidClientAuthorizationResource struct {
	ResourceServerId   string                           `json:"-"`
	RealmId            string                           `json:"-"`
	Id                 string                           `json:"_id,omitempty"`
	DisplayName        string                           `json:"displayName"`
	Name               string                           `json:"name"`
	Uris               []string                         `json:"uris"`
	IconUri            string                           `json:"icon_uri"`
	OwnerManagedAccess bool                             `json:"ownerManagedAccess"`
	Scopes             []OpenidClientAuthorizationScope `json:"scopes"`
	Type               string                           `json:"type"`
	Attributes         map[string][]string              `json:"attributes"`
}

func (keycloakClient *KeycloakClient) NewOpenidClientAuthorizationResource(ctx context.Context, resource *OpenidClientAuthorizationResource) error {
	body, _, err := keycloakClient.post(ctx, fmt.Sprintf("/realms/%s/clients/%s/authz/resource-server/resource", resource.RealmId, resource.ResourceServerId), resource)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &resource)
	if err != nil {
		return err
	}
	return nil
}

func (keycloakClient *KeycloakClient) GetOpenidClientAuthorizationResource(ctx context.Context, realm, resourceServerId, resourceId string) (*OpenidClientAuthorizationResource, error) {
	resource := OpenidClientAuthorizationResource{
		RealmId:          realm,
		ResourceServerId: resourceServerId,
	}
	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/clients/%s/authz/resource-server/resource/%s", realm, resourceServerId, resourceId), &resource, nil)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

func (keycloakClient *KeycloakClient) GetOpenidClientAuthorizationResourceByName(ctx context.Context, realmId, resourceServerId, name string) (*OpenidClientAuthorizationResource, error) {
	var resources []OpenidClientAuthorizationResource
	params := map[string]string{"name": name}
	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/clients/%s/authz/resource-server/resource", realmId, resourceServerId), &resources, params)
	if err != nil {
		return nil, err
	}
	resource := resources[0]
	resource.RealmId = realmId
	resource.ResourceServerId = resourceServerId
	resource.Name = name
	return &resource, nil
}

func (keycloakClient *KeycloakClient) UpdateOpenidClientAuthorizationResource(ctx context.Context, resource *OpenidClientAuthorizationResource) error {
	err := keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/clients/%s/authz/resource-server/resource/%s", resource.RealmId, resource.ResourceServerId, resource.Id), resource)
	if err != nil {
		return err
	}
	return nil
}

func (keycloakClient *KeycloakClient) DeleteOpenidClientAuthorizationResource(ctx context.Context, realmId, clientId, resourceId string) error {
	return keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/clients/%s/authz/resource-server/resource/%s", realmId, clientId, resourceId), nil)
}
