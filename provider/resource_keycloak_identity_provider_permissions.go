package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakIdentityProviderPermissions() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakIdentityProviderPermissionsReconcile,
		ReadContext:   resourceKeycloakIdentityProviderPermissionsRead,
		DeleteContext: resourceKeycloakIdentityProviderPermissionsDelete,
		UpdateContext: resourceKeycloakIdentityProviderPermissionsReconcile,
		Importer: &schema.ResourceImporter{
			StateContext: resourceKeycloakIdentityProviderPermissionsImport,
		},
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"provider_alias": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"authorization_resource_server_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Resource server id representing the realm management client on which this permission is managed",
			},
			"view_scope":           scopePermissionsSchema(),
			"manage_scope":         scopePermissionsSchema(),
			"token_exchange_scope": scopePermissionsSchema(),
		},
	}
}

func identityProviderPermissionsId(realmId, providerAlias string) string {
	return fmt.Sprintf("%s/%s", realmId, providerAlias)
}

func idpScopePermissionId(permissions *keycloak.IdentityProviderPermissions, scope string) string {
	if v, ok := permissions.ScopePermissions[scope]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func resourceKeycloakIdentityProviderPermissionsReconcile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	providerAlias := data.Get("provider_alias").(string)

	err := keycloakClient.EnableIdentityProviderPermissions(ctx, realmId, providerAlias)
	if err != nil {
		return diag.FromErr(err)
	}

	idpPermissions, err := keycloakClient.GetIdentityProviderPermissions(ctx, realmId, providerAlias)
	if err != nil {
		return diag.FromErr(err)
	}

	realmManagementClient, err := keycloakClient.GetOpenidClientByClientId(ctx, realmId, "realm-management")
	if err != nil {
		return diag.FromErr(err)
	}

	if viewScope, ok := data.GetOk("view_scope"); ok {
		if scopeId := idpScopePermissionId(idpPermissions, "view"); scopeId != "" {
			err := setOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, scopeId, viewScope.(*schema.Set))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}
	if manageScope, ok := data.GetOk("manage_scope"); ok {
		if scopeId := idpScopePermissionId(idpPermissions, "manage"); scopeId != "" {
			err := setOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, scopeId, manageScope.(*schema.Set))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}
	if tokenExchangeScope, ok := data.GetOk("token_exchange_scope"); ok {
		if scopeId := idpScopePermissionId(idpPermissions, "token-exchange"); scopeId != "" {
			err := setOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, scopeId, tokenExchangeScope.(*schema.Set))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return resourceKeycloakIdentityProviderPermissionsRead(ctx, data, meta)
}

func resourceKeycloakIdentityProviderPermissionsRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	providerAlias := data.Get("provider_alias").(string)

	realmManagementClient, err := keycloakClient.GetOpenidClientByClientId(ctx, realmId, "realm-management")
	if err != nil {
		return diag.FromErr(err)
	}

	idpPermissions, err := keycloakClient.GetIdentityProviderPermissions(ctx, realmId, providerAlias)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	if !idpPermissions.Enabled {
		tflog.Warn(ctx, "Removing resource from state as it is no longer enabled", map[string]interface{}{
			"id": data.Id(),
		})
		data.SetId("")
		return nil
	}

	data.SetId(identityProviderPermissionsId(idpPermissions.RealmId, idpPermissions.ProviderAlias))
	data.Set("realm_id", idpPermissions.RealmId)
	data.Set("provider_alias", idpPermissions.ProviderAlias)
	data.Set("enabled", idpPermissions.Enabled)
	data.Set("authorization_resource_server_id", realmManagementClient.Id)

	if scopeId := idpScopePermissionId(idpPermissions, "view"); scopeId != "" {
		if viewScope, err := getOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, scopeId); err == nil && viewScope != nil {
			data.Set("view_scope", []interface{}{viewScope})
		} else if err != nil {
			return diag.FromErr(err)
		}
	}

	if scopeId := idpScopePermissionId(idpPermissions, "manage"); scopeId != "" {
		if manageScope, err := getOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, scopeId); err == nil && manageScope != nil {
			data.Set("manage_scope", []interface{}{manageScope})
		} else if err != nil {
			return diag.FromErr(err)
		}
	}

	if scopeId := idpScopePermissionId(idpPermissions, "token-exchange"); scopeId != "" {
		if tokenExchangeScope, err := getOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, scopeId); err == nil && tokenExchangeScope != nil {
			data.Set("token_exchange_scope", []interface{}{tokenExchangeScope})
		} else if err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceKeycloakIdentityProviderPermissionsDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	providerAlias := data.Get("provider_alias").(string)

	return diag.FromErr(keycloakClient.DisableIdentityProviderPermissions(ctx, realmId, providerAlias))
}

func resourceKeycloakIdentityProviderPermissionsImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("Invalid import. Supported import formats: {{realmId}}/{{providerAlias}}")
	}

	_, err := keycloakClient.GetIdentityProviderPermissions(ctx, parts[0], parts[1])
	if err != nil {
		return nil, err
	}

	d.Set("realm_id", parts[0])
	d.Set("provider_alias", parts[1])
	d.SetId(identityProviderPermissionsId(parts[0], parts[1]))

	diagnostics := resourceKeycloakIdentityProviderPermissionsRead(ctx, d, meta)
	if diagnostics.HasError() {
		return nil, errors.New(diagnostics[0].Summary)
	}

	return []*schema.ResourceData{d}, nil
}
