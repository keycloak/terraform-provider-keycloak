package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func dataSourceKeycloakGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKeycloakGroupRead,
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"group_path": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"parent_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"path": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"attributes": {
				Type:     schema.TypeMap,
				Computed: true,
			},
		},
	}
}

func dataSourceKeycloakGroupRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	name := data.Get("name").(string)
	groupPath := data.Get("group_path").(string)

	if name == "" && groupPath == "" {
		return diag.Errorf("one of `name` or `group_path` must be specified")
	}
	if name != "" && groupPath != "" {
		return diag.Errorf("only one of `name` or `group_path` may be specified")
	}

	var group *keycloak.Group
	var err error
	if groupPath != "" {
		group, err = keycloakClient.GetGroupByPath(ctx, realmId, groupPath)
	} else {
		group, err = keycloakClient.GetGroupByName(ctx, realmId, name)
	}

	if err != nil {
		return diag.FromErr(err)
	}

	mapFromGroupToData(data, group)

	return nil
}
