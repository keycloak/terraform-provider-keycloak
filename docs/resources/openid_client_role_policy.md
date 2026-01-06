---
page_title: "keycloak_openid_client_role_policy Resource"
---

# keycloak\_openid\_client\_role\_policy Resource

Allows you to manage role policies.

Role policies allow you to define conditions based on user role assignments. You can specify whether all roles must be present or just one.

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

resource "keycloak_role" "admin_role" {
  realm_id = keycloak_realm.realm.id
  name     = "admin"
}

resource "keycloak_role" "user_role" {
  realm_id = keycloak_realm.realm.id
  name     = "user"
}

resource "keycloak_openid_client_role_policy" "test" {
  resource_server_id = keycloak_openid_client.test.resource_server_id
  realm_id           = keycloak_realm.realm.id
  name               = "role_policy"
  decision_strategy  = "UNANIMOUS"
  logic              = "POSITIVE"
  type               = "role"

  role {
    id       = keycloak_role.admin_role.id
    required = true
  }

  role {
    id       = keycloak_role.user_role.id
    required = false
  }
}
```

### Argument Reference

The following arguments are supported:

- `realm_id` - (Required) The realm this policy exists in.
- `resource_server_id` - (Required) The ID of the resource server.
- `name` - (Required) The name of the policy.
- `type` - (Required) The type of policy. Must be `role`.
- `role` - (Required) A list of roles [role](#role-arguments). At least one role must be defined.
- `decision_strategy` - (Optional) The decision strategy, can be one of `UNANIMOUS`, `AFFIRMATIVE`, or `CONSENSUS`.
- `logic` - (Optional) The logic, can be one of `POSITIVE` or `NEGATIVE`. Defaults to `POSITIVE`.
- `fetch_roles` - (Optional) When `true`, roles will be fetched from the user's claims. Available in Keycloak 25+.
- `description` - (Optional) A description for the authorization policy.

### Role Arguments

- `id` - (Required) The ID of the role.
- `required` - (Required) When `true`, this role must be present for the policy to grant access.

### Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

- `id` - Policy ID representing the role policy.

## Import

Role policies can be imported using the format: `{{realmId}}/{{resourceServerId}}/{{policyId}}`.

Example:

```bash
$ terraform import keycloak_openid_client_role_policy.test my-realm/3bd4a686-1062-4b59-97b8-e4e3f10b99da/63b3cde8-987d-4cd9-9306-1955579281d9
```
