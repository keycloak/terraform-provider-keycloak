package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

var (
	keycloakRealmKeystoreRsaAlgorithm    = []string{"RS256", "RS384", "RS512", "PS256", "PS384", "PS512"}
	keycloakRealmKeystoreRsaEncAlgorithm = []string{"RSA1_5", "RSA-OAEP", "RSA-OAEP-256"}
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
				ValidateFunc: validation.StringInSlice(append(keycloakRealmKeystoreRsaAlgorithm, keycloakRealmKeystoreRsaEncAlgorithm...), false),
				Default:      "RS256",
				Description:  "Intended algorithm for the key",
			},
			"private_key": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "Private RSA Key encoded in PEM format",
				ConflictsWith: []string{"private_key_wo", "private_key_wo_version"},
				ExactlyOneOf:  []string{"private_key", "private_key_wo"},
			},
			"private_key_wo": {
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				WriteOnly:     true,
				Description:   "Private RSA Key encoded in PEM format as write-only argument",
				ConflictsWith: []string{"private_key"},
				RequiredWith:  []string{"private_key_wo_version"},
				ExactlyOneOf:  []string{"private_key", "private_key_wo"},
			},
			"private_key_wo_version": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "Version of the private_key write-only argument",
				ConflictsWith: []string{"private_key"},
				RequiredWith:  []string{"private_key_wo"},
			},
			"certificate": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "X509 Certificate encoded in PEM format",
				ConflictsWith: []string{"certificate_wo", "certificate_wo_version"},
				ExactlyOneOf:  []string{"certificate", "certificate_wo"},
			},
			"certificate_wo": {
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				WriteOnly:     true,
				Description:   "X509 Certificate encoded in PEM format as write-only argument",
				ConflictsWith: []string{"certificate"},
				RequiredWith:  []string{"certificate_wo_version"},
				ExactlyOneOf:  []string{"certificate", "certificate_wo"},
			},
			"certificate_wo_version": {
				Type:          schema.TypeString,
				Optional:      true,
				ValidateFunc:  validation.StringIsNotEmpty,
				Description:   "Version of the certificate write-only argument",
				ConflictsWith: []string{"certificate"},
				RequiredWith:  []string{"certificate_wo"},
			},
			"provider_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "rsa",
				Description: "RSA key provider id",
				ForceNew:    true,
			},
			"extra_config": {
				Type:     schema.TypeMap,
				Optional: true,
			},
		},
	}
}

// keycloakMaskedSecret is the value Keycloak returns for secret config values, and which it
// interprets on update as "keep the stored value".
const keycloakMaskedSecret = "**********"

func getRealmKeystoreRsaWriteOnlyValue(data *schema.ResourceData, attribute string) (string, error) {
	value, diags := data.GetRawConfigAt(cty.GetAttrPath(attribute))
	if diags.HasError() {
		return "", fmt.Errorf("error reading '%s' argument", attribute)
	}
	if value.IsNull() || !value.IsKnown() {
		return "", errors.New("'" + attribute + "' argument is null or unknown")
	}

	return value.AsString(), nil
}

func getRealmKeystoreRsaFromData(data *schema.ResourceData) (*keycloak.RealmKeystoreRsa, error) {
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

	if data.Get("private_key_wo_version").(string) != "" {
		if data.HasChange("private_key_wo_version") {
			privateKey, err := getRealmKeystoreRsaWriteOnlyValue(data, "private_key_wo")
			if err != nil {
				return nil, err
			}
			mapper.PrivateKey = privateKey
		} else {
			// Keycloak keeps the stored private key when it receives the masked value
			mapper.PrivateKey = keycloakMaskedSecret
		}
	}

	// the certificate is not a secret and Keycloak requires it on every update, so it is always sent
	if data.Get("certificate_wo_version").(string) != "" {
		certificate, err := getRealmKeystoreRsaWriteOnlyValue(data, "certificate_wo")
		if err != nil {
			return nil, err
		}
		mapper.Certificate = certificate
	}

	mapper.ExtraConfig = getExtraConfigFromData(data)

	return mapper, nil
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
	if realmKey.PrivateKey != keycloakMaskedSecret {
		// never store values in state when they are managed through write-only arguments
		if data.Get("private_key_wo_version").(string) == "" {
			data.Set("private_key", realmKey.PrivateKey)
		}
		if data.Get("certificate_wo_version").(string) == "" {
			data.Set("certificate", realmKey.Certificate)
		}
	}
	setExtraConfigData(data, realmKey.ExtraConfig)
}

func resourceKeycloakRealmKeystoreRsaCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmKey, err := getRealmKeystoreRsaFromData(data)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.NewRealmKeystoreRsa(ctx, realmKey)
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

	realmKey, err := getRealmKeystoreRsaFromData(data)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.UpdateRealmKeystoreRsa(ctx, realmKey)
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
