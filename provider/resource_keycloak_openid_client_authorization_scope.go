package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"strings"
)

func resourceKeycloakOpenidClientAuthorizationScope() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakOpenidClientAuthorizationScopeCreate,
		ReadContext:   resourceKeycloakOpenidClientAuthorizationScopeRead,
		DeleteContext: resourceKeycloakOpenidClientAuthorizationScopeDelete,
		UpdateContext: resourceKeycloakOpenidClientAuthorizationScopeUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: resourceKeycloakOpenidClientAuthorizationScopeImport,
		},
		Schema: map[string]*schema.Schema{
			"resource_server_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"display_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"icon_uri": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func getOpenidClientAuthorizationScopeFromData(data *schema.ResourceData) *keycloak.OpenidClientAuthorizationScope {
	scope := keycloak.OpenidClientAuthorizationScope{
		DisplayName:      data.Get("display_name").(string),
		Name:             data.Get("name").(string),
		IconUri:          data.Get("icon_uri").(string),
		Id:               data.Id(),
		ResourceServerId: data.Get("resource_server_id").(string),
		RealmId:          data.Get("realm_id").(string),
	}
	return &scope
}

func setOpenidClientAuthorizationScopeData(data *schema.ResourceData, scope *keycloak.OpenidClientAuthorizationScope) {
	data.SetId(scope.Id)
	data.Set("resource_server_id", scope.ResourceServerId)
	data.Set("realm_id", scope.RealmId)
	data.Set("display_name", scope.DisplayName)
	data.Set("name", scope.Name)
	data.Set("icon_uri", scope.IconUri)
}

func resourceKeycloakOpenidClientAuthorizationScopeCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	scope := getOpenidClientAuthorizationScopeFromData(data)

	err := keycloakClient.NewOpenidClientAuthorizationScope(ctx, scope)
	if err != nil {
		return diag.FromErr(err)
	}

	setOpenidClientAuthorizationScopeData(data, scope)

	return resourceKeycloakOpenidClientAuthorizationScopeRead(ctx, data, meta)
}

func resourceKeycloakOpenidClientAuthorizationScopeRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	resourceServerId := data.Get("resource_server_id").(string)
	id := data.Id()

	scope, err := keycloakClient.GetOpenidClientAuthorizationScope(ctx, realmId, resourceServerId, id)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	setOpenidClientAuthorizationScopeData(data, scope)

	return nil
}

func resourceKeycloakOpenidClientAuthorizationScopeUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	scope := getOpenidClientAuthorizationScopeFromData(data)

	err := keycloakClient.UpdateOpenidClientAuthorizationScope(ctx, scope)
	if err != nil {
		return diag.FromErr(err)
	}

	setOpenidClientAuthorizationScopeData(data, scope)

	return nil
}

func resourceKeycloakOpenidClientAuthorizationScopeDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	resourceServerId := data.Get("resource_server_id").(string)
	id := data.Id()

	return diag.FromErr(keycloakClient.DeleteOpenidClientAuthorizationScope(ctx, realmId, resourceServerId, id))
}

func resourceKeycloakOpenidClientAuthorizationScopeImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 3 {
		return nil, fmt.Errorf("Invalid import. Supported import formats: {{realmId}}/{{resourceServerId}}/{{authorizationScopeId}}")
	}
	d.Set("realm_id", parts[0])
	d.Set("resource_server_id", parts[1])
	d.SetId(parts[2])

	return []*schema.ResourceData{d}, nil
}
