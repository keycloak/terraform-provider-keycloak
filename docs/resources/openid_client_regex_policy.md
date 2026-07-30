---
page_title: "keycloak_openid_client_regex_policy Resource"
---

# keycloak\_openid\_client\_regex\_policy Resource

Allows you to manage Regex policies.

Regex policies allow you to define conditions using regular expressions evaluated on attributes of either the current identity or a specific target claim.

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

resource "keycloak_openid_client_regex_policy" "test" {
  resource_server_id = keycloak_openid_client.test.resource_server_id
  realm_id           = keycloak_realm.realm.id
  name               = "regex_policy"
  decision_strategy  = "UNANIMOUS"
  logic              = "POSITIVE"
  target_claim       = "sample-claim"
  pattern            = "^sample.+$"
}
```

### Argument Reference

The following arguments are supported:

- `realm_id` - (Required) The realm this policy exists in.
- `resource_server_id` - (Required) The ID of the resource server.
- `name` - (Required) The name of the policy.
- `decision_strategy` - (Required) The decision strategy, can be one of `UNANIMOUS`, `AFFIRMATIVE`, or `CONSENSUS`.
- `target_claim` - (Required) The name of the target claim in the token.
- `pattern` - (Required) The Regex pattern.
- `target_context_attributes` - (Optional) `true` if policy should be evaluated on context attributes instead of identity attributes.
- `logic` - (Optional) The logic, can be one of `POSITIVE` or `NEGATIVE`. Defaults to `POSITIVE`.
- `type` - (Optional) The type of the policy. Defaults to `regex`.
- `description` - (Optional) A description for the authorization policy.

### Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

- `id` - Policy ID representing the Regex policy.

## Import

Regex policies can be imported using the format: `{{realmId}}/{{resourceServerId}}/{{policyId}}`.

Example:

```bash
$ terraform import keycloak_openid_client_regex_policy.test my-realm/3bd4a686-1062-4b59-97b8-e4e3f10b99da/63b3cde8-987d-4cd9-9306-1955579281d9
```
