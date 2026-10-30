package provider

import (
	"context"
	"fmt"
	"strings"

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
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringIsNotEmpty,
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

	credential := &keycloak.UserOtpCredential{
		RealmId:   data.Get("realm_id").(string),
		UserId:    data.Get("user_id").(string),
		Secret:    data.Get("secret").(string),
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

	// Keycloak never returns the secret, so only the existence of the credential and its label can be checked.
	credential, err := keycloakClient.GetUserOtpCredential(ctx, realmId, userId, userOtpCredentialIdFromResourceId(data.Id()))
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	data.Set("user_label", credential.UserLabel)

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

func userOtpCredentialId(realmId, userId, credentialId string) string {
	return fmt.Sprintf("%s/%s/%s", realmId, userId, credentialId)
}

func userOtpCredentialIdFromResourceId(id string) string {
	parts := strings.Split(id, "/")

	return parts[len(parts)-1]
}
