---
page_title: "keycloak_openid_client_aggregate_policy Resource"
---

# keycloak\_openid\_client\_aggregate\_policy Resource

Allows you to manage aggregate policies.

Aggregate policies combine multiple policies into a single policy, allowing you to reuse existing policies to build more complex authorization logic.

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

resource "keycloak_openid_client_role_policy" "role_policy" {
  resource_server_id = keycloak_openid_client.test.resource_server_id
  realm_id           = keycloak_realm.realm.id
  name               = "role_policy"
  decision_strategy  = "UNANIMOUS"
  logic              = "POSITIVE"

  role {
    id       = keycloak_role.test.id
    required = true
  }
}

resource "keycloak_openid_client_user_policy" "user_policy" {
  resource_server_id = keycloak_openid_client.test.resource_server_id
  realm_id           = keycloak_realm.realm.id
  name               = "user_policy"
  decision_strategy  = "UNANIMOUS"
  logic              = "POSITIVE"
  users              = [keycloak_user.test.id]
}

resource "keycloak_openid_client_aggregate_policy" "test" {
  resource_server_id = keycloak_openid_client.test.resource_server_id
  realm_id           = keycloak_realm.realm.id
  name               = "aggregate_policy"
  decision_strategy  = "AFFIRMATIVE"
  logic              = "POSITIVE"

  policies = [
    keycloak_openid_client_role_policy.role_policy.id,
    keycloak_openid_client_user_policy.user_policy.id,
  ]
}
```

### Argument Reference

The following arguments are supported:

- `realm_id` - (Required) The realm this policy exists in.
- `resource_server_id` - (Required) The ID of the resource server.
- `name` - (Required) The name of the policy.
- `decision_strategy` - (Required) The decision strategy, can be one of `UNANIMOUS`, `AFFIRMATIVE`, or `CONSENSUS`.
- `logic` - (Optional) The logic, can be one of `POSITIVE` or `NEGATIVE`. Defaults to `POSITIVE`.
- `policies` - (Required) A list of policy IDs to aggregate.
- `description` - (Optional) A description for the authorization policy.

### Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

- `id` - Policy ID representing the aggregate policy.

## Import

Aggregate policies can be imported using the format: `{{realmId}}/{{resourceServerId}}/{{policyId}}`.

Example:

```bash
$ terraform import keycloak_openid_client_aggregate_policy.test my-realm/3bd4a686-1062-4b59-97b8-e4e3f10b99da/63b3cde8-987d-4cd9-9306-1955579281d9
```
