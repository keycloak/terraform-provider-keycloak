---
page_title: "keycloak_saml_identity_provider Resource"
---

# keycloak\_saml\_identity\_provider Resource

Allows for creating and managing SAML Identity Providers within Keycloak.

SAML (Security Assertion Markup Language) identity providers allows users to authenticate through a third-party system using the SAML protocol.

## Example Usage

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_saml_identity_provider" "realm_saml_identity_provider" {
  realm = keycloak_realm.realm.id
  alias = "my-saml-idp"

  entity_id                  = "https://domain.com/entity_id"
  single_sign_on_service_url = "https://domain.com/adfs/ls/"
  single_logout_service_url  = "https://domain.com/adfs/ls/?wa=wsignout1.0"

  backchannel_supported      = true
  post_binding_response      = true
  post_binding_logout        = true
  post_binding_authn_request = true
  store_token                = false
  trust_email                = true
  force_authn                = true
}
```

## Argument Reference

- `realm` - (Required) The name of the realm. This is unique across Keycloak.
- `alias` - (Optional) The unique name of identity provider.
- `enabled` - (Optional) When `false`, users and clients will not be able to access this realm. Defaults to `true`.
- `display_name` - (Optional) The display name for the realm that is shown when logging in to the admin console.
- `store_token` - (Optional) When `true`, tokens will be stored after authenticating users. Defaults to `true`.
- `add_read_token_role_on_create` - (Optional) When `true`, new users will be able to read stored tokens. This will automatically assign the `broker.read-token` role. Defaults to `false`.
- `trust_email` - (Optional) When `true`, email addresses for users in this provider will automatically be verified regardless of the realm's email verification policy. Defaults to `false`.
- `link_only` - (Optional) When `true`, users cannot log in using this provider, but their existing accounts will be linked when possible. Defaults to `false`.
- `hide_on_login_page` - (Optional) If hidden, then login with this provider is possible only if requested explicitly, e.g. using the 'kc_idp_hint' parameter.
- `first_broker_login_flow_alias` - (Optional) Alias of authentication flow, which is triggered after first login with this identity provider. Term 'First Login' means that there is not yet existing Keycloak account linked with the authenticated identity provider account. Defaults to `first broker login`.
- `post_broker_login_flow_alias` - (Optional) Alias of authentication flow, which is triggered after each login with this identity provider. Useful if you want additional verification of each user authenticated with this identity provider (for example OTP). Leave this empty if you don't want any additional authenticators to be triggered after login with this identity provider. Also note, that authenticator implementations must assume that user is already set in ClientSession as identity provider already set it. Defaults to empty.
- `authenticate_by_default` - (Optional) Authenticate users by default. Defaults to `false`.
- `entity_id` - (Required) The Entity ID that will be used to uniquely identify this SAML Service Provider.
- `single_sign_on_service_url` - (Required) The Url that must be used to send authentication requests (SAML AuthnRequest).
- `single_logout_service_url` - (Optional) The Url that must be used to send logout requests.
- `backchannel_supported` - (Optional) Does the external IDP support backchannel logout?. Defaults to `false`.
- `provider_id` - (Optional) The ID of the identity provider to use. Defaults to `saml`, which should be used unless you have extended Keycloak and provided your own implementation.
- `name_id_policy_format` - (Optional) Specifies the URI reference corresponding to a name identifier format. Defaults to empty.
- `post_binding_response` - (Optional) Indicates whether to respond to requests using HTTP-POST binding. If false, HTTP-REDIRECT binding will be used.
- `post_binding_authn_request` - (Optional) Indicates whether the AuthnRequest must be sent using HTTP-POST binding. If false, HTTP-REDIRECT binding will be used.
- `post_binding_logout` - (Optional) Indicates whether to respond to requests using HTTP-POST binding. If false, HTTP-REDIRECT binding will be used.
- `want_assertions_signed` - (Optional) Indicates whether this service provider expects a signed Assertion.
- `want_assertions_encrypted` - (Optional) Indicates whether this service provider expects an encrypted Assertion.
- `force_authn` - (Optional) Indicates whether the identity provider must authenticate the presenter directly rather than rely on a previous security context.
- `validate_signature` - (Optional) Enable/disable signature validation of SAML responses.
- `signing_certificate` - (Optional) Signing Certificate.
- `signature_algorithm` - (Optional) Signing Algorithm. Defaults to empty.
- `xml_sign_key_info_key_name_transformer` - (Optional) The SAML signature key name. Can be one of `NONE`, `KEY_ID`, or `CERT_SUBJECT`.
- `sync_mode` - (Optional) The default sync mode to use for all mappers attached to this identity provider. Can be one of `IMPORT`, `FORCE`, or `LEGACY`.
- `gui_order` - (Optional) A number defining the order of this identity provider in the GUI.
- `principal_type` - (Optional) The principal type. Can be one of `SUBJECT`, `ATTRIBUTE` or `FRIENDLY_ATTRIBUTE`.
- `principal_attribute` - (Optional) The principal attribute.
- `authn_context_class_refs` - (Optional) Ordered list of requested AuthnContext ClassRefs.
- `authn_context_decl_refs` - (Optional) Ordered list of requested AuthnContext DeclRefs.
- `authn_context_comparison_type` - (Optional) Specifies the comparison method used to evaluate the requested context classes or statements.
- `extra_config` - (Optional) A map of key/value pairs to add extra configuration to this identity provider. This can be used for custom oidc provider implementations, or to add configuration that is not yet supported by this Terraform provider. Use this attribute at your own risk, as custom attributes may conflict with top-level configuration attributes in future provider updates.

## Import

Identity providers can be imported using the format `{{realm_id}}/{{idp_alias}}`, where `idp_alias` is the identity provider alias.

Example:

```bash
$ terraform import keycloak_saml_identity_provider.realm_saml_identity_provider my-realm/my-saml-idp
```
