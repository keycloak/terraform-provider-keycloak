package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakOpenidClientAuthorizationRolePolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakOpenidClientAuthorizationRolePolicyCreate,
		ReadContext:   resourceKeycloakOpenidClientAuthorizationRolePolicyRead,
		DeleteContext: resourceKeycloakOpenidClientAuthorizationRolePolicyDelete,
		UpdateContext: resourceKeycloakOpenidClientAuthorizationRolePolicyUpdate,
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
				Optional: true,
			},
			"logic": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(keycloakPolicyLogicTypes, false),
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"role": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"required": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func getOpenidClientAuthorizationRolePolicyResourceFromData(data *schema.ResourceData) *keycloak.OpenidClientAuthorizationRolePolicy {
	var rolesList []keycloak.OpenidClientAuthorizationRole
	if v, ok := data.Get("role").(*schema.Set); ok {
		for _, role := range v.List() {
			roleMap := role.(map[string]interface{})
			tempRole := keycloak.OpenidClientAuthorizationRole{
				Id:       roleMap["id"].(string),
				Required: roleMap["required"].(bool),
			}
			rolesList = append(rolesList, tempRole)
		}
	}

	resource := keycloak.OpenidClientAuthorizationRolePolicy{
		Id:               data.Id(),
		ResourceServerId: data.Get("resource_server_id").(string),
		RealmId:          data.Get("realm_id").(string),
		DecisionStrategy: data.Get("decision_strategy").(string),
		Logic:            data.Get("logic").(string),
		Name:             data.Get("name").(string),
		Type:             "role",
		Roles:            rolesList,
		Description:      data.Get("description").(string),
	}

	return &resource
}

func setOpenidClientAuthorizationRolePolicyResourceData(data *schema.ResourceData, policy *keycloak.OpenidClientAuthorizationRolePolicy) {
	data.SetId(policy.Id)

	data.Set("resource_server_id", policy.ResourceServerId)
	data.Set("realm_id", policy.RealmId)
	data.Set("name", policy.Name)
	data.Set("decision_strategy", policy.DecisionStrategy)
	data.Set("logic", policy.Logic)
	data.Set("type", policy.Type)
	data.Set("description", policy.Description)

	var roles []interface{}
	for _, r := range policy.Roles {
		role := map[string]interface{}{
			"id":       r.Id,
			"required": r.Required,
		}

		roles = append(roles, role)
	}

	data.Set("role", roles)
}

func resourceKeycloakOpenidClientAuthorizationRolePolicyCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	resource := getOpenidClientAuthorizationRolePolicyResourceFromData(data)

	err := keycloakClient.NewOpenidClientAuthorizationRolePolicy(ctx, resource)
	if err != nil {
		return diag.FromErr(err)
	}

	setOpenidClientAuthorizationRolePolicyResourceData(data, resource)

	return resourceKeycloakOpenidClientAuthorizationRolePolicyRead(ctx, data, meta)
}

func resourceKeycloakOpenidClientAuthorizationRolePolicyRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	resourceServerId := data.Get("resource_server_id").(string)
	id := data.Id()

	resource, err := keycloakClient.GetOpenidClientAuthorizationRolePolicy(ctx, realmId, resourceServerId, id)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	setOpenidClientAuthorizationRolePolicyResourceData(data, resource)

	return nil
}

func resourceKeycloakOpenidClientAuthorizationRolePolicyUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	resource := getOpenidClientAuthorizationRolePolicyResourceFromData(data)

	err := keycloakClient.UpdateOpenidClientAuthorizationRolePolicy(ctx, resource)
	if err != nil {
		return diag.FromErr(err)
	}

	setOpenidClientAuthorizationRolePolicyResourceData(data, resource)

	return nil
}

func resourceKeycloakOpenidClientAuthorizationRolePolicyDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	resourceServerId := data.Get("resource_server_id").(string)
	id := data.Id()

	return diag.FromErr(keycloakClient.DeleteOpenidClientAuthorizationRolePolicy(ctx, realmId, resourceServerId, id))
}
