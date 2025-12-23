package provider

import (
	"context"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"github.com/keycloak/terraform-provider-keycloak/keycloak/types"
)

var (
	keycloakRealmValidOTPTypes      = []string{"totp", "hotp"}
	keycloakRealmValidOTPAlgorithms = []string{"HmacSHA1", "HmacSHA256", "HmacSHA512"}
)

func resourceKeycloakRealm() *schema.Resource {

	otpPolicySchema := map[string]*schema.Schema{
		"type": {
			Type:         schema.TypeString,
			Description:  "OTP Type, totp for Time-Based One Time Password or hotp for counter base one time password",
			Optional:     true,
			Default:      "totp",
			ValidateFunc: validation.StringInSlice(keycloakRealmValidOTPTypes, false),
		},
		"algorithm": {
			Type:         schema.TypeString,
			Description:  "What hashing algorithm should be used to generate the OTP.",
			Optional:     true,
			Default:      "HmacSHA1",
			ValidateFunc: validation.StringInSlice(keycloakRealmValidOTPAlgorithms, false),
		},
		"digits": {
			Type: schema.TypeInt,
			Elem: &schema.Schema{
				Type: schema.TypeInt,
			},
			Default:  6,
			Optional: true,
		},
		"initial_counter": {
			Type: schema.TypeInt,
			Elem: &schema.Schema{
				Type: schema.TypeInt,
			},
			Default:  2,
			Optional: true,
		},
		"look_ahead_window": {
			Type: schema.TypeInt,
			Elem: &schema.Schema{
				Type: schema.TypeInt,
			},
			Default:  1,
			Optional: true,
		},
		"period": {
			Type: schema.TypeInt,
			Elem: &schema.Schema{
				Type: schema.TypeInt,
			},
			Default:  30,
			Optional: true,
		},
	}

	webAuthnSchema := map[string]*schema.Schema{
		"acceptable_aaguids": {
			Type: schema.TypeSet,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Optional: true,
		},
		"extra_origins": {
			Type: schema.TypeSet,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Optional: true,
		},
		"attestation_conveyance_preference": {
			Type:         schema.TypeString,
			Description:  "Either none, indirect or direct",
			Optional:     true,
			Default:      "not specified",
			ValidateFunc: validation.StringInSlice([]string{"not specified", "none", "indirect", "direct", "enterprise"}, false),
		},
		"authenticator_attachment": {
			Type:         schema.TypeString,
			Description:  "Either platform or cross-platform",
			Optional:     true,
			Default:      "not specified",
			ValidateFunc: validation.StringInSlice([]string{"not specified", "platform", "cross-platform"}, false),
		},
		"avoid_same_authenticator_register": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		"create_timeout": {
			Type:     schema.TypeInt,
			Optional: true,
			Default:  0,
			ValidateFunc: func(i interface{}, k string) ([]string, []error) {
				v := i.(int)

				// https://w3c.github.io/webauthn/#sctn-createCredential
				if v != 0 && (v < 30 || v > 600) {
					return []string{"the recommended timeout value is between 30<->180 seconds (inclusive, userVerification=discouraged) or 30<->600 seconds (inclusive, userVerification=(required || preferred))"}, nil
				}

				return nil, nil
			},
		},
		"require_resident_key": {
			Type:         schema.TypeString,
			Description:  "Either Yes or No",
			Optional:     true,
			Default:      "not specified",
			ValidateFunc: validation.StringInSlice([]string{"not specified", "Yes", "No"}, false),
		},
		"relying_party_entity_name": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "keycloak",
		},
		"relying_party_id": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "",
		},
		"signature_algorithms": {
			Type: schema.TypeSet,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Description: "Keycloak lists ES256, ES384, ES512, RS256, RS384, RS512, RS1 at the time of writing",
			Optional:    true,
			Computed:    true,
		},
		"user_verification_requirement": {
			Type:         schema.TypeString,
			Description:  "Either required, preferred or discouraged",
			Optional:     true,
			Default:      "not specified",
			ValidateFunc: validation.StringInSlice([]string{"not specified", "required", "preferred", "discouraged"}, false),
		},
	}
	return &schema.Resource{
		CreateContext: resourceKeycloakRealmCreate,
		ReadContext:   resourceKeycloakRealmRead,
		DeleteContext: resourceKeycloakRealmDelete,
		UpdateContext: resourceKeycloakRealmUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"realm": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"internal_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"display_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"display_name_html": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"user_managed_access": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"organizations_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"terraform_deletion_protection": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			// Login Config
			"registration_allowed": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"registration_email_as_username": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"edit_username_allowed": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"reset_password_allowed": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"remember_me": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"verify_email": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"login_with_email_allowed": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"duplicate_emails_allowed": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"ssl_required": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "SSL Required: Values can be 'none', 'external' or 'all'.",
				Default:     "external",
			},

			// Smtp server
			"smtp_server": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"starttls": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"port": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"host": {
							Type:     schema.TypeString,
							Required: true,
						},
						"reply_to": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"reply_to_display_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"from": {
							Type:     schema.TypeString,
							Required: true,
						},
						"from_display_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"envelope_from": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ssl": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"allow_utf8": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"auth": {
							Type:          schema.TypeList,
							Optional:      true,
							ConflictsWith: []string{"smtp_server.0.token_auth"},
							MaxItems:      1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"username": {
										Type:     schema.TypeString,
										Required: true,
									},
									"password": {
										Type:      schema.TypeString,
										Required:  true,
										Sensitive: true,
										DiffSuppressFunc: func(_, smtpServerPassword, _ string, _ *schema.ResourceData) bool {
											return smtpServerPassword == "**********"
										},
									},
								},
							},
						},
						"token_auth": {
							Type:          schema.TypeList,
							Optional:      true,
							ConflictsWith: []string{"smtp_server.0.auth"},
							MaxItems:      1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"username": {
										Type:     schema.TypeString,
										Required: true,
									},
									"url": {
										Type:     schema.TypeString,
										Required: true,
									},
									"client_id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"client_secret": {
										Type:      schema.TypeString,
										Required:  true,
										Sensitive: true,
										DiffSuppressFunc: func(_, authTokenClientSecret, _ string, _ *schema.ResourceData) bool {
											return authTokenClientSecret == "**********"
										},
									},
									"scope": {
										Type:     schema.TypeString,
										Required: true,
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
				Optional: true,
			},
			"account_theme": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"admin_theme": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"email_theme": {
				Type:     schema.TypeString,
				Optional: true,
			},

			// Tokens
			"default_signature_algorithm": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"revoke_refresh_token": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"refresh_token_max_reuse": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},
			"sso_session_idle_timeout": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: suppressDurationStringDiff,
			},
			"sso_session_idle_timeout_remember_me": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: suppressDurationStringDiff,
			},
			"sso_session_max_lifespan": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: suppressDurationStringDiff,
			},
			"sso_session_max_lifespan_remember_me": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: suppressDurationStringDiff,
			},
			"offline_session_idle_timeout": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: suppressDurationStringDiff,
			},
			"offline_session_max_lifespan": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: suppressDurationStringDiff,
			},
			"offline_session_max_lifespan_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"client_session_idle_timeout": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: suppressDurationStringDiff,
			},
			"client_session_max_lifespan": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: suppressDurationStringDiff,
			},
			"access_token_lifespan": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: suppressDurationStringDiff,
			},
			"access_token_lifespan_for_implicit_flow": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: suppressDurationStringDiff,
			},
			"access_code_lifespan": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: suppressDurationStringDiff,
			},
			"access_code_lifespan_login": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: suppressDurationStringDiff,
			},
			"access_code_lifespan_user_action": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: suppressDurationStringDiff,
			},
			"action_token_generated_by_user_lifespan": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: suppressDurationStringDiff,
			},
			"action_token_generated_by_admin_lifespan": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: suppressDurationStringDiff,
			},
			"oauth2_device_code_lifespan": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: suppressDurationStringDiff,
			},
			"oauth2_device_polling_interval": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},

			// internationalization
			"internationalization": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"supported_locales": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
							Required: true,
						},
						"default_locale": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},

			// Security Defenses
			"security_defenses": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"headers": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"x_frame_options": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "SAMEORIGIN",
									},
									"content_security_policy": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "frame-src 'self'; frame-ancestors 'self'; object-src 'none';",
									},
									"content_security_policy_report_only": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "",
									},
									"x_content_type_options": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "nosniff",
									},
									"x_robots_tag": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "none",
									},
									"x_xss_protection": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "1; mode=block",
									},
									"strict_transport_security": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "max-age=31536000; includeSubDomains",
									},
									"referrer_policy": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "no-referrer",
									},
								},
							},
						},
						"brute_force_detection": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"permanent_lockout": { //Permanent Lockout
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
									"max_temporary_lockouts": { //Max Temporary Lockouts
										Type:     schema.TypeInt,
										Optional: true,
										Default:  0,
									},
									"max_login_failures": { //failureFactor
										Type:     schema.TypeInt,
										Optional: true,
										Default:  30,
									},
									"wait_increment_seconds": { //Wait Increment
										Type:     schema.TypeInt,
										Optional: true,
										Default:  60,
									},
									"quick_login_check_milli_seconds": { //Quick Login Check Milli Seconds
										Type:     schema.TypeInt,
										Optional: true,
										Default:  1000,
									},
									"minimum_quick_login_wait_seconds": { //Minimum Quick Login Wait
										Type:     schema.TypeInt,
										Optional: true,
										Default:  60,
									},
									"max_failure_wait_seconds": { //Max Wait
										Type:     schema.TypeInt,
										Optional: true,
										Default:  900,
									},
									"failure_reset_time_seconds": { //maxDeltaTimeSeconds
										Type:     schema.TypeInt,
										Optional: true,
										Default:  43200,
									},
								},
							},
						},
					},
				},
			},

			// authentication password policy
			"password_policy": {
				Type:        schema.TypeString,
				Description: "String that represents the passwordPolicies that are in place. Each policy is separated with \" and \". Supported policies can be found in the server-info providers page. example: \"upperCase(1) and length(8) and forceExpiredPasswordChange(365) and notUsername(undefined)\"",
				Optional:    true,
			},

			// authentication flow bindings
			"browser_flow": {
				Type:        schema.TypeString,
				Description: "Which flow should be used for BrowserFlow",
				Optional:    true,
				Computed:    true,
			},
			"registration_flow": {
				Type:        schema.TypeString,
				Description: "Which flow should be used for RegistrationFlow",
				Optional:    true,
				Computed:    true,
			},
			"direct_grant_flow": {
				Type:        schema.TypeString,
				Description: "Which flow should be used for DirectGrantFlow",
				Optional:    true,
				Computed:    true,
			},
			"reset_credentials_flow": {
				Type:        schema.TypeString,
				Description: "Which flow should be used for ResetCredentialsFlow",
				Optional:    true,
				Computed:    true,
			},
			"client_authentication_flow": {
				Type:        schema.TypeString,
				Description: "Which flow should be used for ClientAuthenticationFlow",
				Optional:    true,
				Computed:    true,
			},
			"docker_authentication_flow": {
				Type:        schema.TypeString,
				Description: "Which flow should be used for DockerAuthenticationFlow",
				Optional:    true,
				Computed:    true,
			},
			"first_broker_login_flow": {
				Type:        schema.TypeString,
				Description: "Which flow should be used for FirstBrokerLoginFlow",
				Optional:    true,
				Computed:    true,
			},

			// misc attributes
			"attributes": {
				Type:     schema.TypeMap,
				Optional: true,
			},

			"admin_permissions_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			// default default client scopes
			"default_default_client_scopes": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				ForceNew: false,
			},

			// default optional client scopes
			"default_optional_client_scopes": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
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

