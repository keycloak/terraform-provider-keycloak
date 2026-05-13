---
page_title: "keycloak_group Data Source"
---

# keycloak\_group Data Source

This data source can be used to fetch properties of a Keycloak group for
usage with other resources, such as `keycloak_group_roles`.

## Example Usage

```hcl
resource "keycloak_realm" "realm" {
    realm   = "my-realm"
    enabled = true
}

data "keycloak_role" "offline_access" {
    realm_id = keycloak_realm.realm.id
    name     = "offline_access"
}

data "keycloak_group" "group" {
    realm_id = keycloak_realm.realm.id
    name     = "group"
}

resource "keycloak_group_roles" "group_roles" {
    realm_id = keycloak_realm.realm.id
    group_id = data.keycloak_group.group.id

    role_ids = [
        data.keycloak_role.offline_access.id
    ]
}

# Using group path support: name starting with '/' is treated as full group path
data "keycloak_role" "super_admin" {
    realm_id = keycloak_realm.realm.id
    name     = "super_admin"
}

data "keycloak_group" "admins" {
    realm_id = keycloak_realm.realm.id
    name     = "/Administration/Full Admins"
}

resource "keycloak_group_roles" "admins_roles" {
    realm_id = data.keycloak_realm.realm.id
    group_id = data.keycloak_group.admins.id

    role_ids = [
        keycloak_role.super_admin.id
    ]
}
```

## Argument Reference

- `realm_id` - (Required) The realm this group exists within.
- `name` - (Required) The name of the group or its full path.
   If the value starts with `/`, it is interpreted as the full group path (e.g. `"/parent/child/subgroup"`).
   Otherwise, it is treated as a group name (legacy behavior). When multiple groups match the name, the first result is returned.

## Attributes Reference

- `id` - (Computed) The unique ID of the group, which can be used as an argument to
  other resources supported by this provider.

