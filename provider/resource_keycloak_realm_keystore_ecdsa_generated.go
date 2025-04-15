package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

var (
	keycloakRealmKeystoreEcdsaGeneratedEllipticCurve = []string{"P-256", "P-384", "P-521"}
)

func resourceKeycloakRealmKeystoreEcdsaGenerated() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakRealmKeystoreEcdsaGeneratedCreate,
		ReadContext:   resourceKeycloakRealmKeystoreEcdsaGeneratedRead,
		UpdateContext: resourceKeycloakRealmKeystoreEcdsaGeneratedUpdate,
		DeleteContext: resourceKeycloakRealmKeystoreEcdsaGeneratedDelete,
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
				Description: "Set if the keys can be used for signing",
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
			"elliptic_curve_key": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(keycloakRealmKeystoreEcdsaGeneratedEllipticCurve, false),
				Default:      "P-256",
				Description:  "Elliptic Curve used in ECDSA",
			},
		},
	}
}

func getRealmKeystoreEcdsaGeneratedFromData(data *schema.ResourceData) (*keycloak.RealmKeystoreEcdsaGenerated, error) {
	keystore := &keycloak.RealmKeystoreEcdsaGenerated{
		Id:      data.Id(),
		Name:    data.Get("name").(string),
		RealmId: data.Get("realm_id").(string),

		Active:        data.Get("active").(bool),
		Enabled:       data.Get("enabled").(bool),
		Priority:      data.Get("priority").(int),
		EllipticCurve: data.Get("elliptic_curve_key").(string),
	}

	return keystore, nil
}

func setRealmKeystoreEcdsaGeneratedData(data *schema.ResourceData, realmKey *keycloak.RealmKeystoreEcdsaGenerated) error {
	data.SetId(realmKey.Id)

	data.Set("name", realmKey.Name)
	data.Set("realm_id", realmKey.RealmId)

	data.Set("active", realmKey.Active)
	data.Set("enabled", realmKey.Enabled)
	data.Set("priority", realmKey.Priority)
	data.Set("elliptic_curve_key", realmKey.EllipticCurve)

	return nil
}

func resourceKeycloakRealmKeystoreEcdsaGeneratedCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmKey, err := getRealmKeystoreEcdsaGeneratedFromData(data)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.NewRealmKeystoreEcdsaGenerated(ctx, realmKey)
	if err != nil {
		return diag.FromErr(err)
	}

	err = setRealmKeystoreEcdsaGeneratedData(data, realmKey)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceKeycloakRealmKeystoreEcdsaGeneratedRead(ctx, data, meta)
}

func resourceKeycloakRealmKeystoreEcdsaGeneratedRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	realmKey, err := keycloakClient.GetRealmKeystoreEcdsaGenerated(ctx, realmId, id)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	err = setRealmKeystoreEcdsaGeneratedData(data, realmKey)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakRealmKeystoreEcdsaGeneratedUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmKey, err := getRealmKeystoreEcdsaGeneratedFromData(data)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.UpdateRealmKeystoreEcdsaGenerated(ctx, realmKey)
	if err != nil {
		return diag.FromErr(err)
	}

	err = setRealmKeystoreEcdsaGeneratedData(data, realmKey)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakRealmKeystoreEcdsaGeneratedDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	return diag.FromErr(keycloakClient.DeleteRealmKeystoreEcdsaGenerated(ctx, realmId, id))
}