func getRealmSMTPPasswordFromData(data *schema.ResourceData) (string, bool) {
	if v, ok := data.GetOk("smtp_server"); ok {
		smtpSettings := v.([]interface{})[0].(map[string]interface{})
		authConfig := smtpSettings["auth"].([]interface{})

		if len(authConfig) == 1 {
			return authConfig[0].(map[string]interface{})["password"].(string), true
		}

		return "", false
	}

	return "", false
}

func setRealmFlowBindings(data *schema.ResourceData, realm *keycloak.Realm, keycloakVersion *version.Version) {
	if flow, ok := data.GetOk("browser_flow"); ok {
		realm.BrowserFlow = stringPointer(flow.(string))
	} else {
		realm.BrowserFlow = stringPointer("browser")
	}

	if flow, ok := data.GetOk("registration_flow"); ok {
		realm.RegistrationFlow = stringPointer(flow.(string))
	} else {
		realm.RegistrationFlow = stringPointer("registration")
	}

	if flow, ok := data.GetOk("direct_grant_flow"); ok {
		realm.DirectGrantFlow = stringPointer(flow.(string))
	} else {
		realm.DirectGrantFlow = stringPointer("direct grant")
	}

	if flow, ok := data.GetOk("reset_credentials_flow"); ok {
		realm.ResetCredentialsFlow = stringPointer(flow.(string))
	} else {
		realm.ResetCredentialsFlow = stringPointer("reset credentials")
	}

	if flow, ok := data.GetOk("client_authentication_flow"); ok {
		realm.ClientAuthenticationFlow = stringPointer(flow.(string))
	} else {
		realm.ClientAuthenticationFlow = stringPointer("clients")
	}

	if flow, ok := data.GetOk("docker_authentication_flow"); ok {
		realm.DockerAuthenticationFlow = stringPointer(flow.(string))
	} else {
		realm.DockerAuthenticationFlow = stringPointer("docker auth")
	}

	if keycloakVersion.GreaterThanOrEqual(keycloak.Version_24.AsVersion()) {
		if flow, ok := data.GetOk("first_broker_login_flow"); ok {
			realm.FirstBrokerLoginFlow = stringPointer(flow.(string))
		} else {
			realm.FirstBrokerLoginFlow = stringPointer("first broker login")
		}
	}
}

