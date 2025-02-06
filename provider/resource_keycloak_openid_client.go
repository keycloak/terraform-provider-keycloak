package provider

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"dario.cat/mergo"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"github.com/keycloak/terraform-provider-keycloak/keycloak/types"
)

var (
	keycloakOpenidClientAccessTypes                          = []string{"CONFIDENTIAL", "PUBLIC", "BEARER-ONLY"}
	keycloakOpenidClientAuthorizationPolicyEnforcementMode   = []string{"ENFORCING", "PERMISSIVE", "DISABLED"}
	keycloakOpenidClientResourcePermissionDecisionStrategies = []string{"UNANIMOUS", "AFFIRMATIVE", "CONSENSUS"}
	keycloakOpenidClientPkceCodeChallengeMethod              = []string{"", "plain", "S256"}
)

func resourceKeycloakOpenidClient() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakOpenidClientCreate,
		ReadContext:   resourceKeycloakOpenidClientRead,
		DeleteContext: resourceKeycloakOpenidClientDelete,
		UpdateContext: resourceKeycloakOpenidClientUpdate,
		// This resource can be imported using {{realm}}/{{client_id}}. The Client ID is displayed in the GUI
		Importer: &schema.ResourceImporter{
			StateContext: resourceKeycloakOpenidClientImport,
		},
		Schema: map[string]*schema.Schema{
			"client_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"access_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(keycloakOpenidClientAccessTypes, false),
			},
			"client_secret": {
				Type:      schema.TypeString,
				Optional:  true,
				Computed:  true,
				Sensitive: true,
			},
			"client_authenticator_type": {
				Type:     schema.TypeString,
				Optional: true,
				// No validation is performed since Keycloak plugins can register custom client authenticators
				Default: "client-secret",
			},
			"standard_flow_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"implicit_flow_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"direct_access_grants_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"service_accounts_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"frontchannel_logout_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"valid_redirect_uris": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Optional: true,
				Computed: true,
			},
			"valid_post_logout_redirect_uris": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Optional: true,
				Computed: true,
			},
			"web_origins": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Optional: true,
				Computed: true,
			},
			"root_url": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"admin_url": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"base_url": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"service_account_user_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"pkce_code_challenge_method": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(keycloakOpenidClientPkceCodeChallengeMethod, false),
			},
			"access_token_lifespan": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"client_offline_session_idle_timeout": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"client_offline_session_max_lifespan": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"client_session_idle_timeout": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"client_session_max_lifespan": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"exclude_session_state_from_auth_response": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"resource_server_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authorization": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"policy_enforcement_mode": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(keycloakOpenidClientAuthorizationPolicyEnforcementMode, false),
						},
						"decision_strategy": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(keycloakOpenidClientResourcePermissionDecisionStrategies, false),
							Default:      "UNANIMOUS",
						},
						"allow_remote_resource_management": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"keep_defaults": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"full_scope_allowed": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"consent_required": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"always_display_in_console": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"display_on_consent_screen": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"consent_screen_text": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"authentication_flow_binding_overrides": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"browser_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"direct_grant_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"login_theme": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"use_refresh_tokens": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"use_refresh_tokens_client_credentials": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"frontchannel_logout_url": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"backchannel_logout_url": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"backchannel_logout_session_required": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"backchannel_logout_revoke_offline_sessions": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"extra_config": {
				Type:             schema.TypeMap,
				Optional:         true,
				ValidateDiagFunc: validateExtraConfig(reflect.ValueOf(&keycloak.OpenidClientAttributes{}).Elem()),
			},
			"oauth2_device_authorization_grant_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"oauth2_device_code_lifespan": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"oauth2_device_polling_interval": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"import": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
		},
		CustomizeDiff: customdiff.ComputedIf("service_account_user_id", func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) bool {
			return d.HasChange("service_accounts_enabled")
		}),
	}
}

