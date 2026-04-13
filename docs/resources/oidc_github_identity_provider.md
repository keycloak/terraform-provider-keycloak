---
page_title: "keycloak_oidc_github_identity_provider Resource"
---

# keycloak\_oidc\_github\_identity\_provider Resource

Allows for creating and managing **GitHub**-based OIDC Identity Providers within Keycloak.

OIDC (OpenID Connect) identity providers allows users to authenticate through a third party system using the OIDC standard.

The GitHub variant is specialized for the public GitHub instance (github.com) or GitHub Enterprise deployments.

For example, it will obtain automatically the primary email from the logged in account.

## Example Usage

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_oidc_github_identity_provider" "github" {
  realm         = keycloak_realm.realm.id
  client_id     = var.github_identity_provider_client_id
  client_secret = var.github_identity_provider_client_secret
  trust_email   = true
  sync_mode     = "IMPORT"

  extra_config = {
    "myCustomConfigKey" = "myValue"
  }
}
```

## Argument Reference

- `realm` - (Required) The name of the realm. This is unique across Keycloak.
- `client_id` - (Required) The client or client identifier registered within the identity provider.
- `client_secret` - (Required) The client or client secret registered within the identity provider. This field is able to obtain its value from vault, use $${vault.ID} format.
- `alias` - (Optional) The alias for the GitHub identity provider.
- `display_name` - (Optional) Display name for the GitHub identity provider in the GUI.
- `enabled` - (Optional) When `true`, users will be able to log in to this realm using this identity provider. Defaults to `true`.
- `store_token` - (Optional) When `true`, tokens will be stored after authenticating users. Defaults to `true`.
- `add_read_token_role_on_create` - (Optional) When `true`, new users will be able to read stored tokens. This will automatically assign the `broker.read-token` role. Defaults to `false`.
- `link_only` - (Optional) When `true`, users cannot sign-in using this provider, but their existing accounts will be linked when possible. Defaults to `false`.
- `trust_email` - (Optional) When `true`, email addresses for users in this provider will automatically be verified regardless of the realm's email verification policy. Defaults to `false`.
- `first_broker_login_flow_alias` - (Optional) The authentication flow to use when users log in for the first time through this identity provider. Defaults to `first broker login`.
- `post_broker_login_flow_alias` - (Optional) The authentication flow to use after users have successfully logged in, which can be used to perform additional user verification (such as OTP checking). Defaults to an empty string, which means no post login flow will be used.
- `provider_id` - (Optional) The ID of the identity provider to use. Defaults to `github`, which should be used unless you have extended Keycloak and provided your own implementation.
- `base_url` - (Optional) The GitHub base URL, defaults to `https://github.com`
- `api_url` - (Optional) The GitHub API URL, defaults to `https://api.github.com`.
- `github_json_format` (Optional) When `true`, GitHub API is told explicitly to accept JSON during token authentication requests. Defaults to `false`.
- `default_scopes` - (Optional) The scopes to be sent when asking for authorization. It can be a space-separated list of scopes. Defaults to `user:email`.
- `disable_user_info` - (Optional) When `true`, disables the usage of the user info service to obtain additional user information. Defaults to `false`.
- `hide_on_login_page` - (Optional) When `true`, this identity provider will be hidden on the login page. Defaults to `false`.
- `sync_mode` - (Optional) The default sync mode to use for all mappers attached to this identity provider. Can be once of `IMPORT`, `FORCE`, or `LEGACY`.
- `gui_order` - (Optional) A number defining the order of this identity provider in the GUI.
- `extra_config` - (Optional) A map of key/value pairs to add extra configuration to this identity provider. This can be used for custom oidc provider implementations, or to add configuration that is not yet supported by this Terraform provider. Use this attribute at your own risk, as custom attributes may conflict with top-level configuration attributes in future provider updates.

## Attribute Reference

- `internal_id` - (Computed) The unique ID that Keycloak assigns to the identity provider upon creation.

## Import

GitHub Identity providers can be imported using the format {{realm_id}}/{{idp_alias}}, where idp_alias is the identity provider alias.

Example:

```bash
$ terraform import keycloak_oidc_github_identity_provider.github.github_identity_provider my-realm/my-github-idp
```
