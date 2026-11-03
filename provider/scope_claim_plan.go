package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

// unknownScopeName is the placeholder the SDK uses for set elements that are not known until apply.
const unknownScopeName = "74D93920-ED26-11E3-AC10-0800200C9A66"

// claimScopesOnPlan claims the planned scopes of a scope resource, so that a scope configured as both a default and
// an optional scope of the same owner fails the plan. Terraform plans every resource in the configuration, not only
// the changed ones, so this catches a conflict with a resource that is already applied.
//
// The check is skipped while the owner is not known, for example because the client is created in the same run. The
// reconcile functions claim the scopes again at apply time, which covers that case.
func claimScopesOnPlan(scopeType, scopesAttribute string, ownerKey func(d *schema.ResourceDiff) (string, bool)) schema.CustomizeDiffFunc {
	return func(_ context.Context, d *schema.ResourceDiff, meta interface{}) error {
		key, ok := ownerKey(d)
		if !ok || !d.NewValueKnown(scopesAttribute) {
			return nil
		}

		var scopeNames []string
		for _, name := range d.Get(scopesAttribute).(*schema.Set).List() {
			if name.(string) != unknownScopeName {
				scopeNames = append(scopeNames, name.(string))
			}
		}

		return meta.(*keycloak.KeycloakClient).ClaimScopes(key, scopeType, scopeNames)
	}
}

func openidClientScopeOwnerKeyOnPlan(d *schema.ResourceDiff) (string, bool) {
	if !d.NewValueKnown("realm_id") || !d.NewValueKnown("client_id") {
		return "", false
	}

	return keycloak.OpenidClientScopeOwnerKey(d.Get("realm_id").(string), d.Get("client_id").(string)), true
}

func samlClientScopeOwnerKeyOnPlan(d *schema.ResourceDiff) (string, bool) {
	if !d.NewValueKnown("realm_id") || !d.NewValueKnown("client_id") {
		return "", false
	}

	return keycloak.SamlClientScopeOwnerKey(d.Get("realm_id").(string), d.Get("client_id").(string)), true
}

func realmScopeOwnerKeyOnPlan(d *schema.ResourceDiff) (string, bool) {
	if !d.NewValueKnown("realm_id") {
		return "", false
	}

	return keycloak.RealmScopeOwnerKey(d.Get("realm_id").(string)), true
}
