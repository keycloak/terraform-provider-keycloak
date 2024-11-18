package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakOpenidClientAuthorizationGroupPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakOpenidClientAuthorizationGroupPolicyCreate,
		ReadContext:   resourceKeycloakOpenidClientAuthorizationGroupPolicyRead,
		DeleteContext: resourceKeycloakOpenidClientAuthorizationGroupPolicyDelete,
		UpdateContext: resourceKeycloakOpenidClientAuthorizationGroupPolicyUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: genericResourcePolicyImport,
		},
		Schema: map[string]*schema.Schema{
			"resource_server_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"decision_strategy": {
				Type:     schema.TypeString,
				Required: true,
			},
			"logic": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(keycloakPolicyLogicTypes, false),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"groups_claim": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"groups": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"path": {
							Type:     schema.TypeString,
							Required: true,
						},
						"extend_children": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func getOpenidClientAuthorizationGroupPolicyResourceFromData(data *schema.ResourceData) *keycloak.OpenidClientAuthorizationGroupPolicy {
	var groups []keycloak.OpenidClientAuthorizationGroup
	if v, ok := data.Get("groups").([]interface{}); ok {
		for _, group := range v {
			groupMap := group.(map[string]interface{})
			tempGroup := keycloak.OpenidClientAuthorizationGroup{
				Id:             groupMap["id"].(string),
				Path:           groupMap["path"].(string),
				ExtendChildren: groupMap["extend_children"].(bool),
			}
			groups = append(groups, tempGroup)
		}
	}

	resource := keycloak.OpenidClientAuthorizationGroupPolicy{
		Id:               data.Id(),
		ResourceServerId: data.Get("resource_server_id").(string),
		RealmId:          data.Get("realm_id").(string),
		DecisionStrategy: data.Get("decision_strategy").(string),
		Logic:            data.Get("logic").(string),
		Name:             data.Get("name").(string),
		Type:             "group",
		GroupsClaim:      data.Get("groups_claim").(string),
		Groups:           groups,
		Description:      data.Get("description").(string),
	}

	return &resource
}

func setOpenidClientAuthorizationGroupPolicyResourceData(ctx context.Context, keycloakClient *keycloak.KeycloakClient, policy *keycloak.OpenidClientAuthorizationGroupPolicy, data *schema.ResourceData) error {
	data.SetId(policy.Id)

	data.Set("resource_server_id", policy.ResourceServerId)
	data.Set("realm_id", policy.RealmId)
	data.Set("name", policy.Name)
	data.Set("decision_strategy", policy.DecisionStrategy)
	data.Set("logic", policy.Logic)
	data.Set("description", policy.Description)
	data.Set("groups_claim", policy.GroupsClaim)

	var groups []interface{}
	for _, g := range policy.Groups {
		// the "path" attribute is omitted by Keycloak, so we have to look this group up ourselves to get the path
		group, err := keycloakClient.GetGroup(ctx, policy.RealmId, g.Id)
		if err != nil {
			return err
		}

		groups = append(groups, map[string]interface{}{
			"id":              g.Id,
			"path":            group.Path,
			"extend_children": g.ExtendChildren,
		})
	}

	data.Set("groups", groups)

	return nil
}

func resourceKeycloakOpenidClientAuthorizationGroupPolicyCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	resource := getOpenidClientAuthorizationGroupPolicyResourceFromData(data)

	err := keycloakClient.NewOpenidClientAuthorizationGroupPolicy(ctx, resource)
	if err != nil {
		return diag.FromErr(err)
	}

	err = setOpenidClientAuthorizationGroupPolicyResourceData(ctx, keycloakClient, resource, data)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceKeycloakOpenidClientAuthorizationGroupPolicyRead(ctx, data, meta)
}

func resourceKeycloakOpenidClientAuthorizationGroupPolicyRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	resourceServerId := data.Get("resource_server_id").(string)
	id := data.Id()

	resource, err := keycloakClient.GetOpenidClientAuthorizationGroupPolicy(ctx, realmId, resourceServerId, id)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	err = setOpenidClientAuthorizationGroupPolicyResourceData(ctx, keycloakClient, resource, data)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakOpenidClientAuthorizationGroupPolicyUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	resource := getOpenidClientAuthorizationGroupPolicyResourceFromData(data)

	err := keycloakClient.UpdateOpenidClientAuthorizationGroupPolicy(ctx, resource)
	if err != nil {
		return diag.FromErr(err)
	}

	err = setOpenidClientAuthorizationGroupPolicyResourceData(ctx, keycloakClient, resource, data)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakOpenidClientAuthorizationGroupPolicyDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	resourceServerId := data.Get("resource_server_id").(string)
	id := data.Id()

	return diag.FromErr(keycloakClient.DeleteOpenidClientAuthorizationGroupPolicy(ctx, realmId, resourceServerId, id))
}
