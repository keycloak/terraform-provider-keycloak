package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"reflect"
	"strings"
)

var syncModes = []string{
	"IMPORT",
	"FORCE",
	"LEGACY",
}

type identityProviderDataGetterFunc func(data *schema.ResourceData) (*keycloak.IdentityProvider, error)
type identityProviderDataSetterFunc func(data *schema.ResourceData, identityProvider *keycloak.IdentityProvider) error

func resourceKeycloakIdentityProvider() *schema.Resource {
	return &schema.Resource{
		DeleteContext: resourceKeycloakIdentityProviderDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceKeycloakIdentityProviderImport,
		},
		Schema: map[string]*schema.Schema{
			"alias": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The alias uniquely identifies an identity provider and it is also used to build the redirect uri.",
			},
			"realm": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Realm Name",
			},
			"internal_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Internal Identity Provider Id",
			},
			"display_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Friendly name for Identity Providers.",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Enable/disable this identity provider.",
			},
			"store_token": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Enable/disable if tokens must be stored after authenticating users.",
			},
			"add_read_token_role_on_create": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
				Description: "Enable/disable if new users can read any stored tokens. This assigns the broker.read-token role.",
			},
			"authenticate_by_default": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Enable/disable authenticate users by default.",
			},
			"link_only": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "If true, users cannot log in through this provider.  They can only link to this provider.  This is useful if you don't want to allow login from the provider, but want to integrate with a provider",
			},
			"trust_email": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "If enabled then email provided by this provider is not verified even if verification is enabled for the realm.",
			},
			"first_broker_login_flow_alias": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "first broker login",
				Description: "Alias of authentication flow, which is triggered after first login with this identity provider. Term 'First Login' means that there is not yet existing Keycloak account linked with the authenticated identity provider account.",
			},
			"post_broker_login_flow_alias": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Alias of authentication flow, which is triggered after each login with this identity provider. Useful if you want additional verification of each user authenticated with this identity provider (for example OTP). Leave this empty if you don't want any additional authenticators to be triggered after login with this identity provider. Also note, that authenticator implementations must assume that user is already set in ClientSession as identity provider already set it.",
			},
			// all schema values below this point will be configuration values that are shared among all identity providers
			"extra_config": {
				Type:             schema.TypeMap,
				Optional:         true,
				ValidateDiagFunc: validateExtraConfig(reflect.ValueOf(&keycloak.IdentityProviderConfig{}).Elem()),
			},
			"gui_order": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "GUI Order",
			},
			"sync_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "",
				ValidateFunc: validation.StringInSlice(syncModes, false),
				Description:  "Sync Mode",
			},
		},
	}
}

func getIdentityProviderFromData(data *schema.ResourceData) (*keycloak.IdentityProvider, *keycloak.IdentityProviderConfig) {
	// some identity provider config is shared among all identity providers, so this default config will be used as a base to merge extra config into
	defaultIdentityProviderConfig := &keycloak.IdentityProviderConfig{
		GuiOrder:    data.Get("gui_order").(string),
		SyncMode:    data.Get("sync_mode").(string),
		ExtraConfig: getExtraConfigFromData(data),
	}

	return &keycloak.IdentityProvider{
		Realm:                     data.Get("realm").(string),
		Alias:                     data.Get("alias").(string),
		DisplayName:               data.Get("display_name").(string),
		Enabled:                   data.Get("enabled").(bool),
		StoreToken:                data.Get("store_token").(bool),
		AddReadTokenRoleOnCreate:  data.Get("add_read_token_role_on_create").(bool),
		AuthenticateByDefault:     data.Get("authenticate_by_default").(bool),
		LinkOnly:                  data.Get("link_only").(bool),
		TrustEmail:                data.Get("trust_email").(bool),
		FirstBrokerLoginFlowAlias: data.Get("first_broker_login_flow_alias").(string),
		PostBrokerLoginFlowAlias:  data.Get("post_broker_login_flow_alias").(string),
		InternalId:                data.Get("internal_id").(string),
	}, defaultIdentityProviderConfig
}

