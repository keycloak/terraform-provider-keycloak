package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakRealmOptionalClientScopes() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakRealmOptionalClientScopesReconcile,
		ReadContext:   resourceKeycloakRealmOptionalClientScopesRead,
		DeleteContext: resourceKeycloakRealmOptionalClientScopesDelete,
		UpdateContext: resourceKeycloakRealmOptionalClientScopesReconcile,
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"optional_scopes": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Required: true,
				Set:      schema.HashString,
			},
		},
	}
}

func resourceKeycloakRealmOptionalClientScopesRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)

	optionalClientScopes, err := keycloakClient.GetRealmOptionalClientScopes(ctx, realmId)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	var scopeNames []string
	for _, clientScope := range optionalClientScopes {
		scopeNames = append(scopeNames, clientScope.Name)
	}

	data.Set("optional_scopes", scopeNames)
	data.SetId(realmId)

	return nil
}

func resourceKeycloakRealmOptionalClientScopesReconcile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	tfOptionalClientScopes := data.Get("optional_scopes").(*schema.Set)

	keycloakOptionalClientScopes, err := keycloakClient.GetRealmOptionalClientScopes(ctx, realmId)
	if err != nil {
		return diag.FromErr(err)
	}

	var scopesToUnmark []string
	for _, keycloakOptionalClientScope := range keycloakOptionalClientScopes {
		// if this scope is an optional client scope in keycloak and tf state, no update is required
		if tfOptionalClientScopes.Contains(keycloakOptionalClientScope.Name) {
			tfOptionalClientScopes.Remove(keycloakOptionalClientScope.Name)
		} else {
			// if this scope is marked as optional in keycloak but not in tf state unmark it
			scopesToUnmark = append(scopesToUnmark, keycloakOptionalClientScope.Name)
		}
	}

	// unmark scopes that aren't in tf state
	err = keycloakClient.UnmarkClientScopesAsRealmOptional(ctx, realmId, scopesToUnmark)
	if err != nil {
		return diag.FromErr(err)
	}

	// mark scopes as optional that exist in tf state but not in keycloak
	err = keycloakClient.MarkClientScopesAsRealmOptional(ctx, realmId, interfaceSliceToStringSlice(tfOptionalClientScopes.List()))
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(realmId)

	return resourceKeycloakRealmOptionalClientScopesRead(ctx, data, meta)
}

func resourceKeycloakRealmOptionalClientScopesDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	optionalClientScopes := data.Get("optional_scopes").(*schema.Set)

	return diag.FromErr(keycloakClient.UnmarkClientScopesAsRealmOptional(ctx, realmId, interfaceSliceToStringSlice(optionalClientScopes.List())))
}
