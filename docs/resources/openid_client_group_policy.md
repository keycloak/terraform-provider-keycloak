---
page_title: "keycloak_openid_client_group_policy Resource"
---

# keycloak\_openid\_client\_group\_policy Resource

Allows you to manage group policies.

Group policies allow you to define conditions based on group membership. You can specify whether child groups should be included in the evaluation.

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

resource "keycloak_group" "group1" {
  realm_id = keycloak_realm.realm.id
  name     = "group1"
}

resource "keycloak_group" "group2" {
  realm_id = keycloak_realm.realm.id
  name     = "group2"
}

resource "keycloak_openid_client_group_policy" "test" {
  resource_server_id = keycloak_openid_client.test.resource_server_id
  realm_id           = keycloak_realm.realm.id
  name               = "group_policy"
  decision_strategy  = "UNANIMOUS"
  logic              = "POSITIVE"

  groups {
    id              = keycloak_group.group1.id
    path            = keycloak_group.group1.path
    extend_children = false
  }

  groups {
    id              = keycloak_group.group2.id
    path            = keycloak_group.group2.path
    extend_children = true
  }
}
```

### Argument Reference

The following arguments are supported:

- `realm_id` - (Required) The realm this policy exists in.
- `resource_server_id` - (Required) The ID of the resource server.
- `name` - (Required) The name of the policy.
- `decision_strategy` - (Required) The decision strategy, can be one of `UNANIMOUS`, `AFFIRMATIVE`, or `CONSENSUS`.
- `logic` - (Optional) The logic, can be one of `POSITIVE` or `NEGATIVE`. Defaults to `POSITIVE`.
- `groups` - (Required) A list of groups [group](#group-arguments). At least one group must be defined.
- `groups_claim` - (Optional) The name of the claim in the token that contains the group information.
- `description` - (Optional) A description for the authorization policy.

### Group Arguments

- `id` - (Required) The ID of the group.
- `path` - (Required) The path of the group.
- `extend_children` - (Required) When `true`, the policy will also apply to all child groups of this group.

### Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

- `id` - Policy ID representing the group policy.

## Import

Group policies can be imported using the format: `{{realmId}}/{{resourceServerId}}/{{policyId}}`.

Example:

```bash
$ terraform import keycloak_openid_client_group_policy.test my-realm/3bd4a686-1062-4b59-97b8-e4e3f10b99da/63b3cde8-987d-4cd9-9306-1955579281d9
```
