package provider

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

const clientRegistrationPolicyClientDisabledProviderId = "client-disabled"

func resourceKeycloakClientRegistrationPolicyClientDisabled() *schema.Resource {
	return resourceKeycloakClientRegistrationPolicyConfigless(clientRegistrationPolicyClientDisabledProviderId)
}
