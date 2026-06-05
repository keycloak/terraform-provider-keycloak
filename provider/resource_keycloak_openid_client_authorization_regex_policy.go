package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakOpenidClientAuthorizationRegexPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakOpenidClientAuthorizationRegexPolicyCreate,
		ReadContext:   resourceKeycloakOpenidClientAuthorizationRegexPolicyRead,
		DeleteContext: resourceKeycloakOpenidClientAuthorizationRegexPolicyDelete,
		UpdateContext: resourceKeycloakOpenidClientAuthorizationRegexPolicyUpdate,
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
			"target_claim": {
				Type:     schema.TypeString,
				Required: true,
			},
			"pattern": {
				Type:     schema.TypeString,
				Required: true,
			},
			"target_context_attributes": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func getOpenidClientAuthorizationRegexPolicyResourceFromData(data *schema.ResourceData) *keycloak.OpenidClientAuthorizationRegexPolicy {

	resource := keycloak.OpenidClientAuthorizationRegexPolicy{
		Id:                      data.Id(),
		ResourceServerId:        data.Get("resource_server_id").(string),
		RealmId:                 data.Get("realm_id").(string),
		DecisionStrategy:        data.Get("decision_strategy").(string),
		Logic:                   data.Get("logic").(string),
		Name:                    data.Get("name").(string),
		Type:                    "regex",
		Pattern:                 data.Get("pattern").(string),
		Description:             data.Get("description").(string),
		TargetClaim:             data.Get("target_claim").(string),
		TargetContextAttributes: data.Get("target_context_attributes").(bool),
	}
	return &resource
}

func setOpenidClientAuthorizationRegexPolicyResourceData(data *schema.ResourceData, policy *keycloak.OpenidClientAuthorizationRegexPolicy) {
	data.SetId(policy.Id)

	data.Set("resource_server_id", policy.ResourceServerId)
	data.Set("realm_id", policy.RealmId)
	data.Set("name", policy.Name)
	data.Set("decision_strategy", policy.DecisionStrategy)
	data.Set("logic", policy.Logic)
	data.Set("description", policy.Description)
	data.Set("pattern", policy.Pattern)
	data.Set("target_claim", policy.TargetClaim)
	data.Set("target_context_attributes", policy.TargetContextAttributes)
}

func resourceKeycloakOpenidClientAuthorizationRegexPolicyCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	resource := getOpenidClientAuthorizationRegexPolicyResourceFromData(data)

	err := keycloakClient.NewOpenidClientAuthorizationRegexPolicy(ctx, resource)
	if err != nil {
		return diag.FromErr(err)
	}

	setOpenidClientAuthorizationRegexPolicyResourceData(data, resource)

	return resourceKeycloakOpenidClientAuthorizationRegexPolicyRead(ctx, data, meta)
}

func resourceKeycloakOpenidClientAuthorizationRegexPolicyRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	resourceServerId := data.Get("resource_server_id").(string)
	id := data.Id()

	resource, err := keycloakClient.GetOpenidClientAuthorizationRegexPolicy(ctx, realmId, resourceServerId, id)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	setOpenidClientAuthorizationRegexPolicyResourceData(data, resource)

	return nil
}

func resourceKeycloakOpenidClientAuthorizationRegexPolicyUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	resource := getOpenidClientAuthorizationRegexPolicyResourceFromData(data)

	err := keycloakClient.UpdateOpenidClientAuthorizationRegexPolicy(ctx, resource)
	if err != nil {
		return diag.FromErr(err)
	}

	setOpenidClientAuthorizationRegexPolicyResourceData(data, resource)

	return nil
}

func resourceKeycloakOpenidClientAuthorizationRegexPolicyDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	resourceServerId := data.Get("resource_server_id").(string)
	id := data.Id()

	return diag.FromErr(keycloakClient.DeleteOpenidClientAuthorizationRegexPolicy(ctx, realmId, resourceServerId, id))
}
