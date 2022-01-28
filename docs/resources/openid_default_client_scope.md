---
page_title: "keycloak_openid_default_client_scope Resource"
---

# keycloak\_openid\_default\_client\_scope Resource

Manages a **realm-wide default client scope** in Keycloak. When a default client scope is assigned at the realm level, it is automatically applied to all new clients in the realm that use the OpenID Connect protocol. The protocol mappers defined within the scope are included by default in the claims for all clients, regardless of the provided OAuth2.0 `scope` parameter.

!> Using `keycloak_openid_default_client_scope` will conflict with `keycloak_realm.default_default_client_scopes`.

Unlike the list-based resource (`keycloak_openid_client_default_scopes`), this resource is **not** authoritative for realm-wide default client scopes. Instead, it allows you to add or manage a single client scope without modifying other default client scopes already present in the realm.

!> This resource should be created before any clients that will use the default client scope.

## Example Usage

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_openid_client_scope" "openid_client_scope" {
  realm_id = keycloak_realm.realm.id
  name     = "groups"
}

resource "keycloak_openid_default_client_scope" "openid_default_client_scope" {
  realm_id        = keycloak_realm.realm.id
  client_scope_id = keycloak_openid_client_scope.client_scope.id
}
```

## Argument Reference

- `realm_id` - (Required) The realm this client scope belongs to.
- `client_scope_id` - (Required) The client scope to manage.
