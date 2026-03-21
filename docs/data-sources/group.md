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

# Use group_path
data "keycloak_role" "super_admin" {
    realm_id = keycloak_realm.realm.id
    name     = "super_admin"
}

data "keycloak_group" "admins" {
    realm_id   = keycloak_realm.realm.id
    group_path = "/Administration/Full Admins"
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
- `name` - (Optional) The name of the group (backward compatible). If there are multiple groups match `name`, the first result will be returned.
- `group_path` - (Optional) The **full path** of the group (e.g. `"/parent/child"`). Use this instead of `name` for nested groups or to guarantee uniqueness. **Exactly one** of `name` or `group_path` must be provided.

## Attributes Reference

- `id` - (Computed) The unique ID of the group, which can be used as an argument to
  other resources supported by this provider.