func getRealmFromData(data *schema.ResourceData, keycloakVersion *version.Version) (*keycloak.Realm, error) {
	internationalizationEnabled := false
	supportLocales := make([]string, 0)
	defaultLocale := ""
	if v, ok := data.GetOk("internationalization"); ok {
		internationalizationEnabled = true
		internationalizationSettings := v.([]interface{})[0].(map[string]interface{})
		if v, ok := internationalizationSettings["supported_locales"]; ok {
			for _, supportLocale := range v.(*schema.Set).List() {
				supportLocales = append(supportLocales, supportLocale.(string))
			}
		}
		defaultLocale = internationalizationSettings["default_locale"].(string)
	}

	realmId := data.Get("realm")
	internalId := data.Get("internal_id")
	if internalId != "" {
		realmId = internalId
	}

	realm := &keycloak.Realm{
		Id:                   realmId.(string),
		Realm:                data.Get("realm").(string),
		Enabled:              data.Get("enabled").(bool),
		DisplayName:          data.Get("display_name").(string),
		DisplayNameHtml:      data.Get("display_name_html").(string),
		UserManagedAccess:    data.Get("user_managed_access").(bool),
		OrganizationsEnabled: data.Get("organizations_enabled").(bool),

		// Login Config
		RegistrationAllowed:         data.Get("registration_allowed").(bool),
		RegistrationEmailAsUsername: data.Get("registration_email_as_username").(bool),
		EditUsernameAllowed:         data.Get("edit_username_allowed").(bool),
		ResetPasswordAllowed:        data.Get("reset_password_allowed").(bool),
		RememberMe:                  data.Get("remember_me").(bool),
		VerifyEmail:                 data.Get("verify_email").(bool),
		LoginWithEmailAllowed:       data.Get("login_with_email_allowed").(bool),
		DuplicateEmailsAllowed:      data.Get("duplicate_emails_allowed").(bool),
		SslRequired:                 data.Get("ssl_required").(string),

		//internationalization
		InternationalizationEnabled: internationalizationEnabled,
		SupportLocales:              supportLocales,
		DefaultLocale:               defaultLocale,
	}

	//smtp
	if v, ok := data.GetOk("smtp_server"); ok {
		smtpSettings := v.([]interface{})[0].(map[string]interface{})

		smtpServer := keycloak.SmtpServer{
			StartTls:           types.KeycloakBoolQuoted(smtpSettings["starttls"].(bool)),
			Port:               smtpSettings["port"].(string),
			Host:               smtpSettings["host"].(string),
			ReplyTo:            smtpSettings["reply_to"].(string),
			ReplyToDisplayName: smtpSettings["reply_to_display_name"].(string),
			From:               smtpSettings["from"].(string),
			FromDisplayName:    smtpSettings["from_display_name"].(string),
			EnvelopeFrom:       smtpSettings["envelope_from"].(string),
			Ssl:                types.KeycloakBoolQuoted(smtpSettings["ssl"].(bool)),
			AllowUtf8:          types.KeycloakBoolQuoted(smtpSettings["allow_utf8"].(bool)),
		}

		authConfig := smtpSettings["auth"].([]interface{})
		tokenAuthConfig := smtpSettings["token_auth"].([]interface{})

		if len(authConfig) == 1 {
			auth := authConfig[0].(map[string]interface{})

			smtpServer.Auth = true
			smtpServer.AuthType = "basic"
			smtpServer.User = auth["username"].(string)
			smtpServer.Password = auth["password"].(string)
		} else if len(tokenAuthConfig) == 1 {
			tokenAuth := tokenAuthConfig[0].(map[string]interface{})

			smtpServer.Auth = true
			smtpServer.AuthType = "token"
			smtpServer.User = tokenAuth["username"].(string)
			smtpServer.AuthTokenUrl = tokenAuth["url"].(string)
			smtpServer.AuthTokenClientId = tokenAuth["client_id"].(string)
			smtpServer.AuthTokenClientSecret = tokenAuth["client_secret"].(string)
			smtpServer.AuthTokenScope = tokenAuth["scope"].(string)
		} else {
			smtpServer.Auth = false
		}

		realm.SmtpServer = smtpServer
	}

	// Themes

	if loginTheme, ok := data.GetOk("login_theme"); ok {
		realm.LoginTheme = loginTheme.(string)
	}

	if accountTheme, ok := data.GetOk("account_theme"); ok {
		realm.AccountTheme = accountTheme.(string)
	}

	if adminTheme, ok := data.GetOk("admin_theme"); ok {
		realm.AdminTheme = adminTheme.(string)
	}

	if emailTheme, ok := data.GetOk("email_theme"); ok {
		realm.EmailTheme = emailTheme.(string)
	}

	// Tokens

	if defaultSignatureAlgorithm, ok := data.GetOk("default_signature_algorithm"); ok {
		realm.DefaultSignatureAlgorithm = defaultSignatureAlgorithm.(string)
	}

	if revokeRefreshToken, ok := data.GetOk("revoke_refresh_token"); ok {
		realm.RevokeRefreshToken = revokeRefreshToken.(bool)
	}

	if refreshTokenMaxReuse, ok := data.GetOk("refresh_token_max_reuse"); ok {
		realm.RefreshTokenMaxReuse = refreshTokenMaxReuse.(int)
	}

	if ssoSessionIdleTimeout := data.Get("sso_session_idle_timeout").(string); ssoSessionIdleTimeout != "" {
		ssoSessionIdleTimeoutDurationString, err := getSecondsFromDurationString(ssoSessionIdleTimeout)
		if err != nil {
			return nil, err
		}
		realm.SsoSessionIdleTimeout = ssoSessionIdleTimeoutDurationString
	}

	if ssoSessionMaxLifespan := data.Get("sso_session_max_lifespan").(string); ssoSessionMaxLifespan != "" {
		ssoSessionMaxLifespanDurationString, err := getSecondsFromDurationString(ssoSessionMaxLifespan)
		if err != nil {
			return nil, err
		}
		realm.SsoSessionMaxLifespan = ssoSessionMaxLifespanDurationString
	}

	if ssoSessionIdleTimeoutRememberMe := data.Get("sso_session_idle_timeout_remember_me").(string); ssoSessionIdleTimeoutRememberMe != "" {
		ssoSessionIdleTimeoutRememberMeDurationString, err := getSecondsFromDurationString(ssoSessionIdleTimeoutRememberMe)
		if err != nil {
			return nil, err
		}
		realm.SsoSessionIdleTimeoutRememberMe = ssoSessionIdleTimeoutRememberMeDurationString
	}

	if ssoSessionMaxLifespanRememberMe := data.Get("sso_session_max_lifespan_remember_me").(string); ssoSessionMaxLifespanRememberMe != "" {
		ssoSessionMaxLifespanRememberMeDurationString, err := getSecondsFromDurationString(ssoSessionMaxLifespanRememberMe)
		if err != nil {
			return nil, err
		}
		realm.SsoSessionMaxLifespanRememberMe = ssoSessionMaxLifespanRememberMeDurationString
	}

	if offlineSessionIdleTimeout := data.Get("offline_session_idle_timeout").(string); offlineSessionIdleTimeout != "" {
		offlineSessionIdleTimeoutDurationString, err := getSecondsFromDurationString(offlineSessionIdleTimeout)
		if err != nil {
			return nil, err
		}
		realm.OfflineSessionIdleTimeout = offlineSessionIdleTimeoutDurationString
	}

	if offlineSessionMaxLifespan := data.Get("offline_session_max_lifespan").(string); offlineSessionMaxLifespan != "" {
		offlineSessionMaxLifespanDurationString, err := getSecondsFromDurationString(offlineSessionMaxLifespan)
		if err != nil {
			return nil, err
		}
		realm.OfflineSessionMaxLifespan = offlineSessionMaxLifespanDurationString
	}

	if offlineSessionMaxLifespanEnabled, ok := data.GetOk("offline_session_max_lifespan_enabled"); ok {
		realm.OfflineSessionMaxLifespanEnabled = offlineSessionMaxLifespanEnabled.(bool)
	}

	if clientSessionIdleTimeout := data.Get("client_session_idle_timeout").(string); clientSessionIdleTimeout != "" {
		clientSessionIdleTimeoutDurationString, err := getSecondsFromDurationString(clientSessionIdleTimeout)
		if err != nil {
			return nil, err
		}
		realm.ClientSessionIdleTimeout = clientSessionIdleTimeoutDurationString
	}

	if clientSessionMaxLifespan := data.Get("client_session_max_lifespan").(string); clientSessionMaxLifespan != "" {
		clientSessionMaxLifespanDurationString, err := getSecondsFromDurationString(clientSessionMaxLifespan)
		if err != nil {
			return nil, err
		}
		realm.ClientSessionMaxLifespan = clientSessionMaxLifespanDurationString
	}

	if accessTokenLifespan := data.Get("access_token_lifespan").(string); accessTokenLifespan != "" {
		accessTokenLifespanDurationString, err := getSecondsFromDurationString(accessTokenLifespan)
		if err != nil {
			return nil, err
		}
		realm.AccessTokenLifespan = accessTokenLifespanDurationString
	}

	if accessTokenLifespanForImplicitFlow := data.Get("access_token_lifespan_for_implicit_flow").(string); accessTokenLifespanForImplicitFlow != "" {
		accessTokenLifespanForImplicitFlowDurationString, err := getSecondsFromDurationString(accessTokenLifespanForImplicitFlow)
		if err != nil {
			return nil, err
		}
		realm.AccessTokenLifespanForImplicitFlow = accessTokenLifespanForImplicitFlowDurationString
	}

	if accessCodeLifespan := data.Get("access_code_lifespan").(string); accessCodeLifespan != "" {
		accessCodeLifespanDurationString, err := getSecondsFromDurationString(accessCodeLifespan)
		if err != nil {
			return nil, err
		}
		realm.AccessCodeLifespan = accessCodeLifespanDurationString
	}

	if accessCodeLifespanLogin := data.Get("access_code_lifespan_login").(string); accessCodeLifespanLogin != "" {
		accessCodeLifespanLoginDurationString, err := getSecondsFromDurationString(accessCodeLifespanLogin)
		if err != nil {
			return nil, err
		}
		realm.AccessCodeLifespanLogin = accessCodeLifespanLoginDurationString
	}

	if accessCodeLifespanUserAction := data.Get("access_code_lifespan_user_action").(string); accessCodeLifespanUserAction != "" {
		accessCodeLifespanUserActionDurationString, err := getSecondsFromDurationString(accessCodeLifespanUserAction)
		if err != nil {
			return nil, err
		}
		realm.AccessCodeLifespanUserAction = accessCodeLifespanUserActionDurationString
	}

	if actionTokenGeneratedByUserLifespan := data.Get("action_token_generated_by_user_lifespan").(string); actionTokenGeneratedByUserLifespan != "" {
		actionTokenGeneratedByUserLifespanDurationString, err := getSecondsFromDurationString(actionTokenGeneratedByUserLifespan)
		if err != nil {
			return nil, err
		}
		realm.ActionTokenGeneratedByUserLifespan = actionTokenGeneratedByUserLifespanDurationString
	}

	if actionTokenGeneratedByAdminLifespan := data.Get("action_token_generated_by_admin_lifespan").(string); actionTokenGeneratedByAdminLifespan != "" {
		actionTokenGeneratedByAdminLifespanDurationString, err := getSecondsFromDurationString(actionTokenGeneratedByAdminLifespan)
		if err != nil {
			return nil, err
		}
		realm.ActionTokenGeneratedByAdminLifespan = actionTokenGeneratedByAdminLifespanDurationString
	}

	if oauth2DeviceCodeLifespan := data.Get("oauth2_device_code_lifespan").(string); oauth2DeviceCodeLifespan != "" {
		oauth2DeviceCodeLifespanDurationString, err := getSecondsFromDurationString(oauth2DeviceCodeLifespan)
		if err != nil {
			return nil, err
		}
		realm.Oauth2DeviceCodeLifespan = oauth2DeviceCodeLifespanDurationString
	}

	if oauth2DevicePollingInterval, ok := data.GetOk("oauth2_device_polling_interval"); ok {
		realm.Oauth2DevicePollingInterval = oauth2DevicePollingInterval.(int)
	}

	//security defenses
	if v, ok := data.GetOk("security_defenses"); ok {
		securityDefensesSettings := v.([]interface{})[0].(map[string]interface{})

		headersConfig := securityDefensesSettings["headers"].([]interface{})
		if len(headersConfig) == 1 {
			headerSettings := headersConfig[0].(map[string]interface{})

			realm.BrowserSecurityHeaders = keycloak.BrowserSecurityHeaders{
				ContentSecurityPolicy:           headerSettings["content_security_policy"].(string),
				ContentSecurityPolicyReportOnly: headerSettings["content_security_policy_report_only"].(string),
				StrictTransportSecurity:         headerSettings["strict_transport_security"].(string),
				XContentTypeOptions:             headerSettings["x_content_type_options"].(string),
				XFrameOptions:                   headerSettings["x_frame_options"].(string),
				XRobotsTag:                      headerSettings["x_robots_tag"].(string),
				XXSSProtection:                  headerSettings["x_xss_protection"].(string),
				ReferrerPolicy:                  headerSettings["referrer_policy"].(string),
			}
		} else {
			setDefaultSecuritySettingHeaders(realm)
		}

		bruteForceDetectionConfig := securityDefensesSettings["brute_force_detection"].([]interface{})
		versionOk := keycloakVersion.GreaterThanOrEqual(keycloak.Version_24.AsVersion())
		if len(bruteForceDetectionConfig) == 1 {
			bruteForceDetectionSettings := bruteForceDetectionConfig[0].(map[string]interface{})
			realm.BruteForceProtected = true
			realm.PermanentLockout = bruteForceDetectionSettings["permanent_lockout"].(bool)
			realm.FailureFactor = bruteForceDetectionSettings["max_login_failures"].(int)
			realm.WaitIncrementSeconds = bruteForceDetectionSettings["wait_increment_seconds"].(int)
			realm.QuickLoginCheckMilliSeconds = bruteForceDetectionSettings["quick_login_check_milli_seconds"].(int)
			realm.MinimumQuickLoginWaitSeconds = bruteForceDetectionSettings["minimum_quick_login_wait_seconds"].(int)
			realm.MaxFailureWaitSeconds = bruteForceDetectionSettings["max_failure_wait_seconds"].(int)
			realm.MaxDeltaTimeSeconds = bruteForceDetectionSettings["failure_reset_time_seconds"].(int)

			if versionOk {
				realm.MaxTemporaryLockouts = bruteForceDetectionSettings["max_temporary_lockouts"].(int)
			}
		} else {
			setDefaultSecuritySettingsBruteForceDetection(realm, keycloakVersion)
		}
	} else {
		setDefaultSecuritySettingHeaders(realm)
		setDefaultSecuritySettingsBruteForceDetection(realm, keycloakVersion)
	}

	if passwordPolicy, ok := data.GetOk("password_policy"); ok {
		realm.PasswordPolicy = passwordPolicy.(string)
	}

	setRealmFlowBindings(data, realm, keycloakVersion)

	attributes := map[string]interface{}{}
	if v, ok := data.GetOk("attributes"); ok {
		for key, value := range v.(map[string]interface{}) {
			attributes[key] = value
		}
	}
	realm.Attributes = attributes

	defaultDefaultClientScopes := make([]string, 0)
	if v, ok := data.GetOk("default_default_client_scopes"); ok {
		for _, defaultDefaultClientScope := range v.(*schema.Set).List() {
			defaultDefaultClientScopes = append(defaultDefaultClientScopes, defaultDefaultClientScope.(string))
		}
	}
	realm.DefaultDefaultClientScopes = defaultDefaultClientScopes

	defaultOptionalClientScopes := make([]string, 0)
	if v, ok := data.GetOk("default_optional_client_scopes"); ok {
		for _, defaultOptionalClientScope := range v.(*schema.Set).List() {
			defaultOptionalClientScopes = append(defaultOptionalClientScopes, defaultOptionalClientScope.(string))
		}
	}
	realm.DefaultOptionalClientScopes = defaultOptionalClientScopes

	realm.AdminPermissionsEnabled = data.Get("admin_permissions_enabled").(bool)

	//OTPPolicy
	if v, ok := data.GetOk("otp_policy"); ok {
		otpPolicy := v.([]interface{})[0].(map[string]interface{})

		if otpPolicyAlgorithm, ok := otpPolicy["algorithm"]; ok {
			realm.OTPPolicyAlgorithm = otpPolicyAlgorithm.(string)
		}

		if otpPolicyDigits, ok := otpPolicy["digits"]; ok {
			realm.OTPPolicyDigits = otpPolicyDigits.(int)
		}

		if otpPolicyInitialCounter, ok := otpPolicy["initial_counter"]; ok {
			realm.OTPPolicyInitialCounter = otpPolicyInitialCounter.(int)
		}

		if otpPolicyLookAheadWindow, ok := otpPolicy["look_ahead_window"]; ok {
			realm.OTPPolicyLookAheadWindow = otpPolicyLookAheadWindow.(int)
		}

		if otpPolicyPeriod, ok := otpPolicy["period"]; ok {
			realm.OTPPolicyPeriod = otpPolicyPeriod.(int)
		}

		if otpPolicyType, ok := otpPolicy["type"]; ok {
			realm.OTPPolicyType = otpPolicyType.(string)
		}
	}

	//WebAuthn
	if v, ok := data.GetOk("web_authn_policy"); ok {
		webAuthnPolicy := v.([]interface{})[0].(map[string]interface{})

		realm.WebAuthnPolicyAcceptableAaguids = interfaceSliceToStringSlice(webAuthnPolicy["acceptable_aaguids"].(*schema.Set).List())
		realm.WebAuthnPolicyExtraOrigins = interfaceSliceToStringSlice(webAuthnPolicy["extra_origins"].(*schema.Set).List())

		if webAuthnPolicyAttestationConveyancePreference, ok := webAuthnPolicy["attestation_conveyance_preference"]; ok {
			realm.WebAuthnPolicyAttestationConveyancePreference = webAuthnPolicyAttestationConveyancePreference.(string)
		}

		if webAuthnPolicyAuthenticatorAttachment, ok := webAuthnPolicy["authenticator_attachment"]; ok {
			realm.WebAuthnPolicyAuthenticatorAttachment = webAuthnPolicyAuthenticatorAttachment.(string)
		}

		if webAuthnPolicyAvoidSameAuthenticatorRegister, ok := webAuthnPolicy["avoid_same_authenticator_register"]; ok {
			realm.WebAuthnPolicyAvoidSameAuthenticatorRegister = webAuthnPolicyAvoidSameAuthenticatorRegister.(bool)
		}

		if webAuthnPolicyCreateTimeout, ok := webAuthnPolicy["create_timeout"]; ok {
			realm.WebAuthnPolicyCreateTimeout = webAuthnPolicyCreateTimeout.(int)
		}

		if webAuthnPolicyRequireResidentKey, ok := webAuthnPolicy["require_resident_key"]; ok {
			realm.WebAuthnPolicyRequireResidentKey = webAuthnPolicyRequireResidentKey.(string)
		}

		if webAuthnPolicyRpEntityName, ok := webAuthnPolicy["relying_party_entity_name"]; ok {
			realm.WebAuthnPolicyRpEntityName = webAuthnPolicyRpEntityName.(string)
		}

		if webAuthnPolicyRpId, ok := webAuthnPolicy["relying_party_id"]; ok {
			realm.WebAuthnPolicyRpId = webAuthnPolicyRpId.(string)
		}

		realm.WebAuthnPolicySignatureAlgorithms = interfaceSliceToStringSlice(webAuthnPolicy["signature_algorithms"].(*schema.Set).List())

		if webAuthnPolicyUserVerificationRequirement, ok := webAuthnPolicy["user_verification_requirement"]; ok {
			realm.WebAuthnPolicyUserVerificationRequirement = webAuthnPolicyUserVerificationRequirement.(string)
		}
	}

	//WebAuthn Passwordless
	if v, ok := data.GetOk("web_authn_passwordless_policy"); ok {
		webAuthnPasswordlessPolicy := v.([]interface{})[0].(map[string]interface{})

		realm.WebAuthnPolicyPasswordlessAcceptableAaguids = interfaceSliceToStringSlice(webAuthnPasswordlessPolicy["acceptable_aaguids"].(*schema.Set).List())
		realm.WebAuthnPolicyPasswordlessExtraOrigins = interfaceSliceToStringSlice(webAuthnPasswordlessPolicy["extra_origins"].(*schema.Set).List())

		if webAuthnPolicyPasswordlessAttestationConveyancePreference, ok := webAuthnPasswordlessPolicy["attestation_conveyance_preference"]; ok {
			realm.WebAuthnPolicyPasswordlessAttestationConveyancePreference = webAuthnPolicyPasswordlessAttestationConveyancePreference.(string)
		}

		if webAuthnPolicyPasswordlessAuthenticatorAttachment, ok := webAuthnPasswordlessPolicy["authenticator_attachment"]; ok {
			realm.WebAuthnPolicyPasswordlessAuthenticatorAttachment = webAuthnPolicyPasswordlessAuthenticatorAttachment.(string)
		}

		if webAuthnPolicyPasswordlessAvoidSameAuthenticatorRegister, ok := webAuthnPasswordlessPolicy["avoid_same_authenticator_register"]; ok {
			realm.WebAuthnPolicyPasswordlessAvoidSameAuthenticatorRegister = webAuthnPolicyPasswordlessAvoidSameAuthenticatorRegister.(bool)
		}

		if webAuthnPolicyPasswordlessCreateTimeout, ok := webAuthnPasswordlessPolicy["create_timeout"]; ok {
			realm.WebAuthnPolicyPasswordlessCreateTimeout = webAuthnPolicyPasswordlessCreateTimeout.(int)
		}

		if webAuthnPolicyPasswordlessRequireResidentKey, ok := webAuthnPasswordlessPolicy["require_resident_key"]; ok {
			realm.WebAuthnPolicyPasswordlessRequireResidentKey = webAuthnPolicyPasswordlessRequireResidentKey.(string)
		}

		if webAuthnPolicyPasswordlessRpEntityName, ok := webAuthnPasswordlessPolicy["relying_party_entity_name"]; ok {
			realm.WebAuthnPolicyPasswordlessRpEntityName = webAuthnPolicyPasswordlessRpEntityName.(string)
		}

		if webAuthnPolicyPasswordlessRpId, ok := webAuthnPasswordlessPolicy["relying_party_id"]; ok {
			realm.WebAuthnPolicyPasswordlessRpId = webAuthnPolicyPasswordlessRpId.(string)
		}

		realm.WebAuthnPolicyPasswordlessSignatureAlgorithms = interfaceSliceToStringSlice(webAuthnPasswordlessPolicy["signature_algorithms"].(*schema.Set).List())

		if webAuthnPolicyPasswordlessUserVerificationRequirement, ok := webAuthnPasswordlessPolicy["user_verification_requirement"]; ok {
			realm.WebAuthnPolicyPasswordlessUserVerificationRequirement = webAuthnPolicyPasswordlessUserVerificationRequirement.(string)
		}
	}

	return realm, nil
}

