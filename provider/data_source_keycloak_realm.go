package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func dataSourceKeycloakRealm() *schema.Resource {

	otpPolicySchema := map[string]*schema.Schema{
		"type": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"algorithm": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"digits": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"initial_counter": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"look_ahead_window": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"period": {
			Type:     schema.TypeInt,
			Computed: true,
		},
	}

	webAuthnSchema := map[string]*schema.Schema{
		"acceptable_aaguids": {
			Type: schema.TypeSet,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Computed: true,
		},
		"extra_origins": {
			Type: schema.TypeSet,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Computed: true,
		},
		"attestation_conveyance_preference": {
			Type:        schema.TypeString,
			Description: "Either none, indirect or direct",
			Computed:    true,
		},
		"authenticator_attachment": {
			Type:        schema.TypeString,
			Description: "Either platform or cross-platform",
			Computed:    true,
		},
		"avoid_same_authenticator_register": {
			Type:     schema.TypeBool,
			Computed: true,
		},
		"create_timeout": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"require_resident_key": {
			Type:        schema.TypeString,
			Description: "Either Yes or No",
			Computed:    true,
		},
		"relying_party_entity_name": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"relying_party_id": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"signature_algorithms": {
			Type: schema.TypeSet,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Description: "Keycloak lists ES256, ES384, ES512, RS256, ES384, ES512 at the time of writing",
			Computed:    true,
		},
		"user_verification_requirement": {
			Type:        schema.TypeString,
			Description: "Either required, preferred or discouraged",
			Computed:    true,
		},
	}
	return &schema.Resource{
		ReadContext: dataSourceKeycloakRealmRead,
		Schema: map[string]*schema.Schema{
			"realm": {
				Type:     schema.TypeString,
				Required: true,
			},
			"internal_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"display_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"display_name_html": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"user_managed_access": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"organizations_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"admin_permissions_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			// Login Config

			"registration_allowed": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"registration_email_as_username": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"edit_username_allowed": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"reset_password_allowed": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"remember_me": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"verify_email": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"login_with_email_allowed": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"duplicate_emails_allowed": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"ssl_required": {
				Type:     schema.TypeString,
				Computed: true,
			},

			//Smtp server

			"smtp_server": {
				Type:     schema.TypeList,
				Computed: true,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"starttls": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"port": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"host": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"reply_to": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"reply_to_display_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"from": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"from_display_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"envelope_from": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ssl": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"auth": {
							Type:     schema.TypeList,
							Computed: true,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"username": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"password": {
										Type:      schema.TypeString,
										Computed:  true,
										Sensitive: true,
									},
								},
							},
						},
						"token_auth": {
							Type:     schema.TypeList,
							Computed: true,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"username": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"url": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"client_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"client_secret": {
										Type:      schema.TypeString,
										Computed:  true,
										Sensitive: true,
									},
									"scope": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},

			// Themes

			"login_theme": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"account_theme": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"admin_theme": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"email_theme": {
				Type:     schema.TypeString,
				Computed: true,
			},

			// Tokens

			"default_signature_algorithm": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"revoke_refresh_token": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"refresh_token_max_reuse": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"sso_session_idle_timeout": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sso_session_idle_timeout_remember_me": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sso_session_max_lifespan": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sso_session_max_lifespan_remember_me": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"offline_session_idle_timeout": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"offline_session_max_lifespan": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"offline_session_max_lifespan_enabled": {
				Type:     schema.TypeBool,
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
			"access_token_lifespan": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"access_token_lifespan_for_implicit_flow": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"access_code_lifespan": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"access_code_lifespan_login": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"access_code_lifespan_user_action": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"action_token_generated_by_user_lifespan": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"action_token_generated_by_admin_lifespan": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"oauth2_device_code_lifespan": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"oauth2_device_polling_interval": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			//internationalization
			"internationalization": {
				Type:     schema.TypeList,
				Computed: true,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"supported_locales": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
							Computed: true,
						},
						"default_locale": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			//Security Defenses
			"security_defenses": {
				Type:     schema.TypeList,
				Computed: true,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"headers": {
							Type:     schema.TypeList,
							Computed: true,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"x_frame_options": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"content_security_policy": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"content_security_policy_report_only": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"x_content_type_options": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"x_robots_tag": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"x_xss_protection": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"strict_transport_security": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"referrer_policy": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"brute_force_detection": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"permanent_lockout": { //Permanent Lockout
										Type:     schema.TypeBool,
										Computed: true,
									},
									"max_temporary_lockouts": { //Max Temporary Lockouts
										Type:     schema.TypeInt,
										Computed: true,
									},
									"max_login_failures": { //failureFactor
										Type:     schema.TypeInt,
										Computed: true,
									},
									"wait_increment_seconds": { //Wait Increment
										Type:     schema.TypeInt,
										Computed: true,
									},
									"quick_login_check_milli_seconds": { //Quick Login Check Milli Seconds
										Type:     schema.TypeInt,
										Computed: true,
									},
									"minimum_quick_login_wait_seconds": { //Minimum Quick Login Wait
										Type:     schema.TypeInt,
										Computed: true,
									},
									"max_failure_wait_seconds": { //Max Wait
										Type:     schema.TypeInt,
										Computed: true,
									},
									"failure_reset_time_seconds": { //maxDeltaTimeSeconds
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"password_policy": {
				Type:        schema.TypeString,
				Description: "String that represents the passwordPolicies that are in place. Each policy is separated with \" and \". Supported policies can be found in the server-info providers page. example: \"upperCase(1) and length(8) and forceExpiredPasswordChange(365) and notUsername(undefined)\"",
				Computed:    true,
			},

			//flow bindings
			"browser_flow": {
				Type:        schema.TypeString,
				Description: "Which flow should be used for BrowserFlow",
				Computed:    true,
			},
			"registration_flow": {
				Type:        schema.TypeString,
				Description: "Which flow should be used for RegistrationFlow",
				Computed:    true,
			},
			"direct_grant_flow": {
				Type:        schema.TypeString,
				Description: "Which flow should be used for DirectGrantFlow",
				Computed:    true,
			},
			"reset_credentials_flow": {
				Type:        schema.TypeString,
				Description: "Which flow should be used for ResetCredentialsFlow",
				Computed:    true,
			},
			"client_authentication_flow": {
				Type:        schema.TypeString,
				Description: "Which flow should be used for ClientAuthenticationFlow",
				Computed:    true,
			},
			"docker_authentication_flow": {
				Type:        schema.TypeString,
				Description: "Which flow should be used for DockerAuthenticationFlow",
				Computed:    true,
			},
			"first_broker_login_flow": {
				Type:        schema.TypeString,
				Description: "Which flow should be used for FirstBrokerLoginFlow",
				Computed:    true,
			},
			"attributes": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
			},
			// default default client scopes
			"default_default_client_scopes": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
				Optional: true,
				ForceNew: false,
			},
			// default optional client scopes
			"default_optional_client_scopes": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
				Optional: true,
				ForceNew: false,
			},
			// OTPPolicy
			"otp_policy": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: otpPolicySchema,
				},
			},
			// WebAuthn
			"web_authn_policy": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: webAuthnSchema,
				},
			},

			// WebAuthn Passwordless
			"web_authn_passwordless_policy": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: webAuthnSchema,
				},
			},
		},
	}
}

func dataSourceKeycloakRealmRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	keycloakVersion, err := keycloakClient.Version(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	realmName := data.Get("realm").(string)

	realm, err := keycloakClient.GetRealm(ctx, realmName)
	if err != nil {
		return diag.FromErr(err)
	}

	setRealmData(data, realm, keycloakVersion)

	return nil
}
