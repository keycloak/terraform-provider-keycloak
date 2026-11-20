---
page_title: "keycloak_group Resource"
---

# keycloak\_group Resource

Allows for creating and managing Groups within Keycloak.

Groups provide a logical wrapping for users within Keycloak. Users within a group can share attributes and roles, and
group membership can be mapped to a claim.

Attributes can also be defined on Groups.

Groups can also be federated from external data sources, such as LDAP or Active Directory. This resource **should not**
be used to manage groups that were created this way.

## Example Usage

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_group" "parent_group" {
  realm_id = keycloak_realm.realm.id
  name     = "parent-group"
}

resource "keycloak_group" "child_group" {
  realm_id  = keycloak_realm.realm.id
  parent_id = keycloak_group.parent_group.id
  name      = "child-group"
}

resource "keycloak_group" "child_group_with_optional_attributes" {
  realm_id   = keycloak_realm.realm.id
  parent_id  = keycloak_group.parent_group.id
  name       = "child-group-with-optional-attributes"
  attributes = {
    "foo" = "bar"
    "multivalue" = "value1##value2"
  }
}
```

Organization groups can be managed by setting `organization_id`. Organization groups require Keycloak 26.6.0 or later.

```hcl
resource "keycloak_organization" "organization" {
  realm = keycloak_realm.realm.id
  name  = "my-organization"

  domain {
    name = "example.com"
  }
}

resource "keycloak_group" "organization_group" {
  realm_id        = keycloak_realm.realm.id
  organization_id = keycloak_organization.organization.id
  name            = "organization-group"
}

resource "keycloak_group" "organization_child_group" {
  realm_id        = keycloak_realm.realm.id
  organization_id = keycloak_organization.organization.id
  parent_id       = keycloak_group.organization_group.id
  name            = "organization-child-group"
}
```

## Argument Reference

- `realm_id` - (Required) The realm this group exists in.
- `organization_id` - (Optional) The organization this group exists in. If omitted, this group will be managed as a realm group. Organization groups require Keycloak 26.6.0 or later.
- `parent_id` - (Optional) The ID of this group's parent. If omitted, this group will be defined at the root level.
- `name` - (Required) The name of the group.
- `attributes` - (Optional) A map representing attributes for the group. In order to add multivalued attributes, use `##` to separate the values. Max length for each value is 255 chars

## Attributes Reference

- `path` - (Computed) The complete path of the group. For example, the child group's path in the example configuration would be `/parent-group/child-group`.

## Import

Groups can be imported using the format `{{realm_id}}/{{group_id}}`, where `group_id` is the unique ID that Keycloak
assigns to the group upon creation. This value can be found in the URI when editing this group in the GUI, and is typically a GUID.

Organization groups can be imported using the format `{{realm_id}}/{{organization_id}}/{{group_id}}`.

Example:

```bash
$ terraform import keycloak_group.child_group my-realm/934a4a4e-28bd-4703-a0fa-332df153aabd
$ terraform import keycloak_group.organization_group my-realm/9c9ef4b9-c4f2-4d17-ae03-0c6576df6c11/934a4a4e-28bd-4703-a0fa-332df153aabd
```
