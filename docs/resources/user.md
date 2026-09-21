---
page_title: "keycloak_user Resource"
---

# keycloak\_user Resource

Allows for creating and managing Users within Keycloak.

This resource was created primarily to enable the acceptance tests for the `keycloak_group` resource. Creating users within
Keycloak is not recommended. Instead, users should be federated from external sources by configuring user federation providers
or identity providers.

> **NOTICE:** This resource now supports [write-only arguments](https://developer.hashicorp.com/terraform/language/resources/ephemeral#write-only-arguments)
> for the initial password via the new arguments `initial_password.value_wo` and `initial_password.value_wo_version`. Using
> write-only arguments prevents sensitive values from being stored in plan and state files. You cannot use
> `initial_password.value_wo` and `initial_password.value_wo_version` alongside `initial_password.value` as this will result
> in a validation error due to conflicts.
>
> For backward compatibility, the behavior of the original `initial_password.value` argument remains unchanged: it is only
> respected during user creation. Unlike `initial_password.value`, bumping `initial_password.value_wo_version` resets the
> password of an existing user.

## Example Usage

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_user" "user" {
  realm_id = keycloak_realm.realm.id
  username = "bob"
  enabled  = true

  email      = "bob@domain.com"
  first_name = "Bob"
  last_name  = "Bobson"
}

resource "keycloak_user" "user_with_initial_password" {
  realm_id   = keycloak_realm.realm.id
  username   = "alice"
  enabled    = true

  email      = "alice@domain.com"
  first_name = "Alice"
  last_name  = "Aliceberg"

  attributes = {
    foo = "bar"
    multivalue = "value1##value2"
  }

  initial_password {
    value     = "some password"
    temporary = true
  }
}
```

## Example Usage with `initial_password.value_wo`

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

ephemeral "random_password" "user_password" {
  length           = 16
  override_special = "!#$%&*()-_=+[]{}<>:?"
}

resource "keycloak_user" "user_with_initial_password" {
  realm_id = keycloak_realm.realm.id
  username = "alice"
  enabled  = true

  initial_password {
    value_wo         = ephemeral.random_password.user_password.result
    value_wo_version = "version1"
    temporary        = true
  }
}
```

## Argument Reference

- `realm_id` - (Required) The realm this user belongs to.
- `username` - (Required) The unique username of this user.
- `initial_password` - (Optional) When given, the user's initial password will be set. Exactly one of `value` and `value_wo` must be given.
  - `value` - (Optional) The initial password. This argument is only respected during initial user creation; later changes to it are ignored. Conflicts with `value_wo` and `value_wo_version`.
  - `value_wo` - (Optional, Write-Only) The initial password. This is a write-only argument and Terraform does not store it in state or plan files. Conflicts with `value`. Required when using `value_wo_version`.
  - `value_wo_version` - (Optional) Functions as a flag and/or trigger to indicate Terraform when to use the input value in `value_wo` to execute a Create or Update operation. The value of this argument is stored in the state and plan files. Changing it resets the password of an existing user. Conflicts with `value`. Required when using `value_wo`.
  - `temporary` - (Optional) If set to `true`, the initial password is set up for renewal on first use. Default to `false`.
- `enabled` - (Optional) When false, this user cannot log in. Defaults to `true`.
- `email` - (Optional) The user's email.
- `email_verified` - (Optional) Whether the email address was validated or not. Default to `false`.
- `first_name` - (Optional) The user's first name.
- `last_name` - (Optional) The user's last name.
- `attributes` - (Optional) A map representing attributes for the user. In order to add multivalue attributes, use `##` to seperate the values. Max length for each value is 255 chars
- `required_actions` - (Optional) A list of required user actions.
- `federated_identity` - (Optional) When specified, the user will be linked to a federated identity provider. Refer to the [federated user example](https://github.com/keycloak/terraform-provider-keycloak/blob/master/example/federated_user_example.tf) for more details.
  - `identity_provider` - (Required) The name of the identity provider
  - `user_id` - (Required) The ID of the user defined in the identity provider
  - `user_name` - (Required) The username of the user defined in the identity provider
- `import` - (Optional) When `true`, the user with the specified `username` is assumed to already exist, and it will be imported into state instead of being created. This attribute is useful when dealing with users that Keycloak creates automatically during realm creation, such as `admin`. Note, that the user will not be removed during destruction if `import` is `true`.

## Import

Users can be imported using the format `{{realm_id}}/{{user_id}}`, where `user_id` is the unique ID that Keycloak
assigns to the user upon creation. This value can be found in the GUI when editing the user.

Example:

```bash
$ terraform import keycloak_user.user my-realm/60c3f971-b1d3-4b3a-9035-d16d7540a5e4
```
