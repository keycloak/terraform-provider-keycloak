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

func resourceKeycloakRealmClientRegistrationPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakRealmClientRegistrationPolicyCreate,
		ReadContext:   resourceKeycloakRealmClientRegistrationPolicyRead,
		UpdateContext: resourceKeycloakRealmClientRegistrationPolicyUpdate,
		DeleteContext: resourceKeycloakRealmClientRegistrationPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceKeycloakRealmClientRegistrationPolicyImport,
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
			"provider_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The provider ID of the client registration policy (e.g. 'trusted-hosts', 'consent-required').",
			},
			"sub_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"anonymous", "authenticated"}, false),
				Description:  "Whether this policy applies to anonymous or authenticated client registration.",
			},
			"config": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Policy-specific configuration key-value pairs.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func getRealmClientRegistrationPolicyFromData(data *schema.ResourceData) *keycloak.RealmClientRegistrationPolicy {
	config := map[string]string{}
	if v, ok := data.GetOk("config"); ok {
		for k, val := range v.(map[string]interface{}) {
			config[k] = val.(string)
		}
	}

	return &keycloak.RealmClientRegistrationPolicy{
		Id:         data.Id(),
		Name:       data.Get("name").(string),
		RealmId:    data.Get("realm_id").(string),
		ProviderId: data.Get("provider_id").(string),
		SubType:    data.Get("sub_type").(string),
		Config:     config,
	}
}

func setRealmClientRegistrationPolicyData(data *schema.ResourceData, policy *keycloak.RealmClientRegistrationPolicy) {
	data.SetId(policy.Id)
	data.Set("name", policy.Name)
	data.Set("realm_id", policy.RealmId)
	data.Set("provider_id", policy.ProviderId)
	data.Set("sub_type", policy.SubType)
	data.Set("config", policy.Config)
}

func resourceKeycloakRealmClientRegistrationPolicyCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	policy := getRealmClientRegistrationPolicyFromData(data)

	err := keycloakClient.NewRealmClientRegistrationPolicy(ctx, policy)
	if err != nil {
		return diag.FromErr(err)
	}

	setRealmClientRegistrationPolicyData(data, policy)

	return resourceKeycloakRealmClientRegistrationPolicyRead(ctx, data, meta)
}

func resourceKeycloakRealmClientRegistrationPolicyRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	policy, err := keycloakClient.GetRealmClientRegistrationPolicy(ctx, realmId, id)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	setRealmClientRegistrationPolicyData(data, policy)

	return nil
}

func resourceKeycloakRealmClientRegistrationPolicyUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	policy := getRealmClientRegistrationPolicyFromData(data)

	err := keycloakClient.UpdateRealmClientRegistrationPolicy(ctx, policy)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceKeycloakRealmClientRegistrationPolicyRead(ctx, data, meta)
}

func resourceKeycloakRealmClientRegistrationPolicyDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	return diag.FromErr(keycloakClient.DeleteRealmClientRegistrationPolicy(ctx, realmId, id))
}

func resourceKeycloakRealmClientRegistrationPolicyImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")

	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid import format, expected: {realmId}/{policyId}")
	}

	d.Set("realm_id", parts[0])
	d.SetId(parts[1])

	return []*schema.ResourceData{d}, nil
}
