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
			"organization_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"group_path"},
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"name", "group_path"},
			},
			// group_path enables lookup of nested groups by their full hierarchy path
			// (e.g. "/parent/child/subgroup"). It is separate from `name` to avoid
			// overloading semantics and follows Terraform's convention of one field
			// per concern. Uses the Keycloak /group-by-path endpoint for deterministic,
			// unambiguous resolution regardless of name collisions.
			"group_path": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"name", "group_path"},
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
	organizationId := data.Get("organization_id").(string)
	groupPath := data.Get("group_path").(string)
	groupName := data.Get("name").(string)

	if groupName == "" && groupPath == "" {
		return diag.Errorf("one of `name` or `group_path` must be specified")
	}
	if groupPath != "" && organizationId != "" {
		return diag.Errorf("group_path and organization_id cannot be set together: path-based lookups use the realm-level /group-by-path endpoint and do not support organization scoping")
	}

	var group *keycloak.Group
	var err error
	// group_path is set → precise path-based lookup via /group-by-path endpoint.
	// name is set → legacy name-based lookup (may be ambiguous for nested groups).
	if groupPath != "" {
		group, err = keycloakClient.GetGroupByPath(ctx, realmId, groupPath)
	} else {
		group, err = keycloakClient.GetOrganizationGroupByName(ctx, realmId, organizationId, groupName)
	}
	if err != nil {
		return diag.FromErr(err)
	}

	mapFromGroupToData(data, group)

	return nil
}
