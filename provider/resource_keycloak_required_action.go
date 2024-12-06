package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"strings"
)

func resourceKeycloakRequiredAction() *schema.Resource {

	return &schema.Resource{
		CreateContext: resourceKeycloakRequiredActionsCreate,
		ReadContext:   resourceKeycloakRequiredActionsRead,
		DeleteContext: resourceKeycloakRequiredActionsDelete,
		UpdateContext: resourceKeycloakRequiredActionsUpdate,
		Importer: &schema.ResourceImporter{
			// This resource can be imported using {{realm}}/{{alias}}. The required action aliases are displayed in the server info or GET realms/{{realm}}/authentication/required-actions
			StateContext: resourceKeycloakRequiredActionsImport,
		},
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"alias": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"default_action": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"priority": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func getRequiredActionFromData(data *schema.ResourceData) (*keycloak.RequiredAction, error) {
	action := &keycloak.RequiredAction{
		Id:            fmt.Sprintf("%s/%s", data.Get("realm_id").(string), data.Get("alias").(string)),
		RealmId:       data.Get("realm_id").(string),
		Alias:         data.Get("alias").(string),
		Name:          data.Get("name").(string),
		Enabled:       data.Get("enabled").(bool),
		DefaultAction: data.Get("default_action").(bool),
		Priority:      data.Get("priority").(int),
		Config:        make(map[string][]string),
	}

	return action, nil
}

func setRequiredActionData(data *schema.ResourceData, action *keycloak.RequiredAction) {
	data.SetId(fmt.Sprintf("%s/%s", action.RealmId, action.Alias))
	data.Set("realm_id", action.RealmId)
	data.Set("alias", action.Alias)
	data.Set("name", action.Name)
	data.Set("enabled", action.Enabled)
	data.Set("default_action", action.DefaultAction)
	data.Set("priority", action.Priority)
}

func resourceKeycloakRequiredActionsCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	action, err := getRequiredActionFromData(data)
	if err != nil {
		return diag.FromErr(err)
	}

	unregisteredRequiredActions, err := keycloakClient.GetUnregisteredRequiredActions(ctx, action.RealmId)
	if err != nil {
		return diag.FromErr(err)
	}
	for _, unregisteredRequiredAction := range unregisteredRequiredActions {
		if unregisteredRequiredAction.ProviderId == action.Alias {
			if err := keycloakClient.RegisterRequiredAction(ctx, unregisteredRequiredAction); err != nil {
				return diag.FromErr(err)
			}
			break
		}
	}

	err = keycloakClient.CreateRequiredAction(ctx, action)
	if err != nil {
		return diag.FromErr(err)
	}

	setRequiredActionData(data, action)

	return resourceKeycloakRequiredActionsRead(ctx, data, meta)
}

func resourceKeycloakRequiredActionsRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	action, err := keycloakClient.GetRequiredAction(ctx, data.Get("realm_id").(string), data.Get("alias").(string))
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	setRequiredActionData(data, action)

	return nil
}

func resourceKeycloakRequiredActionsUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	action, err := getRequiredActionFromData(data)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.UpdateRequiredAction(ctx, action)
	if err != nil {
		return diag.FromErr(err)
	}

	setRequiredActionData(data, action)

	return nil
}

func resourceKeycloakRequiredActionsDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmName := data.Get("realm_id").(string)
	alias := data.Get("alias").(string)

	return diag.FromErr(keycloakClient.DeleteRequiredAction(ctx, realmName, alias))
}

func resourceKeycloakRequiredActionsImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid import. Supported import formats: {{realmId}}/{{alias}}")
	}

	_, err := keycloakClient.GetRequiredAction(ctx, parts[0], parts[1])
	if err != nil {
		return nil, err
	}

	d.Set("realm_id", parts[0])
	d.Set("alias", parts[1])
	d.SetId(fmt.Sprintf("%s/%s", parts[0], parts[1]))

	diagnostics := resourceKeycloakRequiredActionsRead(ctx, d, meta)
	if diagnostics.HasError() {
		return nil, errors.New(diagnostics[0].Summary)
	}

	return []*schema.ResourceData{d}, nil
}
