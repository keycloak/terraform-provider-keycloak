---
page_title: "keycloak_group_memberships Resource"
---

# keycloak\_group\_memberships Resource

Allows for managing a Keycloak group's members.

Note that this resource attempts to be an **authoritative** source over group members. When this resource takes control
over a group's members, users that are manually added to the group will be removed, and users that are manually removed
from the group will be added upon the next run of `terraform apply`.

Also note that you should not use `keycloak_group_memberships` with a group has been assigned as a default group via
`keycloak_default_groups`.

This resource **should not** be used to control membership of a group that has its members federated from an external
source via group mapping.

To non-exclusively manage the group's of a user, see the [`keycloak_user_groups` resource][1]

This resource paginates its data loading on refresh by 50 items.

## Example Usage

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_group" "group" {
  realm_id = keycloak_realm.realm.id
  name     = "my-group"
}

resource "keycloak_user" "user" {
  realm_id = keycloak_realm.realm.id
  username = "my-user"
}

resource "keycloak_group_memberships" "group_members" {
  realm_id = keycloak_realm.realm.id
  group_id = keycloak_group.group.id

  members  = [
    keycloak_user.user.username
  ]
}
```


## Organization Group Example

Organization groups require Keycloak 26.6 or newer. Set `organizations_enabled = true` on the
`keycloak_realm.realm` resource before applying this example. This example uses the realm and user defined above.

```hcl
resource "keycloak_organization" "organization" {
  realm = keycloak_realm.realm.id
  name  = "my-organization"

  domain {
    name = "organization.example.com"
  }
}

resource "keycloak_group" "organization_group" {
  realm_id        = keycloak_realm.realm.id
  organization_id = keycloak_organization.organization.id
  name            = "my-organization-group"
}

resource "keycloak_organization_memberships" "organization_members" {
  realm_id        = keycloak_realm.realm.id
  organization_id = keycloak_organization.organization.id
  members         = [keycloak_user.user.username]
}

resource "keycloak_group_memberships" "organization_group_members" {
  realm_id        = keycloak_realm.realm.id
  organization_id = keycloak_organization.organization.id
  group_id        = keycloak_group.organization_group.id
  members         = [keycloak_user.user.username]

  depends_on = [keycloak_organization_memberships.organization_members]
}
```

## Argument Reference

- `realm_id` - (Required) The realm this group exists in.
- `organization_id` - (Optional) The ID of the organization this group belongs to. Omit for realm-level groups.
- `group_id` - (Required) The ID of the group this resource should manage memberships for.
- `members` - (Required) A list of usernames that belong to this group.

## Import

This resource can be imported using the format `{{realm_id}}/{{group_id}}`, or, for a group that belongs to an
organization, `{{realm_id}}/{{organization_id}}/{{group_id}}`.

Examples:

```bash
# realm-level group
$ terraform import keycloak_group_memberships.group_members my-realm/18cc6b87-2ce7-4e59-bdc8-b9d49ec98a94

# organization-scoped group
$ terraform import keycloak_group_memberships.group_members my-realm/b258402a-5e1b-4e53-b05a-6e5b9c4d1e77/18cc6b87-2ce7-4e59-bdc8-b9d49ec98a94
```

[1]: https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/group_memberships
