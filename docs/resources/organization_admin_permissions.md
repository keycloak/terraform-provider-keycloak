---
page_title: "keycloak_organization_admin_permissions Resource"
---

# keycloak_organization_admin_permissions

Allows you to manage a fine-grained admin permission for Keycloak organizations.

This resource requires Fine-Grained Admin Permissions v2 (`admin-fine-grained-authz:v2`) and Keycloak 26.7 or later, the first version whose admin permissions cover organizations. The realm needs organizations enabled (`organizations_enabled = true`). See the [`docker-compose.yml`](https://github.com/keycloak/terraform-provider-keycloak/blob/master/docker-compose.yml) for an example of how to enable this feature.

Each instance of this resource represents **one permission** in Keycloak. A single permission can span both scopes, target multiple specific organizations, and reference multiple policies.

When `admin_permissions_enabled = true` is set on the realm, Keycloak automatically creates an `admin-permissions` client that serves as the authorization resource server for all FGAPv2 admin permissions.

Available scopes:

- `view` — view the organization; also what makes organizations appear in queries
- `manage` — manage the organization, including updating and deleting it

Creating organizations requires `manage` on all organizations (a permission without `organization_ids`), not only on specific ones.

## Example Usage

```hcl
resource "keycloak_realm" "realm" {
  realm                     = "my-realm"
  organizations_enabled     = true
  admin_permissions_enabled = true
}

data "keycloak_openid_client" "admin_permissions" {
  realm_id  = keycloak_realm.realm.id
  client_id = "admin-permissions"
}

resource "keycloak_group" "org_admins" {
  realm_id = keycloak_realm.realm.id
  name     = "org-admins"
}

resource "keycloak_organization" "acme" {
  realm = keycloak_realm.realm.id
  name  = "acme"

  domain {
    name = "acme.example.com"
  }
}

resource "keycloak_openid_client_group_policy" "org_admins_policy" {
  realm_id           = keycloak_realm.realm.id
  resource_server_id = data.keycloak_openid_client.admin_permissions.id
  name               = "org-admins-policy"
  groups {
    id              = keycloak_group.org_admins.id
    path            = keycloak_group.org_admins.path
    extend_children = false
  }
  logic             = "POSITIVE"
  decision_strategy = "UNANIMOUS"
}

# Permission targeting a specific organization.
resource "keycloak_organization_admin_permissions" "org_admins_manage_acme" {
  realm_id          = keycloak_realm.realm.id
  name              = "org-admins-manage-acme"
  description       = "Org admins can view and manage acme"
  decision_strategy = "UNANIMOUS"

  organization_ids = [keycloak_organization.acme.id]
  scopes           = ["view", "manage"]
  policies         = [keycloak_openid_client_group_policy.org_admins_policy.id]
}

# Permission targeting ALL organizations in the realm (organization_ids omitted).
resource "keycloak_organization_admin_permissions" "org_admins_view_all_organizations" {
  realm_id = keycloak_realm.realm.id
  name     = "org-admins-can-view-all-organizations"

  # organization_ids omitted → applies to all organizations in the realm
  scopes   = ["view"]
  policies = [keycloak_openid_client_group_policy.org_admins_policy.id]
}
```

## Argument Reference

- `realm_id` - (Required, ForceNew) The realm in which to manage this permission.
- `name` - (Required) The name of the permission. Must be unique within the `admin-permissions` resource server. On first apply, if a permission with this name already exists it is adopted; otherwise a new one is created.
- `description` - (Optional) Description of the permission.
- `decision_strategy` - (Optional) Decision strategy. One of `UNANIMOUS`, `AFFIRMATIVE`, or `CONSENSUS`. Defaults to `UNANIMOUS`.
- `organization_ids` - (Optional) Set of organization UUIDs (`keycloak_organization.xxx.id`) this permission applies to. When omitted or empty, the permission applies to **all organizations** in the realm.
- `scopes` - (Required) Set of scopes this permission grants. Valid values: `view`, `manage`.
- `policies` - (Optional) Set of policy IDs to attach to the permission.

## Attributes Reference

In addition to the arguments above, the following computed attributes are exported:

- `permission_id` - The internal Keycloak UUID of the permission.
- `enabled` - Always `true` when the resource exists.
- `authorization_resource_server_id` - The ID of the `admin-permissions` client, which acts as the resource server for these permissions.

## Import

Organization admin permissions can be imported using `{{realmId}}/{{permissionId}}`:

```bash
terraform import keycloak_organization_admin_permissions.example my-realm/permission-uuid
```

After import, run `terraform apply` to reconcile `organization_ids`, `scopes`, and `policies` with your configuration.
