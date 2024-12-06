package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakOpenidClientAuthorizationAggregatePolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakOpenidClientAuthorizationAggregatePolicyCreate,
		ReadContext:   resourceKeycloakOpenidClientAuthorizationAggregatePolicyRead,
		DeleteContext: resourceKeycloakOpenidClientAuthorizationAggregatePolicyDelete,
		UpdateContext: resourceKeycloakOpenidClientAuthorizationAggregatePolicyUpdate,
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
			"policies": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func getOpenidClientAuthorizationAggregatePolicyResourceFromData(data *schema.ResourceData) *keycloak.OpenidClientAuthorizationAggregatePolicy {
	var policies []string

	if v, ok := data.GetOk("policies"); ok {
		for _, policy := range v.(*schema.Set).List() {
			policies = append(policies, policy.(string))
		}
	}

	resource := keycloak.OpenidClientAuthorizationAggregatePolicy{
		Id:               data.Id(),
		ResourceServerId: data.Get("resource_server_id").(string),
		RealmId:          data.Get("realm_id").(string),
		DecisionStrategy: data.Get("decision_strategy").(string),
		Logic:            data.Get("logic").(string),
		Name:             data.Get("name").(string),
		Type:             "aggregate",
		Policies:         policies,
		Description:      data.Get("description").(string),
	}
	return &resource
}

func setOpenidClientAuthorizationAggregatePolicyResourceData(data *schema.ResourceData, policy *keycloak.OpenidClientAuthorizationAggregatePolicy) {
	data.SetId(policy.Id)

	data.Set("resource_server_id", policy.ResourceServerId)
	data.Set("realm_id", policy.RealmId)
	data.Set("name", policy.Name)
	data.Set("decision_strategy", policy.DecisionStrategy)
	data.Set("logic", policy.Logic)
	data.Set("policies", policy.Policies)
	data.Set("description", policy.Description)
}

func resourceKeycloakOpenidClientAuthorizationAggregatePolicyCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	resource := getOpenidClientAuthorizationAggregatePolicyResourceFromData(data)
	err := keycloakClient.NewOpenidClientAuthorizationAggregatePolicy(ctx, resource)
	if err != nil {
		return diag.FromErr(err)
	}

	setOpenidClientAuthorizationAggregatePolicyResourceData(data, resource)

	return resourceKeycloakOpenidClientAuthorizationAggregatePolicyRead(ctx, data, meta)
}

func resourceKeycloakOpenidClientAuthorizationAggregatePolicyRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	resourceServerId := data.Get("resource_server_id").(string)
	id := data.Id()

	resource, err := keycloakClient.GetOpenidClientAuthorizationAggregatePolicy(ctx, realmId, resourceServerId, id)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	setOpenidClientAuthorizationAggregatePolicyResourceData(data, resource)

	return nil
}

func resourceKeycloakOpenidClientAuthorizationAggregatePolicyUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	resource := getOpenidClientAuthorizationAggregatePolicyResourceFromData(data)

	err := keycloakClient.UpdateOpenidClientAuthorizationAggregatePolicy(ctx, resource)
	if err != nil {
		return diag.FromErr(err)
	}

	setOpenidClientAuthorizationAggregatePolicyResourceData(data, resource)

	return nil
}

func resourceKeycloakOpenidClientAuthorizationAggregatePolicyDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	resourceServerId := data.Get("resource_server_id").(string)
	id := data.Id()

	return diag.FromErr(keycloakClient.DeleteOpenidClientAuthorizationAggregatePolicy(ctx, realmId, resourceServerId, id))
}
