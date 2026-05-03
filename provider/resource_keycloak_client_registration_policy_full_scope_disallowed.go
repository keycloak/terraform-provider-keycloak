package provider

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

// The wire-level providerId is just "scope" — Keycloak's class is named
// ScopeClientRegistrationPolicyFactory and its help text is "When present,
// then newly registered client won't have full scope allowed". The resource
// is named full_scope_disallowed for parity with the admin UI label.
const clientRegistrationPolicyFullScopeDisallowedProviderId = "scope"

func resourceKeycloakClientRegistrationPolicyFullScopeDisallowed() *schema.Resource {
	return resourceKeycloakClientRegistrationPolicyConfigless(clientRegistrationPolicyFullScopeDisallowedProviderId)
}
