package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

var (
	keycloakRealmKeystoreRsaGeneratedSize      = []int{1024, 2048, 4096}
	keycloakRealmKeystoreRsaGeneratedAlgorithm = []string{"RS256", "RS384", "RS512", "PS256", "PS384", "PS512"}
)

func resourceKeycloakRealmKeystoreRsaGenerated() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakRealmKeystoreRsaGeneratedCreate,
		ReadContext:   resourceKeycloakRealmKeystoreRsaGeneratedRead,
		UpdateContext: resourceKeycloakRealmKeystoreRsaGeneratedUpdate,
		DeleteContext: resourceKeycloakRealmKeystoreRsaGeneratedDelete,
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
				ValidateFunc: validation.StringInSlice(keycloakRealmKeystoreRsaGeneratedAlgorithm, false),
				Default:      "RS256",
				Description:  "Intended algorithm for the key",
			},
			"key_size": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntInSlice(keycloakRealmKeystoreRsaGeneratedSize),
				Default:      2048,
				Description:  "Size for the generated keys",
			},
		},
	}
}

func getRealmKeystoreRsaGeneratedFromData(data *schema.ResourceData) (*keycloak.RealmKeystoreRsaGenerated, error) {
	keystore := &keycloak.RealmKeystoreRsaGenerated{
		Id:      data.Id(),
		Name:    data.Get("name").(string),
		RealmId: data.Get("realm_id").(string),

		Active:    data.Get("active").(bool),
		Enabled:   data.Get("enabled").(bool),
		Priority:  data.Get("priority").(int),
		KeySize:   data.Get("key_size").(int),
		Algorithm: data.Get("algorithm").(string),
	}

	return keystore, nil
}

func setRealmKeystoreRsaGeneratedData(data *schema.ResourceData, realmKey *keycloak.RealmKeystoreRsaGenerated) error {
	data.SetId(realmKey.Id)

	data.Set("name", realmKey.Name)
	data.Set("realm_id", realmKey.RealmId)

	data.Set("active", realmKey.Active)
	data.Set("enabled", realmKey.Enabled)
	data.Set("priority", realmKey.Priority)
	data.Set("key_size", realmKey.KeySize)
	data.Set("algorithm", realmKey.Algorithm)

	return nil
}

func resourceKeycloakRealmKeystoreRsaGeneratedCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmKey, err := getRealmKeystoreRsaGeneratedFromData(data)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.NewRealmKeystoreRsaGenerated(ctx, realmKey)
	if err != nil {
		return diag.FromErr(err)
	}

	err = setRealmKeystoreRsaGeneratedData(data, realmKey)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceKeycloakRealmKeystoreRsaGeneratedRead(ctx, data, meta)
}

func resourceKeycloakRealmKeystoreRsaGeneratedRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	realmKey, err := keycloakClient.GetRealmKeystoreRsaGenerated(ctx, realmId, id)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	err = setRealmKeystoreRsaGeneratedData(data, realmKey)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakRealmKeystoreRsaGeneratedUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmKey, err := getRealmKeystoreRsaGeneratedFromData(data)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.UpdateRealmKeystoreRsaGenerated(ctx, realmKey)
	if err != nil {
		return diag.FromErr(err)
	}

	err = setRealmKeystoreRsaGeneratedData(data, realmKey)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakRealmKeystoreRsaGeneratedDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	return diag.FromErr(keycloakClient.DeleteRealmKeystoreRsaGenerated(ctx, realmId, id))
}
