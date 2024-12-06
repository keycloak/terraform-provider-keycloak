package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

var (
	keycloakRealmKeystoreJavaKeystoreAlgorithm = []string{"RS256", "RS384", "RS512", "PS256", "PS384", "PS512"}
)

func resourceKeycloakRealmKeystoreJavaKeystore() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakRealmKeystoreJavaKeystoreCreate,
		ReadContext:   resourceKeycloakRealmKeystoreJavaKeystoreRead,
		UpdateContext: resourceKeycloakRealmKeystoreJavaKeystoreUpdate,
		DeleteContext: resourceKeycloakRealmKeystoreJavaKeystoreDelete,
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
				ValidateFunc: validation.StringInSlice(keycloakRealmKeystoreJavaKeystoreAlgorithm, false),
				Default:      "RS256",
				Description:  "Intended algorithm for the key",
			},
			"keystore": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Path to keys file",
			},
			"keystore_password": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Password for the keys",
			},
			"key_alias": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Alias for the private key",
			},
			"key_password": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Password for the private key",
			},
		},
	}
}

func getRealmKeystoreJavaKeystoreFromData(data *schema.ResourceData) (*keycloak.RealmKeystoreJavaKeystore, error) {
	keystore := &keycloak.RealmKeystoreJavaKeystore{
		Id:      data.Id(),
		Name:    data.Get("name").(string),
		RealmId: data.Get("realm_id").(string),

		Active:           data.Get("active").(bool),
		Enabled:          data.Get("enabled").(bool),
		Priority:         data.Get("priority").(int),
		Keystore:         data.Get("keystore").(string),
		KeystorePassword: data.Get("keystore_password").(string),
		KeyAlias:         data.Get("key_alias").(string),
		KeyPassword:      data.Get("key_password").(string),
	}

	return keystore, nil
}

func setRealmKeystoreJavaKeystoreData(data *schema.ResourceData, realmKey *keycloak.RealmKeystoreJavaKeystore) error {
	data.SetId(realmKey.Id)

	data.Set("name", realmKey.Name)
	data.Set("realm_id", realmKey.RealmId)

	data.Set("active", realmKey.Active)
	data.Set("enabled", realmKey.Enabled)
	data.Set("priority", realmKey.Priority)
	data.Set("keystore", realmKey.Keystore)
	data.Set("key_alias", realmKey.KeyAlias)
	if realmKey.KeystorePassword != "**********" {
		data.Set("keystore_password", realmKey.KeystorePassword)
	}
	if realmKey.KeyPassword != "**********" {
		data.Set("key_password", realmKey.KeyPassword)
	}
	return nil
}

func resourceKeycloakRealmKeystoreJavaKeystoreCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmKey, err := getRealmKeystoreJavaKeystoreFromData(data)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.NewRealmKeystoreJavaKeystore(ctx, realmKey)
	if err != nil {
		return diag.FromErr(err)
	}

	err = setRealmKeystoreJavaKeystoreData(data, realmKey)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceKeycloakRealmKeystoreJavaKeystoreRead(ctx, data, meta)
}

func resourceKeycloakRealmKeystoreJavaKeystoreRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	realmKey, err := keycloakClient.GetRealmKeystoreJavaKeystore(ctx, realmId, id)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	err = setRealmKeystoreJavaKeystoreData(data, realmKey)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakRealmKeystoreJavaKeystoreUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmKey, err := getRealmKeystoreJavaKeystoreFromData(data)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.UpdateRealmKeystoreJavaKeystore(ctx, realmKey)
	if err != nil {
		return diag.FromErr(err)
	}

	err = setRealmKeystoreJavaKeystoreData(data, realmKey)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakRealmKeystoreJavaKeystoreDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	return diag.FromErr(keycloakClient.DeleteRealmKeystoreJavaKeystore(ctx, realmId, id))
}
