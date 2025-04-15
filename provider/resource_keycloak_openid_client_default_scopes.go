package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakOpenidClientDefaultScopes() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakOpenidClientDefaultScopesReconcile,
		ReadContext:   resourceKeycloakOpenidClientDefaultScopesRead,
		DeleteContext: resourceKeycloakOpenidClientDefaultScopesDelete,
		UpdateContext: resourceKeycloakOpenidClientDefaultScopesReconcile,
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"client_id": {
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

func openidClientDefaultScopesId(realmId string, clientId string) string {
	return fmt.Sprintf("%s/%s", realmId, clientId)
}

func resourceKeycloakOpenidClientDefaultScopesRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)

	clientScopes, err := keycloakClient.GetOpenidClientDefaultScopes(ctx, realmId, clientId)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	var defaultScopes []string
	for _, clientScope := range clientScopes {
		defaultScopes = append(defaultScopes, clientScope.Name)
	}

	data.Set("default_scopes", defaultScopes)
	data.SetId(openidClientDefaultScopesId(realmId, clientId))

	return nil
}

func resourceKeycloakOpenidClientDefaultScopesReconcile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)
	tfOpenidClientDefaultScopes := data.Get("default_scopes").(*schema.Set)

	keycloakOpenidClientDefaultScopes, err := keycloakClient.GetOpenidClientDefaultScopes(ctx, realmId, clientId)
	if err != nil {
		if keycloak.ErrorIs404(err) {
			return diag.FromErr(fmt.Errorf("validation error: client with id %s does not exist", clientId))
		}
		return diag.FromErr(err)
	}

	var openidClientDefaultScopesToDetach []string
	for _, keycloakOpenidClientDefaultScope := range keycloakOpenidClientDefaultScopes {
		// if this scope is attached in keycloak and tf state, no update is required
		// remove it from the set so we can look at scopes that need to be attached later
		if tfOpenidClientDefaultScopes.Contains(keycloakOpenidClientDefaultScope.Name) {
			tfOpenidClientDefaultScopes.Remove(keycloakOpenidClientDefaultScope.Name)
		} else {
			// if this scope is attached in keycloak but not in tf state, add them to a slice containing all scopes to detach
			openidClientDefaultScopesToDetach = append(openidClientDefaultScopesToDetach, keycloakOpenidClientDefaultScope.Name)
		}
	}

	// detach scopes that aren't in tf state
	err = keycloakClient.DetachOpenidClientDefaultScopes(ctx, realmId, clientId, openidClientDefaultScopesToDetach)
	if err != nil {
		return diag.FromErr(err)
	}

	// attach scopes that exist in tf state but not in keycloak
	err = keycloakClient.AttachOpenidClientDefaultScopes(ctx, realmId, clientId, interfaceSliceToStringSlice(tfOpenidClientDefaultScopes.List()))
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(openidClientDefaultScopesId(realmId, clientId))

	return resourceKeycloakOpenidClientDefaultScopesRead(ctx, data, meta)
}

func resourceKeycloakOpenidClientDefaultScopesDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)
	defaultScopes := data.Get("default_scopes").(*schema.Set)

	return diag.FromErr(keycloakClient.DetachOpenidClientDefaultScopes(ctx, realmId, clientId, interfaceSliceToStringSlice(defaultScopes.List())))
}
