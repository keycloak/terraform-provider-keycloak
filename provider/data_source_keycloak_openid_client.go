package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func dataSourceKeycloakOpenidClient() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKeycloakOpenidClientRead,

		Schema: map[string]*schema.Schema{
			"client_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"access_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"client_secret": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"client_authenticator_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"standard_flow_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"implicit_flow_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"direct_access_grants_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"service_accounts_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"frontchannel_logout_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"valid_redirect_uris": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Computed: true,
			},
			"valid_post_logout_redirect_uris": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Computed: true,
			},
			"web_origins": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Computed: true,
			},
			"root_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"admin_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"base_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_account_user_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"pkce_code_challenge_method": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"access_token_lifespan": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"client_offline_session_idle_timeout": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"client_offline_session_max_lifespan": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"client_session_idle_timeout": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"client_session_max_lifespan": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"exclude_session_state_from_auth_response": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"resource_server_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authorization": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"policy_enforcement_mode": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"decision_strategy": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"allow_remote_resource_management": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"keep_defaults": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"full_scope_allowed": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"always_display_in_console" : {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"consent_required": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"display_on_consent_screen": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"consent_screen_text": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"authentication_flow_binding_overrides": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"browser_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"direct_grant_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"login_theme": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"use_refresh_tokens": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"use_refresh_tokens_client_credentials": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"backchannel_logout_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"frontchannel_logout_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"backchannel_logout_session_required": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"backchannel_logout_revoke_offline_sessions": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"extra_config": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
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
		},
	}
}

func dataSourceKeycloakOpenidClientRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)

	client, err := keycloakClient.GetOpenidClientByClientId(ctx, realmId, clientId)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	err = setOpenidClientData(ctx, keycloakClient, data, client)
	if err != nil {
		return diag.FromErr(err)
	}
	data.Set("extra_config", client.Attributes.ExtraConfig)

	return nil
}
