package provider

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

const (
	keycloakClientScopeProtocolOpenidConnect = "openid-connect"
	keycloakClientScopeProtocolSaml          = "saml"

	clientScopeAttributeConsentScreenText               = "consent.screen.text"
	clientScopeAttributeDisplayOnConsentScreen          = "display.on.consent.screen"
	clientScopeAttributeGuiOrder                        = "gui.order"
	clientScopeAttributeIncludeInTokenScope             = "include.in.token.scope"
	clientScopeAttributeIncludeInOpenidProviderMetadata = "include.in.openid.provider.metadata"
)

var clientScopeProtocolBlocks = []string{"openid_connect", "saml"}

func resourceKeycloakClientScope() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakClientScopeCreate,
		ReadContext:   resourceKeycloakClientScopeRead,
		UpdateContext: resourceKeycloakClientScopeUpdate,
		DeleteContext: resourceKeycloakClientScopeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceKeycloakClientScopeImport,
		},
		CustomizeDiff: customizeDiffClientScopeDisplayOnConsentScreen,
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
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressDiffWhenNotInConfig("description"),
			},
			"consent_screen_text": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"display_on_consent_screen": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"gui_order": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"openid_connect": {
				Type:         schema.TypeList,
				Optional:     true,
				ForceNew:     true,
				MaxItems:     1,
				ExactlyOneOf: clientScopeProtocolBlocks,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"include_in_token_scope": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"include_in_openid_provider_metadata": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
					},
				},
			},
			"saml": {
				Type:         schema.TypeList,
				Optional:     true,
				ForceNew:     true,
				MaxItems:     1,
				ExactlyOneOf: clientScopeProtocolBlocks,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{},
				},
			},
			"extra_config": {
				Type:             schema.TypeMap,
				Optional:         true,
				Elem:             &schema.Schema{Type: schema.TypeString},
				ValidateDiagFunc: validateClientScopeExtraConfig,
			},
		},
	}
}

var validateClientScopeModeledExtraConfig = validateExtraConfig(reflect.ValueOf(&keycloak.OpenidClientScopeAttributes{}).Elem())

func validateClientScopeExtraConfig(value interface{}, path cty.Path) diag.Diagnostics {
	diags := validateClientScopeModeledExtraConfig(value, path)
	for key, value := range value.(map[string]interface{}) {
		attribute, ok := value.(string)
		if !ok || strings.TrimSpace(attribute) == "" {
			diags = append(diags, diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "Invalid extra_config value",
				Detail:        fmt.Sprintf("extra_config key %q must not be empty or whitespace-only; remove the key to clear it.", key),
				AttributePath: append(path, cty.IndexStep{Key: cty.StringVal(key)}),
			})
		}
	}
	return diags
}

func getKeycloakClientScopeProtocol(data *schema.ResourceData) string {
	if len(data.Get("openid_connect").([]interface{})) > 0 {
		return keycloakClientScopeProtocolOpenidConnect
	}
	return keycloakClientScopeProtocolSaml
}

func getKeycloakClientScopeFromData(data *schema.ResourceData) (*keycloak.ClientScope, error) {
	attributes := map[string]string{}
	// Keycloak merges attributes on update, so removed keys are sent empty to clear them. Older Keycloak
	// versions keep the key with an empty value, which Keycloak and Read treat as unset.
	for key, value := range getExtraConfigFromData(data) {
		attributes[key] = value.(string)
	}

	displayOnConsentScreen, err := getClientScopeDisplayOnConsentScreen(data)
	if err != nil {
		return nil, err
	}
	attributes[clientScopeAttributeConsentScreenText] = data.Get("consent_screen_text").(string)
	attributes[clientScopeAttributeDisplayOnConsentScreen] = strconv.FormatBool(displayOnConsentScreen)
	// 0 resets the order by clearing the attribute.
	attributes[clientScopeAttributeGuiOrder] = ""
	if guiOrder := data.Get("gui_order").(int); guiOrder != 0 {
		attributes[clientScopeAttributeGuiOrder] = strconv.Itoa(guiOrder)
	}

	protocol := getKeycloakClientScopeProtocol(data)
	if protocol == keycloakClientScopeProtocolOpenidConnect {
		attributes[clientScopeAttributeIncludeInTokenScope] = strconv.FormatBool(data.Get("openid_connect.0.include_in_token_scope").(bool))
		attributes[clientScopeAttributeIncludeInOpenidProviderMetadata] = strconv.FormatBool(data.Get("openid_connect.0.include_in_openid_provider_metadata").(bool))
	}

	return &keycloak.ClientScope{
		Id:          data.Id(),
		RealmId:     data.Get("realm_id").(string),
		Name:        data.Get("name").(string),
		Description: data.Get("description").(string),
		Protocol:    protocol,
		Attributes:  attributes,
	}, nil
}

