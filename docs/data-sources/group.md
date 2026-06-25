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
- `name` - (Required) The name of the group. If there are multiple groups match `name`, the first result will be returned.

## Attributes Reference

- `id` - (Computed) The unique ID of the group, which can be used as an argument to
  other resources supported by this provider.
