package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakOpenidClientAuthorizationTimePolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakOpenidClientAuthorizationTimePolicyCreate,
		ReadContext:   resourceKeycloakOpenidClientAuthorizationTimePolicyRead,
		DeleteContext: resourceKeycloakOpenidClientAuthorizationTimePolicyDelete,
		UpdateContext: resourceKeycloakOpenidClientAuthorizationTimePolicyUpdate,
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
			"not_before": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"not_on_or_after": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"day_month": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"day_month_end": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"month": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"month_end": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"year": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"year_end": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"hour": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"hour_end": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"minute": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"minute_end": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func getOpenidClientAuthorizationTimePolicyResourceFromData(data *schema.ResourceData) *keycloak.OpenidClientAuthorizationTimePolicy {

	resource := keycloak.OpenidClientAuthorizationTimePolicy{
		Id:               data.Id(),
		ResourceServerId: data.Get("resource_server_id").(string),
		RealmId:          data.Get("realm_id").(string),
		DecisionStrategy: data.Get("decision_strategy").(string),
		Logic:            data.Get("logic").(string),
		Name:             data.Get("name").(string),
		Type:             "time",
		NotBefore:        data.Get("not_before").(string),
		NotOnOrAfter:     data.Get("not_on_or_after").(string),
		DayMonth:         data.Get("day_month").(string),
		DayMonthEnd:      data.Get("day_month_end").(string),
		Month:            data.Get("month").(string),
		MonthEnd:         data.Get("month_end").(string),
		Year:             data.Get("year").(string),
		YearEnd:          data.Get("year_end").(string),
		Hour:             data.Get("hour").(string),
		HourEnd:          data.Get("hour_end").(string),
		Minute:           data.Get("minute").(string),
		MinuteEnd:        data.Get("minute_end").(string),
		Description:      data.Get("description").(string),
	}
	return &resource
}

func setOpenidClientAuthorizationTimePolicyResourceData(data *schema.ResourceData, policy *keycloak.OpenidClientAuthorizationTimePolicy) {
	data.SetId(policy.Id)

	data.Set("resource_server_id", policy.ResourceServerId)
	data.Set("realm_id", policy.RealmId)
	data.Set("name", policy.Name)
	data.Set("decision_strategy", policy.DecisionStrategy)
	data.Set("logic", policy.Logic)
	data.Set("description", policy.Description)
	data.Set("not_on_or_after", policy.NotOnOrAfter)
	data.Set("not_before", policy.NotBefore)
	data.Set("day_month", policy.DayMonth)
	data.Set("day_month_end", policy.DayMonthEnd)
	data.Set("month", policy.Month)
	data.Set("month_end", policy.MonthEnd)
	data.Set("year", policy.Year)
	data.Set("year_end", policy.YearEnd)
	data.Set("hour", policy.Hour)
	data.Set("hour_end", policy.HourEnd)
	data.Set("minute", policy.Minute)
	data.Set("minute_end", policy.MinuteEnd)
}

func resourceKeycloakOpenidClientAuthorizationTimePolicyCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	resource := getOpenidClientAuthorizationTimePolicyResourceFromData(data)

	err := keycloakClient.NewOpenidClientAuthorizationTimePolicy(ctx, resource)
	if err != nil {
		return diag.FromErr(err)
	}

	setOpenidClientAuthorizationTimePolicyResourceData(data, resource)

	return resourceKeycloakOpenidClientAuthorizationTimePolicyRead(ctx, data, meta)
}

func resourceKeycloakOpenidClientAuthorizationTimePolicyRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	resourceServerId := data.Get("resource_server_id").(string)
	id := data.Id()

	resource, err := keycloakClient.GetOpenidClientAuthorizationTimePolicy(ctx, realmId, resourceServerId, id)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	setOpenidClientAuthorizationTimePolicyResourceData(data, resource)

	return nil
}

func resourceKeycloakOpenidClientAuthorizationTimePolicyUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	resource := getOpenidClientAuthorizationTimePolicyResourceFromData(data)

	err := keycloakClient.UpdateOpenidClientAuthorizationTimePolicy(ctx, resource)
	if err != nil {
		return diag.FromErr(err)
	}

	setOpenidClientAuthorizationTimePolicyResourceData(data, resource)

	return nil
}

func resourceKeycloakOpenidClientAuthorizationTimePolicyDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	resourceServerId := data.Get("resource_server_id").(string)
	id := data.Id()

	return diag.FromErr(keycloakClient.DeleteOpenidClientAuthorizationTimePolicy(ctx, realmId, resourceServerId, id))
}
