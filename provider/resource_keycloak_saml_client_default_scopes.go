package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakSamlClientDefaultScopes() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakSamlClientDefaultScopesCreate,
		ReadContext:   resourceKeycloakSamlClientDefaultScopesRead,
		DeleteContext: resourceKeycloakSamlClientDefaultScopesDelete,
		UpdateContext: resourceKeycloakSamlClientDefaultScopesUpdate,
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

func resourceKeycloakSamlClientDefaultScopesCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)
	defaultScopes := data.Get("default_scopes").(*schema.Set)

	err := keycloakClient.AttachSamlClientDefaultScopes(ctx, realmId, clientId, interfaceSliceToStringSlice(defaultScopes.List()))
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(samlClientDefaultScopesId(realmId, clientId))

	return resourceKeycloakSamlClientDefaultScopesRead(ctx, data, meta)
}

func samlClientDefaultScopesId(realmId string, clientId string) string {
	return fmt.Sprintf("%s/%s", realmId, clientId)
}

func resourceKeycloakSamlClientDefaultScopesRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)

	clientScopes, err := keycloakClient.GetSamlClientDefaultScopes(ctx, realmId, clientId)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	var defaultScopes []string
	for _, clientScope := range clientScopes {
		defaultScopes = append(defaultScopes, clientScope.Name)
	}

	data.Set("default_scopes", defaultScopes)
	data.SetId(samlClientDefaultScopesId(realmId, clientId))

	return nil
}

func resourceKeycloakSamlClientDefaultScopesUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)
	tfSamlClientDefaultScopes := data.Get("default_scopes").(*schema.Set)

	keycloakSamlClientDefaultScopes, err := keycloakClient.GetSamlClientDefaultScopes(ctx, realmId, clientId)
	if err != nil {
		return diag.FromErr(err)
	}

	var samlClientDefaultScopesToDetach []string
	for _, keycloakSamlClientDefaultScope := range keycloakSamlClientDefaultScopes {
		// if this scope is attached in keycloak and tf state, no update is required
		// remove it from the set so we can look at scopes that need to be attached later
		if tfSamlClientDefaultScopes.Contains(keycloakSamlClientDefaultScope.Name) {
			tfSamlClientDefaultScopes.Remove(keycloakSamlClientDefaultScope.Name)
		} else {
			// if this scope is attached in keycloak but not in tf state, add them to a slice containing all scopes to detach
			samlClientDefaultScopesToDetach = append(samlClientDefaultScopesToDetach, keycloakSamlClientDefaultScope.Name)
		}
	}

	// detach scopes that aren't in tf state
	err = keycloakClient.DetachSamlClientDefaultScopes(ctx, realmId, clientId, samlClientDefaultScopesToDetach)
	if err != nil {
		return diag.FromErr(err)
	}

	// attach scopes that exist in tf state but not in keycloak
	err = keycloakClient.AttachSamlClientDefaultScopes(ctx, realmId, clientId, interfaceSliceToStringSlice(tfSamlClientDefaultScopes.List()))
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(samlClientDefaultScopesId(realmId, clientId))

	return resourceKeycloakSamlClientDefaultScopesRead(ctx, data, meta)
}

func resourceKeycloakSamlClientDefaultScopesDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)
	defaultScopes := data.Get("default_scopes").(*schema.Set)

	return diag.FromErr(keycloakClient.DetachSamlClientDefaultScopes(ctx, realmId, clientId, interfaceSliceToStringSlice(defaultScopes.List())))
}
