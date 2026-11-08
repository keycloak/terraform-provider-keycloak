---
page_title: "keycloak_users_admin_permissions Resource"
---

# keycloak_users_admin_permissions

Allows you to manage a fine-grained admin permission for all users in a Keycloak realm, or for named users.

This resource requires Fine-Grained Admin Permissions v2 (`admin-fine-grained-authz:v2`), available since Keycloak 26.2. See the [`docker-compose.yml`](https://github.com/keycloak/terraform-provider-keycloak/blob/master/docker-compose.yml) for an example of how to enable this feature.

Each instance of this resource represents **one permission** in Keycloak. With `user_ids` omitted the permission is realm-wide and applies to all users in the realm; with `user_ids` set it applies to those users only (for example, a service account's permission to view its own user).

When `admin_permissions_enabled = true` is set on the realm, Keycloak automatically creates an `admin-permissions` client that serves as the authorization resource server for all FGAPv2 admin permissions.

Available scopes:

- `view` — list and view user details
- `manage` — create, update, and delete users
- `map-roles` — assign or remove realm roles on users
- `manage-group-membership` — add or remove users from groups
- `impersonate` — impersonate a user

> **Note:** The `user-impersonated` scope from v1 (`keycloak_users_permissions`) has no equivalent in FGAPv2 and is not available in this resource.

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

resource "keycloak_group" "auditors" {
  realm_id = keycloak_realm.realm.id
  name     = "auditors"
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

resource "keycloak_openid_client_group_policy" "auditors_policy" {
  realm_id           = keycloak_realm.realm.id
  resource_server_id = data.keycloak_openid_client.admin_permissions.id
  name               = "auditors-policy"
  groups {
    id              = keycloak_group.auditors.id
    path            = keycloak_group.auditors.path
    extend_children = false
  }
  logic             = "POSITIVE"
  decision_strategy = "UNANIMOUS"
}

# One permission per logical role — each is a separate Terraform resource.
resource "keycloak_users_admin_permissions" "admins_manage_users" {
  realm_id          = keycloak_realm.realm.id
  name              = "admins-can-manage-users"
  description       = "Admins can view and manage all users"
  decision_strategy = "UNANIMOUS"

  scopes   = ["view", "manage"]
  policies = [keycloak_openid_client_group_policy.admins_policy.id]
}

resource "keycloak_users_admin_permissions" "auditors_view_users" {
  realm_id    = keycloak_realm.realm.id
  name        = "auditors-can-view-users"
  description = "Auditors can view all users"

  scopes   = ["view"]
  policies = [keycloak_openid_client_group_policy.auditors_policy.id]
}

# A permission on named users only: a service account that may view its own
# user (for example, to read its own role mappings) and no other.
resource "keycloak_openid_client" "automation" {
  realm_id                 = keycloak_realm.realm.id
  client_id                = "automation"
  access_type              = "CONFIDENTIAL"
  service_accounts_enabled = true
}

resource "keycloak_openid_client_user_policy" "automation" {
  realm_id           = keycloak_realm.realm.id
  resource_server_id = data.keycloak_openid_client.admin_permissions.id
  name               = "automation"
  users              = [keycloak_openid_client.automation.service_account_user_id]
  logic              = "POSITIVE"
  decision_strategy  = "UNANIMOUS"
}

resource "keycloak_users_admin_permissions" "automation_views_itself" {
  realm_id = keycloak_realm.realm.id
  name     = "automation-views-itself"
  user_ids = [keycloak_openid_client.automation.service_account_user_id]
  scopes   = ["view"]
  policies = [keycloak_openid_client_user_policy.automation.id]
}
```

## Argument Reference

- `realm_id` - (Required, ForceNew) The realm in which to manage this permission.
- `name` - (Required) The name of the permission. Must be unique within the `admin-permissions` resource server. On first apply, if a permission with this name already exists it is adopted; otherwise a new one is created.
- `description` - (Optional) Description of the permission.
- `decision_strategy` - (Optional) Decision strategy. One of `UNANIMOUS`, `AFFIRMATIVE`, or `CONSENSUS`. Defaults to `UNANIMOUS`.
- `user_ids` - (Optional) Set of user UUIDs this permission applies to. Omit to target all users in the realm. Read back from Keycloak, so a target changed outside Terraform shows in the plan.
- `scopes` - (Required) Set of scopes this permission grants. Valid values: `view`, `manage`, `map-roles`, `manage-group-membership`, `impersonate`.
- `policies` - (Optional) Set of policy IDs to attach to the permission.

## Attributes Reference

In addition to the arguments above, the following computed attributes are exported:

- `permission_id` - The internal Keycloak UUID of the permission.
- `enabled` - Always `true` when the resource exists.
- `authorization_resource_server_id` - The ID of the `admin-permissions` client, which acts as the resource server for these permissions.

## Import

Users admin permissions can be imported using `{{realmId}}/{{permissionId}}`:

```bash
terraform import keycloak_users_admin_permissions.example my-realm/permission-uuid
```

After import, run `terraform apply` to reconcile `scopes`, `policies` and `user_ids` with your configuration.
