package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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
		Importer: &schema.ResourceImporter{
			StateContext: resourceKeycloakRealmClientPolicyProfilePolicyImport,
		},
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
	keycloakClient.Mutex.Lock(fmt.Sprintf("resourceKeycloakRealmClientPolicyProfilePolicyUpdate:%s", realmId))
	defer keycloakClient.Mutex.Unlock(fmt.Sprintf("resourceKeycloakRealmClientPolicyProfilePolicyUpdate:%s", realmId))
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
	keycloakClient.Mutex.Lock(fmt.Sprintf("resourceKeycloakRealmClientPolicyProfilePolicyDelete:%s", realmId))
	defer keycloakClient.Mutex.Unlock(fmt.Sprintf("resourceKeycloakRealmClientPolicyProfilePolicyDelete:%s", realmId))
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
	keycloakClient.Mutex.Lock(fmt.Sprintf("resourceKeycloakRealmClientPolicyProfilePolicyCreate:%s", realmId))
	defer keycloakClient.Mutex.Unlock(fmt.Sprintf("resourceKeycloakRealmClientPolicyProfilePolicyCreate:%s", realmId))
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
			return nil
		}
	}

	// the policy has been removed outside of terraform
	data.SetId("")

	return nil
}

func resourceKeycloakRealmClientPolicyProfilePolicyImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	parts := strings.SplitN(d.Id(), "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf("Invalid import. Supported import format: {{realmId}}/{{name}}.")
	}

	realmId := parts[0]
	name := parts[1]

	if _, err := keycloakClient.GetRealmClientPolicyProfilePolicyByName(ctx, realmId, name); err != nil {
		return nil, err
	}

	d.Set("realm_id", realmId)
	d.Set("name", name)
	d.SetId(fmt.Sprintf("%s/realm-client-policy-profile-policies/%s", realmId, name))

	return []*schema.ResourceData{d}, nil
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
				// handle json objects and arrays with exception of the attributes field as that needs to stay json-encoded, see https://github.com/keycloak/keycloak/blob/ca205272ba6360bc808d19b5f8e2af119fa37c5a/services/src/main/java/org/keycloak/services/clientpolicy/condition/ClientAttributesCondition.java#L138-L141
				if cond.Name != "client-attributes" && (strings.HasPrefix(value.(string), "{") || strings.HasPrefix(value.(string), "[")) {
					var t interface{}
					json.Unmarshal([]byte(value.(string)), &t)
					configurations[key] = t
					continue
				}
				configurations[key] = value
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
			conditionMap["configuration"] = mapFromClientPolicyConfigurationToData(cond.Configuration)
		}
		conditions = append(conditions, conditionMap)
	}

	return data.Set("condition", conditions)
}
