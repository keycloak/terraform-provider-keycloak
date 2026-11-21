package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakRealmClientPolicyConfiguration() *schema.Resource {
	jsonArray := func() *schema.Schema {
		return &schema.Schema{
			Type: schema.TypeString, Required: true,
			ValidateFunc: func(v interface{}, k string) ([]string, []error) {
				if _, err := clientPolicyEntries(v.(string)); err != nil {
					return nil, []error{fmt.Errorf("%s: %w", k, err)}
				}
				return nil, nil
			},
			StateFunc: func(v interface{}) string { return canonicalClientPolicyJSON(v.(string)) },
			DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
				return canonicalClientPolicyJSON(old) == canonicalClientPolicyJSON(new)
			},
		}
	}
	return &schema.Resource{
		CreateContext: resourceKeycloakRealmClientPolicyConfigurationWrite,
		UpdateContext: resourceKeycloakRealmClientPolicyConfigurationWrite,
		ReadContext:   resourceKeycloakRealmClientPolicyConfigurationRead,
		DeleteContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
			return diag.Errorf("A realm client-policy singleton cannot be deleted; configure empty arrays to clear it, or remove Terraform ownership without destroying the remote object")
		},
		Importer: &schema.ResourceImporter{StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
			if strings.TrimSpace(d.Id()) == "" || strings.Contains(d.Id(), "/") {
				return nil, fmt.Errorf("import ID must be a single nonempty realm ID")
			}
			if err := d.Set("realm_id", d.Id()); err != nil {
				return nil, err
			}
			ds := resourceKeycloakRealmClientPolicyConfigurationRead(ctx, d, meta)
			if ds.HasError() || d.Id() == "" {
				return nil, fmt.Errorf("cannot read existing realm client-policy configuration: %v", ds)
			}
			return []*schema.ResourceData{d}, nil
		}},
		Schema: map[string]*schema.Schema{
			"realm_id": {Type: schema.TypeString, Required: true, ForceNew: true},
			"profiles": jsonArray(), "policies": jsonArray(),
			"global_profiles": {Type: schema.TypeString, Computed: true},
			"global_policies": {Type: schema.TypeString, Computed: true},
		},
	}
}

func canonicalClientPolicyJSON(s string) string {
	var v interface{}
	dec := json.NewDecoder(strings.NewReader(s))
	dec.UseNumber()
	if dec.Decode(&v) != nil {
		return s
	}
	var extra interface{}
	if dec.Decode(&extra) != io.EOF {
		return s
	}
	b, err := json.Marshal(v)
	if err != nil {
		return s
	}
	return string(b)
}
func clientPolicyEntries(s string) ([]json.RawMessage, error) {
	var entries []json.RawMessage
	dec := json.NewDecoder(strings.NewReader(s))
	if err := dec.Decode(&entries); err != nil || entries == nil {
		return nil, fmt.Errorf("expected a JSON array of named objects")
	}
	var extra interface{}
	if dec.Decode(&extra) != io.EOF {
		return nil, fmt.Errorf("expected a single JSON array")
	}
	names := map[string]bool{}
	for _, e := range entries {
		var item map[string]json.RawMessage
		var name string
		if json.Unmarshal(e, &item) != nil || item == nil || json.Unmarshal(item["name"], &name) != nil || strings.TrimSpace(name) == "" {
			return nil, fmt.Errorf("each entry must have a nonempty string name")
		}
		if names[name] {
			return nil, fmt.Errorf("duplicate entry name")
		}
		names[name] = true
	}
	return entries, nil
}
func clientPolicyName(e json.RawMessage) string {
	var n struct {
		Name string `json:"name"`
	}
	_ = json.Unmarshal(e, &n)
	return n.Name
}
func clientPolicyArrayJSON(v []json.RawMessage) string {
	if v == nil {
		v = []json.RawMessage{}
	}
	b, _ := json.Marshal(v)
	return canonicalClientPolicyJSON(string(b))
}
func sameClientPolicyEntries(a, b []json.RawMessage) bool {
	return clientPolicyArrayJSON(a) == clientPolicyArrayJSON(b)
}

func resourceKeycloakRealmClientPolicyConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*keycloak.KeycloakClient)
	v, err := c.GetRealmClientPolicyConfiguration(ctx, d.Get("realm_id").(string))
	if err != nil {
		return handleNotFoundError(ctx, err, d)
	}
	for k, entries := range map[string][]json.RawMessage{"profiles": v.Profiles, "policies": v.Policies, "global_profiles": v.GlobalProfiles, "global_policies": v.GlobalPolicies} {
		if err := d.Set(k, clientPolicyArrayJSON(entries)); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}
func resourceKeycloakRealmClientPolicyConfigurationWrite(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*keycloak.KeycloakClient)
	realm := d.Get("realm_id").(string)
	profiles, err := clientPolicyEntries(d.Get("profiles").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	policies, err := clientPolicyEntries(d.Get("policies").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	// Share locks with individual resources, always acquiring profiles first.
	c.Mutex.Lock(realmClientPolicyProfilesMutexKey(realm))
	defer c.Mutex.Unlock(realmClientPolicyProfilesMutexKey(realm))
	c.Mutex.Lock(realmClientPolicyProfilePoliciesMutexKey(realm))
	defer c.Mutex.Unlock(realmClientPolicyProfilePoliciesMutexKey(realm))
	current, err := c.GetRealmClientPolicyConfiguration(ctx, realm)
	if err != nil {
		return diag.FromErr(err)
	}
	for _, pair := range []struct{ desired, global []json.RawMessage }{{profiles, current.GlobalProfiles}, {policies, current.GlobalPolicies}} {
		global := map[string]bool{}
		for _, e := range pair.global {
			global[clientPolicyName(e)] = true
		}
		for _, e := range pair.desired {
			if global[clientPolicyName(e)] {
				return diag.Errorf("realm client-policy entries cannot replace global templates")
			}
		}
	}
	// Retain ownership if a later endpoint fails after a partial update.
	d.SetId(realm)
	// Keep existing profiles available until policies have switched to their new
	// references. The endpoints are not atomic; a failed stage returns an error
	// and a subsequent refresh/retry resumes from actual server state.
	staged := append([]json.RawMessage{}, profiles...)
	desiredNames := map[string]bool{}
	for _, e := range profiles {
		desiredNames[clientPolicyName(e)] = true
	}
	for _, e := range current.Profiles {
		if !desiredNames[clientPolicyName(e)] {
			staged = append(staged, e)
		}
	}
	for _, step := range []struct {
		collection    string
		before, after []json.RawMessage
	}{{"profiles", current.Profiles, staged}, {"policies", current.Policies, policies}, {"profiles", staged, profiles}} {
		if !sameClientPolicyEntries(step.before, step.after) {
			if err := c.PutRealmClientPolicyCollection(ctx, realm, step.collection, step.after); err != nil {
				return diag.FromErr(err)
			}
		}
	}
	d.SetId(realm)
	return resourceKeycloakRealmClientPolicyConfigurationRead(ctx, d, meta)
}
