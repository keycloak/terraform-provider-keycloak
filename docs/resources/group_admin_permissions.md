---
page_title: "keycloak_group_admin_permissions Resource"
---

# keycloak_group_admin_permissions

Allows you to manage a fine-grained admin permission for Keycloak groups.

This resource requires Fine-Grained Admin Permissions v2 (`admin-fine-grained-authz:v2`), available since Keycloak 26.2. See the [`docker-compose.yml`](https://github.com/keycloak/terraform-provider-keycloak/blob/master/docker-compose.yml) for an example of how to enable this feature.

Each instance of this resource represents **one permission** in Keycloak. A single permission can span multiple scopes, target multiple specific groups, and reference multiple policies.

When `admin_permissions_enabled = true` is set on the realm, Keycloak automatically creates an `admin-permissions` client that serves as the authorization resource server for all FGAPv2 admin permissions.

Available scopes:

- `view` — view the group's configuration
- `manage` — change the group's configuration
- `view-members` — view user details of the group's members
- `manage-members` — manage the users that belong to this group
- `manage-membership` — add or remove members from this group

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

resource "keycloak_group" "admins" {
  realm_id = keycloak_realm.realm.id
  name     = "admins"
}

resource "keycloak_group" "target_group" {
  realm_id = keycloak_realm.realm.id
  name     = "my-group"
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

# Permission targeting a specific group with multiple scopes.
resource "keycloak_group_admin_permissions" "admins_manage_target_group" {
  realm_id          = keycloak_realm.realm.id
  name              = "admins-manage-my-group"
  description       = "Admins can view and manage members of my-group"
  decision_strategy = "UNANIMOUS"

  group_ids = [keycloak_group.target_group.id]
  scopes    = ["view", "manage-members"]
  policies  = [keycloak_openid_client_group_policy.admins_policy.id]
}

# Permission targeting ALL groups in the realm (group_ids omitted).
resource "keycloak_group_admin_permissions" "admins_view_all_groups" {
  realm_id = keycloak_realm.realm.id
  name     = "admins-can-view-all-groups"

  # group_ids omitted → applies to all groups in the realm
  scopes   = ["view"]
  policies = [keycloak_openid_client_group_policy.admins_policy.id]
}
```

## Argument Reference

- `realm_id` - (Required, ForceNew) The realm in which to manage this permission.
- `name` - (Required) The name of the permission. Must be unique within the `admin-permissions` resource server. On first apply, if a permission with this name already exists it is adopted; otherwise a new one is created.
- `description` - (Optional) Description of the permission.
- `decision_strategy` - (Optional) Decision strategy. One of `UNANIMOUS`, `AFFIRMATIVE`, or `CONSENSUS`. Defaults to `UNANIMOUS`.
- `group_ids` - (Optional) Set of group UUIDs (`keycloak_group.xxx.id`) this permission applies to. When omitted or empty, the permission applies to **all groups** in the realm.
- `scopes` - (Required) Set of scopes this permission grants. Valid values: `view`, `manage`, `view-members`, `manage-members`, `manage-membership`.
- `policies` - (Optional) Set of policy IDs to attach to the permission.

## Attributes Reference

In addition to the arguments above, the following computed attributes are exported:

- `permission_id` - The internal Keycloak UUID of the permission.
- `enabled` - Always `true` when the resource exists.
- `authorization_resource_server_id` - The ID of the `admin-permissions` client, which acts as the resource server for these permissions.

## Import

Group admin permissions can be imported using `{{realmId}}/{{permissionId}}`:

```bash
terraform import keycloak_group_admin_permissions.example my-realm/permission-uuid
```

After import, run `terraform apply` to reconcile `group_ids`, `scopes`, and `policies` with your configuration.
