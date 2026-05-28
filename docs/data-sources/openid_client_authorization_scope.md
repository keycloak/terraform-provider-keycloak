---
page_title: "keycloak_openid_client_authorization_scope Data Source"
---

# keycloak\_openid\_client\_authorization\_scope Data Source

This data source fetches an authorization scope by name from an OpenID client that has authorization enabled.

The primary use case is looking up scopes that Keycloak creates automatically — for example, the `view`, `manage`, `view-members`, `manage-members`, and `manage-membership` scopes that Keycloak creates on the `realm-management` client when Fine-Grained Admin Permissions are enabled for a group. These scopes need to be referenced by ID in `keycloak_openid_client_authorization_permission`, but their IDs are not known until after the permissions have been initialized. Providing a scope name directly (instead of an ID) causes plan instability because Keycloak always stores and returns scope IDs.

## Example Usage

The following example demonstrates the primary use case: resolving an authorization scope by name so it can be referenced by ID in a `keycloak_openid_client_authorization_permission`. Using a scope name directly causes plan instability because Keycloak normalises the value to an ID on write — this data source fixes that.

```hcl
resource "keycloak_realm" "realm" {
  realm = "my-realm"
}

# An application client with authorization enabled.
resource "keycloak_openid_client" "app" {
  realm_id  = keycloak_realm.realm.id
  client_id = "my-app"
  access_type = "CONFIDENTIAL"
  service_accounts_enabled = true
  authorization {
    policy_enforcement_mode = "ENFORCING"
  }
}

# A named scope on the authorization-enabled client.
resource "keycloak_openid_client_authorization_scope" "read_orders" {
  realm_id           = keycloak_realm.realm.id
  resource_server_id = keycloak_openid_client.app.resource_server_id
  name               = "read:orders"
}

# Resolve the scope ID by name — stable across plan/apply cycles.
data "keycloak_openid_client_authorization_scope" "read_orders" {
  realm_id           = keycloak_realm.realm.id
  resource_server_id = keycloak_openid_client.app.resource_server_id
  name               = "read:orders"
  depends_on         = [keycloak_openid_client_authorization_scope.read_orders]
}

resource "keycloak_openid_client_authorization_resource" "orders" {
  realm_id           = keycloak_realm.realm.id
  resource_server_id = keycloak_openid_client.app.resource_server_id
  name               = "orders"
  uris               = ["/orders/*"]
}

resource "keycloak_openid_client_authorization_permission" "read_orders" {
  realm_id           = keycloak_realm.realm.id
  resource_server_id = keycloak_openid_client.app.resource_server_id
  name               = "read-orders-permission"
  type               = "scope"
  decision_strategy  = "UNANIMOUS"
  resources          = [keycloak_openid_client_authorization_resource.orders.id]
  # Use the resolved ID — not the name string — to prevent perpetual diffs.
  scopes             = [data.keycloak_openid_client_authorization_scope.read_orders.id]
}
```

### FGAPv2 example: role-based map-role permission

The following shows how to use FGAPv2 (`admin-fine-grained-authz:v2`) to grant members of an `hr-managers` group permission to map an `hr-viewer` role. The `map-role` scope lives on the `admin-permissions` resource server; this data source resolves its ID for use in a custom scope-based permission alongside the one managed by `keycloak_role_permissions`.

```hcl
resource "keycloak_realm" "realm" {
  realm                     = "my-realm"
  admin_permissions_enabled = true
}

data "keycloak_openid_client" "admin_permissions" {
  realm_id  = keycloak_realm.realm.id
  client_id = "admin-permissions"
}

resource "keycloak_role" "hr_viewer" {
  realm_id = keycloak_realm.realm.id
  name     = "hr-viewer"
}

# Enabling role permissions creates map-role/map-role-client-scope/map-role-composite
# scope permissions on the admin-permissions resource server.
resource "keycloak_role_permissions" "hr_viewer" {
  realm_id = keycloak_realm.realm.id
  role_id  = keycloak_role.hr_viewer.id
}

# Resolve the map-role scope by name for use in a custom permission.
data "keycloak_openid_client_authorization_scope" "map_role" {
  realm_id           = keycloak_realm.realm.id
  resource_server_id = data.keycloak_openid_client.admin_permissions.id
  name               = "map-role"
  depends_on         = [keycloak_role_permissions.hr_viewer]
}

resource "keycloak_group" "hr_managers" {
  realm_id = keycloak_realm.realm.id
  name     = "hr-managers"
}

resource "keycloak_openid_client_group_policy" "hr_managers" {
  realm_id           = keycloak_realm.realm.id
  resource_server_id = data.keycloak_openid_client.admin_permissions.id
  name               = "policy-hr-managers"
  decision_strategy  = "UNANIMOUS"
  logic              = "POSITIVE"
  groups {
    id              = keycloak_group.hr_managers.id
    path            = keycloak_group.hr_managers.path
    extend_children = false
  }
}

resource "keycloak_openid_client_authorization_permission" "hr_managers_map_role" {
  realm_id           = keycloak_realm.realm.id
  resource_server_id = data.keycloak_openid_client.admin_permissions.id
  name               = "hr-managers-map-hr-viewer-role"
  type               = "scope"
  decision_strategy  = "UNANIMOUS"
  resources          = [keycloak_role.hr_viewer.id]
  scopes             = [data.keycloak_openid_client_authorization_scope.map_role.id]
  policies           = [keycloak_openid_client_group_policy.hr_managers.id]
}
```

## Argument Reference

- `realm_id` - (Required) The realm this authorization scope exists within.
- `resource_server_id` - (Required) The ID of the resource server (client) this authorization scope belongs to.
- `name` - (Required) The name of the authorization scope to look up.

## Attributes Reference

- `display_name` - (Computed) The display name of the scope.
- `icon_uri` - (Computed) The icon URI of the scope.
