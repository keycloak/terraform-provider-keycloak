---
page_title: "keycloak_custom_identity_provider_mapper Resource"
---

# keycloak\_custom\_identity\_provider\_mapper Resource

Allows for creating and managing custom identity provider mapper within Keycloak.

The custom identity provider mapper can be used to define custom mapper type for the imported Keycloak user. This can be
useful for extending an existing Keycloak mapper with additional config that is not supported by this provider, or when
configuring a custom identity provider mapper via Terraform.

~> If you are using Keycloak 10 or higher, you will need to specify the `extra_config` argument in order to define a `syncMode` for the mapper.

## Example Usage

Following is provisioning a user attribute importer mapper through the extra_config. Name of the keys can be verified in keycloak source code:
* [user.attribute](https://github.com/keycloak/keycloak/blob/44956e10d00b79b0000583ed964130712abe5353/services/src/main/java/org/keycloak/broker/oidc/mappers/UserAttributeMapper.java#L52)
* [claim](https://github.com/keycloak/keycloak/blob/44956e10d00b79b0000583ed964130712abe5353/services/src/main/java/org/keycloak/broker/oidc/mappers/AbstractClaimMapper.java#L41)

For a strongly typed alternative, use [`keycloak_attribute_importer_identity_provider_mapper`](attribute_importer_identity_provider_mapper.md), which exposes `claim_name` and `user_attribute` directly.

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

resource "keycloak_custom_identity_provider_mapper" "oidc" {
  realm                    = keycloak_realm.realm.id
  name                     = "email-attribute-importer"
  identity_provider_alias  = keycloak_oidc_identity_provider.oidc.alias
  identity_provider_mapper = "%s-user-attribute-idp-mapper"

  # extra_config with syncMode is required in Keycloak 10+
  extra_config = {
    syncMode      = "INHERIT"
    claim         = "my-email-claim"
    "user.attribute" = "email"
  }
}
```

```hcl
resource "keycloak_custom_identity_provider_mapper" "groups" {
  realm                    = keycloak_realm.realm.id
  name                     = "group-mapper
  identity_provider_alias  = keycloak_oidc_identity_provider.oidc.alias
  identity_provider_mapper = "oidc-advanced-group-idp-mapper"

  extra_config = {
    "are.claim.values.regex" = true
    syncMode                 = "FORCE"
    group                    = "/DevOps"
    claims = jsonencode(
      [{
        key = "groups", value = "DevOps"
      }]
    )
  }
}
```

## Argument Reference

The following arguments are supported:

- `realm` - (Required) The name of the realm.
- `name` - (Required) The name of the mapper.
- `identity_provider_alias` - (Required) The alias of the associated identity provider.
- `identity_provider_mapper` - (Required) The type of the identity provider mapper. This can be a format string that includes a `%s` - this will be replaced by the provider id.
- `extra_config` - (Optional) Key/value attributes to add to the identity provider mapper model that is persisted to Keycloak. This can be used to extend the base model with new Keycloak features.

## Import

Identity provider mappers can be imported using the format `{{realm_id}}/{{idp_alias}}/{{idp_mapper_id}}`, where `idp_alias` is the identity provider alias, and `idp_mapper_id` is the unique ID that Keycloak
assigns to the mapper upon creation. This value can be found in the URI when editing this mapper in the GUI, and is typically a GUID.

Example:

```bash
$ terraform import keycloak_custom_identity_provider_mapper.test_mapper my-realm/my-mapper/f446db98-7133-4e30-b18a-3d28fde7ca1b
```
