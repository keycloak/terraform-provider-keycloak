package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mrparkers/terraform-provider-keycloak/keycloak"
	"reflect"
)

func resourceKeycloakRealmKeystoreCustom() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakRealmKeystoreCustomCreate,
		ReadContext:   resourceKeycloakRealmKeystoreCustomRead,
		UpdateContext: resourceKeycloakRealmKeystoreCustomUpdate,
		DeleteContext: resourceKeycloakRealmKeystoreCustomDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceKeycloakRealmKeystoreGenericImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Display name of provider when linked in admin console.",
			},
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"active": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Set if the keys are enabled",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Set if the keys are enabled",
			},
			"priority": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "Priority for the provider",
			},
			"provider_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the custom provider",
			},
			"provider_type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The type of the custom provider",
			},
			"extra_config": {
				Type:             schema.TypeMap,
				Optional:         true,
				ValidateDiagFunc: validateExtraConfig(reflect.ValueOf(&keycloak.RealmKeystoreCustom{}).Elem()),
			},
		},
	}
}

func getRealmKeystoreCustomFromData(data *schema.ResourceData) (*keycloak.RealmKeystoreCustom, error) {
	keystore := &keycloak.RealmKeystoreCustom{
		Id:      data.Id(),
		Name:    data.Get("name").(string),
		RealmId: data.Get("realm_id").(string),

		Active:       data.Get("active").(bool),
		Enabled:      data.Get("enabled").(bool),
		Priority:     data.Get("priority").(int),
		ProviderId:   data.Get("provider_id").(string),
		ProviderType: data.Get("provider_type").(string),

		ExtraConfig: getExtraConfigFromData(data),
	}

	return keystore, nil
}

func setRealmKeystoreCustomData(data *schema.ResourceData, realmKey *keycloak.RealmKeystoreCustom) error {
	data.SetId(realmKey.Id)

	data.Set("name", realmKey.Name)
	data.Set("realm_id", realmKey.RealmId)

	data.Set("active", realmKey.Active)
	data.Set("enabled", realmKey.Enabled)
	data.Set("priority", realmKey.Priority)
	data.Set("provider_id", realmKey.ProviderId)
	data.Set("provider_type", realmKey.ProviderType)

	setExtraConfigData(data, realmKey.ExtraConfig)

	return nil
}

func resourceKeycloakRealmKeystoreCustomCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmKey, err := getRealmKeystoreCustomFromData(data)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.NewRealmKeystoreCustom(ctx, realmKey)
	if err != nil {
		return diag.FromErr(err)
	}

	err = setRealmKeystoreCustomData(data, realmKey)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceKeycloakRealmKeystoreCustomRead(ctx, data, meta)
}

func resourceKeycloakRealmKeystoreCustomRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	realmKey, err := keycloakClient.GetRealmKeystoreCustom(ctx, realmId, id)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	err = setRealmKeystoreCustomData(data, realmKey)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakRealmKeystoreCustomUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmKey, err := getRealmKeystoreCustomFromData(data)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.UpdateRealmKeystoreCustom(ctx, realmKey)
	if err != nil {
		return diag.FromErr(err)
	}

	err = setRealmKeystoreCustomData(data, realmKey)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakRealmKeystoreCustomDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	return diag.FromErr(keycloakClient.DeleteRealmKeystoreCustom(ctx, realmId, id))
}
