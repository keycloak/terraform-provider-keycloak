package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakGroupPermissions() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakGroupPermissionsCreate,
		ReadContext:   resourceKeycloakGroupPermissionsRead,
		DeleteContext: resourceKeycloakGroupPermissionsDelete,
		UpdateContext: resourceKeycloakGroupPermissionsUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: resourceKeycloakGroupPermissionsImport,
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
			"group_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"authorization_resource_server_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Resource server id representing the realm management client on which this permission is managed",
			},
			"view_scope":              scopePermissionsSchema(),
			"manage_scope":            scopePermissionsSchema(),
			"view_members_scope":      scopePermissionsSchema(),
			"manage_members_scope":    scopePermissionsSchema(),
			"manage_membership_scope": scopePermissionsSchema(),
		},
	}
}

func groupPermissionsId(realmId, groupId string) string {
	return fmt.Sprintf("%s/%s", realmId, groupId)
}

func resourceKeycloakGroupPermissionsCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceKeycloakGroupPermissionsUpdate(ctx, data, meta)
}

func resourceKeycloakGroupPermissionsUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	groupId := data.Get("group_id").(string)

	// the existence of this resource implies that it is enabled.
	err := keycloakClient.EnableGroupPermissions(ctx, realmId, groupId)
	if err != nil {
		return diag.FromErr(err)
	}

	// setting scope permissions requires us to fetch the users permissions details, as well as the realm management client
	groupPermissions, err := keycloakClient.GetGroupPermissions(ctx, realmId, groupId)
	if err != nil {
		return diag.FromErr(err)
	}

	realmManagementClient, err := keycloakClient.GetOpenidClientByClientId(ctx, realmId, "realm-management")
	if err != nil {
		return diag.FromErr(err)
	}

	if viewScope, ok := data.GetOk("view_scope"); ok {
		err := setOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, groupPermissions.ScopePermissions["view"].(string), viewScope.(*schema.Set))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if manageScope, ok := data.GetOk("manage_scope"); ok {
		err := setOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, groupPermissions.ScopePermissions["manage"].(string), manageScope.(*schema.Set))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if viewMembersScope, ok := data.GetOk("view_members_scope"); ok {
		err := setOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, groupPermissions.ScopePermissions["view-members"].(string), viewMembersScope.(*schema.Set))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if manageMembersScope, ok := data.GetOk("manage_members_scope"); ok {
		err := setOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, groupPermissions.ScopePermissions["manage-members"].(string), manageMembersScope.(*schema.Set))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if manageMembershipScope, ok := data.GetOk("manage_membership_scope"); ok {
		err := setOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, groupPermissions.ScopePermissions["manage-membership"].(string), manageMembershipScope.(*schema.Set))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceKeycloakGroupPermissionsRead(ctx, data, meta)
}

func resourceKeycloakGroupPermissionsRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	realmId := data.Get("realm_id").(string)
	groupId := data.Get("group_id").(string)

	realmManagementClient, err := keycloakClient.GetOpenidClientByClientId(ctx, realmId, "realm-management")
	if err != nil {
		return diag.FromErr(err)
	}

	groupPermissions, err := keycloakClient.GetGroupPermissions(ctx, realmId, groupId)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	data.SetId(groupPermissionsId(groupPermissions.RealmId, groupPermissions.GroupId))
	data.Set("realm_id", groupPermissions.RealmId)
	data.Set("group_id", groupPermissions.GroupId)
	data.Set("enabled", groupPermissions.Enabled)
	data.Set("authorization_resource_server_id", realmManagementClient.Id)

	if viewScope, err := getOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, groupPermissions.ScopePermissions["view"].(string)); err == nil && viewScope != nil {
		data.Set("view_scope", []interface{}{viewScope})
	} else if err != nil {
		return diag.FromErr(err)
	}

	if manageScope, err := getOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, groupPermissions.ScopePermissions["manage"].(string)); err == nil && manageScope != nil {
		data.Set("manage_scope", []interface{}{manageScope})
	} else if err != nil {
		return diag.FromErr(err)
	}

	if viewMembersScope, err := getOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, groupPermissions.ScopePermissions["view-members"].(string)); err == nil && viewMembersScope != nil {
		data.Set("view_members_scope", []interface{}{viewMembersScope})
	} else if err != nil {
		return diag.FromErr(err)
	}

	if manageMembersScope, err := getOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, groupPermissions.ScopePermissions["manage-members"].(string)); err == nil && manageMembersScope != nil {
		data.Set("manage_members_scope", []interface{}{manageMembersScope})
	} else if err != nil {
		return diag.FromErr(err)
	}

	if manageMembershipScope, err := getOpenidClientScopePermissionPolicy(ctx, keycloakClient, realmId, realmManagementClient.Id, groupPermissions.ScopePermissions["manage-membership"].(string)); err == nil && manageMembershipScope != nil {
		data.Set("manage_membership_scope", []interface{}{manageMembershipScope})
	} else if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakGroupPermissionsDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	groupId := data.Get("group_id").(string)

	return diag.FromErr(keycloakClient.DisableGroupPermissions(ctx, realmId, groupId))
}

func resourceKeycloakGroupPermissionsImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("Invalid import. Supported import formats: {{realmId}}/{{groupId}}")
	}

	_, err := keycloakClient.GetGroupPermissions(ctx, parts[0], parts[1])
	if err != nil {
		return nil, err
	}

	d.Set("realm_id", parts[0])
	d.Set("group_id", parts[1])

	d.SetId(groupPermissionsId(parts[0], parts[1]))

	diagnostics := resourceKeycloakGroupPermissionsRead(ctx, d, meta)
	if diagnostics.HasError() {
		return nil, errors.New(diagnostics[0].Summary)
	}

	return []*schema.ResourceData{d}, nil
}
