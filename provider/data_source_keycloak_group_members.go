package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mrparkers/terraform-provider-keycloak/keycloak"
)

func dataSourceKeycloakGroupMembers() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKeycloakGroupMembersRead,
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"users": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
		},
	}
}

func dataSourceKeycloakGroupMembersRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	groupName := data.Get("name").(string)

	group, err := keycloakClient.GetGroupByName(ctx, realmId, groupName)
	if err != nil {
		return diag.FromErr(err)
	}

	users, err := keycloakClient.GetGroupMembers(ctx, realmId, group.Id)
	if err != nil {
		return diag.FromErr(err)
	}

	usernames := make([]string, len(users))
	for num, user := range users {
		usernames[num] = user.Username
	}

	data.SetId(groupName)
	data.Set("realm_id", realmId)
	data.Set("name", groupName)
	data.Set("users", usernames)

	return nil
}
