package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakOpenidDefaultClientScope() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakOpenidDefaultClientScopeCreate,
		ReadContext:   resourceKeycloakOpenidDefaultClientScopesRead,
		DeleteContext: resourceKeycloakOpenidDefaultClientScopeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceKeycloakOpenidDefaultClientScopeImport,
		},

		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"client_scope_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func parseKeycloakOpenidDefaultClientScopeDataIdtoRealmIdClientScopeId(id string) (string, string, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("Invalid import. Supported import formats: {{realmId}}/{{openidClientScopeId}}")
	}
	return parts[0], parts[1], nil
}

func parseKeycloakOpenidDefaultClientScopeRealmIdClientScopeIdToDataId(realmId string, clientScopeId string) string {
	return fmt.Sprintf("%s/%s", realmId, clientScopeId)
}

func resourceKeycloakOpenidDefaultClientScopeCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientScopeId := data.Get("client_scope_id").(string)

	err := keycloakClient.PutOpenidRealmDefaultClientScope(ctx, realmId, clientScopeId)
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(parseKeycloakOpenidDefaultClientScopeRealmIdClientScopeIdToDataId(realmId, clientScopeId))
	return nil
}

func resourceKeycloakOpenidDefaultClientScopesRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientScopeId := data.Get("client_scope_id").(string)

	clientScope, err := keycloakClient.GetOpenidRealmDefaultClientScope(ctx, realmId, clientScopeId)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	data.SetId(parseKeycloakOpenidDefaultClientScopeRealmIdClientScopeIdToDataId(realmId, clientScopeId))
	data.Set("realm_id", realmId)
	data.Set("client_scope_id", clientScope.Id)

	return nil
}

func resourceKeycloakOpenidDefaultClientScopeDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientScopeId := data.Get("client_scope_id").(string)

	return diag.FromErr(keycloakClient.DeleteOpenidRealmDefaultClientScope(ctx, realmId, clientScopeId))
}

func resourceKeycloakOpenidDefaultClientScopeImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	realmId, clientScopeId, err := parseKeycloakOpenidDefaultClientScopeDataIdtoRealmIdClientScopeId(d.Id())
	if err != nil {
		return nil, err
	}

	d.Set("realm_id", realmId)
	d.Set("client_scope_id", clientScopeId)
	d.SetId(d.Id())

	return []*schema.ResourceData{d}, nil
}
