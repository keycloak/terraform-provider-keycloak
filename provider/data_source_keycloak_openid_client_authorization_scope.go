package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func dataSourceKeycloakOpenidClientAuthorizationScope() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKeycloakOpenidClientAuthorizationScopeRead,
		Schema: map[string]*schema.Schema{
			"resource_server_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"display_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"icon_uri": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceKeycloakOpenidClientAuthorizationScopeRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	resourceServerId := data.Get("resource_server_id").(string)
	name := data.Get("name").(string)

	scope, err := keycloakClient.GetOpenidClientAuthorizationScopeByName(ctx, realmId, resourceServerId, name)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	data.SetId(scope.Id)
	data.Set("resource_server_id", scope.ResourceServerId)
	data.Set("realm_id", scope.RealmId)
	data.Set("name", scope.Name)
	data.Set("display_name", scope.DisplayName)
	data.Set("icon_uri", scope.IconUri)

	return nil
}