func setDefaultSecuritySettingHeaders(realm *keycloak.Realm) {
	realm.BrowserSecurityHeaders = keycloak.BrowserSecurityHeaders{
		ContentSecurityPolicy:           "frame-src 'self'; frame-ancestors 'self'; object-src 'none';",
		ContentSecurityPolicyReportOnly: "",
		StrictTransportSecurity:         "max-age=31536000; includeSubDomains",
		XContentTypeOptions:             "nosniff",
		XFrameOptions:                   "SAMEORIGIN",
		XRobotsTag:                      "none",
		XXSSProtection:                  "1; mode=block",
		ReferrerPolicy:                  "no-referrer",
	}
}

func setDefaultSecuritySettingsBruteForceDetection(realm *keycloak.Realm, keycloakVersion *version.Version) {
	realm.BruteForceProtected = false
	realm.PermanentLockout = false
	realm.FailureFactor = 30
	realm.WaitIncrementSeconds = 60
	realm.QuickLoginCheckMilliSeconds = 1000
	realm.MinimumQuickLoginWaitSeconds = 60
	realm.MaxFailureWaitSeconds = 900
	realm.MaxDeltaTimeSeconds = 43200

	if keycloakVersion.GreaterThanOrEqual(keycloak.Version_24.AsVersion()) {
		realm.MaxTemporaryLockouts = 0
	}
}

