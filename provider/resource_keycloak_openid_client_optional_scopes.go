package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakOpenidClientOptionalScopes() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakOpenidClientOptionalScopesReconcile,
		ReadContext:   resourceKeycloakOpenidClientOptionalScopesRead,
		DeleteContext: resourceKeycloakOpenidClientOptionalScopesDelete,
		UpdateContext: resourceKeycloakOpenidClientOptionalScopesReconcile,
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
			"optional_scopes": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Required: true,
				Set:      schema.HashString,
			},
		},
	}
}

func openidClientOptionalScopesId(realmId string, clientId string) string {
	return fmt.Sprintf("%s/%s", realmId, clientId)
}

func resourceKeycloakOpenidClientOptionalScopesRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)

	clientScopes, err := keycloakClient.GetOpenidClientOptionalScopes(ctx, realmId, clientId)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	var optionalScopes []string
	for _, clientScope := range clientScopes {
		optionalScopes = append(optionalScopes, clientScope.Name)
	}

	data.Set("optional_scopes", optionalScopes)
	data.SetId(openidClientOptionalScopesId(realmId, clientId))

	return nil
}

func resourceKeycloakOpenidClientOptionalScopesReconcile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)
	tfOpenidClientOptionalScopes := data.Get("optional_scopes").(*schema.Set)

	keycloakOpenidClientOptionalScopes, err := keycloakClient.GetOpenidClientOptionalScopes(ctx, realmId, clientId)
	if err != nil {
		if keycloak.ErrorIs404(err) {
			return diag.FromErr(fmt.Errorf("validation error: client with id %s does not exist", clientId))
		}
		return diag.FromErr(err)
	}

	var openidClientOptionalScopesToDetach []string
	for _, keycloakOpenidClientOptionalScope := range keycloakOpenidClientOptionalScopes {
		// if this scope is attached in keycloak and tf state, no update is required
		// remove it from the set so we can look at scopes that need to be attached later
		if tfOpenidClientOptionalScopes.Contains(keycloakOpenidClientOptionalScope.Name) {
			tfOpenidClientOptionalScopes.Remove(keycloakOpenidClientOptionalScope.Name)
		} else {
			// if this scope is attached in keycloak but not in tf state, add them to a slice containing all scopes to detach
			openidClientOptionalScopesToDetach = append(openidClientOptionalScopesToDetach, keycloakOpenidClientOptionalScope.Name)
		}
	}

	// detach scopes that aren't in tf state
	err = keycloakClient.DetachOpenidClientOptionalScopes(ctx, realmId, clientId, openidClientOptionalScopesToDetach)
	if err != nil {
		return diag.FromErr(err)
	}

	// attach scopes that exist in tf state but not in keycloak
	err = keycloakClient.AttachOpenidClientOptionalScopes(ctx, realmId, clientId, interfaceSliceToStringSlice(tfOpenidClientOptionalScopes.List()))
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(openidClientOptionalScopesId(realmId, clientId))

	return resourceKeycloakOpenidClientOptionalScopesRead(ctx, data, meta)
}

func resourceKeycloakOpenidClientOptionalScopesDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)
	optionalScopes := data.Get("optional_scopes").(*schema.Set)

	return diag.FromErr(keycloakClient.DetachOpenidClientOptionalScopes(ctx, realmId, clientId, interfaceSliceToStringSlice(optionalScopes.List())))
}