// clientScopeBoolAttribute mirrors Keycloak, which treats a missing boolean scope attribute as true
// and otherwise uses Boolean.parseBoolean, so present-but-empty values are false.
func clientScopeBoolAttribute(attributes map[string]string, key string) bool {
	value, ok := attributes[key]
	if !ok {
		return true
	}
	return strings.EqualFold(value, "true")
}

func setKeycloakClientScopeData(data *schema.ResourceData, clientScope *keycloak.ClientScope) error {
	var openidConnect, saml []interface{}
	switch clientScope.Protocol {
	case keycloakClientScopeProtocolOpenidConnect:
		openidConnect = []interface{}{map[string]interface{}{
			"include_in_token_scope":              clientScopeBoolAttribute(clientScope.Attributes, clientScopeAttributeIncludeInTokenScope),
			"include_in_openid_provider_metadata": clientScopeBoolAttribute(clientScope.Attributes, clientScopeAttributeIncludeInOpenidProviderMetadata),
		}}
	case keycloakClientScopeProtocolSaml:
		saml = []interface{}{map[string]interface{}{}}
	default:
		return fmt.Errorf("client scope %s has unsupported protocol %q", clientScope.Id, clientScope.Protocol)
	}

	guiOrder, err := strconv.Atoi(clientScope.Attributes[clientScopeAttributeGuiOrder])
	if err != nil {
		guiOrder = 0
	}

	data.SetId(clientScope.Id)
	values := map[string]interface{}{
		"realm_id":                  clientScope.RealmId,
		"name":                      clientScope.Name,
		"description":               clientScope.Description,
		"consent_screen_text":       clientScope.Attributes[clientScopeAttributeConsentScreenText],
		"display_on_consent_screen": clientScopeBoolAttribute(clientScope.Attributes, clientScopeAttributeDisplayOnConsentScreen),
		"gui_order":                 guiOrder,
		"openid_connect":            openidConnect,
		"saml":                      saml,
	}
	for key, value := range values {
		if err := data.Set(key, value); err != nil {
			return err
		}
	}

	extraConfig := map[string]interface{}{}
	modeledKeys := map[string]bool{
		clientScopeAttributeConsentScreenText:               true,
		clientScopeAttributeDisplayOnConsentScreen:          true,
		clientScopeAttributeGuiOrder:                        true,
		clientScopeAttributeIncludeInTokenScope:             true,
		clientScopeAttributeIncludeInOpenidProviderMetadata: true,
	}
	for key, value := range clientScope.Attributes {
		if !modeledKeys[key] && strings.TrimSpace(value) != "" {
			extraConfig[key] = value
		}
	}
	setExtraConfigData(data, extraConfig)

	return nil
}

func resourceKeycloakClientScopeCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	clientScope, err := getKeycloakClientScopeFromData(data)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := keycloakClient.NewClientScope(ctx, clientScope); err != nil {
		return diag.FromErr(err)
	}
	data.SetId(clientScope.Id)

	return resourceKeycloakClientScopeRead(ctx, data, meta)
}

func resourceKeycloakClientScopeRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	clientScope, err := keycloakClient.GetClientScope(ctx, data.Get("realm_id").(string), data.Id())
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	return diag.FromErr(setKeycloakClientScopeData(data, clientScope))
}

func resourceKeycloakClientScopeUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	clientScope, err := getKeycloakClientScopeFromData(data)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := keycloakClient.UpdateClientScope(ctx, clientScope); err != nil {
		return diag.FromErr(err)
	}

	return resourceKeycloakClientScopeRead(ctx, data, meta)
}

func resourceKeycloakClientScopeDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	return diag.FromErr(keycloakClient.DeleteClientScope(ctx, data.Get("realm_id").(string), data.Id()))
}

func resourceKeycloakClientScopeImport(ctx context.Context, data *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(data.Id(), "/")
	if len(parts) != 3 || parts[0] != "realm" {
		return nil, fmt.Errorf("invalid import ID %q; expected realm/<realm_id>/<client_scope_id>", data.Id())
	}

	clientScope, err := meta.(*keycloak.KeycloakClient).GetClientScope(ctx, parts[1], parts[2])
	if err != nil {
		return nil, err
	}
	if err := setKeycloakClientScopeData(data, clientScope); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{data}, nil
}