func setRealmData(data *schema.ResourceData, realm *keycloak.Realm, keycloakVersion *version.Version) {
	data.SetId(realm.Realm)

	data.Set("realm", realm.Realm)
	data.Set("internal_id", realm.Id)
	data.Set("enabled", realm.Enabled)
	data.Set("display_name", realm.DisplayName)
	data.Set("display_name_html", realm.DisplayNameHtml)
	data.Set("user_managed_access", realm.UserManagedAccess)
	data.Set("organizations_enabled", realm.OrganizationsEnabled)
	data.Set("admin_permissions_enabled", realm.AdminPermissionsEnabled)

	// Login Config
	data.Set("registration_allowed", realm.RegistrationAllowed)
	data.Set("registration_email_as_username", realm.RegistrationEmailAsUsername)
	data.Set("edit_username_allowed", realm.EditUsernameAllowed)
	data.Set("reset_password_allowed", realm.ResetPasswordAllowed)
	data.Set("remember_me", realm.RememberMe)
	data.Set("verify_email", realm.VerifyEmail)
	data.Set("login_with_email_allowed", realm.LoginWithEmailAllowed)
	data.Set("duplicate_emails_allowed", realm.DuplicateEmailsAllowed)
	data.Set("ssl_required", realm.SslRequired)

	// Smtp Config

	if (keycloak.SmtpServer{}) == realm.SmtpServer {
		data.Set("smtp_server", nil)
	} else {
		smtpSettings := make(map[string]interface{})

		smtpSettings["starttls"] = realm.SmtpServer.StartTls
		smtpSettings["port"] = realm.SmtpServer.Port
		smtpSettings["host"] = realm.SmtpServer.Host
		smtpSettings["reply_to"] = realm.SmtpServer.ReplyTo
		smtpSettings["reply_to_display_name"] = realm.SmtpServer.ReplyToDisplayName
		smtpSettings["from"] = realm.SmtpServer.From
		smtpSettings["from_display_name"] = realm.SmtpServer.FromDisplayName
		smtpSettings["envelope_from"] = realm.SmtpServer.EnvelopeFrom
		smtpSettings["ssl"] = realm.SmtpServer.Ssl
		smtpSettings["allow_utf8"] = realm.SmtpServer.AllowUtf8

		if realm.SmtpServer.Auth {
			if realm.SmtpServer.AuthType == "token" {
				token_auth := make(map[string]interface{})

				token_auth["username"] = realm.SmtpServer.User
				token_auth["url"] = realm.SmtpServer.AuthTokenUrl
				token_auth["client_id"] = realm.SmtpServer.AuthTokenClientId
				token_auth["client_secret"] = realm.SmtpServer.AuthTokenClientSecret
				token_auth["scope"] = realm.SmtpServer.AuthTokenScope

				smtpSettings["token_auth"] = []interface{}{token_auth}
			} else {
				auth := make(map[string]interface{})

				auth["username"] = realm.SmtpServer.User
				auth["password"] = realm.SmtpServer.Password

				smtpSettings["auth"] = []interface{}{auth}
			}
		}

		data.Set("smtp_server", []interface{}{smtpSettings})
	}

	// Themes
	data.Set("login_theme", realm.LoginTheme)
	data.Set("account_theme", realm.AccountTheme)
	data.Set("admin_theme", realm.AdminTheme)
	data.Set("email_theme", realm.EmailTheme)

	// Tokens
	data.Set("default_signature_algorithm", realm.DefaultSignatureAlgorithm)
	data.Set("revoke_refresh_token", realm.RevokeRefreshToken)
	data.Set("refresh_token_max_reuse", realm.RefreshTokenMaxReuse)
	data.Set("sso_session_idle_timeout", getDurationStringFromSeconds(realm.SsoSessionIdleTimeout))
	data.Set("sso_session_max_lifespan", getDurationStringFromSeconds(realm.SsoSessionMaxLifespan))
	data.Set("sso_session_idle_timeout_remember_me", getDurationStringFromSeconds(realm.SsoSessionIdleTimeoutRememberMe))
	data.Set("sso_session_max_lifespan_remember_me", getDurationStringFromSeconds(realm.SsoSessionMaxLifespanRememberMe))
	data.Set("offline_session_idle_timeout", getDurationStringFromSeconds(realm.OfflineSessionIdleTimeout))
	data.Set("offline_session_max_lifespan", getDurationStringFromSeconds(realm.OfflineSessionMaxLifespan))
	data.Set("offline_session_max_lifespan_enabled", realm.OfflineSessionMaxLifespanEnabled)
	data.Set("client_session_idle_timeout", getDurationStringFromSeconds(realm.ClientSessionIdleTimeout))
	data.Set("client_session_max_lifespan", getDurationStringFromSeconds(realm.ClientSessionMaxLifespan))
	data.Set("access_token_lifespan", getDurationStringFromSeconds(realm.AccessTokenLifespan))
	data.Set("access_token_lifespan_for_implicit_flow", getDurationStringFromSeconds(realm.AccessTokenLifespanForImplicitFlow))
	data.Set("access_code_lifespan", getDurationStringFromSeconds(realm.AccessCodeLifespan))
	data.Set("access_code_lifespan_login", getDurationStringFromSeconds(realm.AccessCodeLifespanLogin))
	data.Set("access_code_lifespan_user_action", getDurationStringFromSeconds(realm.AccessCodeLifespanUserAction))
	data.Set("action_token_generated_by_user_lifespan", getDurationStringFromSeconds(realm.ActionTokenGeneratedByUserLifespan))
	data.Set("action_token_generated_by_admin_lifespan", getDurationStringFromSeconds(realm.ActionTokenGeneratedByAdminLifespan))
	data.Set("oauth2_device_code_lifespan", getDurationStringFromSeconds(realm.Oauth2DeviceCodeLifespan))
	data.Set("oauth2_device_polling_interval", realm.Oauth2DevicePollingInterval)

	//internationalization
	if realm.InternationalizationEnabled {
		internationalizationSettings := make(map[string]interface{})
		internationalizationSettings["supported_locales"] = realm.SupportLocales
		internationalizationSettings["default_locale"] = realm.DefaultLocale
		data.Set("internationalization", []interface{}{internationalizationSettings})
	} else {
		data.Set("internationalization", nil)
	}

	if v, ok := data.GetOk("security_defenses"); ok {
		oldHeadersConfig := v.([]interface{})[0].(map[string]interface{})["headers"].([]interface{})
		if len(oldHeadersConfig) == 0 && !realm.BruteForceProtected {
			data.Set("security_defenses", nil)
		} else if len(oldHeadersConfig) == 1 && realm.BruteForceProtected {
			securityDefensesSettings := make(map[string]interface{})
			securityDefensesSettings["headers"] = []interface{}{getHeaderSettings(realm)}
			securityDefensesSettings["brute_force_detection"] = []interface{}{getBruteForceDetectionSettings(realm, keycloakVersion)}
			data.Set("security_defenses", []interface{}{securityDefensesSettings})
		} else if len(oldHeadersConfig) == 1 {
			securityDefensesSettings := make(map[string]interface{})
			securityDefensesSettings["headers"] = []interface{}{getHeaderSettings(realm)}
			data.Set("security_defenses", []interface{}{securityDefensesSettings})
		} else if realm.BruteForceProtected {
			securityDefensesSettings := make(map[string]interface{})
			securityDefensesSettings["brute_force_detection"] = []interface{}{getBruteForceDetectionSettings(realm, keycloakVersion)}
			data.Set("security_defenses", []interface{}{securityDefensesSettings})
		}
	}

	data.Set("password_policy", realm.PasswordPolicy)

	//Flow Bindings
	data.Set("browser_flow", realm.BrowserFlow)
	data.Set("registration_flow", realm.RegistrationFlow)
	data.Set("direct_grant_flow", realm.DirectGrantFlow)
	data.Set("reset_credentials_flow", realm.ResetCredentialsFlow)
	data.Set("client_authentication_flow", realm.ClientAuthenticationFlow)
	data.Set("docker_authentication_flow", realm.DockerAuthenticationFlow)

	if keycloakVersion.GreaterThanOrEqual(keycloak.Version_24.AsVersion()) {
		data.Set("first_broker_login_flow", realm.FirstBrokerLoginFlow)
	}

	//WebAuthn
	webAuthnPolicy := make(map[string]interface{})
	webAuthnPolicy["acceptable_aaguids"] = realm.WebAuthnPolicyAcceptableAaguids
	webAuthnPolicy["extra_origins"] = realm.WebAuthnPolicyExtraOrigins
	webAuthnPolicy["attestation_conveyance_preference"] = realm.WebAuthnPolicyAttestationConveyancePreference
	webAuthnPolicy["authenticator_attachment"] = realm.WebAuthnPolicyAuthenticatorAttachment
	webAuthnPolicy["avoid_same_authenticator_register"] = realm.WebAuthnPolicyAvoidSameAuthenticatorRegister
	webAuthnPolicy["create_timeout"] = realm.WebAuthnPolicyCreateTimeout
	webAuthnPolicy["require_resident_key"] = realm.WebAuthnPolicyRequireResidentKey
	webAuthnPolicy["relying_party_entity_name"] = realm.WebAuthnPolicyRpEntityName
	webAuthnPolicy["relying_party_id"] = realm.WebAuthnPolicyRpId
	webAuthnPolicy["signature_algorithms"] = realm.WebAuthnPolicySignatureAlgorithms
	webAuthnPolicy["user_verification_requirement"] = realm.WebAuthnPolicyUserVerificationRequirement
	data.Set("web_authn_policy", []interface{}{webAuthnPolicy})

	//OTP Policy
	otpPolicy := make(map[string]interface{})
	otpPolicy["type"] = realm.OTPPolicyType
	otpPolicy["algorithm"] = realm.OTPPolicyAlgorithm
	otpPolicy["digits"] = realm.OTPPolicyDigits
	otpPolicy["initial_counter"] = realm.OTPPolicyInitialCounter
	otpPolicy["look_ahead_window"] = realm.OTPPolicyLookAheadWindow
	otpPolicy["period"] = realm.OTPPolicyPeriod
	data.Set("otp_policy", []interface{}{otpPolicy})

	//WebAuthn Passwordless
	webAuthnPasswordlessPolicy := make(map[string]interface{})
	webAuthnPasswordlessPolicy["acceptable_aaguids"] = realm.WebAuthnPolicyPasswordlessAcceptableAaguids
	webAuthnPasswordlessPolicy["extra_origins"] = realm.WebAuthnPolicyPasswordlessExtraOrigins
	webAuthnPasswordlessPolicy["attestation_conveyance_preference"] = realm.WebAuthnPolicyPasswordlessAttestationConveyancePreference
	webAuthnPasswordlessPolicy["authenticator_attachment"] = realm.WebAuthnPolicyPasswordlessAuthenticatorAttachment
	webAuthnPasswordlessPolicy["avoid_same_authenticator_register"] = realm.WebAuthnPolicyPasswordlessAvoidSameAuthenticatorRegister
	webAuthnPasswordlessPolicy["create_timeout"] = realm.WebAuthnPolicyPasswordlessCreateTimeout
	webAuthnPasswordlessPolicy["require_resident_key"] = realm.WebAuthnPolicyPasswordlessRequireResidentKey
	webAuthnPasswordlessPolicy["relying_party_entity_name"] = realm.WebAuthnPolicyPasswordlessRpEntityName
	webAuthnPasswordlessPolicy["relying_party_id"] = realm.WebAuthnPolicyPasswordlessRpId
	webAuthnPasswordlessPolicy["signature_algorithms"] = realm.WebAuthnPolicyPasswordlessSignatureAlgorithms
	webAuthnPasswordlessPolicy["user_verification_requirement"] = realm.WebAuthnPolicyPasswordlessUserVerificationRequirement
	data.Set("web_authn_passwordless_policy", []interface{}{webAuthnPasswordlessPolicy})

	attributes := map[string]interface{}{}
	if v, ok := data.GetOk("attributes"); ok {
		for key := range v.(map[string]interface{}) {
			attributes[key] = realm.Attributes[key]
			//We are only interested in attributes managed in terraform (Keycloak returns a lot of doubles values in the attributes...)
		}
	}
	data.Set("attributes", attributes)

	// default and optional client scope mappings
	data.Set("default_default_client_scopes", realm.DefaultDefaultClientScopes)
	data.Set("default_optional_client_scopes", realm.DefaultOptionalClientScopes)
}

