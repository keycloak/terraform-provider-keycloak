---
page_title: "keycloak_role_permissions Resource"
---

# keycloak_role_permissions

Allows you to manage fine-grained permissions for a Keycloak role.

This resource is part of the Fine-Grained Admin Permissions v2 feature introduced in Keycloak 26.2, which must be enabled via `admin-fine-grained-authz:v2`. See the [`docker-compose.yml`](https://github.com/keycloak/terraform-provider-keycloak/blob/master/docker-compose.yml) for an example of how to enable this feature.

Enabling permissions for a role causes Keycloak to automatically:

1. Enable authorization on the built-in `realm-management` client (if not already enabled).
1. Create a resource representing the role.
1. Create scopes `view`, `map-role`, and `manage`.
1. Create scope-based permissions for each of those scopes.

## Example Usage

```hcl
resource "keycloak_realm" "realm" {
  realm = "my-realm"
}

data "keycloak_openid_client" "realm_management" {
  realm_id  = keycloak_realm.realm.id
  client_id = "realm-management"
}

resource "keycloak_openid_client_permissions" "realm_management_permission" {
  realm_id  = keycloak_realm.realm.id
  client_id = data.keycloak_openid_client.realm_management.id
}

resource "keycloak_role" "role" {
  realm_id = keycloak_realm.realm.id
  name     = "my-role"
}

resource "keycloak_group" "group" {
  realm_id = keycloak_realm.realm.id
  name     = "my-group"
}

resource "keycloak_openid_client_group_policy" "policy" {
  realm_id           = keycloak_realm.realm.id
  resource_server_id = data.keycloak_openid_client.realm_management.id
  name               = "my-group-policy"
  groups {
    id              = keycloak_group.group.id
    path            = keycloak_group.group.path
    extend_children = false
  }
  logic             = "POSITIVE"
  decision_strategy = "UNANIMOUS"
  depends_on = [
    keycloak_openid_client_permissions.realm_management_permission,
  ]
}

resource "keycloak_role_permissions" "role_permissions" {
  realm_id = keycloak_realm.realm.id
  role_id  = keycloak_role.role.id

  map_role_scope {
    policies          = [keycloak_openid_client_group_policy.policy.id]
    description       = "Allow my-group to map this role"
    decision_strategy = "UNANIMOUS"
  }
}
```

## Argument Reference

- `realm_id` - (Required) The realm in which to manage fine-grained role permissions.
- `role_id` - (Required) The ID of the role.

Each scope block is optional. When specified, the block configures the policy attached to that scope's permission.

- `view_scope` - (Optional) Policies that decide if an admin can view this role.
- `map_role_scope` - (Optional) Policies that decide if an admin can map this role to users, groups, or clients.
- `manage_scope` - (Optional) Policies that decide if an admin can manage the configuration of this role.

Each scope block supports:

- `policies` - (Optional) List of policy IDs to attach to the permission.
- `description` - (Optional) Description of the permission.
- `decision_strategy` - (Optional) Decision strategy of the permission. Can be `UNANIMOUS`, `AFFIRMATIVE`, or `CONSENSUS`.

## Attributes Reference

In addition to the arguments above, the following computed attributes are exported:

- `enabled` - When `true`, fine-grained permissions are enabled for this role. This will always be `true`.
- `authorization_resource_server_id` - The ID of the `realm-management` client, which acts as the resource server for these permissions.

## Import

Role permissions can be imported using the format `{{realmId}}/{{roleId}}`:

```bash
terraform import keycloak_role_permissions.role_permissions my-realm/role-uuid
```
