---
page_title: "keycloak_required_action Resource"
---

# keycloak\_required\_action Resource

Allows for creating and managing required actions within Keycloak.

[Required actions](https://www.keycloak.org/docs/latest/server_admin/#con-required-actions_server_administration_guide) specify actions required before the first login of all new users.


## Example Usage

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_required_action" "required_action" {
  realm_id = keycloak_realm.realm.realm
  alias    = "UPDATE_PASSWORD"
  enabled  = true
  name     = "Update Password"
  config = {
    max_auth_age = "600"
  }
}
```

## Argument Reference

- `realm_id` - (Required) The realm the required action exists in.
- `alias` - (Required) The alias of the action to attach as a required action.
- `name` - (Optional) The name of the required action.
- `enabled` - (Optional) When `false`, the required action is not enabled for new users. Defaults to `false`.
- `default_action` - (Optional) When `true`, the required action is set as the default action for new users. Defaults to `false`.
- `priority`- (Optional) The priority of the required action.
- `config`- (Optional) The configuration. Keys are specific to each configurable required action and not checked when applying.

## Import

Authentication executions can be imported using the formats: `{{realm}}/{{alias}}`.

Example:

```bash
$ terraform import keycloak_required_action.required_action my-realm/my-default-action-alias
```