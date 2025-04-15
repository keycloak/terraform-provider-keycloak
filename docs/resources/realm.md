---
page_title: "keycloak_realm Resource"
---

# keycloak\_realm Resource

Allows for creating and managing Realms within Keycloak.

A realm manages a logical collection of users, credentials, roles, and groups. Users log in to realms and can be federated
from multiple sources.

## Example Usage

```hcl
resource "keycloak_realm" "realm" {
  realm             = "my-realm"
  enabled           = true
  display_name      = "my realm"
  display_name_html = "<b>my realm</b>"

  login_theme = "base"

  access_code_lifespan = "1h"

  ssl_required    = "external"
  password_policy = "upperCase(1) and length(8) and forceExpiredPasswordChange(365) and notUsername"
  attributes      = {
    mycustomAttribute = "myCustomValue"
  }

  smtp_server {
    host = "smtp.example.com"
    from = "example@example.com"

    auth {
      username = "tom"
      password = "password"
    }
  }

  internationalization {
    supported_locales = [
      "en",
      "de",
      "es"
    ]
    default_locale    = "en"
  }

  security_defenses {
    headers {
      x_frame_options                     = "DENY"
      content_security_policy             = "frame-src 'self'; frame-ancestors 'self'; object-src 'none';"
      content_security_policy_report_only = ""
      x_content_type_options              = "nosniff"
      x_robots_tag                        = "none"
      x_xss_protection                    = "1; mode=block"
      strict_transport_security           = "max-age=31536000; includeSubDomains"
    }
    brute_force_detection {
      permanent_lockout                 = false
      max_login_failures                = 30
      wait_increment_seconds            = 60
      quick_login_check_milli_seconds   = 1000
      minimum_quick_login_wait_seconds  = 60
      max_failure_wait_seconds          = 900
      failure_reset_time_seconds        = 43200
    }
  }

  web_authn_policy {
    relying_party_entity_name = "Example"
    relying_party_id          = "keycloak.example.com"
    signature_algorithms      = ["ES256", "RS256"]
  }
}
```

## Argument Reference

- `realm` - (Required) The name of the realm. This is unique across Keycloak. This will also be used as the realm's internal ID within Keycloak.
- `enabled` - (Optional) When `false`, users and clients will not be able to access this realm. Defaults to `true`.
- `display_name` - (Optional) The display name for the realm that is shown when logging in to the admin console.
- `display_name_html` - (Optional) The display name for the realm that is rendered as HTML on the screen when logging in to the admin console.
- `user_managed_access` - (Optional) When `true`, users are allowed to manage their own resources. Defaults to `false`.
- `organizations_enabled` - (Optional) When `true`, organization support is enabled. Defaults to `false`.
- `attributes` - (Optional) A map of custom attributes to add to the realm.
- `internal_id` - (Optional) When specified, this will be used as the realm's internal ID within Keycloak. When not specified, the realm's internal ID will be set to the realm's name.

### Login Settings

The following arguments are all booleans, and can be found in the "Login" tab within the realm settings.
If any of these arguments are not specified, they will default to Keycloak's default settings.

- `registration_allowed` - (Optional) When true, user registration will be enabled, and a link for registration will be displayed on the login page.
- `registration_email_as_username` - (Optional) When true, the user's email will be used as their username during registration.
- `edit_username_allowed` - (Optional) When true, the username field is editable.
- `reset_password_allowed` - (Optional) When true, a "forgot password" link will be displayed on the login page.
- `remember_me` - (Optional) When true, a "remember me" checkbox will be displayed on the login page, and the user's session will not expire between browser restarts.
- `verify_email` - (Optional) When true, users are required to verify their email address after registration and after email address changes.
- `login_with_email_allowed` - (Optional) When true, users may log in with their email address.
- `duplicate_emails_allowed` - (Optional) When true, multiple users will be allowed to have the same email address. This argument must be set to `false` if `login_with_email_allowed` is set to `true`.
- `ssl_required` - (Optional) Can be one of following values: 'none, 'external' or 'all'

