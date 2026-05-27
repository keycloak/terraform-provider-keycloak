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

func resourceKeycloakRolePermissions() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakRolePermissionsReconcile,
		ReadContext:   resourceKeycloakRolePermissionsRead,
		DeleteContext: resourceKeycloakRolePermissionsDelete,
		UpdateContext: resourceKeycloakRolePermissionsReconcile,
		Importer: &schema.ResourceImporter{
			StateContext: resourceKeycloakRolePermissionsImport,
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
			"role_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"authorization_resource_server_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Resource server id representing the realm management client on which this permission is managed",
			},
			"view_scope":     scopePermissionsSchema(),
			"map_role_scope": scopePermissionsSchema(),
			"manage_scope":   scopePermissionsSchema(),
		},
	}
}

func rolePermissionsId(realmId, roleId string) string {
	return fmt.Sprintf("%s/%s", realmId, roleId)
}

func resourceKeycloakRolePermissionsReconcile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	roleId := data.Get("role_id").(string)

	err := keycloakClient.EnableRolePermissions(ctx, realmId, roleId)
	if err != nil {
		return diag.FromErr(err)
	}

	rolePermissions, err := keycloakClient.GetRolePermissions(ctx, realmId, roleId)
	if err != nil {
		return diag.FromErr(err)
	}

	realmManagementClient, err := keycloakClient.GetOpenidClientByClientId(ctx, realmId, "realm-management")
	if err != nil {
		return diag.FromErr(err)
	}

	if viewScope, ok := data.GetOk("view_scope"); ok {
		err := setOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, rolePermissions.ScopePermissions["view"], viewScope.(*schema.Set))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if mapRoleScope, ok := data.GetOk("map_role_scope"); ok {
		err := setOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, rolePermissions.ScopePermissions["map-role"], mapRoleScope.(*schema.Set))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if manageScope, ok := data.GetOk("manage_scope"); ok {
		err := setOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, rolePermissions.ScopePermissions["manage"], manageScope.(*schema.Set))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceKeycloakRolePermissionsRead(ctx, data, meta)
}

func resourceKeycloakRolePermissionsRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	roleId := data.Get("role_id").(string)

	realmManagementClient, err := keycloakClient.GetOpenidClientByClientId(ctx, realmId, "realm-management")
	if err != nil {
		return diag.FromErr(err)
	}

	rolePermissions, err := keycloakClient.GetRolePermissions(ctx, realmId, roleId)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	if !rolePermissions.Enabled {
		tflog.Warn(ctx, "Removing resource from state as it is no longer enabled", map[string]interface{}{
			"id": data.Id(),
		})
		data.SetId("")
		return nil
	}

	data.SetId(rolePermissionsId(rolePermissions.RealmId, rolePermissions.RoleId))
	data.Set("realm_id", rolePermissions.RealmId)
	data.Set("role_id", rolePermissions.RoleId)
	data.Set("enabled", rolePermissions.Enabled)
	data.Set("authorization_resource_server_id", realmManagementClient.Id)

	if scopeId := rolePermissions.ScopePermissions["view"]; scopeId != "" {
		if viewScope, err := getOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, scopeId); err == nil && viewScope != nil {
			data.Set("view_scope", []interface{}{viewScope})
		} else if err != nil {
			return diag.FromErr(err)
		}
	}

	if scopeId := rolePermissions.ScopePermissions["map-role"]; scopeId != "" {
		if mapRoleScope, err := getOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, scopeId); err == nil && mapRoleScope != nil {
			data.Set("map_role_scope", []interface{}{mapRoleScope})
		} else if err != nil {
			return diag.FromErr(err)
		}
	}

	if scopeId := rolePermissions.ScopePermissions["manage"]; scopeId != "" {
		if manageScope, err := getOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, scopeId); err == nil && manageScope != nil {
			data.Set("manage_scope", []interface{}{manageScope})
		} else if err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceKeycloakRolePermissionsDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	roleId := data.Get("role_id").(string)

	return diag.FromErr(keycloakClient.DisableRolePermissions(ctx, realmId, roleId))
}

func resourceKeycloakRolePermissionsImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("Invalid import. Supported import formats: {{realmId}}/{{roleId}}")
	}

	_, err := keycloakClient.GetRolePermissions(ctx, parts[0], parts[1])
	if err != nil {
		return nil, err
	}

	d.Set("realm_id", parts[0])
	d.Set("role_id", parts[1])
	d.SetId(rolePermissionsId(parts[0], parts[1]))

	diagnostics := resourceKeycloakRolePermissionsRead(ctx, d, meta)
	if diagnostics.HasError() {
		return nil, errors.New(diagnostics[0].Summary)
	}

	return []*schema.ResourceData{d}, nil
}
