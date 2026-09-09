package provider

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakOpenidDefaultDefaultClientScopes() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakOpenidDefaultDefaultClientScopesReconcile,
		ReadContext:   resourceKeycloakOpenidDefaultDefaultClientScopesRead,
		DeleteContext: resourceKeycloakOpenidDefaultDefaultClientScopesDelete,
		UpdateContext: resourceKeycloakOpenidDefaultDefaultClientScopesReconcile,
		Importer: &schema.ResourceImporter{
			StateContext: resourceKeycloakOpenidDefaultDefaultClientScopesImport,
		},
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"default_scopes": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Required: true,
			},
		},
	}
}

func resourceKeycloakOpenidDefaultDefaultClientScopesRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)

	clientScopes, err := keycloakClient.GetOpenidRealmDefaultDefaultClientScopes(ctx, realmId)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	var defaultScopes []string
	for _, clientScope := range clientScopes {
		defaultScopes = append(defaultScopes, clientScope.Name)
	}

	err = data.Set("default_scopes", defaultScopes)
	if err != nil {
		return diag.FromErr(err)
	}
	data.SetId(realmId)

	return nil
}

func resourceKeycloakOpenidDefaultDefaultClientScopesReconcile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	tfOpenidDefaultDefaultScopes := interfaceSliceToStringSlice(data.Get("default_scopes").([]any))

	keycloakOpenidDefaultDefaultScopes, err := keycloakClient.GetOpenidRealmDefaultDefaultClientScopes(ctx, realmId)
	if err != nil {
		if keycloak.ErrorIs404(err) {
			return diag.FromErr(fmt.Errorf("validation error: realm with id %s does not exist", realmId))
		}
		return diag.FromErr(fmt.Errorf("validation error: error getting default default client scopes: %s", err.Error()))
	}

	diagnostics := detachDeletedDefaultScopes(ctx, keycloakOpenidDefaultDefaultScopes, tfOpenidDefaultDefaultScopes, err, keycloakClient, realmId)
	if diagnostics != nil {
		return diagnostics
	}

	if len(tfOpenidDefaultDefaultScopes) > 0 {
		diagnostics = attachNewDefaultScopes(ctx, keycloakOpenidDefaultDefaultScopes, keycloakClient, realmId, tfOpenidDefaultDefaultScopes)
		if diagnostics != nil {
			return diagnostics
		}
	}

	data.SetId(realmId)

	return waitForDefaultUpdates(ctx, keycloakClient, realmId, tfOpenidDefaultDefaultScopes, 10)
}

func waitForDefaultUpdates(ctx context.Context, keycloakClient *keycloak.KeycloakClient, realmId string, scopes []string, times int) diag.Diagnostics {
	if times == 0 {
		return nil
	}
	keycloakOpenidDefaultDefaultScopes, err := keycloakClient.GetOpenidRealmDefaultDefaultClientScopes(ctx, realmId)
	if err != nil {
		if keycloak.ErrorIs404(err) {
			return diag.FromErr(fmt.Errorf("validation error: realm with id %s does not exist", realmId))
		}
		return diag.FromErr(fmt.Errorf("validation error: error getting default default client scopes: %s", err.Error()))
	}

	if len(keycloakOpenidDefaultDefaultScopes) != len(scopes) {
		time.Sleep(1 * time.Second)
		return waitForDefaultUpdates(ctx, keycloakClient, realmId, scopes, times-1)
	}
	for _, keycloakOpenidDefaultDefaultScope := range keycloakOpenidDefaultDefaultScopes {
		if !slices.Contains(scopes, keycloakOpenidDefaultDefaultScope.Name) {
			time.Sleep(1 * time.Second)
			return waitForDefaultUpdates(ctx, keycloakClient, realmId, scopes, times-1)
		}
	}
	return nil
}

func detachDeletedDefaultScopes(ctx context.Context, keycloakOpenidDefaultDefaultScopes []*keycloak.OpenidClientScope, tfOpenidDefaultDefaultScopes []string, err error, keycloakClient *keycloak.KeycloakClient, realmId string) diag.Diagnostics {
	for _, keycloakOpenidDefaultDefaultScope := range keycloakOpenidDefaultDefaultScopes {
		if !slices.Contains(tfOpenidDefaultDefaultScopes, keycloakOpenidDefaultDefaultScope.Name) {
			err = keycloakClient.DeleteOpenidRealmDefaultDefaultClientScope(ctx, realmId, keycloakOpenidDefaultDefaultScope.Id)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}
	return nil
}

func attachNewDefaultScopes(ctx context.Context, keycloakOpenidDefaultDefaultScopes []*keycloak.OpenidClientScope, keycloakClient *keycloak.KeycloakClient, realmId string, tfOpenidDefaultDefaultScopes []string) diag.Diagnostics {
	keycloakClientScopes, err := keycloakClient.GetRealmClientScopes(ctx, realmId)
	if err != nil {
		return diag.FromErr(err)
	}
	for _, keycloakClientScope := range keycloakClientScopes {
		if slices.Contains(tfOpenidDefaultDefaultScopes, keycloakClientScope.Name) &&
			!slices.ContainsFunc(keycloakOpenidDefaultDefaultScopes, func(e *keycloak.OpenidClientScope) bool {
				return e.Id == keycloakClientScope.Id
			}) {
			err = keycloakClient.PutOpenidRealmDefaultDefaultClientScope(ctx, realmId, keycloakClientScope.Id)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}
	return nil
}

func resourceKeycloakOpenidDefaultDefaultClientScopesDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)

	keycloakOpenidDefaultDefaultScopes, err := keycloakClient.GetOpenidRealmDefaultDefaultClientScopes(ctx, realmId)
	if err != nil {
		if keycloak.ErrorIs404(err) {
			return diag.FromErr(fmt.Errorf("validation error: realm with id %s does not exist", realmId))
		}
		return diag.FromErr(err)
	}

	for _, keycloakClientScope := range keycloakOpenidDefaultDefaultScopes {
		err = keycloakClient.DeleteOpenidRealmDefaultDefaultClientScope(ctx, realmId, keycloakClientScope.Id)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

func resourceKeycloakOpenidDefaultDefaultClientScopesImport(ctx context.Context, data *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	err := data.Set("realm_id", data.Id())
	if err != nil {
		return nil, err
	}

	diagnostics := resourceKeycloakOpenidDefaultDefaultClientScopesRead(ctx, data, meta)
	if diagnostics.HasError() {
		return nil, errors.New(diagnostics[0].Summary)
	}

	return []*schema.ResourceData{data}, nil
}
