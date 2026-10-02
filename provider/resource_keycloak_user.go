package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"dario.cat/mergo"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

const MULTIVALUE_ATTRIBUTE_SEPARATOR = "##"

func resourceKeycloakUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakUserCreate,
		ReadContext:   resourceKeycloakUserRead,
		DeleteContext: resourceKeycloakUserDelete,
		UpdateContext: resourceKeycloakUserUpdate,
		// This resource can be imported using {{realm}}/({{user_id}}|{{user_name}}). The User's ID is displayed in the GUI when editing
		Importer: &schema.ResourceImporter{
			StateContext: resourceKeycloakUserImport,
		},
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"username": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: func(i interface{}, k string) ([]string, []error) {
					username := i.(string)

					if strings.ToLower(username) != username {
						return nil, []error{fmt.Errorf("expected username %s to be all lowercase", username)}
					}

					return nil, nil
				},
			},
			"email": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"email_verified": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"first_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"last_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"attributes": {
				Type:     schema.TypeMap,
				Optional: true,
				// ignore ordering of multi-valued attributes
				DiffSuppressFunc: suppressDiffForMultivalueAttributeOrder(),
			},
			"required_actions": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"federated_identity": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"identity_provider": {
							Type:     schema.TypeString,
							Required: true,
						},
						"user_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"user_name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"initial_password": {
				Type:             schema.TypeList,
				Optional:         true,
				DiffSuppressFunc: onlyDiffOnCreate,
				MaxItems:         1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"value": {
							Type:          schema.TypeString,
							Optional:      true,
							Sensitive:     true,
							ConflictsWith: []string{"initial_password.0.value_wo", "initial_password.0.value_wo_version"},
							ExactlyOneOf:  []string{"initial_password.0.value", "initial_password.0.value_wo"},
						},
						"value_wo": {
							Type:          schema.TypeString,
							Optional:      true,
							Sensitive:     true,
							WriteOnly:     true,
							ConflictsWith: []string{"initial_password.0.value"},
							RequiredWith:  []string{"initial_password.0.value_wo_version"},
							ExactlyOneOf:  []string{"initial_password.0.value", "initial_password.0.value_wo"},
							ValidateFunc:  validation.StringIsNotEmpty,
							Description:   "The initial password as write-only argument",
						},
						"value_wo_version": {
							Type:          schema.TypeString,
							Optional:      true,
							ConflictsWith: []string{"initial_password.0.value"},
							RequiredWith:  []string{"initial_password.0.value_wo"},
							ValidateFunc:  validation.StringIsNotEmpty,
							Description:   "Version of the initial password write-only argument",
						},
						"temporary": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"import": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
		},
	}
}

// onlyDiffOnCreate suppresses changes to the initial_password block once the user exists, since the
// password is only applied during creation. This does not apply when the write-only argument
// `value_wo` is used: there, `value_wo_version` acts as the trigger to reset the password of an
// existing user, so its changes - and the changes of its sibling arguments - have to be visible.
func onlyDiffOnCreate(_, _, _ string, d *schema.ResourceData) bool {
	if initialPasswordUsesWriteOnly(d) {
		return false
	}

	return d.Id() != ""
}

// initialPasswordUsesWriteOnly returns true when either the state or the configuration carries a
// non-empty `initial_password.value_wo_version`, which is only valid together with
// `initial_password.value_wo`. Looking at both sides keeps changes visible while a user is migrated
// to or away from the write-only argument.
func initialPasswordUsesWriteOnly(d *schema.ResourceData) bool {
	stateVersion, configVersion := d.GetChange("initial_password.0.value_wo_version")

	oldVersion, _ := stateVersion.(string)
	newVersion, _ := configVersion.(string)

	return oldVersion != "" || newVersion != ""
}

func initialPasswordPath(attribute string) cty.Path {
	return cty.GetAttrPath("initial_password").IndexInt(0).GetAttr(attribute)
}

type userInitialPassword struct {
	value     string
	temporary bool
}

// getInitialPasswordFromData returns the initial password to send to Keycloak, or nil when the
// configuration does not contain an initial_password block. The write-only argument `value_wo` is
// never present in state, so it has to be read from the raw configuration.
func getInitialPasswordFromData(data *schema.ResourceData) (*userInitialPassword, error) {
	v, ok := data.GetOk("initial_password")
	if !ok {
		return nil, nil
	}

	passwordBlock := v.([]interface{})[0].(map[string]interface{})
	initialPassword := &userInitialPassword{
		value:     passwordBlock["value"].(string),
		temporary: passwordBlock["temporary"].(bool),
	}

	// an empty version means the legacy `value` argument is in use, since the schema rejects an
	// explicitly empty `value_wo_version`
	if passwordBlock["value_wo_version"].(string) == "" {
		return initialPassword, nil
	}

	valueWriteOnly, valueWriteOnlyDiags := data.GetRawConfigAt(initialPasswordPath("value_wo"))
	if valueWriteOnlyDiags.HasError() {
		return nil, errors.New("error reading 'initial_password.value_wo' argument")
	}
	if !valueWriteOnly.IsKnown() || valueWriteOnly.IsNull() || !valueWriteOnly.Type().Equals(cty.String) {
		return nil, errors.New("'initial_password.value_wo' must be a known, non-null string")
	}

	initialPassword.value = valueWriteOnly.AsString()

	return initialPassword, nil
}

