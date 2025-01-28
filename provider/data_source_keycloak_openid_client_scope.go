package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func dataSourceKeycloakOpenidClientScope() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKeycloakOpenidClientScopeRead,

		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"consent_screen_text": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"include_in_token_scope": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"gui_order": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceKeycloakOpenidClientScopeRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	name := data.Get("name").(string)

	scopes, err := keycloakClient.ListOpenidClientScopesWithFilter(ctx, realmId, keycloak.IncludeOpenidClientScopesMatchingNames([]string{name}))
	if err != nil {
		return diag.FromErr(err)
	}

	if len(scopes) != 1 {
		return diag.Errorf("expected provided client scope name to match 1 scope, but matched %d scopes", len(scopes))
	}

	setOpenidClientScopeData(data, scopes[0])

	return nil
}
