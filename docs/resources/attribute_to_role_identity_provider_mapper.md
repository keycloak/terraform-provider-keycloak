---
page_title: "keycloak_attribute_to_role_identity_provider_mapper Resource"
---

# keycloak\_attribute\_to\_role\_identity\_provider\_mapper Resource

Allows for creating and managing an attribute to role identity provider mapper within Keycloak.

~> If you are using Keycloak 10 or higher, you will need to specify the `extra_config` argument in order to define a `syncMode` for the mapper.

## Example Usage

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_oidc_identity_provider" "oidc" {
  realm             = keycloak_realm.realm.id
  alias             = "oidc"
  authorization_url = "https://example.com/auth"
  token_url         = "https://example.com/token"
  client_id         = "example_id"
  client_secret     = "example_token"
  default_scopes    = "openid random profile"
}

resource "keycloak_role" "realm_role" {
  realm_id    = keycloak_realm.realm.id
  name        = "my-realm-role"
  description = "My Realm Role"
}

resource "keycloak_attribute_to_role_identity_provider_mapper" "oidc" {
  realm                   = keycloak_realm.realm.id
  name                    = "role-attribute"
  identity_provider_alias = keycloak_oidc_identity_provider.oidc.alias
  role                    = "my-realm-role"
  claim_name              = "my-claim"
  claim_value             = "my-value"

  # extra_config with syncMode is required in Keycloak 10+
  extra_config = {
    syncMode = "INHERIT"
  }
}
```

## Argument Reference

The following arguments are supported:

- `realm` - (Required) The name of the realm.
- `name` - (Required) The name of the mapper.
- `identity_provider_alias` - (Required) The alias of the associated identity provider.
- `role` - (Required) Role Name.
- `attribute_name` - (Optional) Attribute Name.
- `attribute_friendly_name` - (Optional) Attribute Friendly Name. Conflicts with `attribute_name`.
- `attribute_value` - (Optional) Attribute Value.
- `claim_name` - (Optional) OIDC Claim Name
- `claim_value` - (Optional) OIDC Claim Value
- `extra_config` - (Optional) Key/value attributes to add to the identity provider mapper model that is persisted to Keycloak. This can be used to extend the base model with new Keycloak features.

## Import

Identity provider mappers can be imported using the format `{{realm_id}}/{{idp_alias}}/{{idp_mapper_id}}`, where `idp_alias` is the identity provider alias, and `idp_mapper_id` is the unique ID that Keycloak
assigns to the mapper upon creation. This value can be found in the URI when editing this mapper in the GUI, and is typically a GUID.

Example:

```bash
$ terraform import keycloak_attribute_to_role_identity_provider_mapper.test_mapper my-realm/my-mapper/f446db98-7133-4e30-b18a-3d28fde7ca1b
```