func getOpenidClientFromData(data *schema.ResourceData) (*keycloak.OpenidClient, error) {
	validRedirectUris := make([]string, 0)
	webOrigins := make([]string, 0)
	validPostLogoutRedirectUris := make([]string, 0)

	rootUrlData, rootUrlOk := data.GetOkExists("root_url")
	validRedirectUrisData, validRedirectUrisOk := data.GetOk("valid_redirect_uris")
	webOriginsData, webOriginsOk := data.GetOk("web_origins")
	validPostLogoutRedirectUrisData, validPostLogoutRedirectUrisOk := data.GetOk("valid_post_logout_redirect_uris")

	rootUrlString := rootUrlData.(string)

	if validRedirectUrisOk {
		for _, validRedirectUri := range validRedirectUrisData.(*schema.Set).List() {
			validRedirectUris = append(validRedirectUris, validRedirectUri.(string))
		}
	}

	if webOriginsOk {
		for _, webOrigin := range webOriginsData.(*schema.Set).List() {
			webOrigins = append(webOrigins, webOrigin.(string))
		}
	}

	if validPostLogoutRedirectUrisOk {
		for _, validPostLogoutRedirectUri := range validPostLogoutRedirectUrisData.(*schema.Set).List() {
			validPostLogoutRedirectUris = append(validPostLogoutRedirectUris, validPostLogoutRedirectUri.(string))
		}
	}

	openidClient := &keycloak.OpenidClient{
		Id:                        data.Id(),
		ClientId:                  data.Get("client_id").(string),
		RealmId:                   data.Get("realm_id").(string),
		Name:                      data.Get("name").(string),
		Enabled:                   data.Get("enabled").(bool),
		Description:               data.Get("description").(string),
		ClientSecret:              data.Get("client_secret").(string),
		ClientAuthenticatorType:   data.Get("client_authenticator_type").(string),
		StandardFlowEnabled:       data.Get("standard_flow_enabled").(bool),
		ImplicitFlowEnabled:       data.Get("implicit_flow_enabled").(bool),
		DirectAccessGrantsEnabled: data.Get("direct_access_grants_enabled").(bool),
		ServiceAccountsEnabled:    data.Get("service_accounts_enabled").(bool),
		FrontChannelLogoutEnabled: data.Get("frontchannel_logout_enabled").(bool),
		FullScopeAllowed:          data.Get("full_scope_allowed").(bool),
		Attributes: keycloak.OpenidClientAttributes{
			PkceCodeChallengeMethod:               data.Get("pkce_code_challenge_method").(string),
			ExcludeSessionStateFromAuthResponse:   types.KeycloakBoolQuoted(data.Get("exclude_session_state_from_auth_response").(bool)),
			AccessTokenLifespan:                   data.Get("access_token_lifespan").(string),
			LoginTheme:                            data.Get("login_theme").(string),
			ClientOfflineSessionIdleTimeout:       data.Get("client_offline_session_idle_timeout").(string),
			ClientOfflineSessionMaxLifespan:       data.Get("client_offline_session_max_lifespan").(string),
			ClientSessionIdleTimeout:              data.Get("client_session_idle_timeout").(string),
			ClientSessionMaxLifespan:              data.Get("client_session_max_lifespan").(string),
			UseRefreshTokens:                      types.KeycloakBoolQuoted(data.Get("use_refresh_tokens").(bool)),
			UseRefreshTokensClientCredentials:     types.KeycloakBoolQuoted(data.Get("use_refresh_tokens_client_credentials").(bool)),
			FrontchannelLogoutUrl:                 data.Get("frontchannel_logout_url").(string),
			BackchannelLogoutUrl:                  data.Get("backchannel_logout_url").(string),
			BackchannelLogoutRevokeOfflineTokens:  types.KeycloakBoolQuoted(data.Get("backchannel_logout_revoke_offline_sessions").(bool)),
			BackchannelLogoutSessionRequired:      types.KeycloakBoolQuoted(data.Get("backchannel_logout_session_required").(bool)),
			ExtraConfig:                           getExtraConfigFromData(data),
			Oauth2DeviceAuthorizationGrantEnabled: types.KeycloakBoolQuoted(data.Get("oauth2_device_authorization_grant_enabled").(bool)),
			Oauth2DeviceCodeLifespan:              data.Get("oauth2_device_code_lifespan").(string),
			Oauth2DevicePollingInterval:           data.Get("oauth2_device_polling_interval").(string),
			ConsentScreenText:                     data.Get("consent_screen_text").(string),
			DisplayOnConsentScreen:                types.KeycloakBoolQuoted(data.Get("display_on_consent_screen").(bool)),
			PostLogoutRedirectUris:                types.KeycloakSliceHashDelimited(validPostLogoutRedirectUris),
		},
		ValidRedirectUris: validRedirectUris,
		WebOrigins:        webOrigins,
		AdminUrl:          data.Get("admin_url").(string),
		BaseUrl:           data.Get("base_url").(string),
		ConsentRequired:   data.Get("consent_required").(bool),
		AlwaysDisplayInConsole: data.Get("always_display_in_console").(bool),
	}

	if rootUrlOk {
		openidClient.RootUrl = &rootUrlString
	}

	if !openidClient.ImplicitFlowEnabled && !openidClient.StandardFlowEnabled {
		if _, ok := data.GetOk("valid_redirect_uris"); ok {
			return nil, errors.New("valid_redirect_uris cannot be set when standard or implicit flow is not enabled")
		}
	}

	if !openidClient.ImplicitFlowEnabled && !openidClient.StandardFlowEnabled && !openidClient.DirectAccessGrantsEnabled {
		if _, ok := data.GetOk("web_origins"); ok {
			return nil, errors.New("web_origins cannot be set when standard or implicit flow is not enabled")
		}
	}

	// access type
	if accessType := data.Get("access_type").(string); accessType == "PUBLIC" {
		openidClient.PublicClient = true
	} else if accessType == "BEARER-ONLY" {
		openidClient.BearerOnly = true
	}

	if v, ok := data.GetOk("authorization"); ok {
		openidClient.AuthorizationServicesEnabled = true
		authorizationSettingsData := v.(*schema.Set).List()[0]
		authorizationSettings := authorizationSettingsData.(map[string]interface{})
		openidClient.AuthorizationSettings = &keycloak.OpenidClientAuthorizationSettings{
			PolicyEnforcementMode:         authorizationSettings["policy_enforcement_mode"].(string),
			DecisionStrategy:              authorizationSettings["decision_strategy"].(string),
			AllowRemoteResourceManagement: authorizationSettings["allow_remote_resource_management"].(bool),
			KeepDefaults:                  authorizationSettings["keep_defaults"].(bool),
		}
	} else {
		openidClient.AuthorizationServicesEnabled = false
	}

	if v, ok := data.GetOk("authentication_flow_binding_overrides"); ok {
		authenticationFlowBindingOverridesData := v.(*schema.Set).List()[0]
		authenticationFlowBindingOverrides := authenticationFlowBindingOverridesData.(map[string]interface{})
		openidClient.AuthenticationFlowBindingOverrides = keycloak.OpenidAuthenticationFlowBindingOverrides{
			BrowserId:     authenticationFlowBindingOverrides["browser_id"].(string),
			DirectGrantId: authenticationFlowBindingOverrides["direct_grant_id"].(string),
		}
	}

	return openidClient, nil
}