### Themes

The following arguments can be used to configure themes for the realm. Custom themes can be specified here.
If any of these arguments are not specified, they will default to Keycloak's default settings. Typically, the `keycloak` theme is used by default.

- `login_theme` - (Optional) Used for the login, forgot password, and registration pages.
- `account_theme` - (Optional) Used for account management pages.
- `admin_theme` - (Optional) Used for the admin console.
- `email_theme` - (Optional) Used for emails that are sent by Keycloak.

### Tokens

The following arguments can be found in the "Tokens" tab within the realm settings. Each of these settings are top level arguments for the `keycloak_realm` resource.

- `default_signature_algorithm` - (Optional) Default algorithm used to sign tokens for the realm.
- `revoke_refresh_token` - (Optional) If enabled a refresh token can only be used number of times specified in 'refresh_token_max_reuse' before they are revoked. If unspecified, refresh tokens can be reused.
- `refresh_token_max_reuse` - (Optional) Maximum number of times a refresh token can be reused before they are revoked. If unspecified and 'revoke_refresh_token' is enabled the default value is 0 and refresh tokens can not be reused.

The arguments below should be specified as [Go duration strings](https://golang.org/pkg/time/#Duration.String). They will default to Keycloak's default settings.

- `sso_session_idle_timeout` - (Optional) The amount of time a session can be idle before it expires.
- `sso_session_max_lifespan` - (Optional) The maximum amount of time before a session expires regardless of activity.
- `sso_session_idle_timeout_remember_me` - (Optional) Similar to `sso_session_idle_timeout`, but used when a user clicks "Remember Me". If not set, Keycloak will default to the value of `sso_session_idle_timeout`.
- `sso_session_max_lifespan_remember_me` - (Optional) Similar to `sso_session_max_lifespan`, but used when a user clicks "Remember Me". If not set, Keycloak will default to the value of `sso_session_max_lifespan`.
- `offline_session_idle_timeout` - (Optional) The amount of time an offline session can be idle before it expires.
- `offline_session_max_lifespan` - (Optional) The maximum amount of time before an offline session expires regardless of activity.
- `offline_session_max_lifespan_enabled` - (Optional) Enable `offline_session_max_lifespan`.
- `client_session_idle_timeout` - (Optional) The amount of time a session can be idle before it expires. Users can override it for individual clients.
- `client_session_max_lifespan` - (Optional) The maximum amount of time before a session expires regardless of activity. Users can override it for individual clients.
- `access_token_lifespan` - (Optional) The amount of time an access token can be used before it expires.
- `access_token_lifespan_for_implicit_flow` - (Optional) The amount of time an access token issued with the OpenID Connect Implicit Flow can be used before it expires.
- `access_code_lifespan` - (Optional) The maximum amount of time a client has to finish the authorization code flow.
- `access_code_lifespan_login` - (Optional) The maximum amount of time a user is permitted to stay on the login page before the authentication process must be restarted.
- `access_code_lifespan_user_action` - (Optional) The maximum amount of time a user has to complete login related actions, such as updating a password.
- `action_token_generated_by_user_lifespan` - (Optional) The maximum time a user has to use a user-generated permit before it expires.
- `action_token_generated_by_admin_lifespan` - (Optional) The maximum time a user has to use an admin-generated permit before it expires.
- `oauth2_device_code_lifespan` - (Optional) The maximum amount of time a client has to finish the device code flow before it expires.

The attributes below should be specified in seconds.

- `oauth2_device_polling_interval` - (Optional) The minimum amount of time in seconds that the client should wait between polling requests to the token endpoint.

### SMTP

The `smtp_server` block can be used to configure the realm's SMTP settings, which can be found in the "Email" tab in the GUI.
This block supports the following arguments:

- `host` - (Required) The host of the SMTP server.
- `port` - (Optional) The port of the SMTP server (defaults to 25).
- `from` - (Required) The email address for the sender.
- `from_display_name` - (Optional) The display name of the sender email address.
- `reply_to` - (Optional) The "reply to" email address.
- `reply_to_display_name` - (Optional) The display name of the "reply to" email address.
- `envelope_from` - (Optional) The email address uses for bounces.
- `starttls` - (Optional) When `true`, enables StartTLS. Defaults to `false`.
- `ssl` - (Optional) When `true`, enables SSL. Defaults to `false`.
- `auth` - (Optional) Enables authentication to the SMTP server.  This block supports the following arguments:
    - `username` - (Required) The SMTP server username.
    - `password` - (Required) The SMTP server password.

### Internationalization

Internationalization support can be configured by using the `internationalization` block, which supports the following arguments:

- `supported_locales` - (Required) A list of [ISO 639-1](https://en.wikipedia.org/wiki/List_of_ISO_639-1_codes) locale codes that the realm should support.
- `default_locale` - (Required) The locale to use by default. This locale code must be present within the `supported_locales` list.

### Security Defenses

The `security_defenses` argument can be used to configure the realm's security defenses via the `headers` and `brute_force_detection` sub-blocks.

The `headers` block supports the following arguments:

- `x_frame_options` - (Optional) Sets the x-frame-option, which can be used to prevent pages from being included by non-origin iframes. More information can be found in the [RFC7034](https://tools.ietf.org/html/rfc7034)
- `content_security_policy` - (Optional) Sets the Content Security Policy, which can be used for prevent pages from being included by non-origin iframes. More information can be found in the [W3C-CSP](https://www.w3.org/TR/CSP/) Abstract.
- `content_security_policy_report_only` - (Optional) Used for testing Content Security Policies.
- `x_content_type_options` - (Optional) Sets the X-Content-Type-Options, which can be used for prevent MIME-sniffing a response away from the declared content-type
- `x_robots_tag` - (Optional) Prevent pages from appearing in search engines.
- `x_xss_protection` - (Optional) This header configures the Cross-site scripting (XSS) filter in your browser.
- `strict_transport_security` - (Optional) The Script-Transport-Security HTTP header tells browsers to always use HTTPS.
- `referrer_policy` - (Optional) The Referrer-Policy HTTP header controls how much referrer information (sent with the Referer header) should be included with requests.

The `brute_force_detection` block supports the following arguments:

- `permanent_lockout` - (Optional) When `true`, this will lock the user permanently when the user exceeds the maximum login failures.
- `max_login_failures` - (Optional) How many failures before wait is triggered.
- `wait_increment_seconds` - (Optional) This represents the amount of time a user should be locked out when the login failure threshold has been met.
- `quick_login_check_milli_seconds` - (Optional) Configures the amount of time, in milliseconds, for consecutive failures to lock a user out.
- `minimum_quick_login_wait_seconds` - (Optional) How long to wait after a quick login failure.
- `max_failure_wait_seconds ` - (Optional) Max. time a user will be locked out.
- `failure_reset_time_seconds` - (Optional) When will failure count be reset?

### Authentication Settings

The following authentication settings can also be configured. Note that these are top level arguments for the `keycloak_realm` resource.

- `password_policy` - (Optional) The password policy for users within the realm.

The arguments below can be used to configure authentication flow bindings:

- `browser_flow` - (Optional) The desired flow for browser authentication. Defaults to `browser`.
- `registration_flow` - (Optional) The desired flow for user registration. Defaults to `registration`.
- `direct_grant_flow` - (Optional) The desired flow for direct access authentication. Defaults to `direct grant`.
- `reset_credentials_flow` - (Optional) The desired flow to use when a user attempts to reset their credentials. Defaults to `reset credentials`.
- `client_authentication_flow` - (Optional) The desired flow for client authentication. Defaults to `clients`.
- `docker_authentication_flow` - (Optional) The desired flow for Docker authentication. Defaults to `docker auth`.
- `first_broker_login_flow` - (Optional) The desired flow for First Broker Login (since Keycloak 24). Defaults to `first broker login`.

### OTP Policy

The `otp_policy` block with following arguments can be found in the "OTP Policy" tab within the realm settings.

- `type` - (Optional) One Time Password Type, supported Values are `totp` for Time-Based One Time Password and `hotp` for Counter Based. Defaults to `totp`.
- `algorithm` - (Optional) What hashing algorithm should be used to generate the OTP, Valid options are `HmacSHA1`,`HmacSHA256` and `HmacSHA512`. Defaults to `HmacSHA1`.
- `digits` - (Optional) How many digits the OTP have. Defaults to `6`.
- `initial_counter` - (Optional) What should the initial counter value be. Defaults to `2`.
- `look_ahead_window` - (Optional) How far ahead should the server look just in case the token generator and server are out of time sync or counter sync. Defaults to `1`.
- `period` - (Optional) How many seconds should an OTP token be valid. Defaults to `30`.

### WebAuthn

The following settings can be used to modify the "WebAuthn Policy" and "WebAuthn Passwordless Policy" settings found within
the "Authentication" section of the realm configuration UI. These top level attributes can be used:

- `web_authn_policy` - (Optional) Configuration for WebAuthn Policy authentication.
- `web_authn_passwordless_policy` - (Optional) Configuration for WebAuthn Passwordless Policy authentication.

Each of these attributes are blocks with the following attributes:

- `relying_party_entity_name` - (Optional) A human-readable server name for the WebAuthn Relying Party. Defaults to `keycloak`.
- `relying_party_id` - (Optional) The WebAuthn relying party ID.
- `signature_algorithms` - (Optional) A set of signature algorithms that should be used for the authentication assertion. Valid options at the time these docs were written are `ES256`, `ES384`, `ES512`, `RS256`, `RS384`, `RS512`, and `RS1`.
- `attestation_conveyance_preference` - (Optional) The preference of how to generate a WebAuthn attestation statement. Valid options are `not specified`, `none`, `indirect`, `direct`, or `enterprise`. Defaults to `not specified`.
- `authenticator_attachment` - (Optional) The acceptable attachment pattern for the WebAuthn authenticator. Valid options are `not specified`, `platform`, or `cross-platform`. Defaults to `not specified`.
- `require_resident_key` - (Optional) Specifies whether a public key should be created to represent the resident key. Valid options are `not specified`, `Yes`, or `No`. Defaults to `not specified`.
- `user_verification_requirement` - (Optional) Specifies the policy for verifying a user logging in via WebAuthn. Valid options are `not specified`, `required`, `preferred`, or `discouraged`. Defaults to `not specified`.
- `create_timeout` - (Optional) The timeout value for creating a user's public key credential in seconds. When set to `0`, this timeout option is not adapted. Defaults to `0`.
- `avoid_same_authenticator_register` - (Optional) When `true`, Keycloak will avoid registering the authenticator for WebAuthn if it has already been registered. Defaults to `false`.
- `acceptable_aaguids` - (Optional) A set of AAGUIDs for which an authenticator can be registered.
- `extra_origins` - (Optional) A set of extra origins for non-web applications.

## Default Client Scopes

- `default_default_client_scopes` - (Optional) A list of default `default client scopes` to be used for client definitions. Defaults to `[]` or keycloak's built-in default `default client-scopes`. For an alternative, please refer to the dedicated resource `keycloak_realm_default_client_scopes`.
- `default_optional_client_scopes` - (Optional) A list of default `optional client scopes` to be used for client definitions. Defaults to `[]` or keycloak's built-in default `optional client-scopes`. For an alternative, please refer to the dedicated resource `keycloak_realm_optional_client_scopes`.

## Import

Realms can be imported using their name.

Example:

```bash
$ terraform import keycloak_realm.realm my-realm
```
