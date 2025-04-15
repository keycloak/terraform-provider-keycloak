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

func resourceKeycloakUserGroups() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakUserGroupsReconcile,
		ReadContext:   resourceKeycloakUserGroupsRead,
		DeleteContext: resourceKeycloakUserGroupsDelete,
		UpdateContext: resourceKeycloakUserGroupsReconcile,
		// This resource can be imported using {{realm}}/{{userId}}.
		Importer: &schema.ResourceImporter{
			StateContext: resourceKeycloakUserGroupsImport,
		},
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"user_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"group_ids": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Required: true,
			},
			"exhaustive": {
				Type:     schema.TypeBool,
				Default:  true,
				Optional: true,
			},
		},
	}
}

func resourceKeycloakUserGroupsRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	userId := data.Get("user_id").(string)
	groupIds := data.Get("group_ids").(*schema.Set)
	exhaustive := data.Get("exhaustive").(bool)

	userGroups, err := keycloakClient.GetUserGroups(ctx, realmId, userId)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	var groups []string
	for _, group := range userGroups {
		//only add groups that we care about
		if exhaustive || groupIds.Contains(group.Id) {
			groups = append(groups, group.Id)
		}
	}

	data.Set("group_ids", groups)
	data.SetId(userGroupsId(realmId, userId))

	return nil
}

func resourceKeycloakUserGroupsReconcile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	userId := data.Get("user_id").(string)
	groupIds := interfaceSliceToStringSlice(data.Get("group_ids").(*schema.Set).List())
	exhaustive := data.Get("exhaustive").(bool)

	if data.HasChange("group_ids") {
		o, n := data.GetChange("group_ids")
		os := o.(*schema.Set)
		ns := n.(*schema.Set)
		remove := interfaceSliceToStringSlice(os.Difference(ns).List())

		if err := keycloakClient.RemoveUserFromGroups(ctx, remove, userId, realmId); err != nil {
			return diag.FromErr(err)
		}
	}

	userGroups, err := keycloakClient.GetUserGroups(ctx, realmId, userId)
	if err != nil {
		return diag.FromErr(err)
	}

	var userGroupsIds []string
	for _, group := range userGroups {
		userGroupsIds = append(userGroupsIds, group.Id)
	}

	remove := stringArrayDifference(userGroupsIds, groupIds)
	add := stringArrayDifference(groupIds, userGroupsIds)

	if err := keycloakClient.AddUserToGroups(ctx, add, userId, realmId); err != nil {
		return diag.FromErr(err)
	}

	if exhaustive {
		if err := keycloakClient.RemoveUserFromGroups(ctx, remove, userId, realmId); err != nil {
			return diag.FromErr(err)
		}
	}

	data.SetId(userGroupsId(realmId, userId))
	return resourceKeycloakUserGroupsRead(ctx, data, meta)
}

func resourceKeycloakUserGroupsDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	userId := data.Get("user_id").(string)
	groupIds := interfaceSliceToStringSlice(data.Get("group_ids").(*schema.Set).List())

	return diag.FromErr(keycloakClient.RemoveUserFromGroups(ctx, groupIds, userId, realmId))
}

func resourceKeycloakUserGroupsImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("Invalid import. Supported import format: {{realm}}/{{userId}}.")
	}

	realmId := parts[0]
	userId := parts[1]

	_, err := keycloakClient.GetUserGroups(ctx, realmId, userId)
	if err != nil {
		return nil, err
	}

	d.Set("realm_id", realmId)
	d.Set("user_id", userId)
	d.Set("exhaustive", true)

	diagnostics := resourceKeycloakUserGroupsRead(ctx, d, meta)
	if diagnostics.HasError() {
		return nil, errors.New(diagnostics[0].Summary)
	}

	return []*schema.ResourceData{d}, nil
}

func userGroupsId(realmId, userId string) string {
	return fmt.Sprintf("%s/%s", realmId, userId)
}
