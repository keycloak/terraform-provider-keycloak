package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakRealmClientPolicyProfilePolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakRealmClientPolicyProfilePolicyCreate,
		ReadContext:   resourceKeycloakRealmClientPolicyProfilePolicyRead,
		DeleteContext: resourceKeycloakRealmClientPolicyProfilePolicyDelete,
		UpdateContext: resourceKeycloakRealmClientPolicyProfilePolicyUpdate,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"condition": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"configuration": {
							Type:     schema.TypeMap,
							Optional: true,
						},
					},
				},
			},
			"profiles": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				ForceNew: true,
				Required: true,
			},
		},
	}
}

func resourceKeycloakRealmClientPolicyProfilePolicyUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	policy := mapFromDataToRealmClientPolicyProfilePolicy(data)
	realmId := policy.RealmId
	realmClientPolicyProfilePolicies, err := keycloakClient.GetAllRealmClientPolicyProfilePolices(ctx, realmId)
	if err != nil {
		return diag.FromErr(err)
	}

	for i, p := range realmClientPolicyProfilePolicies.Policies {
		if p.Name == policy.Name {
			realmClientPolicyProfilePolicies.Policies[i] = *policy
		}
	}

	err = keycloakClient.UpdateRealmClientPolicyProfilePolicies(ctx, realmId, realmClientPolicyProfilePolicies)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakRealmClientPolicyProfilePolicyDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	slicedPolicies := []keycloak.RealmClientPolicyProfilePolicy{}
	policy := mapFromDataToRealmClientPolicyProfilePolicy(data)
	realmId := policy.RealmId
	realmClientPolicyProfilePolicies, err := keycloakClient.GetAllRealmClientPolicyProfilePolices(ctx, realmId)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, p := range realmClientPolicyProfilePolicies.Policies {
		if p.Name != policy.Name {
			slicedPolicies = append(slicedPolicies, p)
		}
	}

	realmClientPolicyProfilePolicies.Policies = slicedPolicies

	err = keycloakClient.UpdateRealmClientPolicyProfilePolicies(ctx, realmId, realmClientPolicyProfilePolicies)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakRealmClientPolicyProfilePolicyCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	policy := mapFromDataToRealmClientPolicyProfilePolicy(data)

	realmId := policy.RealmId
	name := policy.Name
	data.SetId(fmt.Sprintf("%s/realm-client-policy-profile-policies/%s", realmId, name))

	realmClientPolicyProfilyProfilyPolicies, err := keycloakClient.GetAllRealmClientPolicyProfilePolices(ctx, realmId)
	if err != nil {
		return diag.FromErr(err)
	}

	realmClientPolicyProfilyProfilyPolicies.Policies = append(realmClientPolicyProfilyProfilyPolicies.Policies, *policy)

	err = keycloakClient.UpdateRealmClientPolicyProfilePolicies(ctx, realmId, realmClientPolicyProfilyProfilyPolicies)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceKeycloakRealmClientPolicyProfilePolicyRead(ctx, data, meta)
}

func resourceKeycloakRealmClientPolicyProfilePolicyRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	realmId := data.Get("realm_id").(string)
	name := data.Get("name").(string)
	realmClientPolicyProfilePolicies, err := keycloakClient.GetAllRealmClientPolicyProfilePolices(ctx, realmId)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, policy := range realmClientPolicyProfilePolicies.Policies {
		if policy.Name == name {
			policy.RealmId = realmId
			err = mapFromRealmClientPolicyProfilePolicyToData(data, &policy)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return nil
}

func mapFromDataToRealmClientPolicyProfilePolicy(data *schema.ResourceData) *keycloak.RealmClientPolicyProfilePolicy {
	conditions := []keycloak.RealmClientPolicyProfilePolicyCondition{}
	profiles := make([]string, 0)

	for _, condition := range data.Get("condition").([]interface{}) {
		conditionMap := condition.(map[string]interface{})

		cond := keycloak.RealmClientPolicyProfilePolicyCondition{
			Name: conditionMap["name"].(string),
		}

		if v, ok := conditionMap["configuration"]; ok {
			configurations := make(map[string]interface{})
			for key, value := range v.(map[string]interface{}) {
				configurations[key] = value.(string)
			}
			cond.Configuration = configurations
		}

		conditions = append(conditions, cond)
	}

	for _, profile := range data.Get("profiles").(*schema.Set).List() {
		profiles = append(profiles, profile.(string))
	}

	return &keycloak.RealmClientPolicyProfilePolicy{
		Name:        data.Get("name").(string),
		RealmId:     data.Get("realm_id").(string),
		Description: data.Get("description").(string),
		Enabled:     data.Get("enabled").(bool),
		Profiles:    profiles,
		Conditions:  conditions,
	}
}

func mapFromRealmClientPolicyProfilePolicyToData(data *schema.ResourceData, policy *keycloak.RealmClientPolicyProfilePolicy) error {
	data.Set("name", policy.Name)
	data.Set("realm_id", policy.RealmId)
	data.Set("description", policy.Description)
	data.Set("enabled", policy.Enabled)
	data.Set("profiles", policy.Profiles)

	conditions := make([]interface{}, 0)
	for _, cond := range policy.Conditions {

		conditionMap := map[string]interface{}{
			"name": cond.Name,
		}

		if cond.Configuration != nil {
			configurations := make(map[string]interface{})
			for k, v := range cond.Configuration {
				configurations[k] = v
			}
			conditionMap["configuration"] = configurations
		}
		conditions = append(conditions, conditionMap)
	}

	data.Set("condition", conditions)

	return nil
}
