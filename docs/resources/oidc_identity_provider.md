---
page_title: "keycloak_oidc_identity_provider Resource"
---

# keycloak\_oidc\_identity\_provider Resource

Allows for creating and managing OIDC Identity Providers within Keycloak.

OIDC (OpenID Connect) identity providers allows users to authenticate through a third party system using the OIDC standard.

> **NOTICE:** This resource now supports [write-only arguments](https://developer.hashicorp.com/terraform/language/resources/ephemeral#write-only-arguments)
> for client secret via the new arguments `client_secret_wo` and `client_secret_wo_version`. Using write-only arguments
> prevents sensitive values from being stored in plan and state files. You cannot use `client_secret_wo` and
> `client_secret_wo_version` alongside `client_secret` as this will result in a validation error due to conflicts.
>
> For backward compatibility, the behavior of the original `client_secret` argument remains unchanged.


## Example Usage

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_oidc_identity_provider" "realm_identity_provider" {
  realm             = keycloak_realm.realm.id
  alias             = "my-idp"
  authorization_url = "https://authorizationurl.com"
  client_id         = "clientID"
  client_secret     = "clientSecret"
  token_url         = "https://tokenurl.com"

  extra_config = {
    "clientAuthMethod" = "client_secret_post"
  }
}
```

## Example Usage with `client_secret_wo`

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

ephemeral "random_password" "openid_client_secret" {
  length           = 16
  override_special = "!#$%&*()-_=+[]{}<>:?"
}

resource "keycloak_oidc_identity_provider" "realm_identity_provider" {
  realm                    = keycloak_realm.realm.id
  alias                    = "my-idp"
  authorization_url        = "https://authorizationurl.com"
  client_id                = "clientID"
  client_secret_wo         = ephemeral.random_password.openid_client_secret.result
  client_secret_wo_version = 1
  token_url                = "https://tokenurl.com"

  extra_config = {
    "clientAuthMethod" = "client_secret_post"
  }
}
```

## Argument Reference

- `realm` - (Required) The name of the realm. This is unique across Keycloak.
- `alias` - (Required) The alias uniquely identifies an identity provider, and it is also used to build the redirect uri.
- `authorization_url` - (Required) The Authorization Url.
- `client_id` - (Required) The client or client identifier registered within the identity provider.
- `client_secret` - (Optional) The client or client secret registered within the identity provider. This field is able to obtain its value from vault, use $${vault.ID} format. Required without `client_secret_wo` and `client_secret_wo_version`.
- `client_secret_wo` - (Optional, Write-Only) The secret for clients with an `access_type` of `CONFIDENTIAL` or `BEARER-ONLY`. This is a write-only argument and Terraform does not store them in state or plan files. If omitted, this will fallback to use `client_secret`.
- `client_secret_wo_version` - (Optional) Functions as a flag and/or trigger to indicate Terraform when to use the input value in `client_secret_wo` to execute a Create or Update operation. The value of this argument is stored in the state and plan files. Required when using `client_secret_wo`.
- `token_url` - (Required) The Token URL.
- `display_name` - (Optional) Display name for the identity provider in the GUI.
- `enabled` - (Optional) When `true`, users will be able to log in to this realm using this identity provider. Defaults to `true`.
- `store_token` - (Optional) When `true`, tokens will be stored after authenticating users. Defaults to `true`.
- `add_read_token_role_on_create` - (Optional) When `true`, new users will be able to read stored tokens. This will automatically assign the `broker.read-token` role. Defaults to `false`.
- `link_only` - (Optional) When `true`, users cannot sign-in using this provider, but their existing accounts will be linked when possible. Defaults to `false`.
- `trust_email` - (Optional) When `true`, email addresses for users in this provider will automatically be verified regardless of the realm's email verification policy. Defaults to `false`.
- `first_broker_login_flow_alias` - (Optional) The authentication flow to use when users log in for the first time through this identity provider. Defaults to `first broker login`.
- `post_broker_login_flow_alias` - (Optional) The authentication flow to use after users have successfully logged in, which can be used to perform additional user verification (such as OTP checking). Defaults to an empty string, which means no post login flow will be used.
- `provider_id` - (Optional) The ID of the identity provider to use. Defaults to `oidc`, which should be used unless you have extended Keycloak and provided your own implementation.
- `backchannel_supported` - (Optional) Does the external IDP support backchannel logout? Defaults to `true`.
- `validate_signature` - (Optional) Enable/disable signature validation of external IDP signatures. Defaults to `false`.
- `user_info_url` - (Optional) User Info URL.
- `jwks_url` - (Optional) JSON Web Key Set URL.
- `issuer` - (Optional) The issuer identifier for the issuer of the response. If not provided, no validation will be performed.
- `disable_user_info` - (Optional) When `true`, disables the usage of the user info service to obtain additional user information. Defaults to `false`.
- `hide_on_login_page` - (Optional) When `true`, this provider will be hidden on the login page, and is only accessible when requested explicitly. Defaults to `false`.
- `disable_type_claim_check` - (Optional) When `true`, disables the check for the `typ` claim of tokens received from the identity provider. Defaults to `false`.
- `logout_url` - (Optional) The Logout URL is the end session endpoint to use to sign-out the user from external identity provider.
- `login_hint` - (Optional) Pass login hint to identity provider.
- `ui_locales` - (Optional) Pass current locale to identity provider. Defaults to `false`.
- `accepts_prompt_none_forward_from_client` (Optional) When `true`, the IDP will accept forwarded authentication requests that contain the `prompt=none` query parameter. Defaults to `false`.
- `default_scopes` - (Optional) The scopes to be sent when asking for authorization. It can be a space-separated list of scopes. Defaults to `openid`.
- `organization_id` - (Optional) The ID of the organization to link this identity provider to.
- `org_domain` - (Optional) The organization domain to associate this identity provider with. it is used to map users to an organization based on their email domain and to authenticate them accordingly in the scope of the organization.
- `org_redirect_mode_email_matches` - (Optional) Indicates whether to automatically redirect user to this identity provider when email domain matches domain.
- `sync_mode` - (Optional) The default sync mode to use for all mappers attached to this identity provider. Can be once of `IMPORT`, `FORCE`, or `LEGACY`.
- `gui_order` - (Optional) A number defining the order of this identity provider in the GUI.
- `extra_config` - (Optional) A map of key/value pairs to add extra configuration to this identity provider. This can be used for custom oidc provider implementations, or to add configuration that is not yet supported by this Terraform provider. Use this attribute at your own risk, as custom attributes may conflict with top-level configuration attributes in future provider updates.
    - `clientAuthMethod` (Optional) The client authentication method. Since Keycloak 8, this is a required attribute if OIDC provider is created using the Keycloak GUI. It accepts the values `client_secret_post` (Client secret sent as post), `client_secret_basic` (Client secret sent as basic auth), `client_secret_jwt` (Client secret as jwt) and `private_key_jwt ` (JTW signed with private key)

## Attribute Reference

- `internal_id` - (Computed) The unique ID that Keycloak assigns to the identity provider upon creation.

## Import

Identity providers can be imported using the format `{{realm_id}}/{{idp_alias}}`, where `idp_alias` is the identity provider alias.

Example:

```bash
$ terraform import keycloak_oidc_identity_provider.realm_identity_provider my-realm/my-idp
```
