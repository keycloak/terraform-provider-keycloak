package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

// Note: the wire-level providerId is "allowed-client-templates" — a legacy
// name from when client scopes were called templates. The config keys use the
// modern "scopes" naming. The resource is named after the modern UI label.
const clientRegistrationPolicyAllowedClientScopesProviderId = "allowed-client-templates"

func resourceKeycloakClientRegistrationPolicyAllowedClientScopes() *schema.Resource {
	s := commonClientRegistrationPolicySchema()
	s["allowed_client_scopes"] = &schema.Schema{
		Type:        schema.TypeSet,
		Optional:    true,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Description: "Names of client scopes that newly-registered clients are allowed to claim.",
	}
	s["allow_default_scopes"] = &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
		Description: "If true, the realm's default client scopes are always permitted regardless of allowed_client_scopes.",
	}

	return &schema.Resource{
		CreateContext: resourceKeycloakClientRegistrationPolicyAllowedClientScopesCreate,
		ReadContext:   resourceKeycloakClientRegistrationPolicyAllowedClientScopesRead,
		UpdateContext: resourceKeycloakClientRegistrationPolicyAllowedClientScopesUpdate,
		DeleteContext: deleteClientRegistrationPolicy,
		Importer:      clientRegistrationPolicyImporter(clientRegistrationPolicyAllowedClientScopesProviderId),
		Schema:        s,
	}
}

func getClientRegistrationPolicyAllowedClientScopesFromData(d *schema.ResourceData) *keycloak.ClientRegistrationPolicy {
	return &keycloak.ClientRegistrationPolicy{
		Id:         d.Id(),
		RealmId:    d.Get("realm_id").(string),
		Name:       d.Get("name").(string),
		ProviderId: clientRegistrationPolicyAllowedClientScopesProviderId,
		SubType:    d.Get("sub_type").(string),
		Config: map[string][]string{
			"allowed-client-scopes": configStringSliceFromSet(d.Get("allowed_client_scopes").(*schema.Set)),
			"allow-default-scopes":  configBoolValue(d.Get("allow_default_scopes").(bool)),
		},
	}
}

func setClientRegistrationPolicyAllowedClientScopesData(d *schema.ResourceData, policy *keycloak.ClientRegistrationPolicy) {
	readClientRegistrationPolicyCommon(d, policy)
	d.Set("allowed_client_scopes", policy.Config["allowed-client-scopes"])
	d.Set("allow_default_scopes", configBool(policy.Config["allow-default-scopes"], true))
}

func resourceKeycloakClientRegistrationPolicyAllowedClientScopesCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	policy := getClientRegistrationPolicyAllowedClientScopesFromData(d)
	if err := keycloakClient.NewClientRegistrationPolicy(ctx, policy); err != nil {
		return diag.FromErr(err)
	}
	setClientRegistrationPolicyAllowedClientScopesData(d, policy)
	return resourceKeycloakClientRegistrationPolicyAllowedClientScopesRead(ctx, d, meta)
}

func resourceKeycloakClientRegistrationPolicyAllowedClientScopesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	policy, err := keycloakClient.GetClientRegistrationPolicy(ctx, d.Get("realm_id").(string), d.Id())
	if err != nil {
		return handleNotFoundError(ctx, err, d)
	}
	setClientRegistrationPolicyAllowedClientScopesData(d, policy)
	return nil
}

func resourceKeycloakClientRegistrationPolicyAllowedClientScopesUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	policy := getClientRegistrationPolicyAllowedClientScopesFromData(d)
	if err := keycloakClient.UpdateClientRegistrationPolicy(ctx, policy); err != nil {
		return diag.FromErr(err)
	}
	setClientRegistrationPolicyAllowedClientScopesData(d, policy)
	return nil
}
