---
page_title: "keycloak_oidc_identity_provider Resource"
---

# keycloak\_oidc\_identity\_provider Resource

Allows for creating and managing OIDC Identity Providers within Keycloak.

OIDC (OpenID Connect) identity providers allows users to authenticate through a third party system using the OIDC standard.

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

## Argument Reference

- `realm` - (Required) The name of the realm. This is unique across Keycloak.
- `alias` - (Required) The alias uniquely identifies an identity provider, and it is also used to build the redirect uri.
- `authorization_url` - (Required) The Authorization Url.
- `client_id` - (Required) The client or client identifier registered within the identity provider.
- `client_secret` - (Required) The client or client secret registered within the identity provider. This field is able to obtain its value from vault, use $${vault.ID} format.
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
- `logout_url` - (Optional) The Logout URL is the end session endpoint to use to sign-out the user from external identity provider.
- `login_hint` - (Optional) Pass login hint to identity provider.
- `ui_locales` - (Optional) Pass current locale to identity provider. Defaults to `false`.
- `accepts_prompt_none_forward_from_client` (Optional) When `true`, the IDP will accept forwarded authentication requests that contain the `prompt=none` query parameter. Defaults to `false`.
- `default_scopes` - (Optional) The scopes to be sent when asking for authorization. It can be a space-separated list of scopes. Defaults to `openid`.
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