func getBruteForceDetectionSettings(realm *keycloak.Realm, keycloakVersion *version.Version) map[string]interface{} {
	bruteForceDetectionSettings := make(map[string]interface{})
	bruteForceDetectionSettings["permanent_lockout"] = realm.PermanentLockout
	bruteForceDetectionSettings["max_login_failures"] = realm.FailureFactor
	bruteForceDetectionSettings["wait_increment_seconds"] = realm.WaitIncrementSeconds
	bruteForceDetectionSettings["quick_login_check_milli_seconds"] = realm.QuickLoginCheckMilliSeconds
	bruteForceDetectionSettings["minimum_quick_login_wait_seconds"] = realm.MinimumQuickLoginWaitSeconds
	bruteForceDetectionSettings["max_failure_wait_seconds"] = realm.MaxFailureWaitSeconds
	bruteForceDetectionSettings["failure_reset_time_seconds"] = realm.MaxDeltaTimeSeconds

	if keycloakVersion.GreaterThanOrEqual(keycloak.Version_24.AsVersion()) {
		bruteForceDetectionSettings["max_temporary_lockouts"] = realm.MaxTemporaryLockouts
	}
	return bruteForceDetectionSettings
}

func getHeaderSettings(realm *keycloak.Realm) map[string]interface{} {
	headersSettings := make(map[string]interface{})
	headersSettings["content_security_policy"] = realm.BrowserSecurityHeaders.ContentSecurityPolicy
	headersSettings["content_security_policy_report_only"] = realm.BrowserSecurityHeaders.ContentSecurityPolicyReportOnly
	headersSettings["strict_transport_security"] = realm.BrowserSecurityHeaders.StrictTransportSecurity
	headersSettings["x_content_type_options"] = realm.BrowserSecurityHeaders.XContentTypeOptions
	headersSettings["x_frame_options"] = realm.BrowserSecurityHeaders.XFrameOptions
	headersSettings["x_robots_tag"] = realm.BrowserSecurityHeaders.XRobotsTag
	headersSettings["x_xss_protection"] = realm.BrowserSecurityHeaders.XXSSProtection
	headersSettings["referrer_policy"] = realm.BrowserSecurityHeaders.ReferrerPolicy
	return headersSettings
}

func resourceKeycloakRealmCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	keycloakVersion, err := keycloakClient.Version(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	realm, err := getRealmFromData(data, keycloakVersion)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.ValidateRealm(ctx, realm)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.NewRealm(ctx, realm)
	if err != nil {
		return diag.FromErr(err)
	}

	err = meta.(*keycloak.KeycloakClient).Refresh(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	setRealmData(data, realm, keycloakVersion)

	return resourceKeycloakRealmRead(ctx, data, meta)
}

func resourceKeycloakRealmRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	keycloakVersion, err := keycloakClient.Version(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	realm, err := keycloakClient.GetRealm(ctx, data.Id())
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	// we can't trust the API to set this field correctly since it just responds with "**********" this implies a 'password only' change will not be detected
	if smtpPassword, ok := getRealmSMTPPasswordFromData(data); ok {
		realm.SmtpServer.Password = smtpPassword
	}

	if _, ok := data.GetOk("terraform_deletion_protection"); !ok {
		data.Set("terraform_deletion_protection", false)
	}

	setRealmData(data, realm, keycloakVersion)

	return nil
}

func resourceKeycloakRealmUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	keycloakVersion, err := keycloakClient.Version(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	realm, err := getRealmFromData(data, keycloakVersion)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.ValidateRealm(ctx, realm)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.UpdateRealm(ctx, realm)
	if err != nil {
		return diag.FromErr(err)
	}

	setRealmData(data, realm, keycloakVersion)

	return nil
}

func resourceKeycloakRealmDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if data.Get("terraform_deletion_protection").(bool) {
		return diag.Errorf("Deletion protection is enabled for keycloak_realm resource with realm %s (ID: %s). To delete this resource, first set `terraform_deletion_protection` to `false`.", data.Id(), data.Get("internal_id").(string))
	}
	keycloakClient := meta.(*keycloak.KeycloakClient)

	return diag.FromErr(keycloakClient.DeleteRealm(ctx, data.Id()))
}
