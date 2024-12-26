---
page_title: "keycloak_realm_user_profile Resource"
---

# keycloak_realm_user_profile Resource

Allows for managing Realm User Profiles within Keycloak.

A user profile defines a schema for representing user attributes and how they are managed within a realm.

Information for Keycloak versions < 24:
The realm linked to the `keycloak_realm_user_profile` resource must have the user profile feature enabled.
It can be done via the administration UI, or by setting the `userProfileEnabled` realm attribute to `true`.

## Example Usage

```hcl
resource "keycloak_realm" "realm" {
  realm = "my-realm"
}

resource "keycloak_realm_user_profile" "userprofile" {
  realm_id = keycloak_realm.my_realm.id
  unmanaged_attribute_policy = "ENABLED"

  attribute {
    name         = "field1"
    display_name = "Field 1"
    group        = "group1"

    enabled_when_scope = ["offline_access"]

    required_for_roles  = ["user"]
    required_for_scopes = ["offline_access"]

    permissions {
      view = ["admin", "user"]
      edit = ["admin", "user"]
    }

    validator {
      name = "person-name-prohibited-characters"
    }

    validator {
      name   = "pattern"
      config = {
        pattern       = "^[a-z]+$"
        error-message = "Nope"
      }
    }

    annotations = {
      foo = "bar"
    }
  }

  attribute {
    name = "field2"

    validator {
      name   = "options"
      config = {
        options = jsonencode ( [ "opt1" ])
      }
    }

    annotations = {
      foo = jsonencode ( {"key": "val" } )
    }
  }

  group {
    name                = "group1"
    display_header      = "Group 1"
    display_description = "A first group"

    annotations = {
      foo = "bar"
      foo2 = jsonencode ( { "key": "val" } )
    }
  }

  group {
    name = "group2"
  }
}
```

## Argument Reference

- `realm_id` - (Required) The ID of the realm the user profile applies to.
- `attribute` - (Optional) An ordered list of [attributes](#attribute-arguments).
- `group` - (Optional) A list of [groups](#group-arguments).
- `unmanaged_attribute_policy` - (Optional) Unmanaged attributes are user attributes not explicitly defined in the user profile configuration. By default, unmanaged attributes are not enabled. Value could be one of `DISABLED`, `ENABLED`, `ADMIN_EDIT` or `ADMIN_VIEW`. If value is not specified it means `DISABLED`

### Attribute Arguments

- `name` - (Required) The name of the attribute.
- `display_name` - (Optional) The display name of the attribute.
- `group` - (Optional) The group that the attribute belong to.
- `enabled_when_scope` - (Optional) A list of scopes. The attribute will only be enabled when these scopes are requested by clients.
- `required_for_roles` - (Optional) A list of roles for which the attribute will be required.
- `required_for_scopes` - (Optional) A list of scopes for which the attribute will be required.
- `permissions` - (Optional) The [permissions](#permissions-arguments) configuration information.
- `validator` - (Optional) A list of [validators](#validator-arguments) for the attribute.
- `annotations` - (Optional) A map of annotations for the attribute. Values can be a String or a json object.

#### Permissions Arguments

- `edit` - (Optional) A list of profiles that will be able to edit the attribute. One of `admin`, `user`.
- `view` - (Optional) A list of profiles that will be able to view the attribute. One of `admin`, `user`.

#### Validator Arguments

- `name` - (Required) The name of the validator.
- `config` - (Optional) A map defining the configuration of the validator. Values can be a String or a json object.

### Group Arguments

- `name` - (Required) The name of the group.
- `display_header` - (Optional) The display header of the group.
- `display_description` - (Optional) The display description of the group.
- `annotations` - (Optional) A map of annotations for the group. Values can be a String or a json object.

## Import

This resource currently does not support importing.
