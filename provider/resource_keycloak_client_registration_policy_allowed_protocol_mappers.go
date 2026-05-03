package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

const clientRegistrationPolicyAllowedProtocolMappersProviderId = "allowed-protocol-mappers"

func resourceKeycloakClientRegistrationPolicyAllowedProtocolMappers() *schema.Resource {
	s := commonClientRegistrationPolicySchema()
	// The legal values for this list are populated dynamically from registered
	// ProtocolMapper provider factories at runtime, so the schema deliberately
	// does not enforce a closed enum here.
	s["allowed_protocol_mapper_types"] = &schema.Schema{
		Type:        schema.TypeSet,
		Optional:    true,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Description: "Whitelist of protocol mapper provider IDs that may appear on a registered client.",
	}

	return &schema.Resource{
		CreateContext: resourceKeycloakClientRegistrationPolicyAllowedProtocolMappersCreate,
		ReadContext:   resourceKeycloakClientRegistrationPolicyAllowedProtocolMappersRead,
		UpdateContext: resourceKeycloakClientRegistrationPolicyAllowedProtocolMappersUpdate,
		DeleteContext: deleteClientRegistrationPolicy,
		Importer:      clientRegistrationPolicyImporter(clientRegistrationPolicyAllowedProtocolMappersProviderId),
		Schema:        s,
	}
}

func getClientRegistrationPolicyAllowedProtocolMappersFromData(d *schema.ResourceData) *keycloak.ClientRegistrationPolicy {
	return &keycloak.ClientRegistrationPolicy{
		Id:         d.Id(),
		RealmId:    d.Get("realm_id").(string),
		Name:       d.Get("name").(string),
		ProviderId: clientRegistrationPolicyAllowedProtocolMappersProviderId,
		SubType:    d.Get("sub_type").(string),
		Config: map[string][]string{
			"allowed-protocol-mapper-types": configStringSliceFromSet(d.Get("allowed_protocol_mapper_types").(*schema.Set)),
		},
	}
}

func setClientRegistrationPolicyAllowedProtocolMappersData(d *schema.ResourceData, policy *keycloak.ClientRegistrationPolicy) {
	readClientRegistrationPolicyCommon(d, policy)
	d.Set("allowed_protocol_mapper_types", policy.Config["allowed-protocol-mapper-types"])
}

func resourceKeycloakClientRegistrationPolicyAllowedProtocolMappersCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	policy := getClientRegistrationPolicyAllowedProtocolMappersFromData(d)
	if err := keycloakClient.NewClientRegistrationPolicy(ctx, policy); err != nil {
		return diag.FromErr(err)
	}
	setClientRegistrationPolicyAllowedProtocolMappersData(d, policy)
	return resourceKeycloakClientRegistrationPolicyAllowedProtocolMappersRead(ctx, d, meta)
}

func resourceKeycloakClientRegistrationPolicyAllowedProtocolMappersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	policy, err := keycloakClient.GetClientRegistrationPolicy(ctx, d.Get("realm_id").(string), d.Id())
	if err != nil {
		return handleNotFoundError(ctx, err, d)
	}
	setClientRegistrationPolicyAllowedProtocolMappersData(d, policy)
	return nil
}

func resourceKeycloakClientRegistrationPolicyAllowedProtocolMappersUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	policy := getClientRegistrationPolicyAllowedProtocolMappersFromData(d)
	if err := keycloakClient.UpdateClientRegistrationPolicy(ctx, policy); err != nil {
		return diag.FromErr(err)
	}
	setClientRegistrationPolicyAllowedProtocolMappersData(d, policy)
	return nil
}
