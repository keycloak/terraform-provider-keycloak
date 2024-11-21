package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakOpenidClientAuthorizationClientPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakOpenidClientAuthorizationClientPolicyCreate,
		ReadContext:   resourceKeycloakOpenidClientAuthorizationClientPolicyRead,
		DeleteContext: resourceKeycloakOpenidClientAuthorizationClientPolicyDelete,
		UpdateContext: resourceKeycloakOpenidClientAuthorizationClientPolicyUpdate,
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
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"clients": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Required: true,
			},
		},
	}
}

func getOpenidClientAuthorizationClientAuthorizationClientPolicyResourceFromData(data *schema.ResourceData) *keycloak.OpenidClientAuthorizationClientPolicy {
	var clients []string

	if v, ok := data.GetOk("clients"); ok {
		for _, client := range v.(*schema.Set).List() {
			clients = append(clients, client.(string))
		}
	}

	resource := keycloak.OpenidClientAuthorizationClientPolicy{
		Id:               data.Id(),
		ResourceServerId: data.Get("resource_server_id").(string),
		RealmId:          data.Get("realm_id").(string),
		DecisionStrategy: data.Get("decision_strategy").(string),
		Logic:            data.Get("logic").(string),
		Name:             data.Get("name").(string),
		Type:             "client",
		Clients:          clients,
		Description:      data.Get("description").(string),
	}
	return &resource
}

func setOpenidClientAuthorizationClientAuthorizationClientPolicyResourceData(data *schema.ResourceData, policy *keycloak.OpenidClientAuthorizationClientPolicy) {
	data.SetId(policy.Id)

	data.Set("resource_server_id", policy.ResourceServerId)
	data.Set("realm_id", policy.RealmId)
	data.Set("name", policy.Name)
	data.Set("decision_strategy", policy.DecisionStrategy)
	data.Set("logic", policy.Logic)
	data.Set("description", policy.Description)
	data.Set("clients", policy.Clients)
}

func resourceKeycloakOpenidClientAuthorizationClientPolicyCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	resource := getOpenidClientAuthorizationClientAuthorizationClientPolicyResourceFromData(data)

	err := keycloakClient.NewOpenidClientAuthorizationClientPolicy(ctx, resource)
	if err != nil {
		return diag.FromErr(err)
	}

	setOpenidClientAuthorizationClientAuthorizationClientPolicyResourceData(data, resource)

	return resourceKeycloakOpenidClientAuthorizationClientPolicyRead(ctx, data, meta)
}

func resourceKeycloakOpenidClientAuthorizationClientPolicyRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	resourceServerId := data.Get("resource_server_id").(string)
	id := data.Id()

	resource, err := keycloakClient.GetOpenidClientAuthorizationClientPolicy(ctx, realmId, resourceServerId, id)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	setOpenidClientAuthorizationClientAuthorizationClientPolicyResourceData(data, resource)

	return nil
}

func resourceKeycloakOpenidClientAuthorizationClientPolicyUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	resource := getOpenidClientAuthorizationClientAuthorizationClientPolicyResourceFromData(data)

	err := keycloakClient.UpdateOpenidClientAuthorizationClientPolicy(ctx, resource)
	if err != nil {
		return diag.FromErr(err)
	}

	setOpenidClientAuthorizationClientAuthorizationClientPolicyResourceData(data, resource)

	return nil
}

func resourceKeycloakOpenidClientAuthorizationClientPolicyDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	resourceServerId := data.Get("resource_server_id").(string)
	id := data.Id()

	return diag.FromErr(keycloakClient.DeleteOpenidClientAuthorizationClientPolicy(ctx, realmId, resourceServerId, id))
}
