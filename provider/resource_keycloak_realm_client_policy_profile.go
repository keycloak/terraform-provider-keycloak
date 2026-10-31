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

func resourceKeycloakRealmClientPolicyProfile() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakRealmClientPolicyProfileCreate,
		ReadContext:   resourceKeycloakRealmClientPolicyProfileRead,
		DeleteContext: resourceKeycloakRealmClientPolicyProfileDelete,
		UpdateContext: resourceKeycloakRealmClientPolicyProfileUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: resourceKeycloakRealmClientPolicyProfileImport,
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
			"executor": {
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
		},
	}
}

func resourceKeycloakRealmClientPolicyProfileUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	profile := mapFromDataToRealmClientPolicyProfile(data)
	realmId := profile.RealmId
	keycloakClient.Mutex.Lock(fmt.Sprintf("resourceKeycloakRealmClientPolicyProfileUpdate:%s", realmId))
	defer keycloakClient.Mutex.Unlock(fmt.Sprintf("resourceKeycloakRealmClientPolicyProfileUpdate:%s", realmId))
	realmClientPolicyProfiles, err := keycloakClient.GetAllRealmClientPolicyProfiles(ctx, realmId)
	if err != nil {
		return diag.FromErr(err)
	}

	for i, p := range realmClientPolicyProfiles.Profiles {
		if p.Name == profile.Name {
			realmClientPolicyProfiles.Profiles[i] = *profile
		}
	}

	err = keycloakClient.UpdateRealmClientPolicyProfiles(ctx, realmId, realmClientPolicyProfiles)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakRealmClientPolicyProfileDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	slicedProfiles := []keycloak.RealmClientPolicyProfile{}
	profile := mapFromDataToRealmClientPolicyProfile(data)
	realmId := profile.RealmId
	keycloakClient.Mutex.Lock(fmt.Sprintf("resourceKeycloakRealmClientPolicyProfileDelete:%s", realmId))
	defer keycloakClient.Mutex.Unlock(fmt.Sprintf("resourceKeycloakRealmClientPolicyProfileDelete:%s", realmId))
	realmClientPolicyProfiles, err := keycloakClient.GetAllRealmClientPolicyProfiles(ctx, realmId)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, p := range realmClientPolicyProfiles.Profiles {
		if p.Name != profile.Name {
			slicedProfiles = append(slicedProfiles, p)
		}
	}

	realmClientPolicyProfiles.Profiles = slicedProfiles

	err = keycloakClient.UpdateRealmClientPolicyProfiles(ctx, realmId, realmClientPolicyProfiles)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakRealmClientPolicyProfileCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	profile := mapFromDataToRealmClientPolicyProfile(data)

	realmId := profile.RealmId
	keycloakClient.Mutex.Lock(fmt.Sprintf("resourceKeycloakRealmClientPolicyProfileCreate:%s", realmId))
	defer keycloakClient.Mutex.Unlock(fmt.Sprintf("resourceKeycloakRealmClientPolicyProfileCreate:%s", realmId))
	name := profile.Name
	data.SetId(fmt.Sprintf("%s/realm-client-policy-profiles/%s", realmId, name))

	realmClientPolicyProfiles, err := keycloakClient.GetAllRealmClientPolicyProfiles(ctx, realmId)
	if err != nil {
		return diag.FromErr(err)
	}

	realmClientPolicyProfiles.Profiles = append(realmClientPolicyProfiles.Profiles, *profile)

	err = keycloakClient.UpdateRealmClientPolicyProfiles(ctx, realmId, realmClientPolicyProfiles)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceKeycloakRealmClientPolicyProfileRead(ctx, data, meta)
}

func resourceKeycloakRealmClientPolicyProfileRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	realmId := data.Get("realm_id").(string)
	name := data.Get("name").(string)
	realmClientPolicyProfiles, err := keycloakClient.GetAllRealmClientPolicyProfiles(ctx, realmId)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, profile := range realmClientPolicyProfiles.Profiles {
		if profile.Name == name {
			profile.RealmId = realmId
			err = mapFromRealmClientPolicyProfileToData(data, &profile)
			if err != nil {
				return diag.FromErr(err)
			}
			return nil
		}
	}

	// the profile has been removed outside of terraform
	data.SetId("")

	return nil
}

func resourceKeycloakRealmClientPolicyProfileImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	parts := strings.SplitN(d.Id(), "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf("Invalid import. Supported import format: {{realmId}}/{{name}}.")
	}

	realmId := parts[0]
	name := parts[1]

	if _, err := keycloakClient.GetRealmClientPolicyProfileByName(ctx, realmId, name); err != nil {
		return nil, err
	}

	d.Set("realm_id", realmId)
	d.Set("name", name)
	d.SetId(fmt.Sprintf("%s/realm-client-policy-profiles/%s", realmId, name))

	return []*schema.ResourceData{d}, nil
}

func mapFromDataToRealmClientPolicyProfile(data *schema.ResourceData) *keycloak.RealmClientPolicyProfile {
	executors := []keycloak.RealmClientPolicyProfileExecutor{}

	for _, executor := range data.Get("executor").([]interface{}) {
		executorMap := executor.(map[string]interface{})

		exec := keycloak.RealmClientPolicyProfileExecutor{
			Name: executorMap["name"].(string),
		}

		if v, ok := executorMap["configuration"]; ok {
			configurations := make(map[string]interface{})
			for key, value := range v.(map[string]interface{}) {
				// handle json objects and arrays
				if strings.HasPrefix(value.(string), "{") || strings.HasPrefix(value.(string), "[") {
					var t interface{}
					json.Unmarshal([]byte(value.(string)), &t)
					configurations[key] = t
					continue
				}
				configurations[key] = value
			}
			exec.Configuration = configurations
		}

		executors = append(executors, exec)
	}

	return &keycloak.RealmClientPolicyProfile{
		Name:        data.Get("name").(string),
		RealmId:     data.Get("realm_id").(string),
		Description: data.Get("description").(string),
		Executors:   executors,
	}
}

func mapFromRealmClientPolicyProfileToData(data *schema.ResourceData, profile *keycloak.RealmClientPolicyProfile) error {
	data.Set("name", profile.Name)
	data.Set("realm_id", profile.RealmId)
	data.Set("description", profile.Description)

	executors := make([]interface{}, 0)
	for _, ex := range profile.Executors {

		executorMap := map[string]interface{}{
			"name": ex.Name,
		}

		if ex.Configuration != nil {
			executorMap["configuration"] = mapFromClientPolicyConfigurationToData(ex.Configuration)
		}
		executors = append(executors, executorMap)
	}

	return data.Set("executor", executors)
}

// mapFromClientPolicyConfigurationToData converts the configuration of a client policy executor or condition into
// a map of strings. Keycloak may return json objects, arrays, booleans or numbers (e.g. for objects created via the
// admin console), which are stored as their json representation.
func mapFromClientPolicyConfigurationToData(configuration map[string]interface{}) map[string]interface{} {
	configurations := make(map[string]interface{})
	for k, v := range configuration {
		switch v := v.(type) {
		case string:
			configurations[k] = v
		default:
			s, err := json.Marshal(v)
			if err != nil {
				configurations[k] = fmt.Sprint(v)
				continue
			}
			configurations[k] = string(s)
		}
	}

	return configurations
}
