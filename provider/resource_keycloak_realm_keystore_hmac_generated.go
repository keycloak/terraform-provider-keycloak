package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

var (
	keycloakRealmKeystoreHmacGeneratedSize      = []int{16, 24, 32, 64, 128, 256, 512}
	keycloakRealmKeystoreHmacGeneratedAlgorithm = []string{"HS256", "HS384", "HS512"}
)

func resourceKeycloakRealmKeystoreHmacGenerated() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakRealmKeystoreHmacGeneratedCreate,
		ReadContext:   resourceKeycloakRealmKeystoreHmacGeneratedRead,
		UpdateContext: resourceKeycloakRealmKeystoreHmacGeneratedUpdate,
		DeleteContext: resourceKeycloakRealmKeystoreHmacGeneratedDelete,
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
			"algorithm": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(keycloakRealmKeystoreHmacGeneratedAlgorithm, false),
				Default:      "HS256",
				Description:  "Intended algorithm for the key",
			},
			"secret_size": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntInSlice(keycloakRealmKeystoreHmacGeneratedSize),
				Default:      64,
				Description:  "Size in bytes for the generated secret",
			},
		},
	}
}

func getRealmKeystoreHmacGeneratedFromData(data *schema.ResourceData) (*keycloak.RealmKeystoreHmacGenerated, error) {
	keystore := &keycloak.RealmKeystoreHmacGenerated{
		Id:      data.Id(),
		Name:    data.Get("name").(string),
		RealmId: data.Get("realm_id").(string),

		Active:     data.Get("active").(bool),
		Enabled:    data.Get("enabled").(bool),
		Priority:   data.Get("priority").(int),
		SecretSize: data.Get("secret_size").(int),
		Algorithm:  data.Get("algorithm").(string),
	}

	return keystore, nil
}

func setRealmKeystoreHmacGeneratedData(data *schema.ResourceData, realmKey *keycloak.RealmKeystoreHmacGenerated) error {
	data.SetId(realmKey.Id)

	data.Set("name", realmKey.Name)
	data.Set("realm_id", realmKey.RealmId)

	data.Set("active", realmKey.Active)
	data.Set("enabled", realmKey.Enabled)
	data.Set("priority", realmKey.Priority)
	data.Set("secret_size", realmKey.SecretSize)
	data.Set("algorithm", realmKey.Algorithm)

	return nil
}

func resourceKeycloakRealmKeystoreHmacGeneratedCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmKey, err := getRealmKeystoreHmacGeneratedFromData(data)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.NewRealmKeystoreHmacGenerated(ctx, realmKey)
	if err != nil {
		return diag.FromErr(err)
	}

	err = setRealmKeystoreHmacGeneratedData(data, realmKey)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceKeycloakRealmKeystoreHmacGeneratedRead(ctx, data, meta)
}

func resourceKeycloakRealmKeystoreHmacGeneratedRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	realmKey, err := keycloakClient.GetRealmKeystoreHmacGenerated(ctx, realmId, id)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	err = setRealmKeystoreHmacGeneratedData(data, realmKey)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakRealmKeystoreHmacGeneratedUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmKey, err := getRealmKeystoreHmacGeneratedFromData(data)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.UpdateRealmKeystoreHmacGenerated(ctx, realmKey)
	if err != nil {
		return diag.FromErr(err)
	}

	err = setRealmKeystoreHmacGeneratedData(data, realmKey)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakRealmKeystoreHmacGeneratedDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	return diag.FromErr(keycloakClient.DeleteRealmKeystoreHmacGenerated(ctx, realmId, id))
}
