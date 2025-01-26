package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/keycloak/terraform-provider-keycloak/keycloak/types"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakOpenidClientScope() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakOpenidClientScopeCreate,
		ReadContext:   resourceKeycloakOpenidClientScopeRead,
		DeleteContext: resourceKeycloakOpenidClientScopeDelete,
		UpdateContext: resourceKeycloakOpenidClientScopeUpdate,
		// This resource can be imported using {{realm}}/{{client_scope_id}}. The Client Scope ID is displayed in the GUI
		Importer: &schema.ResourceImporter{
			StateContext: resourceKeycloakOpenidClientScopeImport,
		},
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"consent_screen_text": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"include_in_token_scope": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"gui_order": {
				Type:     schema.TypeInt,
				Optional: true,
			},
		},
	}
}

func getOpenidClientScopeFromData(data *schema.ResourceData) *keycloak.OpenidClientScope {
	clientScope := &keycloak.OpenidClientScope{
		Id:          data.Id(),
		RealmId:     data.Get("realm_id").(string),
		Name:        data.Get("name").(string),
		Description: data.Get("description").(string),
	}

	if consentScreenText, ok := data.GetOk("consent_screen_text"); ok {
		clientScope.Attributes.ConsentScreenText = consentScreenText.(string)
		clientScope.Attributes.DisplayOnConsentScreen = true
	} else {
		clientScope.Attributes.DisplayOnConsentScreen = false
	}

	clientScope.Attributes.IncludeInTokenScope = types.KeycloakBoolQuoted(data.Get("include_in_token_scope").(bool))

	// Treat 0 as an empty string for the purpose of omitting the attribute to reset the order
	if guiOrder := data.Get("gui_order").(int); guiOrder != 0 {
		clientScope.Attributes.GuiOrder = strconv.Itoa(guiOrder)
	}

	return clientScope
}

func setOpenidClientScopeData(data *schema.ResourceData, clientScope *keycloak.OpenidClientScope) {
	data.SetId(clientScope.Id)

	data.Set("realm_id", clientScope.RealmId)
	data.Set("name", clientScope.Name)
	data.Set("description", clientScope.Description)

	if clientScope.Attributes.DisplayOnConsentScreen {
		data.Set("consent_screen_text", clientScope.Attributes.ConsentScreenText)
	}

	data.Set("include_in_token_scope", clientScope.Attributes.IncludeInTokenScope)
	if guiOrder, err := strconv.Atoi(clientScope.Attributes.GuiOrder); err == nil {
		data.Set("gui_order", guiOrder)
	}
}

func resourceKeycloakOpenidClientScopeCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	clientScope := getOpenidClientScopeFromData(data)

	err := keycloakClient.NewOpenidClientScope(ctx, clientScope)
	if err != nil {
		return diag.FromErr(err)
	}

	setOpenidClientScopeData(data, clientScope)

	return resourceKeycloakOpenidClientScopeRead(ctx, data, meta)
}

func resourceKeycloakOpenidClientScopeRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	clientScope, err := keycloakClient.GetOpenidClientScope(ctx, realmId, id)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	setOpenidClientScopeData(data, clientScope)

	return nil
}

func resourceKeycloakOpenidClientScopeUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	clientScope := getOpenidClientScopeFromData(data)

	err := keycloakClient.UpdateOpenidClientScope(ctx, clientScope)
	if err != nil {
		return diag.FromErr(err)
	}

	setOpenidClientScopeData(data, clientScope)

	return nil
}

func resourceKeycloakOpenidClientScopeDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	return diag.FromErr(keycloakClient.DeleteOpenidClientScope(ctx, realmId, id))
}

func resourceKeycloakOpenidClientScopeImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("Invalid import. Supported import formats: {{realmId}}/{{openidClientScopeId}}")
	}

	d.Set("realm_id", parts[0])
	d.SetId(parts[1])

	return []*schema.ResourceData{d}, nil
}
