package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakDefaultGroups() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakDefaultGroupsCreate,
		ReadContext:   resourceKeycloakDefaultGroupsRead,
		UpdateContext: resourceKeycloakDefaultGroupsUpdate,
		DeleteContext: resourceKeycloakDefaultGroupsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceKeycloakDefaultGroupsImport,
		},
		Schema: map[string]*schema.Schema{
			"realm_id": {
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
		},
	}
}

func defaultGroupId(realmId string) string {
	return realmId + "/default-groups"
}

func resourceKeycloakDefaultGroupsCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	groupIds := interfaceSliceToStringSlice(data.Get("group_ids").(*schema.Set).List())

	for _, groupId := range groupIds {
		err := keycloakClient.PutDefaultGroup(ctx, realmId, groupId)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	data.SetId(defaultGroupId(realmId))

	return nil
}

func resourceKeycloakDefaultGroupsRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)

	groups, err := keycloakClient.GetDefaultGroups(ctx, realmId)
	if err != nil {
		return diag.FromErr(err)
	}

	var groupIds []string
	for _, group := range groups {
		groupIds = append(groupIds, group.Id)
	}

	data.SetId(defaultGroupId(realmId))
	data.Set("group_ids", groupIds)

	return nil
}

func resourceKeycloakDefaultGroupsUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	newGroupIds := data.Get("group_ids").(*schema.Set)

	originalGroups, err := keycloakClient.GetDefaultGroups(ctx, realmId)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, originalGroup := range originalGroups {
		if newGroupIds.Contains(originalGroup.Id) {
			newGroupIds.Remove(originalGroup.Id)
		} else {
			err := keycloakClient.DeleteDefaultGroup(ctx, realmId, originalGroup.Id)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	// at this point newGroupIds should contain only users that need to be created
	for _, group := range interfaceSliceToStringSlice(newGroupIds.List()) {
		err := keycloakClient.PutDefaultGroup(ctx, realmId, group)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	data.SetId(defaultGroupId(realmId))

	return resourceKeycloakDefaultGroupsRead(ctx, data, meta)
}

func resourceKeycloakDefaultGroupsDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	groupIds := interfaceSliceToStringSlice(data.Get("group_ids").(*schema.Set).List())

	for _, groupId := range groupIds {
		err := keycloakClient.DeleteDefaultGroup(ctx, realmId, groupId)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceKeycloakDefaultGroupsImport(_ context.Context, data *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	data.Set("realm_id", data.Id())
	data.SetId(data.Id())
	return []*schema.ResourceData{data}, nil
}
