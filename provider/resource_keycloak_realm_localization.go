package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakRealmLocalization() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakRealmLocalizationTextsUpdate,
		ReadContext:   resourceKeycloakRealmLocalizationTextsRead,
		DeleteContext: resourceKeycloakRealmLocalizationTextsDelete,
		UpdateContext: resourceKeycloakRealmLocalizationTextsUpdate,
		Description:   "Manage realm-level localization texts.",
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The realm in which the texts exists.",
			},
			"locale": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The locale for the localization texts.",
			},
			"texts": {
				Optional: true,
				Type:     schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "The mapping of localization texts keys to values.",
			},
		},
	}
}

func resourceKeycloakRealmLocalizationTextsRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	realmId := data.Get("realm_id").(string)
	locale := data.Get("locale").(string)
	realmLocaleTexts, err := keycloakClient.GetRealmLocalizationTexts(ctx, realmId, locale)
	if err != nil {
		return diag.FromErr(err)
	}
	data.Set("texts", realmLocaleTexts)
	return nil
}

func resourceKeycloakRealmLocalizationTextsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*keycloak.KeycloakClient)
	realm := d.Get("realm_id").(string)
	locale := d.Get("locale").(string)
	texts := d.Get("texts").(map[string]interface{})
	textsConverted := convertTexts(texts)

	err := client.UpdateRealmLocalizationTexts(ctx, realm, locale, textsConverted)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s/%s", realm, locale)) // Set resource ID as "realm/locale"
	return resourceKeycloakRealmLocalizationTextsRead(ctx, d, meta)
}

func resourceKeycloakRealmLocalizationTextsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*keycloak.KeycloakClient)
	realm := d.Get("realm_id").(string)
	locale := d.Get("locale").(string)
	texts := d.Get("texts").(map[string]interface{})
	textsConverted := convertTexts(texts)

	err := client.DeleteRealmLocalizationTexts(ctx, realm, locale, textsConverted)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func convertTexts(texts map[string]interface{}) map[string]string {
	translionsConverted := make(map[string]string)
	for key, value := range texts {
		strValue, ok := value.(string)
		if !ok {
			panic(fmt.Sprintf("expected string, got %T for key %s", value, key))
		}
		translionsConverted[key] = strValue
	}

	return translionsConverted
}
