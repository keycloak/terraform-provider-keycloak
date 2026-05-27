---
page_title: "keycloak_openid_client_authorization_scope Data Source"
---

# keycloak\_openid\_client\_authorization\_scope Data Source

This data source fetches an authorization scope by name from an OpenID client that has authorization enabled.

The primary use case is looking up scopes that Keycloak creates automatically — for example, the `view`, `manage`, `view-members`, `manage-members`, and `manage-membership` scopes that Keycloak creates on the `realm-management` client when Fine-Grained Admin Permissions are enabled for a group. These scopes need to be referenced by ID in `keycloak_openid_client_authorization_permission`, but their IDs are not known until after the permissions have been initialized. Providing a scope name directly (instead of an ID) causes plan instability because Keycloak always stores and returns scope IDs.

## Example Usage

The following example uses Fine-Grained Admin Permissions v2 to allow both holders of an `hr-auditor` role and members of an `hr-managers` group to view the members of an `hr-team` group. The `view-members` scope is created automatically by Keycloak when group permissions are enabled; this data source resolves its ID so it can be used in a scope-based permission.

```hcl
resource "keycloak_realm" "realm" {
  realm                     = "my-realm"
  admin_permissions_enabled = true
}

data "keycloak_openid_client" "realm_management" {
  realm_id  = keycloak_realm.realm.id
  client_id = "realm-management"
}

# Authorization must be enabled on realm-management before policies can be attached.
resource "keycloak_openid_client_permissions" "realm_management" {
  realm_id  = keycloak_realm.realm.id
  client_id = data.keycloak_openid_client.realm_management.id
}

# The group whose members we want to expose to specific admins.
resource "keycloak_group" "hr_team" {
  realm_id = keycloak_realm.realm.id
  name     = "hr-team"
}

# Enabling group permissions causes Keycloak to auto-create authorization scopes
# on realm-management: view, manage, view-members, manage-members, manage-membership.
resource "keycloak_group_permissions" "hr_team" {
  realm_id   = keycloak_realm.realm.id
  group_id   = keycloak_group.hr_team.id
  depends_on = [keycloak_openid_client_permissions.realm_management]
}

# Resolve the ID of the auto-created view-members scope.
data "keycloak_openid_client_authorization_scope" "view_members" {
  realm_id           = keycloak_realm.realm.id
  resource_server_id = data.keycloak_openid_client.realm_management.id
  name               = "view-members"
  depends_on         = [keycloak_group_permissions.hr_team]
}

# Who should be allowed to view hr-team members.
resource "keycloak_role" "hr_auditor" {
  realm_id = keycloak_realm.realm.id
  name     = "hr-auditor"
}

resource "keycloak_group" "hr_managers" {
  realm_id = keycloak_realm.realm.id
  name     = "hr-managers"
}

# Policy: holders of the hr-auditor role.
resource "keycloak_openid_client_role_policy" "hr_auditor" {
  realm_id           = keycloak_realm.realm.id
  resource_server_id = data.keycloak_openid_client.realm_management.id
  name               = "policy-hr-auditor-role"
  decision_strategy  = "UNANIMOUS"
  logic              = "POSITIVE"
  roles {
    id       = keycloak_role.hr_auditor.id
    required = false
  }
  depends_on = [keycloak_openid_client_permissions.realm_management]
}

# Policy: members of the hr-managers group.
resource "keycloak_openid_client_group_policy" "hr_managers" {
  realm_id           = keycloak_realm.realm.id
  resource_server_id = data.keycloak_openid_client.realm_management.id
  name               = "policy-hr-managers-group"
  decision_strategy  = "UNANIMOUS"
  logic              = "POSITIVE"
  groups {
    id              = keycloak_group.hr_managers.id
    path            = keycloak_group.hr_managers.path
    extend_children = false
  }
  depends_on = [keycloak_openid_client_permissions.realm_management]
}

# Grant both the hr-auditor role and the hr-managers group the view-members
# permission on hr-team. The scope must be provided as an ID — using the
# scope name directly would cause Terraform to detect a diff on every plan
# because Keycloak normalises the value to an ID on write.
resource "keycloak_openid_client_authorization_permission" "view_hr_team_members" {
  realm_id           = keycloak_realm.realm.id
  resource_server_id = data.keycloak_openid_client.realm_management.id
  name               = "view-hr-team-members"
  type               = "scope"
  decision_strategy  = "AFFIRMATIVE"
  scopes             = [data.keycloak_openid_client_authorization_scope.view_members.id]
  policies = [
    keycloak_openid_client_role_policy.hr_auditor.id,
    keycloak_openid_client_group_policy.hr_managers.id,
  ]
}
```

## Argument Reference

- `realm_id` - (Required) The realm this authorization scope exists within.
- `resource_server_id` - (Required) The ID of the resource server (client) this authorization scope belongs to.
- `name` - (Required) The name of the authorization scope to look up.

## Attributes Reference

- `display_name` - (Computed) The display name of the scope.
- `icon_uri` - (Computed) The icon URI of the scope.
