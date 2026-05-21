package provider

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakUserPassword() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakUserPasswordCreate,
		ReadContext:   resourceKeycloakUserPasswordRead,
		UpdateContext: resourceKeycloakUserPasswordUpdate,
		DeleteContext: resourceKeycloakUserPasswordDelete,
		CustomizeDiff: resourceKeycloakUserPasswordDiff,
		// This resource can be imported using {{realm}}/{{userId}}.
		Importer: &schema.ResourceImporter{
			StateContext: resourceKeycloakUserPasswordImport,
		},
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
			"value": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
				WriteOnly: true,
			},
			"temporary": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			// credential_id is the UUID of the Keycloak password credential.
			// It changes every time reset-password is called, so it is used
			// to track remote state.
			"credential_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			// value_hash stores a SHA-256 hash of the currently applied
			// password so that the provider can detect changes to the
			// WriteOnly value argument across Terraform runs.
			"value_hash": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func hashValue(value string) string {
	sum := sha256.Sum256([]byte(value))
	return fmt.Sprintf("%x", sum)
}

// resourceKeycloakUserPasswordDiff detects changes to the WriteOnly
// value argument by comparing its SHA-256 hash with the value_hash
// stored in state. When they differ, value_hash is set to the new hash,
// producing a plan diff that triggers an update.
func resourceKeycloakUserPasswordDiff(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
	raw := d.GetRawConfig()
	val := raw.GetAttr("value")
	if val.IsNull() {
		return nil
	}

	newHash := hashValue(val.AsString())
	oldHash := d.Get("value_hash").(string)

	if oldHash != newHash {
		d.SetNew("value_hash", newHash)
	}

	return nil
}

// getPasswordValue returns the password value from the raw config
// since value is WriteOnly and cannot be read from state.
func getPasswordValue(data *schema.ResourceData) (string, diag.Diagnostics) {
	raw, d := data.GetRawConfigAt(cty.GetAttrPath("value"))
	if d.HasError() {
		return "", d
	}
	if raw.IsNull() {
		return "", diag.Errorf("value must be set")
	}
	return raw.AsString(), nil
}

func resourceKeycloakUserPasswordId(realmId, userId string) string {
	return fmt.Sprintf("%s/%s", realmId, userId)
}

// setCredentialState fetches the current password credential after a
// reset-password call and writes credential_id and value_hash into state.
func setCredentialState(ctx context.Context, data *schema.ResourceData, keycloakClient *keycloak.KeycloakClient, value string) error {
	realmId := data.Get("realm_id").(string)
	userId := data.Get("user_id").(string)

	cred, err := keycloakClient.GetUserPasswordCredential(ctx, realmId, userId)
	if err != nil {
		return err
	}
	if cred == nil {
		return fmt.Errorf("password credential not found for user %s in realm %s after reset", userId, realmId)
	}

	data.Set("credential_id", cred.Id)
	data.Set("value_hash", hashValue(value))
	return nil
}

func resourceKeycloakUserPasswordCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	userId := data.Get("user_id").(string)
	temporary := data.Get("temporary").(bool)

	value, valueDiags := getPasswordValue(data)
	if valueDiags.HasError() {
		return valueDiags
	}

	if err := keycloakClient.ResetUserPassword(ctx, realmId, userId, value, temporary); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(resourceKeycloakUserPasswordId(realmId, userId))

	if err := setCredentialState(ctx, data, keycloakClient, value); err != nil {
		return diag.FromErr(err)
	}

	return resourceKeycloakUserPasswordRead(ctx, data, meta)
}

// resourceKeycloakUserPasswordRead checks whether the password credential
// still exists and whether it has been reset out-of-band.
//
// Because reset-password always produces a brand-new credential ID, a change
// in credential_id is the reliable signal that someone reset the password
// outside Terraform.  When that happens we clear the resource ID so Terraform
// will plan a new apply of the configured value.
func resourceKeycloakUserPasswordRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	userId := data.Get("user_id").(string)

	cred, err := keycloakClient.GetUserPasswordCredential(ctx, realmId, userId)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	// Credential was deleted outside Terraform — remove from state.
	if cred == nil {
		data.SetId("")
		return nil
	}

	storedCredentialId := data.Get("credential_id").(string)

	// credential_id changed → password was reset out-of-band.
	// Clear state so Terraform plans a re-apply of the configured value.
	if storedCredentialId != "" && cred.Id != storedCredentialId {
		data.SetId("")
		return nil
	}

	data.Set("credential_id", cred.Id)
	data.Set("temporary", cred.Temporary)

	return nil
}

func resourceKeycloakUserPasswordUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	userId := data.Get("user_id").(string)
	temporary := data.Get("temporary").(bool)

	value, valueDiags := getPasswordValue(data)
	if valueDiags.HasError() {
		return valueDiags
	}

	if err := keycloakClient.ResetUserPassword(ctx, realmId, userId, value, temporary); err != nil {
		return diag.FromErr(err)
	}

	if err := setCredentialState(ctx, data, keycloakClient, value); err != nil {
		return diag.FromErr(err)
	}

	return resourceKeycloakUserPasswordRead(ctx, data, meta)
}

// resourceKeycloakUserPasswordDelete removes the password credential from
// Keycloak using the stored credential_id.
func resourceKeycloakUserPasswordDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	userId := data.Get("user_id").(string)
	credentialId := data.Get("credential_id").(string)

	if credentialId == "" {
		return nil
	}

	return diag.FromErr(keycloakClient.DeleteUserCredential(ctx, realmId, userId, credentialId))
}

func resourceKeycloakUserPasswordImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid import. Supported import format: {{realmId}}/{{userId}}")
	}

	realmId := parts[0]
	userId := parts[1]

	if _, err := keycloakClient.GetUser(ctx, realmId, userId); err != nil {
		return nil, err
	}

	cred, err := keycloakClient.GetUserPasswordCredential(ctx, realmId, userId)
	if err != nil {
		return nil, err
	}

	d.Set("realm_id", realmId)
	d.Set("user_id", userId)
	d.SetId(resourceKeycloakUserPasswordId(realmId, userId))

	if cred != nil {
		d.Set("credential_id", cred.Id)
	}
	// value and value_hash cannot be recovered from the API;
	// callers must supply value after import or accept a plan
	// diff on next apply.

	return []*schema.ResourceData{d}, nil
}
