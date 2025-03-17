---
page_title: "keycloak_openid_client_authorization_client_scope_policy Resource"
---

# keycloak\_openid\_client\_authorization\_client\_scope\_policy Resource

Allows you to manage openid Client Authorization Client Scope type Policies.

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

resource "keycloak_openid_client_scope" "test1" {
    realm_id    = keycloak_realm.realm.id
    name        = "test1"
    description = "test1"
}

resource "keycloak_openid_client_scope" "test2" {
    realm_id    = keycloak_realm.realm.id
    name        = "test2"
    description = "test2"
}

resource "keycloak_openid_client_authorization_client_scope_policy" "test" {
    resource_server_id = keycloak_openid_client.test.resource_server_id
    realm_id           = keycloak_realm.realm.id
    name               = "test_policy_single"
    description        = "test"
    decision_strategy  = "AFFIRMATIVE"
    logic              = "POSITIVE"

    scope {
      id       = keycloak_openid_client_scope.test1.id
      required = false
    }
}

resource "keycloak_openid_client_authorization_client_scope_policy" "test_multiple" {
    resource_server_id = keycloak_openid_client.test.resource_server_id
    realm_id           = keycloak_realm.realm.id
    name               = "test_policy_multiple"
    description        = "test"
    decision_strategy  = "AFFIRMATIVE"
    logic              = "POSITIVE"

    scope {
      id       = keycloak_openid_client_scope.test1.id
      required = false
    }

    scope {
      id       = keycloak_openid_client_scope.test2.id
      required = true
    }
}

```

### Argument Reference

The following arguments are supported:

- `realm_id` - (Required) The realm this group exists in.
- `resource_server_id` - (Required) The ID of the resource server.
- `name` - (Required) The name of the policy.
- `description` - (Optional) A description for the authorization policy.
- `decision_strategy` - (Optional) The decision strategy, can be one of `UNANIMOUS`, `AFFIRMATIVE`, or `CONSENSUS`. Defaults to `UNANIMOUS`.
- `logic` - (Optional) The logic, can be one of `POSITIVE` or `NEGATIVE`. Defaults to `POSITIVE`.
- `scope` - An client scope to add [client scope](#scope-arguments). At least one should be defined.

### Scope Arguments

- `id` - (Required) Id of client scope.
- `required` - (Optional) When `true`, then this client scope will be set as required. Defaults to `false`.

### Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

- `id` - Policy ID representing the policy.

## Import

Client authorization policies can be imported using the format: `{{realmId}}/{{resourceServerId}}/{{policyId}}`.

Example:

```bash
$ terraform import keycloak_openid_client_authorization_client_scope_policy.test my-realm/3bd4a686-1062-4b59-97b8-e4e3f10b99da/63b3cde8-987d-4cd9-9306-1955579281d9
```
