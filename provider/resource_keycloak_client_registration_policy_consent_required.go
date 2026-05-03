package provider

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

const clientRegistrationPolicyConsentRequiredProviderId = "consent-required"

func resourceKeycloakClientRegistrationPolicyConsentRequired() *schema.Resource {
	return resourceKeycloakClientRegistrationPolicyConfigless(clientRegistrationPolicyConsentRequiredProviderId)
}