func setOpenidClientData(ctx context.Context, keycloakClient *keycloak.KeycloakClient, data *schema.ResourceData, client *keycloak.OpenidClient) error {
	var serviceAccountUserId string
	if client.ServiceAccountsEnabled {
		serviceAccountUser, err := keycloakClient.GetOpenidClientServiceAccountUserId(ctx, client.RealmId, client.Id)
		if err != nil {
			return err
		}
		serviceAccountUserId = serviceAccountUser.Id
	}
	data.SetId(client.Id)
	data.Set("client_id", client.ClientId)
	data.Set("realm_id", client.RealmId)
	data.Set("name", client.Name)
	data.Set("enabled", client.Enabled)
	data.Set("description", client.Description)
	data.Set("client_secret", client.ClientSecret)
	data.Set("client_authenticator_type", client.ClientAuthenticatorType)
	data.Set("standard_flow_enabled", client.StandardFlowEnabled)
	data.Set("implicit_flow_enabled", client.ImplicitFlowEnabled)
	data.Set("direct_access_grants_enabled", client.DirectAccessGrantsEnabled)
	data.Set("service_accounts_enabled", client.ServiceAccountsEnabled)
	data.Set("frontchannel_logout_enabled", client.FrontChannelLogoutEnabled)
	data.Set("valid_redirect_uris", client.ValidRedirectUris)
	data.Set("valid_post_logout_redirect_uris", client.Attributes.PostLogoutRedirectUris)
	data.Set("web_origins", client.WebOrigins)
	data.Set("admin_url", client.AdminUrl)
	data.Set("base_url", client.BaseUrl)
	data.Set("root_url", &client.RootUrl)
	data.Set("full_scope_allowed", client.FullScopeAllowed)
	data.Set("consent_required", client.ConsentRequired)

	data.Set("always_display_in_console", client.AlwaysDisplayInConsole)
	data.Set("access_token_lifespan", client.Attributes.AccessTokenLifespan)
	data.Set("login_theme", client.Attributes.LoginTheme)
	data.Set("use_refresh_tokens", client.Attributes.UseRefreshTokens)
	data.Set("use_refresh_tokens_client_credentials", client.Attributes.UseRefreshTokensClientCredentials)
	data.Set("oauth2_device_authorization_grant_enabled", client.Attributes.Oauth2DeviceAuthorizationGrantEnabled)
	data.Set("oauth2_device_code_lifespan", client.Attributes.Oauth2DeviceCodeLifespan)
	data.Set("oauth2_device_polling_interval", client.Attributes.Oauth2DevicePollingInterval)
	data.Set("client_offline_session_idle_timeout", client.Attributes.ClientOfflineSessionIdleTimeout)
	data.Set("client_offline_session_max_lifespan", client.Attributes.ClientOfflineSessionMaxLifespan)
	data.Set("client_session_idle_timeout", client.Attributes.ClientSessionIdleTimeout)
	data.Set("client_session_max_lifespan", client.Attributes.ClientSessionMaxLifespan)
	data.Set("display_on_consent_screen", client.Attributes.DisplayOnConsentScreen)
	data.Set("consent_screen_text", client.Attributes.ConsentScreenText)
	data.Set("frontchannel_logout_url", client.Attributes.FrontchannelLogoutUrl)
	data.Set("backchannel_logout_url", client.Attributes.BackchannelLogoutUrl)
	data.Set("backchannel_logout_revoke_offline_sessions", client.Attributes.BackchannelLogoutRevokeOfflineTokens)
	data.Set("backchannel_logout_session_required", client.Attributes.BackchannelLogoutSessionRequired)
	setExtraConfigData(data, client.Attributes.ExtraConfig)

	if client.AuthorizationServicesEnabled {
		data.Set("resource_server_id", client.Id)
	}

	if client.ServiceAccountsEnabled {
		data.Set("service_account_user_id", serviceAccountUserId)
	} else {
		data.Set("service_account_user_id", "")
	}

	// access type
	if client.PublicClient {
		data.Set("access_type", "PUBLIC")
	} else if client.BearerOnly {
		data.Set("access_type", "BEARER-ONLY")
	} else {
		data.Set("access_type", "CONFIDENTIAL")
	}

	if (keycloak.OpenidAuthenticationFlowBindingOverrides{}) == client.AuthenticationFlowBindingOverrides {
		data.Set("authentication_flow_binding_overrides", nil)
	} else {
		authenticationFlowBindingOverridesSettings := make(map[string]interface{})
		authenticationFlowBindingOverridesSettings["browser_id"] = client.AuthenticationFlowBindingOverrides.BrowserId
		authenticationFlowBindingOverridesSettings["direct_grant_id"] = client.AuthenticationFlowBindingOverrides.DirectGrantId
		data.Set("authentication_flow_binding_overrides", []interface{}{authenticationFlowBindingOverridesSettings})
	}

	return nil
}

func resourceKeycloakOpenidClientCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	client, err := getOpenidClientFromData(data)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.ValidateOpenidClient(ctx, client)
	if err != nil {
		return diag.FromErr(err)
	}

	if data.Get("import").(bool) {
		existingClient, err := keycloakClient.GetOpenidClientByClientId(ctx, client.RealmId, client.ClientId)
		if err != nil {
			return diag.FromErr(err)
		}

		if err = mergo.Merge(client, existingClient); err != nil {
			return diag.FromErr(err)
		}

		err = keycloakClient.UpdateOpenidClient(ctx, client)
		if err != nil {
			return diag.FromErr(err)
		}
	} else {
		err = keycloakClient.NewOpenidClient(ctx, client)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	err = setOpenidClientData(ctx, keycloakClient, data, client)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceKeycloakOpenidClientRead(ctx, data, meta)
}

func resourceKeycloakOpenidClientRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	client, err := keycloakClient.GetOpenidClient(ctx, realmId, id)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	err = setOpenidClientData(ctx, keycloakClient, data, client)
	if err != nil {
		return diag.FromErr(err)
	}

	if _, ok := data.GetOk("import"); !ok {
		data.Set("import", false)
	}

	return nil
}

func resourceKeycloakOpenidClientUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	client, err := getOpenidClientFromData(data)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.ValidateOpenidClient(ctx, client)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.UpdateOpenidClient(ctx, client)
	if err != nil {
		return diag.FromErr(err)
	}

	err = setOpenidClientData(ctx, keycloakClient, data, client)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakOpenidClientDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if data.Get("import").(bool) {
		return nil
	}
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	return diag.FromErr(keycloakClient.DeleteOpenidClient(ctx, realmId, id))
}

func resourceKeycloakOpenidClientImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("Invalid import. Supported import formats: {{realmId}}/{{openidClientId}}")
	}

	_, err := keycloakClient.GetOpenidClient(ctx, parts[0], parts[1])
	if err != nil {
		return nil, err
	}

	d.Set("realm_id", parts[0])
	d.Set("import", false)
	d.SetId(parts[1])

	diagnostics := resourceKeycloakOpenidClientRead(ctx, d, meta)
	if diagnostics.HasError() {
		return nil, errors.New(diagnostics[0].Summary)
	}

	return []*schema.ResourceData{d}, nil
}
