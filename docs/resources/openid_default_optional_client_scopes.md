---
page_title: "keycloak_openid_default_optional_client_scopes Resource"
---

# keycloak\_openid\_default\_optional\_client\_scopes Resource

Allows for managing the realm-level optional client scopes that Keycloak attaches as **optional** scopes to every newly
created client. This is the realm equivalent of `keycloak_openid_client_optional_scopes`, which works at the individual
client level.

The resource is **authoritative** over the realm's default-optional scopes: scopes attached manually outside Terraform
will be detached on the next apply, and scopes listed here that are missing from the realm will be added.

This resource maps to the Keycloak Admin REST API endpoint
`/admin/realms/{realm}/default-optional-client-scopes`.

## Example Usage

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_openid_client_scope" "extra_optional" {
  realm_id = keycloak_realm.realm.id
  name     = "extra-optional"
}

resource "keycloak_openid_default_optional_client_scopes" "realm_optionals" {
  realm_id = keycloak_realm.realm.id

  optional_scopes = [
    "address",
    "phone",
    "offline_access",
    "microprofile-jwt",
    keycloak_openid_client_scope.extra_optional.name,
  ]
}
```

## Argument Reference

- `realm_id` - (Required) The realm to manage default-optional client scopes for.
- `optional_scopes` - (Required) An ordered list of client scope names that should be configured as optional scopes for the realm.

## Import

This resource can be imported using the realm id.

```bash
terraform import keycloak_openid_default_optional_client_scopes.realm_optionals my-realm
```
