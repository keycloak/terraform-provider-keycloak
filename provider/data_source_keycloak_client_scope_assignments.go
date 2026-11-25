package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func dataSourceKeycloakClientScopeAssignments() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKeycloakClientScopeAssignmentsRead,
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"client_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"default_scopes": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
				Set:      schema.HashString,
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

func dataSourceKeycloakClientScopeAssignmentsRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)

	assignments, err := readClientScopeAssignments(ctx, keycloakClient, realmId, clientId)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := data.Set("default_scopes", assignments.defaultScopes); err != nil {
		return diag.FromErr(err)
	}
	if err := data.Set("optional_scopes", assignments.optionalScopes); err != nil {
		return diag.FromErr(err)
	}
	data.SetId(clientScopeAssignmentsId(realmId, clientId))

	return nil
}
