package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

var groupAdminPermissionScopes = []string{
	"view",
	"manage",
	"view-members",
	"manage-members",
	"manage-membership",
}

func resourceKeycloakGroupAdminPermissions() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakGroupAdminPermissionsCreate,
		ReadContext:   resourceKeycloakGroupAdminPermissionsRead,
		DeleteContext: resourceKeycloakGroupAdminPermissionsDelete,
		UpdateContext: resourceKeycloakGroupAdminPermissionsUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"decision_strategy": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "UNANIMOUS",
				ValidateFunc: validation.StringInSlice(keycloakOpenidClientResourcePermissionDecisionStrategies, false),
			},
			"group_ids": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Group UUIDs this permission applies to. Leave empty to target all groups in the realm.",
			},
			"scopes": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(groupAdminPermissionScopes, false),
				},
			},
			"policies": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"permission_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"authorization_resource_server_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Resource server id of the admin-permissions client on which this permission is managed",
			},
		},
	}
}

func resourceKeycloakGroupAdminPermissionsCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	kc := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	apClientId, err := kc.GetAdminPermissionsClientId(ctx, realmId)
	if err != nil {
		return diag.FromErr(err)
	}

	perm := buildGroupAdminPermission(data, realmId, apClientId, "")
	permId, err := createOrAdoptFGAPv2Permission(ctx, kc, perm)
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(realmId + "/" + permId)
	data.Set("permission_id", permId)
	return resourceKeycloakGroupAdminPermissionsRead(ctx, data, meta)
}

func resourceKeycloakGroupAdminPermissionsUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	kc := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	apClientId, err := kc.GetAdminPermissionsClientId(ctx, realmId)
	if err != nil {
		return diag.FromErr(err)
	}

	permId := data.Get("permission_id").(string)
	perm := buildGroupAdminPermission(data, realmId, apClientId, permId)
	if err := kc.UpdateOpenidClientAuthorizationPermission(ctx, perm); err != nil {
		return diag.FromErr(err)
	}

	return resourceKeycloakGroupAdminPermissionsRead(ctx, data, meta)
}

func resourceKeycloakGroupAdminPermissionsRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	kc := meta.(*keycloak.KeycloakClient)

	parts := strings.SplitN(data.Id(), "/", 2)
	if len(parts) != 2 {
		return diag.Errorf("invalid resource id %q, expected realmId/permissionId", data.Id())
	}
	realmId, permId := parts[0], parts[1]

	apClientId, err := kc.GetAdminPermissionsClientId(ctx, realmId)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	perm, err := readFGAPv2ScopePermission(ctx, kc, realmId, apClientId, permId)
	if err != nil {
		return diag.FromErr(err)
	}
	if perm == nil {
		data.SetId("")
		return nil
	}

	ds := perm.DecisionStrategy
	if ds == "" {
		ds = "UNANIMOUS"
	}

	data.Set("realm_id", realmId)
	data.Set("permission_id", permId)
	data.Set("name", perm.Name)
	data.Set("description", perm.Description)
	data.Set("decision_strategy", ds)
	// group_ids are not re-read: Keycloak's resources sub-endpoint returns internal
	// authorization resource UUIDs that differ from the group UUIDs in the config.
	data.Set("scopes", schema.NewSet(schema.HashString, stringsToInterfaces(perm.Scopes)))
	data.Set("policies", schema.NewSet(schema.HashString, stringsToInterfaces(perm.Policies)))
	data.Set("enabled", true)
	data.Set("authorization_resource_server_id", apClientId)

	return nil
}

func resourceKeycloakGroupAdminPermissionsDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	kc := meta.(*keycloak.KeycloakClient)

	parts := strings.SplitN(data.Id(), "/", 2)
	if len(parts) != 2 {
		return nil
	}
	realmId, permId := parts[0], parts[1]

	apClientId, err := kc.GetAdminPermissionsClientId(ctx, realmId)
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(deleteFGAPv2Permission(ctx, kc, realmId, apClientId, permId))
}

func buildGroupAdminPermission(data *schema.ResourceData, realmId, apClientId, permId string) *keycloak.OpenidClientAuthorizationPermission {
	return &keycloak.OpenidClientAuthorizationPermission{
		Id:               permId,
		RealmId:          realmId,
		ResourceServerId: apClientId,
		Name:             data.Get("name").(string),
		Description:      data.Get("description").(string),
		DecisionStrategy: data.Get("decision_strategy").(string),
		Type:             "scope",
		ResourceType:     "Groups",
		Resources:        setToStringSlice(data.Get("group_ids").(*schema.Set)),
		Scopes:           setToStringSlice(data.Get("scopes").(*schema.Set)),
		Policies:         setToStringSlice(data.Get("policies").(*schema.Set)),
	}
}
