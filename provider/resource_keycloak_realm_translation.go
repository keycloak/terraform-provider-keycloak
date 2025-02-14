package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakRealmTranslation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakRealmTranslationUpdate,
		ReadContext:   resourceKeycloakRealmTranslationRead,
		DeleteContext: resourceKeycloakRealmTranslationDelete,
		UpdateContext: resourceKeycloakRealmTranslationUpdate,
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"language": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"translations": {
				Optional: true,
				Type:     schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceKeycloakRealmTranslationRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	realmId := data.Get("realm_id").(string)
	language := data.Get("language").(string)
	realmLanguageTranslations, err := keycloakClient.GetRealmTranslations(ctx, realmId, language)
	if err != nil {
		return diag.FromErr(err)
	}
	data.Set("translations", realmLanguageTranslations)
	return nil
}

func resourceKeycloakRealmTranslationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*keycloak.KeycloakClient)
	realm := d.Get("realm_id").(string)
	language := d.Get("language").(string)
	translations := d.Get("translations").(map[string]interface{})
	translationsConverted := convertTranslations(translations)

	err := client.UpdateRealmTranslations(ctx, realm, language, translationsConverted)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s/%s", realm, language)) // Set resource ID as "realm/language"
	return resourceKeycloakRealmTranslationRead(ctx, d, meta)
}

func resourceKeycloakRealmTranslationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*keycloak.KeycloakClient)
	realm := d.Get("realm_id").(string)
	language := d.Get("language").(string)
	translations := d.Get("translations").(map[string]interface{})
	translationsConverted := convertTranslations(translations)

	err := client.DeleteRealmTranslations(ctx, realm, language, translationsConverted)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func convertTranslations(translations map[string]interface{}) map[string]string {
	translionsConverted := make(map[string]string)
	for key, value := range translations {
		strValue, ok := value.(string)
		if !ok {
			panic(fmt.Sprintf("expected string, got %T for key %s", value, key))
		}
		translionsConverted[key] = strValue
	}

	return translionsConverted
}
