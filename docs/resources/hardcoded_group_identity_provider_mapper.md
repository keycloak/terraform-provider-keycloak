---
page_title: "keycloak_hardcoded_group_identity_provider_mapper Resource"
---

# keycloak_hardcoded_group_identity_provider_mapper Resource

Allows for creating and managing hardcoded group mappers for Keycloak identity provider.

The identity provider hardcoded group mapper grants a specified Keycloak group to each Keycloak user from the identity provider.

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

resource "keycloak_group" "realm_group" {
  realm_id    = keycloak_realm.realm.id
  name        = "my-realm-group"
  description = "My Realm Group"
}

resource "keycloak_hardcoded_group_identity_provider_mapper" "oidc" {
  realm                   = keycloak_realm.realm.id
  name                    = "hardcodedGroup"
  identity_provider_alias = keycloak_oidc_identity_provider.oidc.alias
  group                   = "my-realm-group"

  extra_config = {
    syncMode = "INHERIT"
  }
}
```

## Argument Reference

- `realm` - (Required) The realm ID that this mapper will exist in.
- `name` - (Required) Display name of this mapper when displayed in the console.
- `identity_provider_alias` - (Required) The IDP alias of the attribute to set.
- `group` - (Optional) The name of the group which should be assigned to the users.
- `extra_config` - (Optional) A map of key/value pairs to add extra configuration attributes to this mapper. This can be used for custom attributes, or to add configuration attributes that are not yet supported by this Terraform provider. Use this attribute at your own risk, as it may conflict with top-level configuration attributes in future provider updates.
