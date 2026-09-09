---
page_title: "keycloak_openid_default_default_client_scopes Resource"
---

# keycloak\_openid\_default\_default\_client\_scopes Resource

Allows for managing the realm-level default client scopes that Keycloak attaches as **default** scopes to every newly
created client. This is the realm equivalent of `keycloak_openid_client_default_scopes`, which works at the individual
client level.

The resource is **authoritative** over the realm's default-default scopes: scopes attached manually outside Terraform
will be detached on the next apply, and scopes listed here that are missing from the realm will be added.

This resource maps to the Keycloak Admin REST API endpoint
`/admin/realms/{realm}/default-default-client-scopes`.

## Example Usage

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_openid_client_scope" "extra_scope" {
  realm_id = keycloak_realm.realm.id
  name     = "extra-scope"
}

resource "keycloak_openid_default_default_client_scopes" "realm_defaults" {
  realm_id = keycloak_realm.realm.id

  default_scopes = [
    "profile",
    "email",
    "roles",
    "web-origins",
    keycloak_openid_client_scope.extra_scope.name,
  ]
}
```

## Argument Reference

- `realm_id` - (Required) The realm to manage default-default client scopes for.
- `default_scopes` - (Required) An ordered list of client scope names that should be configured as default scopes for the realm.

## Import

This resource can be imported using the realm id.

```bash
terraform import keycloak_openid_default_default_client_scopes.realm_defaults my-realm
```
