---
page_title: "keycloak_openid_default_client_scope Resource"
---

# keycloak\_openid\_default\_client\_scope Resource

Allows for creating or removing Keycloak client scopes from default that use the OpenID Connect protocol.

A default client scope will be assigned automatically to each new client.

## Example Usage

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_openid_client_scope" "openid_client_scope" {
  realm_id               = keycloak_realm.realm.id
  name                   = "groups"
}

resource "keycloak_openid_default_client_scope" "openid_default_client_scope" {
	realm_id = keycloak_realm.realm.id
	client_scope_id = keycloak_openid_client_scope.client_scope.id
}
```

## Argument Reference

- `realm_id` - (Required) The realm this client scope belongs to.
- `client_scope_id` - (Required) The client scope to manage.
