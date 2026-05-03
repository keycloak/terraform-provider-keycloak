package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

const clientRegistrationPolicyWebOriginsProviderId = "registration-web-origins"

func resourceKeycloakClientRegistrationPolicyWebOrigins() *schema.Resource {
	s := commonClientRegistrationPolicySchema()
	s["web_origins"] = &schema.Schema{
		Type:        schema.TypeSet,
		Optional:    true,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Description: "Allowed web origins for client registration requests.",
	}

	return &schema.Resource{
		CreateContext: resourceKeycloakClientRegistrationPolicyWebOriginsCreate,
		ReadContext:   resourceKeycloakClientRegistrationPolicyWebOriginsRead,
		UpdateContext: resourceKeycloakClientRegistrationPolicyWebOriginsUpdate,
		DeleteContext: deleteClientRegistrationPolicy,
		Importer:      clientRegistrationPolicyImporter(clientRegistrationPolicyWebOriginsProviderId),
		Schema:        s,
	}
}

func getClientRegistrationPolicyWebOriginsFromData(d *schema.ResourceData) *keycloak.ClientRegistrationPolicy {
	return &keycloak.ClientRegistrationPolicy{
		Id:         d.Id(),
		RealmId:    d.Get("realm_id").(string),
		Name:       d.Get("name").(string),
		ProviderId: clientRegistrationPolicyWebOriginsProviderId,
		SubType:    d.Get("sub_type").(string),
		Config: map[string][]string{
			"web-origins": configStringSliceFromSet(d.Get("web_origins").(*schema.Set)),
		},
	}
}

func setClientRegistrationPolicyWebOriginsData(d *schema.ResourceData, policy *keycloak.ClientRegistrationPolicy) {
	readClientRegistrationPolicyCommon(d, policy)
	d.Set("web_origins", policy.Config["web-origins"])
}

func resourceKeycloakClientRegistrationPolicyWebOriginsCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	policy := getClientRegistrationPolicyWebOriginsFromData(d)
	if err := keycloakClient.NewClientRegistrationPolicy(ctx, policy); err != nil {
		return diag.FromErr(err)
	}
	setClientRegistrationPolicyWebOriginsData(d, policy)
	return resourceKeycloakClientRegistrationPolicyWebOriginsRead(ctx, d, meta)
}

func resourceKeycloakClientRegistrationPolicyWebOriginsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	policy, err := keycloakClient.GetClientRegistrationPolicy(ctx, d.Get("realm_id").(string), d.Id())
	if err != nil {
		return handleNotFoundError(ctx, err, d)
	}
	setClientRegistrationPolicyWebOriginsData(d, policy)
	return nil
}

func resourceKeycloakClientRegistrationPolicyWebOriginsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	policy := getClientRegistrationPolicyWebOriginsFromData(d)
	if err := keycloakClient.UpdateClientRegistrationPolicy(ctx, policy); err != nil {
		return diag.FromErr(err)
	}
	setClientRegistrationPolicyWebOriginsData(d, policy)
	return nil
}
