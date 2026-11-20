package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakOpenidClientAuthorizationClientScopePolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakOpenidClientAuthorizationClientScopePolicyCreate,
		ReadContext:   resourceKeycloakOpenidClientAuthorizationClientScopePolicyRead,
		DeleteContext: resourceKeycloakOpenidClientAuthorizationClientScopePolicyDelete,
		UpdateContext: resourceKeycloakOpenidClientAuthorizationClientScopePolicyUpdate,
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
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(keycloakOpenidClientResourcePermissionDecisionStrategies, false),
				Default:      "UNANIMOUS",
			},
			"logic": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(keycloakPolicyLogicTypes, false),
				Default:      "POSITIVE",
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"scope": {
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
							Optional: true,
							Default:  false,
						},
					},
				},
			},
		},
	}
}

func getOpenidClientAuthorizationClientScopePolicyResourceFromData(data *schema.ResourceData) *keycloak.OpenidClientAuthorizationClientScopePolicy {
	var clientScopes []keycloak.OpenidClientAuthorizationClientScope

	if v, ok := data.Get("scope").(*schema.Set); ok {
		for _, clientScope := range v.List() { // Use List() for TypeSet
			clientScopeMap := clientScope.(map[string]interface{})
			clientScopes = append(clientScopes, keycloak.OpenidClientAuthorizationClientScope{
				Id:       clientScopeMap["id"].(string),
				Required: clientScopeMap["required"].(bool),
			})
		}
	}

	resource := keycloak.OpenidClientAuthorizationClientScopePolicy{
		Id:               data.Id(),
		ResourceServerId: data.Get("resource_server_id").(string),
		RealmId:          data.Get("realm_id").(string),
		DecisionStrategy: data.Get("decision_strategy").(string),
		Logic:            data.Get("logic").(string),
		Name:             data.Get("name").(string),
		Type:             "client-scope",
		Scope:            clientScopes,
		Description:      data.Get("description").(string),
	}

	return &resource
}

func setOpenidClientAuthorizationClientScopePolicyResourceData(ctx context.Context, keycloakClient *keycloak.KeycloakClient, policy *keycloak.OpenidClientAuthorizationClientScopePolicy, data *schema.ResourceData) error {
	data.SetId(policy.Id)

	data.Set("resource_server_id", policy.ResourceServerId)
	data.Set("realm_id", policy.RealmId)
	data.Set("name", policy.Name)
	data.Set("decision_strategy", policy.DecisionStrategy)
	data.Set("logic", policy.Logic)
	data.Set("description", policy.Description)

	// Convert scope slice to a set for consistent Terraform state
	clientScopesSet := make([]interface{}, len(policy.Scope))
	for i, g := range policy.Scope {
		clientScopesSet[i] = map[string]interface{}{
			"id":       g.Id,
			"required": g.Required,
		}
	}

	if err := data.Set("scope", clientScopesSet); err != nil {
		return err
	}

	return nil
}

func resourceKeycloakOpenidClientAuthorizationClientScopePolicyCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	resource := getOpenidClientAuthorizationClientScopePolicyResourceFromData(data)

	err := keycloakClient.NewOpenidClientAuthorizationClientScopePolicy(ctx, resource)
	if err != nil {
		return diag.FromErr(err)
	}

	err = setOpenidClientAuthorizationClientScopePolicyResourceData(ctx, keycloakClient, resource, data)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceKeycloakOpenidClientAuthorizationClientScopePolicyRead(ctx, data, meta)
}

func resourceKeycloakOpenidClientAuthorizationClientScopePolicyRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	resourceServerId := data.Get("resource_server_id").(string)
	id := data.Id()

	resource, err := keycloakClient.GetOpenidClientAuthorizationClientScopePolicy(ctx, realmId, resourceServerId, id)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	err = setOpenidClientAuthorizationClientScopePolicyResourceData(ctx, keycloakClient, resource, data)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakOpenidClientAuthorizationClientScopePolicyUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	resource := getOpenidClientAuthorizationClientScopePolicyResourceFromData(data)

	err := keycloakClient.UpdateOpenidClientAuthorizationClientScopePolicy(ctx, resource)
	if err != nil {
		return diag.FromErr(err)
	}

	err = setOpenidClientAuthorizationClientScopePolicyResourceData(ctx, keycloakClient, resource, data)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakOpenidClientAuthorizationClientScopePolicyDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	resourceServerId := data.Get("resource_server_id").(string)
	id := data.Id()

	return diag.FromErr(keycloakClient.DeleteOpenidClientAuthorizationClientScopePolicy(ctx, realmId, resourceServerId, id))
}
