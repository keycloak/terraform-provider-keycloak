package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakOpenidClientPermissions() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakOpenidClientPermissionsReconcile,
		ReadContext:   resourceKeycloakOpenidClientPermissionsRead,
		DeleteContext: resourceKeycloakOpenidClientPermissionsDelete,
		UpdateContext: resourceKeycloakOpenidClientPermissionsReconcile,
		// This resource can be imported using {{realm}}/{{client_id}}. The Client ID is displayed in the GUI
		Importer: &schema.ResourceImporter{
			StateContext: resourceKeycloakOpenidClientPermissionsImport,
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
			"client_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"authorization_resource_server_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Resource server id representing the realm management client on which this permission is managed",
			},
			"view_scope":                   scopePermissionsSchema(),
			"manage_scope":                 scopePermissionsSchema(),
			"configure_scope":              scopePermissionsSchema(),
			"map_roles_scope":              scopePermissionsSchema(),
			"map_roles_client_scope_scope": scopePermissionsSchema(),
			"map_roles_composite_scope":    scopePermissionsSchema(),
			"token_exchange_scope":         scopePermissionsSchema(),
		},
	}
}

func clientPermissionsId(realmId, clientId string) string {
	return fmt.Sprintf("%s/%s", realmId, clientId)
}

func resourceKeycloakOpenidClientPermissionsReconcile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)

	// the existence of this resource implies that permissions are enabled for this client.
	err := keycloakClient.EnableOpenidClientPermissions(ctx, realmId, clientId)
	if err != nil {
		return diag.FromErr(err)
	}

	openidClientPermissions, err := keycloakClient.GetOpenidClientPermissions(ctx, realmId, clientId)
	if err != nil {
		return diag.FromErr(err)
	}

	realmManagementClient, err := keycloakClient.GetOpenidClientByClientId(ctx, realmId, "realm-management")
	if err != nil {
		return diag.FromErr(err)
	}

	if viewScope, ok := data.GetOk("view_scope"); ok {
		err := setOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, openidClientPermissions.ScopePermissions["view"], viewScope.(*schema.Set))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if manageScope, ok := data.GetOk("manage_scope"); ok {
		err := setOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, openidClientPermissions.ScopePermissions["manage"], manageScope.(*schema.Set))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if configureScope, ok := data.GetOk("configure_scope"); ok {
		err := setOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, openidClientPermissions.ScopePermissions["configure"], configureScope.(*schema.Set))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if mapRolesScope, ok := data.GetOk("map_roles_scope"); ok {
		err := setOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, openidClientPermissions.ScopePermissions["map-roles"], mapRolesScope.(*schema.Set))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if mapRolesClientsScope, ok := data.GetOk("map_roles_client_scope_scope"); ok {
		err := setOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, openidClientPermissions.ScopePermissions["map-roles-client-scope"], mapRolesClientsScope.(*schema.Set))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if mapRolesCompositeScope, ok := data.GetOk("map_roles_composite_scope"); ok {
		err := setOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, openidClientPermissions.ScopePermissions["map-roles-composite"], mapRolesCompositeScope.(*schema.Set))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if tokenExchangeScope, ok := data.GetOk("token_exchange_scope"); ok {
		err := setOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, openidClientPermissions.ScopePermissions["token-exchange"], tokenExchangeScope.(*schema.Set))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceKeycloakOpenidClientPermissionsRead(ctx, data, meta)
}

func resourceKeycloakOpenidClientPermissionsRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)

	openidClientPermissions, err := keycloakClient.GetOpenidClientPermissions(ctx, realmId, clientId)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	if !openidClientPermissions.Enabled {
		tflog.Warn(ctx, "Removing resource from state as it is no longer enabled", map[string]interface{}{
			"id": data.Id(),
		})
		data.SetId("")
		return nil
	}

	realmManagementClient, err := keycloakClient.GetOpenidClientByClientId(ctx, realmId, "realm-management")
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(clientPermissionsId(openidClientPermissions.RealmId, openidClientPermissions.ClientId))
	data.Set("realm_id", openidClientPermissions.RealmId)
	data.Set("client_id", openidClientPermissions.ClientId)
	data.Set("enabled", openidClientPermissions.Enabled)
	data.Set("authorization_resource_server_id", realmManagementClient.Id)

	if viewScope, err := getOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, openidClientPermissions.ScopePermissions["view"]); err == nil && viewScope != nil {
		data.Set("view_scope", []interface{}{viewScope})
	} else if err != nil {
		return diag.FromErr(err)
	}

	if manageScope, err := getOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, openidClientPermissions.ScopePermissions["manage"]); err == nil && manageScope != nil {
		data.Set("manage_scope", []interface{}{manageScope})
	} else if err != nil {
		return diag.FromErr(err)
	}

	if mapRolesScope, err := getOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, openidClientPermissions.ScopePermissions["configure"]); err == nil && mapRolesScope != nil {
		data.Set("configure_scope", []interface{}{mapRolesScope})
	} else if err != nil {
		return diag.FromErr(err)
	}

	if manageGroupMembershipScope, err := getOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, openidClientPermissions.ScopePermissions["map-roles"]); err == nil && manageGroupMembershipScope != nil {
		data.Set("map_roles_scope", []interface{}{manageGroupMembershipScope})
	} else if err != nil {
		return diag.FromErr(err)
	}

	if impersonateScope, err := getOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, openidClientPermissions.ScopePermissions["map-roles-client-scope"]); err == nil && impersonateScope != nil {
		data.Set("map_roles_client_scope_scope", []interface{}{impersonateScope})
	} else if err != nil {
		return diag.FromErr(err)
	}

	if userImpersonatedScope, err := getOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, openidClientPermissions.ScopePermissions["map-roles-composite"]); err == nil && userImpersonatedScope != nil {
		data.Set("map_roles_composite_scope", []interface{}{userImpersonatedScope})
	} else if err != nil {
		return diag.FromErr(err)
	}

	if tokenExchangeScope, err := getOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, openidClientPermissions.ScopePermissions["token-exchange"]); err == nil && tokenExchangeScope != nil {
		data.Set("token_exchange_scope", []interface{}{tokenExchangeScope})
	} else if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakOpenidClientPermissionsDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)

	return diag.FromErr(keycloakClient.DisableOpenidClientPermissions(ctx, realmId, clientId))
}

func resourceKeycloakOpenidClientPermissionsImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("Invalid import. Supported import formats: {{realmId}}/{{openidClientId}}")
	}
	d.Set("realm_id", parts[0])
	d.Set("client_id", parts[1])

	d.SetId(clientPermissionsId(parts[0], parts[1]))

	return []*schema.ResourceData{d}, nil
}
