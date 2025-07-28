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
  name        = "my-profile"
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
      is-negative-logic = false
      attributes        = jsonencode([{ "key" : "test-key", "value" : "test-value" }])
    }
  }
}

```

### Attribute Arguments

- `name` - (Required) The name of the attribute.
- `realm_id` - (Required) The realm id.
- `condition` - (Optional) An ordered list of [condition](#condition-arguments)

#### Condition Arguments

- `name` - (Required) The name of the executor. NOTE! The executor needs to exist
- `configuration` - (Optional) - A map of configuration values

## Import

This resource currently does not support importing.
