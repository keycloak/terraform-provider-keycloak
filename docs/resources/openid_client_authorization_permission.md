---
page_title: "keycloak_openid_client_authorization_permission Resource"
---

# keycloak\_openid\_client\_authorization\_permission Resource

Allows you to manage openid Client Authorization Permissions.

### Example Usage

```hcl
resource "keycloak_realm" "realm" {
	realm   = "my-realm"
	enabled = true
}

resource keycloak_openid_client test {
	client_id                = "client_id"
	realm_id                 = keycloak_realm.realm.id
	access_type              = "CONFIDENTIAL"
	service_accounts_enabled = true
	authorization {
		policy_enforcement_mode = "ENFORCING"
	}
}

data keycloak_openid_client_authorization_policy default {
	realm_id           = keycloak_realm.realm.id
	resource_server_id = keycloak_openid_client.test.resource_server_id
	name               = "default"
}

resource keycloak_openid_client_authorization_resource test {
	resource_server_id = keycloak_openid_client.test.resource_server_id
	name               = "resource_name"
	realm_id           = keycloak_realm.realm.id

	uris = [
		"/endpoint/*"
	]
}

resource keycloak_openid_client_authorization_scope test {
	resource_server_id = keycloak_openid_client.test.resource_server_id
	name               = "scope_name"
	realm_id           = keycloak_realm.realm.id
}

resource keycloak_openid_client_authorization_permission test {
	resource_server_id = keycloak_openid_client.test.resource_server_id
	realm_id           = keycloak_realm.realm.id
	name               = "permission_name"
	policies           = [data.keycloak_openid_client_authorization_policy.default.id]
	resources          = [keycloak_openid_client_authorization_resource.test.id]

}
```

### Argument Reference

The following arguments are supported:

- `realm_id` - (Required) The realm this group exists in.
- `resource_server_id` - (Required) The ID of the resource server.
- `name` - (Required) The name of the permission.
- `description` - (Optional) A description for the authorization permission.
- `decision_strategy` - (Optional) The decision strategy, can be one of `UNANIMOUS`, `AFFIRMATIVE`, or `CONSENSUS`. Defaults to `UNANIMOUS`.
- `policies` - (Optional) A list of policy IDs that must be applied to the scopes defined by this permission.
- `resources` - (Optional) A list of resource IDs that this permission must be applied to. Conflicts with `resource_type`.
- `resource_type` - (Optional) When specified, this permission will be evaluated for all instances of a given resource type. Conflicts with `resources`.
- `scopes` - (Optional) A list of scope IDs that this permission must be applied to.
- `type` - (Optional) The type of permission, can be one of `resource` or `scope`.

### Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

- `id` - Permission ID representing the permission.

## Import

Client authorization permissions can be imported using the format: `{{realmId}}/{{resourceServerId}}/{{permissionId}}`.

Example:

```bash
$ terraform import keycloak_openid_client_authorization_permission.test my-realm/3bd4a686-1062-4b59-97b8-e4e3f10b99da/63b3cde8-987d-4cd9-9306-1955579281d9
```
