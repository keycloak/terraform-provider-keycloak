---
page_title: "keycloak_openid_client_admin_permissions Resource"
---

# keycloak_openid_client_admin_permissions

Allows you to manage a fine-grained admin permission for Keycloak OpenID Connect clients.

This resource requires Fine-Grained Admin Permissions v2 (`admin-fine-grained-authz:v2`), available since Keycloak 26.2. See the [`docker-compose.yml`](https://github.com/keycloak/terraform-provider-keycloak/blob/master/docker-compose.yml) for an example of how to enable this feature.

Each instance of this resource represents **one permission** in Keycloak. A single permission can span multiple scopes, target multiple specific clients, and reference multiple policies.

When `admin_permissions_enabled = true` is set on the realm, Keycloak automatically creates an `admin-permissions` client that serves as the authorization resource server for all FGAPv2 admin permissions.

Available scopes:

- `view` — read the client's configuration
- `manage` — change the client's configuration
- `map-roles` — map the client's roles to users or groups
- `map-roles-client-scope` — use the client's roles as client scopes
- `map-roles-composite` — add the client's roles as composites to other roles

> **Note:** The `configure` and `token-exchange` scopes from v1 (`keycloak_openid_client_permissions`) have no equivalent in FGAPv2 and are not available in this resource.

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

resource "keycloak_openid_client" "my_client" {
  realm_id    = keycloak_realm.realm.id
  client_id   = "my-client"
  access_type = "CONFIDENTIAL"
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

# Permission targeting a specific client with multiple scopes.
resource "keycloak_openid_client_admin_permissions" "admins_manage_my_client" {
  realm_id          = keycloak_realm.realm.id
  name              = "admins-manage-my-client"
  description       = "Admins can view and manage my-client"
  decision_strategy = "UNANIMOUS"

  client_ids = [keycloak_openid_client.my_client.id]
  scopes     = ["view", "manage"]
  policies   = [keycloak_openid_client_group_policy.admins_policy.id]
}

# Permission targeting ALL clients in the realm (client_ids omitted).
resource "keycloak_openid_client_admin_permissions" "admins_view_all_clients" {
  realm_id = keycloak_realm.realm.id
  name     = "admins-can-view-all-clients"

  # client_ids omitted → applies to all clients in the realm
  scopes   = ["view"]
  policies = [keycloak_openid_client_group_policy.admins_policy.id]
}
```

## Argument Reference

- `realm_id` - (Required, ForceNew) The realm in which to manage this permission.
- `name` - (Required) The name of the permission. Must be unique within the `admin-permissions` resource server. On first apply, if a permission with this name already exists it is adopted; otherwise a new one is created.
- `description` - (Optional) Description of the permission.
- `decision_strategy` - (Optional) Decision strategy. One of `UNANIMOUS`, `AFFIRMATIVE`, or `CONSENSUS`. Defaults to `UNANIMOUS`.
- `client_ids` - (Optional) Set of client UUIDs (`keycloak_openid_client.xxx.id`) this permission applies to. When omitted or empty, the permission applies to **all clients** in the realm.
- `scopes` - (Required) Set of scopes this permission grants. Valid values: `view`, `manage`, `map-roles`, `map-roles-client-scope`, `map-roles-composite`.
- `policies` - (Optional) Set of policy IDs to attach to the permission.

## Attributes Reference

In addition to the arguments above, the following computed attributes are exported:

- `permission_id` - The internal Keycloak UUID of the permission.
- `enabled` - Always `true` when the resource exists.
- `authorization_resource_server_id` - The ID of the `admin-permissions` client, which acts as the resource server for these permissions.

## Import

OpenID client admin permissions can be imported using `{{realmId}}/{{permissionId}}`:

```bash
terraform import keycloak_openid_client_admin_permissions.example my-realm/permission-uuid
```

After import, run `terraform apply` to reconcile `client_ids`, `scopes`, and `policies` with your configuration.
