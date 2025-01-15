package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func dataSourceKeycloakAuthenticationFlow() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKeycloakAuthenticationFlowRead,
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"alias": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceKeycloakAuthenticationFlowRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmID := data.Get("realm_id").(string)
	alias := data.Get("alias").(string)

	authenticationFlowInfo, err := keycloakClient.GetAuthenticationFlowFromAlias(ctx, realmID, alias)
	if err != nil {
		return diag.FromErr(err)
	}

	mapFromAuthenticationFlowInfoToData(data, authenticationFlowInfo)

	return nil
}
