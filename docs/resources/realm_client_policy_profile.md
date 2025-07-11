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

- `name` - (Required) The name of the attribute.
- `realm_id` - (Required) The realm id.
- `executor` - (Optional) An ordered list of [executors](#executor-arguments)

#### Executor Arguments

- `name` - (Required) The name of the executor. NOTE! The executor needs to exist
- `configuration` - (Optional) - A map of configuration values

## Import

This resource currently does not support importing.
