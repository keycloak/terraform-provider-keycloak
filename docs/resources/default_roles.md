---
page_title: "keycloak_default_roles Resource"
---

# keycloak\_default\_roles Resource

Allows managing default roles within Keycloak.

Note: This feature was added in Keycloak v13, so this resource will not work on older versions of Keycloak.

## Example Usage (Realm role)

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_default_roles" "default_roles" {
  realm_id      = keycloak_realm.realm.id
  default_roles = ["uma_authorization"]
}
```

## Example Usage (Client roles)

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_default_roles" "default_roles" {
  realm_id      = keycloak_realm.realm.id
  default_roles = ["account/manage-account", "account/view-groups"]
}
```

## Argument Reference

- `realm_id` - (Required) The realm this role exists within.
- `default_roles` - (Required) Roles assigned to new users by default.

## Import

Default roles can be imported using the format `{{realm_id}}/{{default_role_id}}`, where `default_role_id` is the unique ID of the composite
role that Keycloak uses to control default realm level roles. The ID is not easy to find in the GUI, but it appears in the dev tools when editing
the default roles.

Example:

```bash
$ terraform import keycloak_default_roles.default_roles my-realm/a04c35c2-e95a-4dc5-bd32-e83a21be9e7d
```
