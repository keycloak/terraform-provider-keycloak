---
page_title: "keycloak_client_scope Resource"
---

# keycloak\_client\_scope Resource

Creates and manages a realm-owned Keycloak client scope for either the OpenID Connect or SAML protocol. Client scopes
share protocol mappers between clients. OpenID Connect clients can request optional scopes using the OAuth 2.0 `scope` parameter.

Protocol mappers remain managed by the existing protocol mapper resources and can be attached to this scope using its
`id`.

## Example Usage

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_client_scope" "groups" {
  realm_id            = keycloak_realm.realm.id
  name                = "groups"
  description         = "When requested, this scope will map a user's group memberships to a claim"
  consent_screen_text = "Your groups"
  gui_order           = 1

  openid_connect {
    include_in_token_scope              = true
    include_in_openid_provider_metadata = true
  }
}

resource "keycloak_client_scope" "saml_groups" {
  realm_id = keycloak_realm.realm.id
  name     = "saml-groups"

  saml {}
}
```

## Argument Reference

- `realm_id` - (Required) The realm this client scope belongs to.
- `name` - (Required) The display name of this client scope.
- `description` - (Optional) The description displayed in the Keycloak GUI.
- `consent_screen_text` - (Optional) When set, a consent screen will be displayed to users authenticating to clients with
  this scope attached. The consent screen will display the string value of this attribute.
- `display_on_consent_screen` - (Optional) When `true`, this scope is shown on the consent screen. When not set, it defaults
  to `true` if `consent_screen_text` is set and `false` otherwise.
- `gui_order` - (Optional) The order of this client scope in the GUI, such as on the consent page.
- `openid_connect` - (Optional) Creates an OpenID Connect scope. Exactly one of `openid_connect` or `saml` must be set, and
  switching between them replaces the scope. Supports the following arguments:
  - `include_in_token_scope` - (Optional) When `true`, the name of this client scope will be added to the access token
    property `scope` as well as to the Token Introspection Endpoint response. Defaults to `true`.
  - `include_in_openid_provider_metadata` - (Optional) When `true`, this client scope will be listed in the
    `scopes_supported` field of the realm's OpenID Provider Metadata (the `.well-known/openid-configuration` discovery
    document). Defaults to `true`.
- `saml` - (Optional) Creates a SAML scope. Set it as an empty block, `saml {}`; it currently has no arguments.
- `extra_config` - (Optional) A map of key/value pairs to add extra configuration attributes to this client scope. This can
  be used for custom attributes or attributes not yet supported by this provider, such as preview features. Keys managed
  by the arguments above are rejected, as are empty values; remove a key to clear it. Only keys set in configuration are
  tracked, so keys added outside Terraform are ignored.

Replacing a scope without changing its name removes its assignments and may not trigger an update to resources that
reference only its name. See [Replacing scope definitions](client_scope_assignments.md#replacing-scope-definitions)
for lifecycle configuration and recovery guidance.

## Migrating protocol-specific scopes

Replace `keycloak_openid_client_scope` with this resource and an `openid_connect` block, moving
`include_in_token_scope` and `include_in_openid_provider_metadata` into that block. Replace `keycloak_saml_client_scope`
with this resource and an empty `saml {}` block. All other arguments, including `extra_config`, keep their names and
behavior.

## Import

Import using `realm/<realm_id>/<client_scope_id>`, where `client_scope_id` is the Keycloak-generated scope ID.

```shell
terraform import keycloak_client_scope.groups realm/my-realm/8e8f7fe1-df9b-40ed-bed3-4597aa0dac52
```

When migrating from a deprecated protocol-specific scope resource, remove that resource from configuration and state
with `terraform state rm` before importing the existing scope here. Use the `realm/<realm_id>/<client_scope_id>` import
format above and review the plan before applying.
