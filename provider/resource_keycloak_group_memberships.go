package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakGroupMemberships() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakGroupMembershipsCreate,
		ReadContext:   resourceKeycloakGroupMembershipsRead,
		DeleteContext: resourceKeycloakGroupMembershipsDelete,
		UpdateContext: resourceKeycloakGroupMembershipsUpdate,
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"group_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"members": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Required: true,
			},
		},
	}
}

func resourceKeycloakGroupMembershipsCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	groupId := data.Get("group_id").(string)
	members := data.Get("members").(*schema.Set).List()
	realmId := data.Get("realm_id").(string)

	err := keycloakClient.ValidateGroupMembers(members)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.AddUsersToGroup(ctx, realmId, groupId, members)
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(groupMembershipsId(realmId, groupId))

	return resourceKeycloakGroupMembershipsRead(ctx, data, meta)
}

func resourceKeycloakGroupMembershipsRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	groupId := data.Get("group_id").(string)

	usersInGroup, err := keycloakClient.GetGroupMembers(ctx, realmId, groupId)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	var members []string
	for _, userInGroup := range usersInGroup {
		members = append(members, userInGroup.Username)
	}

	data.Set("members", members)
	data.SetId(groupMembershipsId(realmId, groupId))

	return nil
}

func resourceKeycloakGroupMembershipsUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	groupId := data.Get("group_id").(string)
	tfMembers := data.Get("members").(*schema.Set)

	err := keycloakClient.ValidateGroupMembers(tfMembers.List())
	if err != nil {
		return diag.FromErr(err)
	}

	keycloakMembers, err := keycloakClient.GetGroupMembers(ctx, realmId, groupId)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, keycloakMember := range keycloakMembers {
		if tfMembers.Contains(keycloakMember.Username) {
			// if the user exists in keycloak and tf state, no update is required for this member
			// remove them from the set so we can look at members that need to be added later
			tfMembers.Remove(keycloakMember.Username)
		} else {
			// if the user exists in keycloak and not in tf state, they need to be removed from the group
			err = keycloakClient.RemoveUserFromGroup(ctx, keycloakMember, groupId)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	// at this point, `tfMembers` should only contain users that exist in tf state but not keycloak. these users need to be added
	err = keycloakClient.AddUsersToGroup(ctx, realmId, groupId, tfMembers.List())
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(groupMembershipsId(realmId, groupId))

	return resourceKeycloakGroupMembershipsRead(ctx, data, meta)
}

func resourceKeycloakGroupMembershipsDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	groupId := data.Get("group_id").(string)

	return diag.FromErr(keycloakClient.RemoveUsersFromGroup(ctx, realmId, groupId, data.Get("members").(*schema.Set).List()))
}

func groupMembershipsId(realmId, groupId string) string {
	return fmt.Sprintf("%s/group-memberships/%s", realmId, groupId)
}
