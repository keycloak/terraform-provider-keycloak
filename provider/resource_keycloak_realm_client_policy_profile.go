package provider

import (
	"context"
	"fmt"

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
		}
	}

	return nil
}

func mapFromDataToRealmClientPolicyProfile(data *schema.ResourceData) *keycloak.RealmClientPolicyProfile {
	executors := []keycloak.RealmClientPolicyProfileExecutor{}

	for _, executor := range data.Get("executor").([]interface{}) {
		executorMap := executor.(map[string]interface{})

		exec := keycloak.RealmClientPolicyProfileExecutor{
			Name:          executorMap["name"].(string),
			Configuration: executorMap["configuration"].(map[string]interface{}),
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
			"name":          ex.Name,
			"configuration": ex.Configuration,
		}
		executors = append(executors, executorMap)
	}

	data.Set("executor", executors)

	return nil
}
