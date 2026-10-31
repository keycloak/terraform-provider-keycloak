package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

var (
	keycloakRealmKeystoreEcdhGeneratedEllipticCurve = []string{"P-256", "P-384", "P-521"}
	keycloakRealmKeystoreEcdhGeneratedAlgorithm     = []string{"ECDH-ES", "ECDH-ES+A128KW", "ECDH-ES+A192KW", "ECDH-ES+A256KW"}
)

func resourceKeycloakRealmKeystoreEcdhGenerated() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakRealmKeystoreEcdhGeneratedCreate,
		ReadContext:   resourceKeycloakRealmKeystoreEcdhGeneratedRead,
		UpdateContext: resourceKeycloakRealmKeystoreEcdhGeneratedUpdate,
		DeleteContext: resourceKeycloakRealmKeystoreEcdhGeneratedDelete,
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
				Description: "Set if the keys can be used for encryption",
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
				ValidateFunc: validation.StringInSlice(keycloakRealmKeystoreEcdhGeneratedEllipticCurve, false),
				Default:      "P-256",
				Description:  "Elliptic Curve used in ECDH",
			},
			"algorithm": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(keycloakRealmKeystoreEcdhGeneratedAlgorithm, false),
				Default:      "ECDH-ES",
				Description:  "Algorithm for processing the Content Encryption Key",
			},
		},
	}
}

func getRealmKeystoreEcdhGeneratedFromData(data *schema.ResourceData) (*keycloak.RealmKeystoreEcdhGenerated, error) {
	keystore := &keycloak.RealmKeystoreEcdhGenerated{
		Id:      data.Id(),
		Name:    data.Get("name").(string),
		RealmId: data.Get("realm_id").(string),

		Active:        data.Get("active").(bool),
		Enabled:       data.Get("enabled").(bool),
		Priority:      data.Get("priority").(int),
		EllipticCurve: data.Get("elliptic_curve_key").(string),
		Algorithm:     data.Get("algorithm").(string),
	}

	return keystore, nil
}

func setRealmKeystoreEcdhGeneratedData(data *schema.ResourceData, realmKey *keycloak.RealmKeystoreEcdhGenerated) error {
	data.SetId(realmKey.Id)

	data.Set("name", realmKey.Name)
	data.Set("realm_id", realmKey.RealmId)

	data.Set("active", realmKey.Active)
	data.Set("enabled", realmKey.Enabled)
	data.Set("priority", realmKey.Priority)
	data.Set("elliptic_curve_key", realmKey.EllipticCurve)
	data.Set("algorithm", realmKey.Algorithm)

	return nil
}

func resourceKeycloakRealmKeystoreEcdhGeneratedCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmKey, err := getRealmKeystoreEcdhGeneratedFromData(data)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.NewRealmKeystoreEcdhGenerated(ctx, realmKey)
	if err != nil {
		return diag.FromErr(err)
	}

	err = setRealmKeystoreEcdhGeneratedData(data, realmKey)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceKeycloakRealmKeystoreEcdhGeneratedRead(ctx, data, meta)
}

func resourceKeycloakRealmKeystoreEcdhGeneratedRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	realmKey, err := keycloakClient.GetRealmKeystoreEcdhGenerated(ctx, realmId, id)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	err = setRealmKeystoreEcdhGeneratedData(data, realmKey)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakRealmKeystoreEcdhGeneratedUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmKey, err := getRealmKeystoreEcdhGeneratedFromData(data)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.UpdateRealmKeystoreEcdhGenerated(ctx, realmKey)
	if err != nil {
		return diag.FromErr(err)
	}

	err = setRealmKeystoreEcdhGeneratedData(data, realmKey)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakRealmKeystoreEcdhGeneratedDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	return diag.FromErr(keycloakClient.DeleteRealmKeystoreEcdhGenerated(ctx, realmId, id))
}
