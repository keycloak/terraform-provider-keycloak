package provider

import (
	"context"
	"encoding/base32"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakUserOtpCredential() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakUserOtpCredentialCreate,
		ReadContext:   resourceKeycloakUserOtpCredentialRead,
		DeleteContext: resourceKeycloakUserOtpCredentialDelete,
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"user_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"secret": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Sensitive:     true,
				ValidateFunc:  validateBase32Secret,
				Description:   "The Base32-encoded TOTP secret. It is stored in the state, use secret_wo to keep it out of the state.",
				ConflictsWith: []string{"secret_wo", "secret_wo_version"},
				ExactlyOneOf:  []string{"secret", "secret_wo"},
			},
			"secret_wo": {
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				WriteOnly:     true,
				ValidateFunc:  validateBase32Secret,
				Description:   "The Base32-encoded TOTP secret as write-only argument, which is not stored in the state.",
				ConflictsWith: []string{"secret"},
				RequiredWith:  []string{"secret_wo_version"},
				ExactlyOneOf:  []string{"secret", "secret_wo"},
			},
			"secret_wo_version": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ValidateFunc:  validation.StringIsNotEmpty,
				Description:   "Version of the secret_wo write-only argument. Changing it replaces the credential with one made from the current secret_wo.",
				ConflictsWith: []string{"secret"},
				RequiredWith:  []string{"secret_wo"},
			},
			"algorithm": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "HmacSHA1",
				ValidateFunc: validation.StringInSlice([]string{"HmacSHA1", "HmacSHA256", "HmacSHA512"}, false),
			},
			"digits": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Default:      6,
				ValidateFunc: validation.IntInSlice([]int{6, 8}),
			},
			"period": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Default:      30,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"user_label": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceKeycloakUserOtpCredentialCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	secret := data.Get("secret").(string)
	if data.Get("secret_wo_version").(string) != "" {
		var err error
		if secret, err = getUserOtpCredentialWriteOnlySecret(data); err != nil {
			return diag.FromErr(err)
		}
	}

	credential := &keycloak.UserOtpCredential{
		RealmId:   data.Get("realm_id").(string),
		UserId:    data.Get("user_id").(string),
		Secret:    secret,
		Algorithm: data.Get("algorithm").(string),
		Digits:    data.Get("digits").(int),
		Period:    data.Get("period").(int),
		UserLabel: data.Get("user_label").(string),
	}

	if err := keycloakClient.CreateUserOtpCredential(ctx, credential); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(userOtpCredentialId(credential.RealmId, credential.UserId, credential.Id))

	return resourceKeycloakUserOtpCredentialRead(ctx, data, meta)
}

func resourceKeycloakUserOtpCredentialRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	userId := data.Get("user_id").(string)

	// Keycloak never returns the secret, so only the existence of the credential can be checked. The label is
	// not read back: users can rename their credentials, which must not replace the credential.
	_, err := keycloakClient.GetUserOtpCredential(ctx, realmId, userId, userOtpCredentialIdFromResourceId(data.Id()))
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	return nil
}

func resourceKeycloakUserOtpCredentialDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	userId := data.Get("user_id").(string)

	err := keycloakClient.DeleteUserOtpCredential(ctx, realmId, userId, userOtpCredentialIdFromResourceId(data.Id()))
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	return nil
}

func getUserOtpCredentialWriteOnlySecret(data *schema.ResourceData) (string, error) {
	value, diags := data.GetRawConfigAt(cty.GetAttrPath("secret_wo"))
	if diags.HasError() {
		return "", errors.New("error reading 'secret_wo' argument")
	}
	if value.IsNull() || !value.IsKnown() {
		return "", errors.New("'secret_wo' argument is null or unknown")
	}

	return value.AsString(), nil
}

func userOtpCredentialId(realmId, userId, credentialId string) string {
	return fmt.Sprintf("%s/%s/%s", realmId, userId, credentialId)
}

func userOtpCredentialIdFromResourceId(id string) string {
	parts := strings.Split(id, "/")

	return parts[len(parts)-1]
}

// validateBase32Secret checks that the secret decodes as Base32. The secret is left out of the error, as it is a credential.
func validateBase32Secret(i interface{}, k string) ([]string, []error) {
	secret, ok := i.(string)
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %s to be string", k)}
	}

	unpadded := strings.ToUpper(strings.TrimRight(secret, "="))
	padding := len(secret) - len(strings.TrimRight(secret, "="))
	if padding > 0 && padding != (8-len(unpadded)%8)%8 {
		return nil, []error{fmt.Errorf("%s is not valid Base32: it has the wrong amount of = padding", k)}
	}

	if decoded, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(unpadded); err != nil || len(decoded) == 0 {
		return nil, []error{fmt.Errorf("%s must be a non-empty, valid Base32 string (A-Z, 2-7, optionally padded with =)", k)}
	}

	return nil, nil
}