func setIdentityProviderData(data *schema.ResourceData, identityProvider *keycloak.IdentityProvider) {
	data.SetId(identityProvider.Alias)

	data.Set("internal_id", identityProvider.InternalId)
	data.Set("realm", identityProvider.Realm)
	data.Set("alias", identityProvider.Alias)
	data.Set("display_name", identityProvider.DisplayName)
	data.Set("enabled", identityProvider.Enabled)
	data.Set("store_token", identityProvider.StoreToken)
	data.Set("add_read_token_role_on_create", identityProvider.AddReadTokenRoleOnCreate)
	data.Set("authenticate_by_default", identityProvider.AuthenticateByDefault)
	data.Set("link_only", identityProvider.LinkOnly)
	data.Set("trust_email", identityProvider.TrustEmail)
	data.Set("first_broker_login_flow_alias", identityProvider.FirstBrokerLoginFlowAlias)
	data.Set("post_broker_login_flow_alias", identityProvider.PostBrokerLoginFlowAlias)

	// identity provider config
	data.Set("gui_order", identityProvider.Config.GuiOrder)
	data.Set("sync_mode", identityProvider.Config.SyncMode)
	setExtraConfigData(data, identityProvider.Config.ExtraConfig)
}

func resourceKeycloakIdentityProviderDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realm := data.Get("realm").(string)
	alias := data.Get("alias").(string)

	return diag.FromErr(keycloakClient.DeleteIdentityProvider(ctx, realm, alias))
}

func resourceKeycloakIdentityProviderImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")

	if len(parts) != 2 {
		return nil, fmt.Errorf("Invalid import. Supported import formats: {{realm}}/{{identityProviderAlias}}")
	}

	d.Set("realm", parts[0])
	d.Set("alias", parts[1])
	d.SetId(parts[1])

	return []*schema.ResourceData{d}, nil
}

func resourceKeycloakIdentityProviderCreate(getIdentityProviderFromData identityProviderDataGetterFunc, setDataFromIdentityProvider identityProviderDataSetterFunc) func(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return func(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
		keycloakClient := meta.(*keycloak.KeycloakClient)
		identityProvider, err := getIdentityProviderFromData(data)
		if err != nil {
			return diag.FromErr(err)
		}

		if err = keycloakClient.NewIdentityProvider(ctx, identityProvider); err != nil {
			return diag.FromErr(err)
		}
		if err = setDataFromIdentityProvider(data, identityProvider); err != nil {
			return diag.FromErr(err)
		}
		return resourceKeycloakIdentityProviderRead(setDataFromIdentityProvider)(ctx, data, meta)
	}
}

func resourceKeycloakIdentityProviderRead(setDataFromIdentityProvider identityProviderDataSetterFunc) func(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return func(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
		keycloakClient := meta.(*keycloak.KeycloakClient)
		realm := data.Get("realm").(string)
		alias := data.Get("alias").(string)
		identityProvider, err := keycloakClient.GetIdentityProvider(ctx, realm, alias)
		if err != nil {
			return handleNotFoundError(ctx, err, data)
		}

		return diag.FromErr(setDataFromIdentityProvider(data, identityProvider))
	}
}

func resourceKeycloakIdentityProviderUpdate(getIdentityProviderFromData identityProviderDataGetterFunc, setDataFromIdentityProvider identityProviderDataSetterFunc) func(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return func(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
		keycloakClient := meta.(*keycloak.KeycloakClient)
		identityProvider, err := getIdentityProviderFromData(data)
		if err != nil {
			return diag.FromErr(err)
		}

		err = keycloakClient.UpdateIdentityProvider(ctx, identityProvider)
		if err != nil {
			return diag.FromErr(err)
		}

		return diag.FromErr(setDataFromIdentityProvider(data, identityProvider))
	}
}
