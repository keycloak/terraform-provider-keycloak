package provider

import (
	"context"
	"fmt"
	"sort"
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
				Type:             schema.TypeMap,
				Optional:         true,
				Description:      "Policy-specific configuration key-value pairs.",
				DiffSuppressFunc: suppressMultiValueClientRegistrationConfigOrder,
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

func resourceKeycloakRealmClientRegistrationPolicyImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")

	switch len(parts) {
	case 2:
		// Format: {realmId}/{policyId} — direct import by component ID.
		d.Set("realm_id", parts[0])
		d.SetId(parts[1])
	case 4:
		// Format: {realmId}/{name}/{providerId}/{subType} — import by attributes.
		// Keycloak auto-creates the default policies ("Trusted Hosts", etc.) with
		// server-generated UUIDs that aren't known at plan time, so this format lets
		// users take ownership of them with static values in an import block.
		keycloakClient := meta.(*keycloak.KeycloakClient)
		realmId, name, providerId, subType := parts[0], parts[1], parts[2], parts[3]

		policies, err := keycloakClient.GetRealmClientRegistrationPolicies(ctx, realmId)
		if err != nil {
			return nil, fmt.Errorf("error fetching client registration policies: %w", err)
		}

		var matchingPolicy *keycloak.RealmClientRegistrationPolicy
		for _, policy := range policies {
			if policy.Name == name && policy.ProviderId == providerId && policy.SubType == subType {
				matchingPolicy = policy
				break
			}
		}

		if matchingPolicy == nil {
			return nil, fmt.Errorf("no client registration policy found with name=%q, providerId=%q, subType=%q in realm %q", name, providerId, subType, realmId)
		}

		d.Set("realm_id", realmId)
		d.SetId(matchingPolicy.Id)
	default:
		return nil, fmt.Errorf("Invalid import. Supported import formats: {{realmId}}/{{policyId}}, {{realmId}}/{{name}}/{{providerId}}/{{subType}}")
	}

	return []*schema.ResourceData{d}, nil
}

// suppressMultiValueClientRegistrationConfigOrder suppresses spurious diffs on
// multi-value config fields (those in keycloak.MultiValueClientRegistrationConfigKeys) whose
// comma-separated old and new values contain the same elements in a different order. Keycloak does not
// preserve element order across writes, so without this the provider would plan an update
// on every run even when nothing meaningfully changed.
//
// k is the flattened map element key, e.g. "config.trusted-hosts"; the synthetic
// "config.%" element-count key and any non-multi-value field are never suppressed.
func suppressMultiValueClientRegistrationConfigOrder(k, old, new string, _ *schema.ResourceData) bool {
	key := strings.TrimPrefix(k, "config.")
	if !keycloak.MultiValueClientRegistrationConfigKeys[key] {
		return false
	}

	return equalCommaSeparatedSet(old, new)
}

// equalCommaSeparatedSet reports whether two comma-separated strings contain the same
// multiset of whitespace-trimmed elements, ignoring order. A differing element count
// (including added, removed, or duplicated values) is treated as a genuine change.
func equalCommaSeparatedSet(a, b string) bool {
	as := splitTrimComma(a)
	bs := splitTrimComma(b)
	if len(as) != len(bs) {
		return false
	}

	sort.Strings(as)
	sort.Strings(bs)
	for i := range as {
		if as[i] != bs[i] {
			return false
		}
	}

	return true
}

// splitTrimComma splits a comma-separated string into its trimmed elements. An empty
// or whitespace-only string yields no elements.
func splitTrimComma(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}

	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		out = append(out, strings.TrimSpace(p))
	}

	return out
}
