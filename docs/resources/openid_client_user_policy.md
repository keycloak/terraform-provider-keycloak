---
page_title: "keycloak_openid_client_user_policy Resource"
---

# keycloak\_openid\_client\_user\_policy Resource

Allows you to manage user policies.

User policies allow you to define conditions based on specific users. This is useful when you need to grant access to individual users rather than based on roles or groups.

## Example Usage

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_openid_client" "test" {
  client_id                = "client_id"
  realm_id                 = keycloak_realm.realm.id
  access_type              = "CONFIDENTIAL"
  service_accounts_enabled = true
  authorization {
    policy_enforcement_mode = "ENFORCING"
  }
}

resource "keycloak_user" "alice" {
  realm_id = keycloak_realm.realm.id
  username = "alice"
  enabled  = true

  email      = "alice@example.com"
  first_name = "Alice"
  last_name  = "Smith"
}

resource "keycloak_user" "bob" {
  realm_id = keycloak_realm.realm.id
  username = "bob"
  enabled  = true

  email      = "bob@example.com"
  first_name = "Bob"
  last_name  = "Jones"
}

resource "keycloak_openid_client_user_policy" "test" {
  resource_server_id = keycloak_openid_client.test.resource_server_id
  realm_id           = keycloak_realm.realm.id
  name               = "user_policy"
  decision_strategy  = "UNANIMOUS"
  logic              = "POSITIVE"

  users = [
    keycloak_user.alice.id,
    keycloak_user.bob.id,
  ]
}
```

### Argument Reference

The following arguments are supported:

- `realm_id` - (Required) The realm this policy exists in.
- `resource_server_id` - (Required) The ID of the resource server.
- `name` - (Required) The name of the policy.
- `decision_strategy` - (Required) The decision strategy, can be one of `UNANIMOUS`, `AFFIRMATIVE`, or `CONSENSUS`.
- `users` - (Required) A list of user IDs that this policy applies to.
- `logic` - (Optional) The logic, can be one of `POSITIVE` or `NEGATIVE`. Defaults to `POSITIVE`.
- `description` - (Optional) A description for the authorization policy.

### Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

- `id` - Policy ID representing the user policy.

## Import

User policies can be imported using the format: `{{realmId}}/{{resourceServerId}}/{{policyId}}`.

Example:

```bash
$ terraform import keycloak_openid_client_user_policy.test my-realm/3bd4a686-1062-4b59-97b8-e4e3f10b99da/63b3cde8-987d-4cd9-9306-1955579281d9
```
