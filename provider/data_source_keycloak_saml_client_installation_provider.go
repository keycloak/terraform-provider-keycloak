package provider

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func dataSourceKeycloakSamlClientInstallationProvider() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKeycloakSamlClientInstallationProviderRead,
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"client_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"provider_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"value": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceKeycloakSamlClientInstallationProviderRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)
	providerId := data.Get("provider_id").(string)

	value, err := keycloakClient.GetSamlClientInstallationProvider(ctx, realmId, clientId, providerId)
	if err != nil {
		return diag.FromErr(err)
	}

	h := sha1.New()
	h.Write(value)
	id := base64.URLEncoding.EncodeToString(h.Sum(nil))

	data.SetId(id)
	data.Set("realm_id", realmId)
	data.Set("client_id", clientId)
	data.Set("provider_id", providerId)
	data.Set("value", string(value))

	return nil
}
