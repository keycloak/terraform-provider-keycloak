---
page_title: "keycloak_realm_client_policy_profile Resource"
---

# keycloak_realm_client_policy_profile Resource

Allows for managing Realm Client Policy Profiles.

## Example Usage

```hcl
resource "keycloak_realm" "realm" {
  realm = "my-realm"
}

resource "keycloak_realm_client_policy_profile" "profile" {
  name     = "my-profile"
  realm_id = keycloak_realm.realm.id

  executor {
    name = "intent-client-bind-checker"

    configuration = {
      auto-configure = true
    }
  }

  executor {
    name = "secure-session"
  }
}

```

### Attribute Arguments

- `name` - (Required) The name of the profile.
- `realm_id` - (Required) The realm id.
- `description` - (Optional) The description of the profile.
- `executor` - (Optional) An ordered list of [executors](#executor-arguments)

#### Executor Arguments

- `name` - (Required) The name of the executor. NOTE! The executor needs to exist
- `configuration` - (Optional) - A map of configuration values

## Import

This resource can be imported using the format `{{realm_id}}/{{name}}`. Keycloak does not assign an ID to client profiles, the name is the identifier within the realm.

Example:

```bash
$ terraform import keycloak_realm_client_policy_profile.profile my-realm/my-profile
```

Note: values in `configuration` are stored as strings in the state. JSON objects and arrays are stored as JSON strings, booleans and numbers as `"true"`, `"false"` or `"5"`.
