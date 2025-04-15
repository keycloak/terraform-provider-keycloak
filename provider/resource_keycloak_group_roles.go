package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"strings"
)

func resourceKeycloakGroupRoles() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakGroupRolesReconcile,
		ReadContext:   resourceKeycloakGroupRolesRead,
		UpdateContext: resourceKeycloakGroupRolesReconcile,
		DeleteContext: resourceKeycloakGroupRolesDelete,
		// This resource can be imported using {{realm}}/{{groupId}}.
		Importer: &schema.ResourceImporter{
			StateContext: resourceKeycloakGroupRolesImport,
		},
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"group_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"role_ids": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Required: true,
			},
			"exhaustive": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func groupRolesId(realmId, groupId string) string {
	return fmt.Sprintf("%s/%s", realmId, groupId)
}

func addRolesToGroup(ctx context.Context, keycloakClient *keycloak.KeycloakClient, clientRolesToAdd map[string][]*keycloak.Role, realmRolesToAdd []*keycloak.Role, group *keycloak.Group) error {
	if len(realmRolesToAdd) != 0 {
		err := keycloakClient.AddRealmRolesToGroup(ctx, group.RealmId, group.Id, realmRolesToAdd)
		if err != nil {
			return err
		}
	}

	for k, roles := range clientRolesToAdd {
		if len(roles) != 0 {
			err := keycloakClient.AddClientRolesToGroup(ctx, group.RealmId, group.Id, k, roles)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func removeRolesFromGroup(ctx context.Context, keycloakClient *keycloak.KeycloakClient, clientRolesToRemove map[string][]*keycloak.Role, realmRolesToRemove []*keycloak.Role, group *keycloak.Group) error {
	if len(realmRolesToRemove) != 0 {
		err := keycloakClient.RemoveRealmRolesFromGroup(ctx, group.RealmId, group.Id, realmRolesToRemove)
		if err != nil {
			return err
		}
	}

	for k, roles := range clientRolesToRemove {
		if len(roles) != 0 {
			err := keycloakClient.RemoveClientRolesFromGroup(ctx, group.RealmId, group.Id, k, roles)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func resourceKeycloakGroupRolesReconcile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	groupId := data.Get("group_id").(string)
	roleIds := interfaceSliceToStringSlice(data.Get("role_ids").(*schema.Set).List())
	exhaustive := data.Get("exhaustive").(bool)

	group, err := keycloakClient.GetGroup(ctx, realmId, groupId)
	if err != nil {
		return diag.FromErr(err)
	}

	if data.HasChange("role_ids") {
		o, n := data.GetChange("role_ids")
		os := o.(*schema.Set)
		ns := n.(*schema.Set)
		remove := interfaceSliceToStringSlice(os.Difference(ns).List())

		tfRolesToRemove, err := getExtendedRoleMapping(ctx, keycloakClient, realmId, remove)
		if err != nil {
			return diag.FromErr(err)
		}

		if err = removeRolesFromGroup(ctx, keycloakClient, tfRolesToRemove.clientRoles, tfRolesToRemove.realmRoles, group); err != nil {
			return diag.FromErr(err)
		}
	}

	tfRoles, err := getExtendedRoleMapping(ctx, keycloakClient, realmId, roleIds)
	if err != nil {
		return diag.FromErr(err)
	}

	// get the list of currently assigned roles. Due to default realm and client roles
	roleMappings, err := keycloakClient.GetGroupRoleMappings(ctx, realmId, groupId)

	// sort into roles we need to add and roles we need to remove
	updates := calculateRoleMappingUpdates(tfRoles, intoRoleMapping(roleMappings))

	// add roles
	err = addRolesToGroup(ctx, keycloakClient, updates.clientRolesToAdd, updates.realmRolesToAdd, group)
	if err != nil {
		return diag.FromErr(err)
	}

	// remove roles if exhaustive (authoritative)
	if exhaustive {
		err = removeRolesFromGroup(ctx, keycloakClient, updates.clientRolesToRemove, updates.realmRolesToRemove, group)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	data.SetId(groupRolesId(realmId, groupId))

	return resourceKeycloakGroupRolesRead(ctx, data, meta)
}

func resourceKeycloakGroupRolesRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	groupId := data.Get("group_id").(string)
	sortedRoleIds := interfaceSliceToStringSlice(data.Get("role_ids").(*schema.Set).List())
	exhaustive := data.Get("exhaustive").(bool)

	// check if group exists, remove from state if not found
	if _, err := keycloakClient.GetGroup(ctx, realmId, groupId); err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	roles, err := keycloakClient.GetGroupRoleMappings(ctx, realmId, groupId)
	if err != nil {
		return diag.FromErr(err)
	}

	var roleIds []string

	for _, realmRole := range roles.RealmMappings {
		if exhaustive || stringSliceContains(sortedRoleIds, realmRole.Id) {
			roleIds = append(roleIds, realmRole.Id)
		}
	}

	for _, clientRoleMapping := range roles.ClientMappings {
		for _, clientRole := range clientRoleMapping.Mappings {
			if exhaustive || stringSliceContains(sortedRoleIds, clientRole.Id) {
				roleIds = append(roleIds, clientRole.Id)
			}
		}
	}

	data.Set("role_ids", roleIds)
	data.SetId(groupRolesId(realmId, groupId))

	return nil
}

func resourceKeycloakGroupRolesDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	groupId := data.Get("group_id").(string)

	group, err := keycloakClient.GetGroup(ctx, realmId, groupId)

	roleIds := interfaceSliceToStringSlice(data.Get("role_ids").(*schema.Set).List())
	rolesToRemove, err := getExtendedRoleMapping(ctx, keycloakClient, realmId, roleIds)
	if err != nil {
		return diag.FromErr(err)
	}

	err = removeRolesFromGroup(ctx, keycloakClient, rolesToRemove.clientRoles, rolesToRemove.realmRoles, group)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakGroupRolesImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("Invalid import. Supported import format: {{realm}}/{{groupId}}.")
	}

	realmId := parts[0]
	groupId := parts[1]

	if _, err := keycloakClient.GetGroup(ctx, realmId, groupId); err != nil {
		return nil, err
	}

	_, err := keycloakClient.GetGroupRoleMappings(ctx, realmId, groupId)
	if err != nil {
		return nil, err
	}

	d.Set("realm_id", realmId)
	d.Set("group_id", groupId)
	d.Set("exhaustive", true)

	diagnostics := resourceKeycloakGroupRolesRead(ctx, d, meta)
	if diagnostics.HasError() {
		return nil, errors.New(diagnostics[0].Summary)
	}

	return []*schema.ResourceData{d}, nil
}
