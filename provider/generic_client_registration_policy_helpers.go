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

var clientRegistrationPolicySubTypes = []string{"anonymous", "authenticated"}

// commonClientRegistrationPolicySchema returns the schema fields shared by
// every keycloak_client_registration_policy_* resource: name, realm_id,
// sub_type. Callers add their own typed config fields.
func commonClientRegistrationPolicySchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"realm_id": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Display name of the client registration policy.",
		},
		"sub_type": {
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringInSlice(clientRegistrationPolicySubTypes, false),
			Description:  `Either "anonymous" (applied to anonymous DCR requests) or "authenticated" (applied to requests using an initial access token / registration access token).`,
		},
	}
}

// clientRegistrationPolicyImporter returns an Importer whose StateFunc parses
// {realmId}/{componentId}, fetches the component, and asserts its providerId
// matches expectedProviderId. This protects against importing e.g. a
// trusted-hosts policy into a max_clients resource (which would silently
// produce nonsense state).
func clientRegistrationPolicyImporter(expectedProviderId string) *schema.ResourceImporter {
	return &schema.ResourceImporter{
		StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
			parts := strings.Split(d.Id(), "/")
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid import id %q: expected format {{realmId}}/{{componentId}}", d.Id())
			}
			realmId, id := parts[0], parts[1]

			keycloakClient := meta.(*keycloak.KeycloakClient)
			policy, err := keycloakClient.GetClientRegistrationPolicy(ctx, realmId, id)
			if err != nil {
				return nil, fmt.Errorf("import %s: %w", d.Id(), err)
			}
			if policy.ProviderId != expectedProviderId {
				return nil, fmt.Errorf("component %s in realm %s has providerId=%q; this resource type expects providerId=%q. Use the matching keycloak_client_registration_policy_* resource for this component", id, realmId, policy.ProviderId, expectedProviderId)
			}

			d.Set("realm_id", realmId)
			d.SetId(id)
			return []*schema.ResourceData{d}, nil
		},
	}
}

// readClientRegistrationPolicyCommon populates name, realm_id, and sub_type
// onto d from policy, and SetId. Callers handle their own typed config fields.
func readClientRegistrationPolicyCommon(d *schema.ResourceData, policy *keycloak.ClientRegistrationPolicy) {
	d.SetId(policy.Id)
	d.Set("name", policy.Name)
	d.Set("realm_id", policy.RealmId)
	d.Set("sub_type", policy.SubType)
}

// deleteClientRegistrationPolicy is the shared Delete callback. All 8
// resource types share the same delete behavior.
func deleteClientRegistrationPolicy(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	return diag.FromErr(keycloakClient.DeleteClientRegistrationPolicy(ctx, d.Get("realm_id").(string), d.Id()))
}

// configBoolValue returns the canonical Keycloak string representation of a
// bool config value: "true" or "false".
func configBoolValue(b bool) []string {
	if b {
		return []string{"true"}
	}
	return []string{"false"}
}

// configStringSliceFromSet converts a *schema.Set of strings into the []string
// shape expected by Keycloak's component config map.
func configStringSliceFromSet(set *schema.Set) []string {
	raw := set.List()
	out := make([]string, 0, len(raw))
	for _, v := range raw {
		out = append(out, v.(string))
	}
	return out
}

// configBool reads a single bool value from a Keycloak config slice, defaulting
// to defaultValue if the slice is empty or unparseable.
func configBool(values []string, defaultValue bool) bool {
	if len(values) == 0 {
		return defaultValue
	}
	switch strings.ToLower(values[0]) {
	case "true":
		return true
	case "false":
		return false
	default:
		return defaultValue
	}
}

// resourceKeycloakClientRegistrationPolicyConfigless returns a fully-wired
// *schema.Resource for a config-less policy (only realm_id, name, sub_type
// — no per-policy config keys). Used by consent_required, full_scope_disallowed,
// and client_disabled, which are byte-for-byte identical except for providerId.
func resourceKeycloakClientRegistrationPolicyConfigless(providerId string) *schema.Resource {
	create := func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		keycloakClient := meta.(*keycloak.KeycloakClient)
		policy := &keycloak.ClientRegistrationPolicy{
			Id:         d.Id(),
			RealmId:    d.Get("realm_id").(string),
			Name:       d.Get("name").(string),
			ProviderId: providerId,
			SubType:    d.Get("sub_type").(string),
			Config:     map[string][]string{},
		}
		if err := keycloakClient.NewClientRegistrationPolicy(ctx, policy); err != nil {
			return diag.FromErr(err)
		}
		readClientRegistrationPolicyCommon(d, policy)
		return nil
	}
	read := func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		keycloakClient := meta.(*keycloak.KeycloakClient)
		policy, err := keycloakClient.GetClientRegistrationPolicy(ctx, d.Get("realm_id").(string), d.Id())
		if err != nil {
			return handleNotFoundError(ctx, err, d)
		}
		readClientRegistrationPolicyCommon(d, policy)
		return nil
	}
	update := func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		keycloakClient := meta.(*keycloak.KeycloakClient)
		policy := &keycloak.ClientRegistrationPolicy{
			Id:         d.Id(),
			RealmId:    d.Get("realm_id").(string),
			Name:       d.Get("name").(string),
			ProviderId: providerId,
			SubType:    d.Get("sub_type").(string),
			Config:     map[string][]string{},
		}
		if err := keycloakClient.UpdateClientRegistrationPolicy(ctx, policy); err != nil {
			return diag.FromErr(err)
		}
		readClientRegistrationPolicyCommon(d, policy)
		return nil
	}
	return &schema.Resource{
		CreateContext: create,
		ReadContext:   read,
		UpdateContext: update,
		DeleteContext: deleteClientRegistrationPolicy,
		Importer:      clientRegistrationPolicyImporter(providerId),
		Schema:        commonClientRegistrationPolicySchema(),
	}
}