func mapFromDataToUser(data *schema.ResourceData) *keycloak.User {
	attributes := map[string][]string{}
	var requiredActions []string

	if v, ok := data.GetOk("required_actions"); ok {
		for _, requiredAction := range v.(*schema.Set).List() {
			requiredActions = append(requiredActions, requiredAction.(string))
		}
	}
	if v, ok := data.GetOk("attributes"); ok {
		for key, value := range v.(map[string]interface{}) {
			attributes[key] = strings.Split(value.(string), MULTIVALUE_ATTRIBUTE_SEPARATOR)
		}
	}

	federatedIdentities := &keycloak.FederatedIdentities{}

	if v, ok := data.GetOk("federated_identity"); ok {
		federatedIdentities = getUserFederatedIdentitiesFromData(v.(*schema.Set).List())
	}

	return &keycloak.User{
		Id:                  data.Id(),
		RealmId:             data.Get("realm_id").(string),
		Username:            data.Get("username").(string),
		Email:               data.Get("email").(string),
		EmailVerified:       data.Get("email_verified").(bool),
		FirstName:           data.Get("first_name").(string),
		LastName:            data.Get("last_name").(string),
		Enabled:             data.Get("enabled").(bool),
		Attributes:          attributes,
		FederatedIdentities: *federatedIdentities,
		RequiredActions:     requiredActions,
	}
}

func getUserFederatedIdentitiesFromData(data []interface{}) *keycloak.FederatedIdentities {
	var federatedIdentities keycloak.FederatedIdentities
	for _, d := range data {
		federatedIdentitiesData := d.(map[string]interface{})
		federatedIdentity := &keycloak.FederatedIdentity{
			IdentityProvider: federatedIdentitiesData["identity_provider"].(string),
			UserId:           federatedIdentitiesData["user_id"].(string),
			UserName:         federatedIdentitiesData["user_name"].(string),
		}
		federatedIdentities = append(federatedIdentities, federatedIdentity)
	}
	return &federatedIdentities
}

func mapFromUserToData(data *schema.ResourceData, user *keycloak.User) {
	var federatedIdentities []interface{}
	for _, federatedIdentity := range user.FederatedIdentities {
		identity := map[string]interface{}{
			"identity_provider": federatedIdentity.IdentityProvider,
			"user_id":           federatedIdentity.UserId,
			"user_name":         federatedIdentity.UserName,
		}
		federatedIdentities = append(federatedIdentities, identity)
	}
	attributes := map[string]string{}
	for k, v := range user.Attributes {
		attributes[k] = strings.Join(v, MULTIVALUE_ATTRIBUTE_SEPARATOR)
	}
	data.SetId(user.Id)
	data.Set("realm_id", user.RealmId)
	data.Set("username", user.Username)
	data.Set("email", user.Email)
	data.Set("email_verified", user.EmailVerified)
	data.Set("first_name", user.FirstName)
	data.Set("last_name", user.LastName)
	data.Set("enabled", user.Enabled)
	data.Set("attributes", attributes)
	data.Set("federated_identity", federatedIdentities)
	data.Set("required_actions", user.RequiredActions)
}

func resourceKeycloakUserCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	user := mapFromDataToUser(data)

	if !data.Get("import").(bool) {
		err := keycloakClient.NewUser(ctx, user)
		if err != nil {
			return diag.FromErr(err)
		}

		initialPassword, err := getInitialPasswordFromData(data)
		if err != nil {
			return diag.FromErr(err)
		}
		if initialPassword != nil {
			err := keycloakClient.ResetUserPassword(ctx, user.RealmId, user.Id, initialPassword.value, initialPassword.temporary)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	} else {
		username := data.Get("username").(string)
		existingUser, err := keycloakClient.GetUserByUsername(ctx, data.Get("realm_id").(string), username)
		if err != nil {
			return diag.FromErr(err)
		}
		if existingUser == nil {
			return diag.FromErr(fmt.Errorf("no user found for username %s", username))
		}

		if err = mergo.Merge(user, existingUser); err != nil {
			return diag.FromErr(err)
		}
		err = keycloakClient.UpdateUser(ctx, user)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	mapFromUserToData(data, user)

	return resourceKeycloakUserRead(ctx, data, meta)
}

func resourceKeycloakUserRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	user, err := keycloakClient.GetUser(ctx, realmId, id)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	mapFromUserToData(data, user)

	if _, ok := data.GetOk("import"); !ok {
		data.Set("import", false)
	}

	return nil
}

func resourceKeycloakUserUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	user := mapFromDataToUser(data)

	err := keycloakClient.UpdateUser(ctx, user)
	if err != nil {
		return diag.FromErr(err)
	}

	// a change of the write-only version is the only way to reset the password of an existing user
	if data.Get("initial_password.0.value_wo_version").(string) != "" && data.HasChange("initial_password.0.value_wo_version") {
		initialPassword, err := getInitialPasswordFromData(data)
		if err != nil {
			return diag.FromErr(err)
		}
		if initialPassword != nil {
			err := keycloakClient.ResetUserPassword(ctx, user.RealmId, user.Id, initialPassword.value, initialPassword.temporary)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	mapFromUserToData(data, user)

	return nil
}

func resourceKeycloakUserDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if data.Get("import").(bool) {
		return nil
	}

	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	return diag.FromErr(keycloakClient.DeleteUser(ctx, realmId, id))
}

func resourceKeycloakUserImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("Invalid import. Supported import formats: {{realmId}}/({{userId}}|{{userName}})")
	}

	user, err := keycloakClient.GetUser(ctx, parts[0], parts[1])
	if err != nil {
		user, err = keycloakClient.GetUserByUsername(ctx, parts[0], parts[1])
		if err != nil {
			return nil, err
		}
	}

	d.Set("realm_id", parts[0])
	d.Set("import", false)
	d.SetId(user.Id)

	diagnostics := resourceKeycloakUserRead(ctx, d, meta)
	if diagnostics.HasError() {
		return nil, errors.New(diagnostics[0].Summary)
	}

	return []*schema.ResourceData{d}, nil
}
