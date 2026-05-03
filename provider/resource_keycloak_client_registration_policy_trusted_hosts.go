package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

const clientRegistrationPolicyTrustedHostsProviderId = "trusted-hosts"

func resourceKeycloakClientRegistrationPolicyTrustedHosts() *schema.Resource {
	s := commonClientRegistrationPolicySchema()
	s["trusted_hosts"] = &schema.Schema{
		Type:        schema.TypeSet,
		Optional:    true,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Description: "List of allowed hostnames or IPs that may submit DCR requests and that may appear in registered client URIs.",
	}
	s["host_sending_registration_request_must_match"] = &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
		Description: "If true, the source host of the DCR request must be in trusted_hosts.",
	}
	s["client_uris_must_match"] = &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
		Description: "If true, every URI on a registered/updated client must match a host in trusted_hosts.",
	}

	return &schema.Resource{
		CreateContext: resourceKeycloakClientRegistrationPolicyTrustedHostsCreate,
		ReadContext:   resourceKeycloakClientRegistrationPolicyTrustedHostsRead,
		UpdateContext: resourceKeycloakClientRegistrationPolicyTrustedHostsUpdate,
		DeleteContext: deleteClientRegistrationPolicy,
		Importer:      clientRegistrationPolicyImporter(clientRegistrationPolicyTrustedHostsProviderId),
		Schema:        s,
	}
}

func getClientRegistrationPolicyTrustedHostsFromData(d *schema.ResourceData) *keycloak.ClientRegistrationPolicy {
	return &keycloak.ClientRegistrationPolicy{
		Id:         d.Id(),
		RealmId:    d.Get("realm_id").(string),
		Name:       d.Get("name").(string),
		ProviderId: clientRegistrationPolicyTrustedHostsProviderId,
		SubType:    d.Get("sub_type").(string),
		Config: map[string][]string{
			"trusted-hosts":                                configStringSliceFromSet(d.Get("trusted_hosts").(*schema.Set)),
			"host-sending-registration-request-must-match": configBoolValue(d.Get("host_sending_registration_request_must_match").(bool)),
			"client-uris-must-match":                       configBoolValue(d.Get("client_uris_must_match").(bool)),
		},
	}
}

func setClientRegistrationPolicyTrustedHostsData(d *schema.ResourceData, policy *keycloak.ClientRegistrationPolicy) {
	readClientRegistrationPolicyCommon(d, policy)
	d.Set("trusted_hosts", policy.Config["trusted-hosts"])
	d.Set("host_sending_registration_request_must_match", configBool(policy.Config["host-sending-registration-request-must-match"], true))
	d.Set("client_uris_must_match", configBool(policy.Config["client-uris-must-match"], true))
}

func resourceKeycloakClientRegistrationPolicyTrustedHostsCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	policy := getClientRegistrationPolicyTrustedHostsFromData(d)
	if err := keycloakClient.NewClientRegistrationPolicy(ctx, policy); err != nil {
		return diag.FromErr(err)
	}
	setClientRegistrationPolicyTrustedHostsData(d, policy)
	return resourceKeycloakClientRegistrationPolicyTrustedHostsRead(ctx, d, meta)
}

func resourceKeycloakClientRegistrationPolicyTrustedHostsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	policy, err := keycloakClient.GetClientRegistrationPolicy(ctx, d.Get("realm_id").(string), d.Id())
	if err != nil {
		return handleNotFoundError(ctx, err, d)
	}
	setClientRegistrationPolicyTrustedHostsData(d, policy)
	return nil
}

func resourceKeycloakClientRegistrationPolicyTrustedHostsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	policy := getClientRegistrationPolicyTrustedHostsFromData(d)
	if err := keycloakClient.UpdateClientRegistrationPolicy(ctx, policy); err != nil {
		return diag.FromErr(err)
	}
	setClientRegistrationPolicyTrustedHostsData(d, policy)
	return nil
}
