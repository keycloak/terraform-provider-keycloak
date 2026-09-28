package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

var (
	keycloakRealmKeystoreRsaEncGeneratedSize      = []int{1024, 2048, 4096}
	keycloakRealmKeystoreRsaEncGeneratedAlgorithm = []string{"RSA-OAEP", "RSA-OAEP-256", "RSA1_5"}
)

func resourceKeycloakRealmKeystoreRsaEncGenerated() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakRealmKeystoreRsaEncGeneratedCreate,
		ReadContext:   resourceKeycloakRealmKeystoreRsaEncGeneratedRead,
		UpdateContext: resourceKeycloakRealmKeystoreRsaEncGeneratedUpdate,
		DeleteContext: resourceKeycloakRealmKeystoreRsaEncGeneratedDelete,
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
			"algorithm": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(keycloakRealmKeystoreRsaEncGeneratedAlgorithm, false),
				Default:      "RSA-OAEP",
				Description:  "Intended algorithm for the key",
			},
			"key_size": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntInSlice(keycloakRealmKeystoreRsaEncGeneratedSize),
				Default:      2048,
				Description:  "Size for the generated keys",
			},
		},
	}
}

func getRealmKeystoreRsaEncGeneratedFromData(data *schema.ResourceData) (*keycloak.RealmKeystoreRsaEncGenerated, error) {
	keystore := &keycloak.RealmKeystoreRsaEncGenerated{
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

func setRealmKeystoreRsaEncGeneratedData(data *schema.ResourceData, realmKey *keycloak.RealmKeystoreRsaEncGenerated) error {
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

func resourceKeycloakRealmKeystoreRsaEncGeneratedCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmKey, err := getRealmKeystoreRsaEncGeneratedFromData(data)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.NewRealmKeystoreRsaEncGenerated(ctx, realmKey)
	if err != nil {
		return diag.FromErr(err)
	}

	err = setRealmKeystoreRsaEncGeneratedData(data, realmKey)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceKeycloakRealmKeystoreRsaEncGeneratedRead(ctx, data, meta)
}

func resourceKeycloakRealmKeystoreRsaEncGeneratedRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	realmKey, err := keycloakClient.GetRealmKeystoreRsaEncGenerated(ctx, realmId, id)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	err = setRealmKeystoreRsaEncGeneratedData(data, realmKey)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakRealmKeystoreRsaEncGeneratedUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmKey, err := getRealmKeystoreRsaEncGeneratedFromData(data)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.UpdateRealmKeystoreRsaEncGenerated(ctx, realmKey)
	if err != nil {
		return diag.FromErr(err)
	}

	err = setRealmKeystoreRsaEncGeneratedData(data, realmKey)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakRealmKeystoreRsaEncGeneratedDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	return diag.FromErr(keycloakClient.DeleteRealmKeystoreRsaEncGenerated(ctx, realmId, id))
}
