---
page_title: "keycloak_openid_client_authorization_scope Resource"
---

# keycloak\_openid\_client\_authorization\_scope Resource

Allows you to manage openid Client Authorization Scopes.

Authorization scopes represent the actions that can be performed on resources. They are used in permissions to define what operations are allowed.

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

resource "keycloak_openid_client_authorization_scope" "read" {
  resource_server_id = keycloak_openid_client.test.resource_server_id
  realm_id           = keycloak_realm.realm.id
  name               = "read"
  display_name       = "Read Access"
  icon_uri           = "https://example.com/icons/read.png"
}

resource "keycloak_openid_client_authorization_scope" "write" {
  resource_server_id = keycloak_openid_client.test.resource_server_id
  realm_id           = keycloak_realm.realm.id
  name               = "write"
  display_name       = "Write Access"
}

resource "keycloak_openid_client_authorization_scope" "delete" {
  resource_server_id = keycloak_openid_client.test.resource_server_id
  realm_id           = keycloak_realm.realm.id
  name               = "delete"
  display_name       = "Delete Access"
}
```

### Argument Reference

The following arguments are supported:

- `realm_id` - (Required) The realm this scope exists in.
- `resource_server_id` - (Required) The ID of the resource server.
- `name` - (Required) The name of the scope.
- `display_name` - (Optional) The display name of the scope.
- `icon_uri` - (Optional) An icon URI for the scope.

### Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

- `id` - Scope ID representing the authorization scope.

## Import

Client authorization scopes can be imported using the format: `{{realmId}}/{{resourceServerId}}/{{authorizationScopeId}}`.

Example:

```bash
$ terraform import keycloak_openid_client_authorization_scope.test my-realm/3bd4a686-1062-4b59-97b8-e4e3f10b99da/63b3cde8-987d-4cd9-9306-1955579281d9
```
