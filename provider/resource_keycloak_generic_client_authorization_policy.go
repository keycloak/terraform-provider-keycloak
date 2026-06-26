package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakGenericClientAuthorizationPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakGenericClientAuthorizationPolicyCreate,
		ReadContext:   resourceKeycloakGenericClientAuthorizationPolicyRead,
		DeleteContext: resourceKeycloakGenericClientAuthorizationPolicyDelete,
		UpdateContext: resourceKeycloakGenericClientAuthorizationPolicyUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: genericResourcePolicyImport,
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
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
		},
	}
}

func getGenericClientAuthorizationPolicyFromData(data *schema.ResourceData) *keycloak.GenericClientAuthorizationPolicy {
	return &keycloak.GenericClientAuthorizationPolicy{
		Id:               data.Id(),
		ResourceServerId: data.Get("resource_server_id").(string),
		RealmId:          data.Get("realm_id").(string),
		DecisionStrategy: data.Get("decision_strategy").(string),
		Logic:            data.Get("logic").(string),
		Name:             data.Get("name").(string),
		Type:             data.Get("type").(string),
		Description:      data.Get("description").(string),
	}
}

func setGenericClientAuthorizationPolicyData(data *schema.ResourceData, policy *keycloak.GenericClientAuthorizationPolicy) {
	data.SetId(policy.Id)

	data.Set("resource_server_id", policy.ResourceServerId)
	data.Set("realm_id", policy.RealmId)
	data.Set("name", policy.Name)
	data.Set("type", policy.Type)
	data.Set("decision_strategy", policy.DecisionStrategy)
	data.Set("logic", policy.Logic)
	data.Set("description", policy.Description)
}

func resourceKeycloakGenericClientAuthorizationPolicyCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	resource := getGenericClientAuthorizationPolicyFromData(data)

	err := keycloakClient.NewGenericClientAuthorizationPolicy(ctx, resource)
	if err != nil {
		return diag.FromErr(err)
	}

	setGenericClientAuthorizationPolicyData(data, resource)

	return resourceKeycloakGenericClientAuthorizationPolicyRead(ctx, data, meta)
}

func resourceKeycloakGenericClientAuthorizationPolicyRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	resourceServerId := data.Get("resource_server_id").(string)
	id := data.Id()

	resource, err := keycloakClient.GetGenericClientAuthorizationPolicy(ctx, realmId, resourceServerId, id)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	setGenericClientAuthorizationPolicyData(data, resource)

	return nil
}

func resourceKeycloakGenericClientAuthorizationPolicyUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	resource := getGenericClientAuthorizationPolicyFromData(data)

	err := keycloakClient.UpdateGenericClientAuthorizationPolicy(ctx, resource)
	if err != nil {
		return diag.FromErr(err)
	}

	setGenericClientAuthorizationPolicyData(data, resource)

	return nil
}

func resourceKeycloakGenericClientAuthorizationPolicyDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	resourceServerId := data.Get("resource_server_id").(string)
	id := data.Id()

	return diag.FromErr(keycloakClient.DeleteGenericClientAuthorizationPolicy(ctx, realmId, resourceServerId, id))
}
