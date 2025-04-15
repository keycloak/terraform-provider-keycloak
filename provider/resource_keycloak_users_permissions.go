package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakUsersPermissions() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakUsersPermissionsReconcile,
		ReadContext:   resourceKeycloakUsersPermissionsRead,
		DeleteContext: resourceKeycloakUsersPermissionsDelete,
		UpdateContext: resourceKeycloakUsersPermissionsReconcile,
		Importer: &schema.ResourceImporter{
			StateContext: resourceKeycloakUsersPermissionsImport,
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
			"authorization_resource_server_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Resource server id representing the realm management client on which this permission is managed",
			},
			"view_scope":                    scopePermissionsSchema(),
			"manage_scope":                  scopePermissionsSchema(),
			"map_roles_scope":               scopePermissionsSchema(),
			"manage_group_membership_scope": scopePermissionsSchema(),
			"impersonate_scope":             scopePermissionsSchema(),
			"user_impersonated_scope":       scopePermissionsSchema(),
		},
	}
}

func resourceKeycloakUsersPermissionsReconcile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)

	// the existence of this resource implies that it is enabled.
	err := keycloakClient.EnableUsersPermissions(ctx, realmId)
	if err != nil {
		return diag.FromErr(err)
	}

	// setting scope permissions requires us to fetch the users permissions details, as well as the realm management client
	usersPermissions, err := keycloakClient.GetUsersPermissions(ctx, realmId)
	if err != nil {
		return diag.FromErr(err)
	}

	realmManagementClient, err := keycloakClient.GetOpenidClientByClientId(ctx, realmId, "realm-management")
	if err != nil {
		return diag.FromErr(err)
	}

	if viewScope, ok := data.GetOk("view_scope"); ok {
		err := setOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, usersPermissions.ScopePermissions["view"], viewScope.(*schema.Set))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if manageScope, ok := data.GetOk("manage_scope"); ok {
		err := setOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, usersPermissions.ScopePermissions["manage"], manageScope.(*schema.Set))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if mapRolesScope, ok := data.GetOk("map_roles_scope"); ok {
		err := setOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, usersPermissions.ScopePermissions["map-roles"], mapRolesScope.(*schema.Set))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if manageGroupMembershipScope, ok := data.GetOk("manage_group_membership_scope"); ok {
		err := setOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, usersPermissions.ScopePermissions["manage-group-membership"], manageGroupMembershipScope.(*schema.Set))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if impersonateScope, ok := data.GetOk("impersonate_scope"); ok {
		err := setOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, usersPermissions.ScopePermissions["impersonate"], impersonateScope.(*schema.Set))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if userImpersonatedScope, ok := data.GetOk("user_impersonated_scope"); ok {
		err := setOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, usersPermissions.ScopePermissions["user-impersonated"], userImpersonatedScope.(*schema.Set))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceKeycloakUsersPermissionsRead(ctx, data, meta)
}

func resourceKeycloakUsersPermissionsRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	realmId := data.Get("realm_id").(string)

	realmManagementClient, err := keycloakClient.GetOpenidClientByClientId(ctx, realmId, "realm-management")
	if err != nil {
		return diag.FromErr(err)
	}

	usersPermissions, err := keycloakClient.GetUsersPermissions(ctx, realmId)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	if !usersPermissions.Enabled {
		tflog.Warn(ctx, "Removing resource with id from state as it is no longer enabled", map[string]interface{}{
			"id": data.Id(),
		})
		data.SetId("")
		return nil
	}

	data.SetId(usersPermissions.RealmId)
	data.Set("realm_id", usersPermissions.RealmId)
	data.Set("enabled", usersPermissions.Enabled)
	data.Set("authorization_resource_server_id", realmManagementClient.Id)

	if viewScope, err := getOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, usersPermissions.ScopePermissions["view"]); err == nil && viewScope != nil {
		data.Set("view_scope", []interface{}{viewScope})
	} else if err != nil {
		return diag.FromErr(err)
	}

	if manageScope, err := getOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, usersPermissions.ScopePermissions["manage"]); err == nil && manageScope != nil {
		data.Set("manage_scope", []interface{}{manageScope})
	} else if err != nil {
		return diag.FromErr(err)
	}

	if mapRolesScope, err := getOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, usersPermissions.ScopePermissions["map-roles"]); err == nil && mapRolesScope != nil {
		data.Set("map_roles_scope", []interface{}{mapRolesScope})
	} else if err != nil {
		return diag.FromErr(err)
	}

	if manageGroupMembershipScope, err := getOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, usersPermissions.ScopePermissions["manage-group-membership"]); err == nil && manageGroupMembershipScope != nil {
		data.Set("manage_group_membership_scope", []interface{}{manageGroupMembershipScope})
	} else if err != nil {
		return diag.FromErr(err)
	}

	if impersonateScope, err := getOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, usersPermissions.ScopePermissions["impersonate"]); err == nil && impersonateScope != nil {
		data.Set("impersonate_scope", []interface{}{impersonateScope})
	} else if err != nil {
		return diag.FromErr(err)
	}

	if userImpersonatedScope, err := getOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, usersPermissions.ScopePermissions["user-impersonated"]); err == nil && userImpersonatedScope != nil {
		data.Set("user_impersonated_scope", []interface{}{userImpersonatedScope})
	} else if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakUsersPermissionsDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)

	return diag.FromErr(keycloakClient.DisableUsersPermissions(ctx, realmId))
}

func resourceKeycloakUsersPermissionsImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	d.Set("realm_id", d.Id())
	d.SetId(d.Id())

	return []*schema.ResourceData{d}, nil
}
