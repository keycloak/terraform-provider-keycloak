---
page_title: "keycloak_openid_client_permissions Resource"
---

# keycloak\_openid\_client\_permissions Resource

Allows you to manage all openid client Scope Based Permissions.

This is part of a preview keycloak feature. You need to enable this feature to be able to use this resource. More
information about enabling the preview feature can be found
here: https://www.keycloak.org/securing-apps/token-exchange

When enabling Openid Client Permissions, Keycloak does several things automatically:

1. Enable Authorization on build-in realm-management client
1. Create scopes "view", "manage", "configure", "map-roles", "map-roles-client-scope", "map-roles-composite", "
   token-exchange"
1. Create a resource representing the openid client
1. Create all scope based permission for the scopes and openid client resource

If the realm-management Authorization is not enable, you have to ceate a dependency (`depends_on`) with the policy and
the openid client.

### Example Usage

```hcl
resource "keycloak_realm" "realm" {
	realm = "realm"
}

resource "keycloak_openid_client" "my_openid_client" {
	realm_id              = keycloak_realm.realm.id
	name                  = "my_openid_client"
	client_id             = "my_openid_client"
	client_secret         = "secret"
	access_type           = "CONFIDENTIAL"
	standard_flow_enabled = true
	valid_redirect_uris   = [
		"http://localhost:8080/*",
	]
}

data "keycloak_openid_client" "realm_management" {
	realm_id  = keycloak_realm.realm.id
	client_id = "realm-management"
}

resource keycloak_user test {
	realm_id = keycloak_realm.realm.id
	username = "test-user"

	email      = "test-user@fakedomain.com"
	first_name = "Testy"
	last_name  = "Tester"
}

resource keycloak_openid_client_user_policy test {
	resource_server_id = data.keycloak_openid_client.realm_management.id
	realm_id           = keycloak_realm.realm.id
	name               = "client_user_policy_test"
	users              = [
		keycloak_user.test.id
	]
	logic              = "POSITIVE"
	decision_strategy  = "UNANIMOUS"
	depends_on         = [
		keycloak_openid_client.my_openid_client
	]
}

resource "keycloak_openid_client_permissions" "my_permission" {
	realm_id  = keycloak_realm.realm.id
	client_id = keycloak_openid_client.my_openid_client.id

	view_scope {
		policies          = [
			keycloak_openid_client_user_policy.test.id,
		]
		description       = "my description"
		decision_strategy = "UNANIMOUS"
	}
}
```

### Argument Reference

The following arguments are supported:

- `realm_id` - (Required) The realm this group exists in.
- `client_id` - (Required) The id of the client that provides the role.

#### Permission Scopes

Permission scopes can be defined using the following attributes:

- `view_scope`
- `manage_scope`
- `configure_scope`
- `map_roles_scope`
- `map_roles_client_scope_scope`
- `map_roles_composite_scope`
- `token_exchange_scope`

Each of these attributes have the following schema:

- `policies` - (Optional) A list of policy IDs
- `description` - (Optional) A description for the permission scope
- `decision_strategy` - (Optional) The decision strategy, can be one of `UNANIMOUS`, `AFFIRMATIVE`, or `CONSENSUS`.

### Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

- `authorization_resource_server_id` - Resource server id representing the realm management client on which this
  permission is managed.

