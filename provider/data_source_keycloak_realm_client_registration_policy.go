package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func dataSourceKeycloakRealmClientRegistrationPolicy() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKeycloakRealmClientRegistrationPolicyRead,
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The realm this client registration policy belongs to.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the client registration policy to find.",
			},
			"provider_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Filter by provider ID (e.g. 'trusted-hosts', 'max-clients').",
			},
			"sub_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Filter by sub-type ('anonymous' or 'authenticated').",
			},
			"config": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Policy-specific configuration key-value pairs.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceKeycloakRealmClientRegistrationPolicyRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	name := data.Get("name").(string)
	providerId := data.Get("provider_id").(string)
	subType := data.Get("sub_type").(string)

	policies, err := keycloakClient.GetRealmClientRegistrationPolicies(ctx, realmId)
	if err != nil {
		return diag.FromErr(err)
	}

	var matchingPolicies []*keycloak.RealmClientRegistrationPolicy
	for _, policy := range policies {
		if policy.Name != name {
			continue
		}
		if providerId != "" && policy.ProviderId != providerId {
			continue
		}
		if subType != "" && policy.SubType != subType {
			continue
		}
		matchingPolicies = append(matchingPolicies, policy)
	}

	if len(matchingPolicies) == 0 {
		return diag.Errorf("no client registration policy found with name %q in realm %q", name, realmId)
	}

	if len(matchingPolicies) > 1 {
		var ids []string
		for _, p := range matchingPolicies {
			ids = append(ids, fmt.Sprintf("%s (provider_id=%s, sub_type=%s)", p.Id, p.ProviderId, p.SubType))
		}
		return diag.Errorf("multiple client registration policies found with name %q: %s. Use provider_id and/or sub_type to filter.", name, strings.Join(ids, ", "))
	}

	policy := matchingPolicies[0]

	data.SetId(policy.Id)
	data.Set("realm_id", policy.RealmId)
	data.Set("name", policy.Name)
	data.Set("provider_id", policy.ProviderId)
	data.Set("sub_type", policy.SubType)
	data.Set("config", policy.Config)

	return nil
}
