package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakOpenidOptionalClientScope() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakOpenidOptionalClientScopeCreate,
		ReadContext:   resourceKeycloakOpenidOptionalClientScopesRead,
		DeleteContext: resourceKeycloakOpenidOptionalClientScopeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceKeycloakOpenidOptionalClientScopeImport,
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

func parseKeycloakOpenidOptionalClientScopeDataIdtoRealmIdClientId(id string) (string, string, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("Invalid import. Supported import formats: {{realmId}}/{{openidClientScopeId}}")
	}
	return parts[0], parts[1], nil
}

func parseKeycloakOpenidOptionalClientScopeRealmIdClientIdToDataId(realmId string, clientScopeId string) string {
	return fmt.Sprintf("%s/%s", realmId, clientScopeId)
}

func resourceKeycloakOpenidOptionalClientScopeCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientScopeId := data.Get("client_scope_id").(string)

	err := keycloakClient.PutOpenidRealmOptionalClientScope(ctx, realmId, clientScopeId)
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(parseKeycloakOpenidOptionalClientScopeRealmIdClientIdToDataId(realmId, clientScopeId))
	return nil
}

func resourceKeycloakOpenidOptionalClientScopesRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientScopeId := data.Get("client_scope_id").(string)

	clientScope, err := keycloakClient.GetOpenidRealmOptionalClientScope(ctx, realmId, clientScopeId)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	data.SetId(parseKeycloakOpenidOptionalClientScopeRealmIdClientIdToDataId(realmId, clientScopeId))
	data.Set("realm_id", realmId)
	data.Set("client_scope_id", clientScope.Id)

	return nil
}

func resourceKeycloakOpenidOptionalClientScopeDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientScopeId := data.Get("client_scope_id").(string)

	return diag.FromErr(keycloakClient.DeleteOpenidRealmOptionalClientScope(ctx, realmId, clientScopeId))
}

func resourceKeycloakOpenidOptionalClientScopeImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	realmId, clientScopeId, err := parseKeycloakOpenidOptionalClientScopeDataIdtoRealmIdClientId(d.Id())
	if err != nil {
		return nil, err
	}

	d.Set("realm_id", realmId)
	d.Set("client_scope_id", clientScopeId)
	d.SetId(d.Id())

	return []*schema.ResourceData{d}, nil
}
