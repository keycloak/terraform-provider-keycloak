---
page_title: "keycloak_group_permissions Resource"
---

# keycloak_group_permissions

Allows you to manage all group Scope Based Permissions https://www.keycloak.org/docs/latest/server_admin/#group.

This is part of a preview Keycloak feature: `admin_fine_grained_authz` (see https://www.keycloak.org/docs/latest/server_admin/#_fine_grain_permissions).
This feature can be enabled with the Keycloak option `-Dkeycloak.profile.feature.admin_fine_grained_authz=enabled`. See the
example [`docker-compose.yml`](https://github.com/keycloak/terraform-provider-keycloak/blob/898094df6b3e01c3404981ce7ca268142d6ff0e5/docker-compose.yml#L21) file for an example.

When enabling Roles Permissions, Keycloak does several things automatically:
1. Enable Authorization on built-in `realm-management` client (if not already enabled).
1. Create a resource representing the role permissions.
1. Create scopes `view`, `manage`, `view-members`, `manage-members`, `manage-membership`.
1. Create all scope based permission for the scopes and role resource


### Example Usage

```hcl
resource "keycloak_realm" "realm" {
	realm = "my_realm"
}

data "keycloak_openid_client" "realm_management" {
  realm_id  = keycloak_realm.realm.id
  client_id = "realm-management"
}

resource "keycloak_openid_client_permissions" "realm-management_permission" {
	realm_id   = keycloak_realm.realm.id
	client_id  = data.keycloak_openid_client.realm_management.id
}

resource "keycloak_group" "group" {
	realm_id = keycloak_realm.realm.id
	name     = "%s"
}

resource "keycloak_openid_client_group_policy" "test" {
	realm_id           = keycloak_realm.realm.id
	resource_server_id = data.keycloak_openid_client.realm_management.id
	name 			   = "client_group_policy_test"
	groups {
		id              = keycloak_group.group.id
		path            = keycloak_group.group.path
		extend_children = false
	}
	logic             = "POSITIVE"
	decision_strategy = "UNANIMOUS"
	depends_on = [
		keycloak_openid_client_permissions.realm-management_permission,
	]
}

resource "keycloak_group_permissions" "test" {
	realm_id                               = keycloak_realm.realm.id
	group_id                               = keycloak_group.group.id
	manage_members_scope {
		policies          = [
			keycloak_openid_client_group_policy.test.id
		]
		description       = "mangage_members_scope"
		decision_strategy = "UNANIMOUS"
	}

}
```

### Argument Reference

The following arguments are supported:

- `realm_id` - (Required) The realm in which to manage fine-grained role permissions.
- `group_id` - (Required) The id of the group.


Each of the scopes that can be managed are defined below:

- `view_scope` - (Optional) Policies that decide if the admin can view information about the group.
- `manage_scope` - (Optional) Policies that decide if the admin can manage the configuration of the group.
- `view_members_scope` - (Optional) Policies that decide if the admin can view the user details of members of the group.
- `manage_members_scope` - (Optional) Policies that decide if the admin can manage the users that belong to this group.
- `manage_membership_scope` - (Optional) Policies that decide if an admin can change the membership of the group. Add or remove members from the group.

The configuration block for each of these scopes supports the following arguments:

- `policies` - (Optional) Assigned policies to the permission. Each element within this list should be a policy ID.
- `description` - (Optional) Description of the permission.
- `decision_strategy` - (Optional) Decision strategy of the permission.

### Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

- `enabled` - When true, this indicates that fine-grained role permissions are enabled. This will always be `true`.
- `authorization_resource_server_id` - Resource server id representing the realm management client on which these permissions are managed.
