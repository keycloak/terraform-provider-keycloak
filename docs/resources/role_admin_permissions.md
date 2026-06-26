---
page_title: "keycloak_role_admin_permissions Resource"
---

# keycloak_role_admin_permissions

Allows you to manage a fine-grained admin permission for Keycloak roles.

This resource requires Fine-Grained Admin Permissions v2 (`admin-fine-grained-authz:v2`), available since Keycloak 26.2. See the [`docker-compose.yml`](https://github.com/keycloak/terraform-provider-keycloak/blob/master/docker-compose.yml) for an example of how to enable this feature.

Each instance of this resource represents **one permission** in Keycloak (matching what you see when you click "Create permission" in the Keycloak UI). A single permission can span multiple scopes, target multiple specific roles, and reference multiple policies.

When `admin_permissions_enabled = true` is set on the realm, Keycloak automatically creates an `admin-permissions` client that serves as the authorization resource server for all FGAPv2 admin permissions.

Available scopes:

- `map-role` — map this role to users, groups, or clients
- `map-role-client-scope` — use this role as a client scope
- `map-role-composite` — add this role as a composite to another role

## Example Usage

```hcl
resource "keycloak_realm" "realm" {
  realm                     = "my-realm"
  admin_permissions_enabled = true
}

data "keycloak_openid_client" "admin_permissions" {
  realm_id  = keycloak_realm.realm.id
  client_id = "admin-permissions"
}

resource "keycloak_role" "role_a" {
  realm_id = keycloak_realm.realm.id
  name     = "role-a"
}

resource "keycloak_role" "role_b" {
  realm_id = keycloak_realm.realm.id
  name     = "role-b"
}

resource "keycloak_group" "admins" {
  realm_id = keycloak_realm.realm.id
  name     = "admins"
}

resource "keycloak_openid_client_group_policy" "admins_policy" {
  realm_id           = keycloak_realm.realm.id
  resource_server_id = data.keycloak_openid_client.admin_permissions.id
  name               = "admins-policy"
  groups {
    id              = keycloak_group.admins.id
    path            = keycloak_group.admins.path
    extend_children = false
  }
  logic             = "POSITIVE"
  decision_strategy = "UNANIMOUS"
}

# Permission targeting two specific roles with multiple scopes.
resource "keycloak_role_admin_permissions" "admins_map_specific_roles" {
  realm_id          = keycloak_realm.realm.id
  name              = "admins-can-map-specific-roles"
  description       = "Admins can map or make composite role-a and role-b"
  decision_strategy = "UNANIMOUS"

  role_ids = [
    keycloak_role.role_a.id,
    keycloak_role.role_b.id,
  ]
  scopes   = ["map-role", "map-role-composite"]
  policies = [keycloak_openid_client_group_policy.admins_policy.id]
}

# Permission targeting ALL roles in the realm (role_ids omitted).
resource "keycloak_role_admin_permissions" "admins_map_any_role" {
  realm_id = keycloak_realm.realm.id
  name     = "admins-can-map-any-role"

  # role_ids omitted → applies to all roles in the realm
  scopes   = ["map-role"]
  policies = [keycloak_openid_client_group_policy.admins_policy.id]
}
```

## Argument Reference

- `realm_id` - (Required, ForceNew) The realm in which to manage this permission.
- `name` - (Required) The name of the permission. Must be unique within the `admin-permissions` resource server. On first apply, if a permission with this name already exists it is adopted; otherwise a new one is created.
- `description` - (Optional) Description of the permission.
- `decision_strategy` - (Optional) Decision strategy. One of `UNANIMOUS`, `AFFIRMATIVE`, or `CONSENSUS`. Defaults to `UNANIMOUS`.
- `role_ids` - (Optional) Set of role UUIDs (`keycloak_role.xxx.id`) this permission applies to. When omitted or empty, the permission applies to **all roles** in the realm.
- `scopes` - (Required) Set of scopes this permission grants. Valid values: `map-role`, `map-role-client-scope`, `map-role-composite`.
- `policies` - (Optional) Set of policy IDs to attach to the permission.

## Attributes Reference

In addition to the arguments above, the following computed attributes are exported:

- `permission_id` - The internal Keycloak UUID of the permission.
- `enabled` - Always `true` when the resource exists.
- `authorization_resource_server_id` - The ID of the `admin-permissions` client, which acts as the resource server for these permissions.

## Import

Role admin permissions can be imported using `{{realmId}}/{{permissionId}}`:

```bash
terraform import keycloak_role_admin_permissions.example my-realm/permission-uuid
```

After import, run `terraform apply` to reconcile `role_ids`, `scopes`, and `policies` with your configuration.
