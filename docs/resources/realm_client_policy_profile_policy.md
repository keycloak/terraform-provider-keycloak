---
page_title: "keycloak_realm_client_policy_profile_policy Resource"
---

# keycloak_realm_client_policy_profile_policy Resource

Allows for managing Realm Client Policy Profile Policies.

## Example Usage

```hcl
resource "keycloak_realm" "realm" {
  realm = "my-realm"
}

resource "keycloak_realm_client_policy_profile" "profile" {
  name     = "my-profile"
  realm_id = keycloak_realm.realm.id
  description = "Some desc"

  executor {
    name = "intent-client-bind-checker"

    configuration = {
      auto-configure = "true"
    }
  }

  executor {
    name = "secret-rotation"
    configuration = {
      expiration-period = 2505600,
      rotated-expiration-period = 172800,
      remaining-rotation-period = 864000
    }
  }
}

resource "keycloak_realm_client_policy_profile_policy" "policy" {
  name        = "my-policy"
  realm_id    = keycloak_realm.realm.id
  description = "Some desc"
  profiles = [
    keycloak_realm_client_policy_profile.profile.name
  ]

  condition {
    name = "client-type"
    configuration = {
      "protocol" = "openid-connect"
    }
  }

  condition {
    name = "client-attributes"
    configuration = {
      is_negative_logic = false
      attributes        = jsonencode([{ "key" : "test-key", "value" : "test-value" }])
    }
  }
}

```

### Attribute Arguments

- `name` - (Required) The name of the policy.
- `realm_id` - (Required) The realm id.
- `description` - (Optional) The description of the policy.
- `enabled` - (Optional) When `false`, the policy is not applied. Defaults to `true`.
- `profiles` - (Required) The names of the client policy profiles that are applied when the conditions of this policy match.
- `condition` - (Optional) An ordered list of [condition](#condition-arguments)

#### Condition Arguments

- `name` - (Required) The name of the condition. NOTE! The condition needs to exist
- `configuration` - (Optional) - A map of configuration values

## Import

This resource can be imported using the format `{{realm_id}}/{{name}}`. Keycloak does not assign an ID to client policies, the name is the identifier within the realm.

Example:

```bash
$ terraform import keycloak_realm_client_policy_profile_policy.policy my-realm/my-policy
```

Note: values in `configuration` are stored as strings in the state. JSON objects and arrays are stored as JSON strings, booleans and numbers as `"true"`, `"false"` or `"5"`.
