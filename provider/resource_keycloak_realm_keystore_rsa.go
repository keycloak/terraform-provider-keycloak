package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

var (
	keycloakRealmKeystoreRsaAlgorithm = []string{"RS256", "RS384", "RS512", "PS256", "PS384", "PS512", "RSA-OAEP"}
)

func resourceKeycloakRealmKeystoreRsa() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakRealmKeystoreRsaCreate,
		ReadContext:   resourceKeycloakRealmKeystoreRsaRead,
		UpdateContext: resourceKeycloakRealmKeystoreRsaUpdate,
		DeleteContext: resourceKeycloakRealmKeystoreRsaDelete,
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
				ValidateFunc: validation.StringInSlice(keycloakRealmKeystoreRsaAlgorithm, false),
				Default:      "RS256",
				Description:  "Intended algorithm for the key",
			},
			"private_key": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Private RSA Key encoded in PEM format",
			},
			"certificate": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "X509 Certificate encoded in PEM format",
			},
			"provider_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "rsa",
				Description: "RSA key provider id",
				ForceNew:    true,
			},
		},
	}
}

func getRealmKeystoreRsaFromData(data *schema.ResourceData) *keycloak.RealmKeystoreRsa {
	mapper := &keycloak.RealmKeystoreRsa{
		Id:      data.Id(),
		Name:    data.Get("name").(string),
		RealmId: data.Get("realm_id").(string),

		Active:      data.Get("active").(bool),
		Enabled:     data.Get("enabled").(bool),
		Priority:    data.Get("priority").(int),
		Algorithm:   data.Get("algorithm").(string),
		PrivateKey:  data.Get("private_key").(string),
		Certificate: data.Get("certificate").(string),
		ProviderId:  data.Get("provider_id").(string),
	}

	return mapper
}

func setRealmKeystoreRsaData(data *schema.ResourceData, realmKey *keycloak.RealmKeystoreRsa) {
	data.SetId(realmKey.Id)

	data.Set("name", realmKey.Name)
	data.Set("realm_id", realmKey.RealmId)

	data.Set("active", realmKey.Active)
	data.Set("enabled", realmKey.Enabled)
	data.Set("priority", realmKey.Priority)
	data.Set("algorithm", realmKey.Algorithm)
	data.Set("provider_id", realmKey.ProviderId)
	if realmKey.PrivateKey != "**********" {
		data.Set("private_key", realmKey.PrivateKey)
		data.Set("certificate", realmKey.Certificate)
	}
}

func resourceKeycloakRealmKeystoreRsaCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmKey := getRealmKeystoreRsaFromData(data)

	err := keycloakClient.NewRealmKeystoreRsa(ctx, realmKey)
	if err != nil {
		return diag.FromErr(err)
	}

	setRealmKeystoreRsaData(data, realmKey)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceKeycloakRealmKeystoreRsaRead(ctx, data, meta)
}

func resourceKeycloakRealmKeystoreRsaRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	realmKey, err := keycloakClient.GetRealmKeystoreRsa(ctx, realmId, id)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	setRealmKeystoreRsaData(data, realmKey)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakRealmKeystoreRsaUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmKey := getRealmKeystoreRsaFromData(data)

	err := keycloakClient.UpdateRealmKeystoreRsa(ctx, realmKey)
	if err != nil {
		return diag.FromErr(err)
	}

	setRealmKeystoreRsaData(data, realmKey)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakRealmKeystoreRsaDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	return diag.FromErr(keycloakClient.DeleteRealmKeystoreRsa(ctx, realmId, id))
}
