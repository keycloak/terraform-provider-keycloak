package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

var (
	keycloakOpenidClientPermissionTypes = []string{"resource", "scope"}
)

func resourceKeycloakOpenidClientAuthorizationPermission() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakOpenidClientAuthorizationPermissionCreate,
		ReadContext:   resourceKeycloakOpenidClientAuthorizationPermissionRead,
		DeleteContext: resourceKeycloakOpenidClientAuthorizationPermissionDelete,
		UpdateContext: resourceKeycloakOpenidClientAuthorizationPermissionUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: resourceKeycloakOpenidClientAuthorizationPermissionImport,
		},
		Schema: map[string]*schema.Schema{
			"resource_server_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
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
				ValidateFunc: validation.StringInSlice(keycloakOpenidClientResourcePermissionDecisionStrategies, false),
				Default:      "UNANIMOUS",
			},
			"policies": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"resources": {
				Type:          schema.TypeSet,
				Elem:          &schema.Schema{Type: schema.TypeString},
				Optional:      true,
				ConflictsWith: []string{"resource_type"},
			},
			"resource_type": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"resources"},
			},
			"scopes": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "resource",
				ValidateFunc: validation.StringInSlice(keycloakOpenidClientPermissionTypes, false),
			},
		},
	}
}

func getOpenidClientAuthorizationPermissionFromData(data *schema.ResourceData) *keycloak.OpenidClientAuthorizationPermission {
	var policies []string
	var resources []string
	var scopes []string
	if v, ok := data.GetOk("resources"); ok {
		for _, resource := range v.(*schema.Set).List() {
			resources = append(resources, resource.(string))
		}
	}
	if v, ok := data.GetOk("policies"); ok {
		for _, policy := range v.(*schema.Set).List() {
			policies = append(policies, policy.(string))
		}
	}
	if v, ok := data.GetOk("scopes"); ok {
		for _, scope := range v.(*schema.Set).List() {
			scopes = append(scopes, scope.(string))
		}
	}

	permission := keycloak.OpenidClientAuthorizationPermission{
		Id:               data.Id(),
		ResourceServerId: data.Get("resource_server_id").(string),
		RealmId:          data.Get("realm_id").(string),
		Description:      data.Get("description").(string),
		Name:             data.Get("name").(string),
		DecisionStrategy: data.Get("decision_strategy").(string),
		Type:             data.Get("type").(string),
		Policies:         policies,
		Scopes:           scopes,
		Resources:        resources,
		ResourceType:     data.Get("resource_type").(string),
	}
	return &permission
}

func setOpenidClientAuthorizationPermissionData(data *schema.ResourceData, permission *keycloak.OpenidClientAuthorizationPermission) {
	data.SetId(permission.Id)
	data.Set("resource_server_id", permission.ResourceServerId)
	data.Set("realm_id", permission.RealmId)
	data.Set("description", permission.Description)
	data.Set("name", permission.Name)
	data.Set("decision_strategy", permission.DecisionStrategy)
	data.Set("type", permission.Type)
	data.Set("policies", permission.Policies)
	data.Set("scopes", permission.Scopes)
	data.Set("resources", permission.Resources)
	data.Set("resource_type", permission.ResourceType)
}

func resourceKeycloakOpenidClientAuthorizationPermissionCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	permission := getOpenidClientAuthorizationPermissionFromData(data)

	err := keycloakClient.NewOpenidClientAuthorizationPermission(ctx, permission)
	if err != nil {
		return diag.FromErr(err)
	}

	setOpenidClientAuthorizationPermissionData(data, permission)

	return resourceKeycloakOpenidClientAuthorizationPermissionRead(ctx, data, meta)
}

func resourceKeycloakOpenidClientAuthorizationPermissionRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	resourceServerId := data.Get("resource_server_id").(string)
	id := data.Id()

	permission, err := keycloakClient.GetOpenidClientAuthorizationPermission(ctx, realmId, resourceServerId, id)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	setOpenidClientAuthorizationPermissionData(data, permission)

	return nil
}

func resourceKeycloakOpenidClientAuthorizationPermissionUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	permission := getOpenidClientAuthorizationPermissionFromData(data)

	err := keycloakClient.UpdateOpenidClientAuthorizationPermission(ctx, permission)
	if err != nil {
		return diag.FromErr(err)
	}

	setOpenidClientAuthorizationPermissionData(data, permission)

	return nil
}

func resourceKeycloakOpenidClientAuthorizationPermissionDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	resourceServerId := data.Get("resource_server_id").(string)
	id := data.Id()

	return diag.FromErr(keycloakClient.DeleteOpenidClientAuthorizationPermission(ctx, realmId, resourceServerId, id))
}

func resourceKeycloakOpenidClientAuthorizationPermissionImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 3 {
		return nil, fmt.Errorf("Invalid import. Supported import formats: {{realmId}}/{{resourceServerId}}/{{permissionId}}")
	}
	d.Set("realm_id", parts[0])
	d.Set("resource_server_id", parts[1])
	d.SetId(parts[2])

	return []*schema.ResourceData{d}, nil
}
