package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func dataSourceKeycloakRealmOptionalClientScopes() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKeycloakRealmOptionalClientScopesRead,
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"optional_scopes": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
				Set:      schema.HashString,
			},
		},
	}
}

func dataSourceKeycloakRealmOptionalClientScopesRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)

	clientScopes, err := keycloakClient.GetRealmOptionalClientScopes(ctx, realmId)
	if err != nil {
		return diag.FromErr(err)
	}

	scopeNames := make([]string, 0, len(clientScopes))
	for _, clientScope := range clientScopes {
		scopeNames = append(scopeNames, clientScope.Name)
	}

	data.Set("optional_scopes", scopeNames)
	data.SetId(realmId)

	return nil
}
