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

# Using group_path to look up nested groups by their full path
data "keycloak_role" "super_admin" {
    realm_id = keycloak_realm.realm.id
    name     = "super_admin"
}

data "keycloak_group" "admins" {
    realm_id   = keycloak_realm.realm.id
    group_path = "/Administration/Full Admins"
}

resource "keycloak_group_roles" "admins_roles" {
    realm_id = keycloak_realm.realm.id
    group_id = data.keycloak_group.admins.id

    role_ids = [
        data.keycloak_role.super_admin.id
    ]
}
```

Organization groups can be looked up by setting `organization_id`. Organization groups require Keycloak 26.6.0 or later.

```hcl
data "keycloak_group" "organization_group" {
    realm_id        = keycloak_realm.realm.id
    organization_id = keycloak_organization.organization.id
    name            = "organization-group"
}
```

## Argument Reference

- `realm_id` - (Required) The realm this group exists within.
- `organization_id` - (Optional) The organization this group exists within. If omitted, the data source looks up realm groups.
- `name` - (Optional) The name of the group. Mutually exclusive with `group_path`. If there are multiple groups matching `name`, the first result is returned.
- `group_path` - (Optional) The full path of the group (e.g. `"/parent/child/subgroup"`). Mutually exclusive with `name`. This uses the Keycloak `/group-by-path` endpoint for a precise lookup.

## Attributes Reference

- `id` - (Computed) The unique ID of the group, which can be used as an argument to
  other resources supported by this provider.
