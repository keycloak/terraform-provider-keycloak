package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func dataSourceKeycloakOpenidClientAuthorizationPolicy() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKeycloakOpenidClientAuthorizationPolicyRead,

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
				Computed: true,
			},
			"owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"logic": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"policies": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"resources": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"scopes": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func setOpenidClientAuthorizationPolicyData(data *schema.ResourceData, policy *keycloak.OpenidClientAuthorizationPolicy) {
	data.SetId(policy.Id)

	data.Set("resource_server_id", policy.ResourceServerId)
	data.Set("realm_id", policy.RealmId)
	data.Set("name", policy.Name)
	data.Set("decision_strategy", policy.DecisionStrategy)
	data.Set("owner", policy.Owner)
	data.Set("logic", policy.Logic)
	data.Set("policies", policy.Policies)
	data.Set("resources", policy.Resources)
	data.Set("scopes", policy.Scopes)
	data.Set("type", policy.Type)
}

func dataSourceKeycloakOpenidClientAuthorizationPolicyRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	resourceServerId := data.Get("resource_server_id").(string)
	name := data.Get("name").(string)

	client, err := keycloakClient.GetClientAuthorizationPolicyByName(ctx, realmId, resourceServerId, name)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	setOpenidClientAuthorizationPolicyData(data, client)

	return nil
}
