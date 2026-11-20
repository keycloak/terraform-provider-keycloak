---
page_title: "keycloak_hardcoded_attribute_identity_provider_mapper Resource"
---

# keycloak_hardcoded_attribute_identity_provider_mapper Resource

Allows for creating and managing hardcoded attribute mappers for Keycloak identity provider.

The identity provider hardcoded attribute mapper will set the specified value to the IDP attribute.

## Example Usage

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_oidc_identity_provider" "oidc" {
  realm             = keycloak_realm.realm.id
  alias             = "my-idp"
  authorization_url = "https://authorizationurl.com"
  client_id         = "clientID"
  client_secret     = "clientSecret"
  token_url         = "https://tokenurl.com"
}

resource "keycloak_hardcoded_attribute_identity_provider_mapper" "oidc" {
  realm                   = keycloak_realm.realm.id
  name                    = "hardcodedUserSessionAttribute"
  identity_provider_alias = keycloak_oidc_identity_provider.oidc.alias
  attribute_name          = "attribute"
  attribute_value         = "value"
  user_session            = true

  extra_config = {
    syncMode = "INHERIT"
  }
}
```

## Argument Reference

- `realm` - (Required) The realm ID that this mapper will exist in.
- `name` - (Required) Display name of this mapper when displayed in the console.
- `identity_provider_alias` - (Required) The IDP alias of the attribute to set.
- `attribute_name` - (Required) The name of the IDP attribute to set.
- `attribute_value` - (Optional) The value to set to the attribute. You can hardcode any value like 'foo'.
- `user_session` - (Required) Is Attribute related to a User Session.
- `extra_config` - (Optional) A map of key/value pairs to add extra configuration attributes to this mapper. This can be used for custom attributes, or to add configuration attributes that are not yet supported by this Terraform provider. Use this attribute at your own risk, as it may conflict with top-level configuration attributes in future provider updates.
