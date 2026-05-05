package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

const clientRegistrationPolicyMaxClientsProviderId = "max-clients"
const clientRegistrationPolicyMaxClientsDefault = 200

func resourceKeycloakClientRegistrationPolicyMaxClients() *schema.Resource {
	s := commonClientRegistrationPolicySchema()
	s["max_clients"] = &schema.Schema{
		Type:         schema.TypeInt,
		Optional:     true,
		Default:      clientRegistrationPolicyMaxClientsDefault,
		ValidateFunc: validation.IntAtLeast(1),
		Description:  "Maximum number of clients allowed in the realm. Once reached, further DCR requests fail.",
	}

	return &schema.Resource{
		CreateContext: resourceKeycloakClientRegistrationPolicyMaxClientsCreate,
		ReadContext:   resourceKeycloakClientRegistrationPolicyMaxClientsRead,
		UpdateContext: resourceKeycloakClientRegistrationPolicyMaxClientsUpdate,
		DeleteContext: deleteClientRegistrationPolicy,
		Importer:      clientRegistrationPolicyImporter(clientRegistrationPolicyMaxClientsProviderId),
		Schema:        s,
	}
}

func getClientRegistrationPolicyMaxClientsFromData(d *schema.ResourceData) *keycloak.ClientRegistrationPolicy {
	return &keycloak.ClientRegistrationPolicy{
		Id:         d.Id(),
		RealmId:    d.Get("realm_id").(string),
		Name:       d.Get("name").(string),
		ProviderId: clientRegistrationPolicyMaxClientsProviderId,
		SubType:    d.Get("sub_type").(string),
		Config: map[string][]string{
			"max-clients": {strconv.Itoa(d.Get("max_clients").(int))},
		},
	}
}

func setClientRegistrationPolicyMaxClientsData(d *schema.ResourceData, policy *keycloak.ClientRegistrationPolicy) {
	readClientRegistrationPolicyCommon(d, policy)
	max := clientRegistrationPolicyMaxClientsDefault
	if values := policy.Config["max-clients"]; len(values) > 0 {
		if n, err := strconv.Atoi(values[0]); err == nil {
			max = n
		}
	}
	d.Set("max_clients", max)
}

func resourceKeycloakClientRegistrationPolicyMaxClientsCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	policy := getClientRegistrationPolicyMaxClientsFromData(d)
	if err := keycloakClient.NewClientRegistrationPolicy(ctx, policy); err != nil {
		return diag.FromErr(err)
	}
	setClientRegistrationPolicyMaxClientsData(d, policy)
	return resourceKeycloakClientRegistrationPolicyMaxClientsRead(ctx, d, meta)
}

func resourceKeycloakClientRegistrationPolicyMaxClientsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	policy, err := keycloakClient.GetClientRegistrationPolicy(ctx, d.Get("realm_id").(string), d.Id())
	if err != nil {
		return handleNotFoundError(ctx, err, d)
	}
	setClientRegistrationPolicyMaxClientsData(d, policy)
	return nil
}

func resourceKeycloakClientRegistrationPolicyMaxClientsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	policy := getClientRegistrationPolicyMaxClientsFromData(d)
	if err := keycloakClient.UpdateClientRegistrationPolicy(ctx, policy); err != nil {
		return diag.FromErr(err)
	}
	setClientRegistrationPolicyMaxClientsData(d, policy)
	return nil
}
