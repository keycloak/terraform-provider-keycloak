---
page_title: "keycloak_realm_optional_client_scopes Resource"
---

# keycloak\_realm\_optional\_client\_scopes Resource

Allows you to manage the set of optional client scopes for a Keycloak realm, which are used when new clients are created.

Note that this resource attempts to be an **authoritative** source over the optional client scopes for a Keycloak realm,
so any Keycloak defaults and manual adjustments will be overwritten.


## Example Usage

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_openid_client_scope" "client_scope" {
  realm_id = keycloak_realm.realm.id
  name     = "test-client-scope"
}

resource "keycloak_realm_optional_client_scopes" "optional_scopes" {
  realm_id  = keycloak_realm.realm.id

  optional_scopes = [
    "address",
    "phone",
    "offline_access",
    "microprofile-jwt",
    keycloak_openid_client_scope.client_scope.name
  ]
}
```

## Argument Reference

- `realm_id` - (Required) The realm this client and scopes exists in.
- `optional_scopes` - (Required) An array of optional client scope names that should be used when creating new Keycloak clients.

## Import

This resource does not support import. Instead of importing, feel free to create this resource
as if it did not already exist on the server.
