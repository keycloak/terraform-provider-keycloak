package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakRealmDefaultClientScopes() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakRealmDefaultClientScopesReconcile,
		ReadContext:   resourceKeycloakRealmDefaultClientScopesRead,
		DeleteContext: resourceKeycloakRealmDefaultClientScopesDelete,
		UpdateContext: resourceKeycloakRealmDefaultClientScopesReconcile,
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"default_scopes": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Required: true,
				Set:      schema.HashString,
			},
		},
	}
}

func resourceKeycloakRealmDefaultClientScopesRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)

	defaultClientScopes, err := keycloakClient.GetRealmDefaultClientScopes(ctx, realmId)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	var scopeNames []string
	for _, clientScope := range defaultClientScopes {
		scopeNames = append(scopeNames, clientScope.Name)
	}

	data.Set("default_scopes", scopeNames)
	data.SetId(realmId)

	return nil
}

func resourceKeycloakRealmDefaultClientScopesReconcile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	tfDefaultClientScopes := data.Get("default_scopes").(*schema.Set)

	keycloakDefaultClientScopes, err := keycloakClient.GetRealmDefaultClientScopes(ctx, realmId)
	if err != nil {
		return diag.FromErr(err)
	}

	var scopesToUnmark []string
	for _, keycloakDefaultClientScope := range keycloakDefaultClientScopes {
		// if this scope is a default client scope in keycloak and tf state, no update is required
		if tfDefaultClientScopes.Contains(keycloakDefaultClientScope.Name) {
			tfDefaultClientScopes.Remove(keycloakDefaultClientScope.Name)
		} else {
			// if this scope is marked as default in keycloak but not in tf state unmark it
			scopesToUnmark = append(scopesToUnmark, keycloakDefaultClientScope.Name)
		}
	}

	// unmark scopes that aren't in tf state
	err = keycloakClient.UnmarkClientScopesAsRealmDefault(ctx, realmId, scopesToUnmark)
	if err != nil {
		return diag.FromErr(err)
	}

	// mark scopes as default that exist in tf state but not in keycloak
	err = keycloakClient.MarkClientScopesAsRealmDefault(ctx, realmId, interfaceSliceToStringSlice(tfDefaultClientScopes.List()))
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(realmId)

	return resourceKeycloakRealmDefaultClientScopesRead(ctx, data, meta)
}

func resourceKeycloakRealmDefaultClientScopesDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	defaultClientScopes := data.Get("default_scopes").(*schema.Set)

	return diag.FromErr(keycloakClient.UnmarkClientScopesAsRealmDefault(ctx, realmId, interfaceSliceToStringSlice(defaultClientScopes.List())))
}
