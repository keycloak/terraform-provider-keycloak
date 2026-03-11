---
page_title: "keycloak_openid_client_authorization_resource Resource"
---

# keycloak\_openid\_client\_authorization\_resource Resource

Allows you to manage openid Client Authorization Resources.

Authorization resources represent the protected resources in your application. Each resource can have associated scopes, URIs, and attributes.

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

resource "keycloak_openid_client_authorization_scope" "read_scope" {
  resource_server_id = keycloak_openid_client.test.resource_server_id
  realm_id           = keycloak_realm.realm.id
  name               = "read"
}

resource "keycloak_openid_client_authorization_scope" "write_scope" {
  resource_server_id = keycloak_openid_client.test.resource_server_id
  realm_id           = keycloak_realm.realm.id
  name               = "write"
}

resource "keycloak_openid_client_authorization_resource" "test" {
  resource_server_id = keycloak_openid_client.test.resource_server_id
  realm_id           = keycloak_realm.realm.id
  name               = "my_resource"
  display_name       = "My Resource"

  uris = [
    "/api/resource/*",
    "/api/resource/**"
  ]

  scopes = [
    keycloak_openid_client_authorization_scope.read_scope.name,
    keycloak_openid_client_authorization_scope.write_scope.name,
  ]

  type = "http://example.com/resource-type"

  attributes = {
    "key1" = "value1,value2"
    "key2" = "value3"
  }
}
```

### Argument Reference

The following arguments are supported:

- `realm_id` - (Required) The realm this resource exists in.
- `resource_server_id` - (Required) The ID of the resource server.
- `name` - (Required) The name of the resource.
- `display_name` - (Optional) The display name of the resource.
- `uris` - (Optional) A set of URIs that this resource represents.
- `icon_uri` - (Optional) An icon URI for the resource.
- `owner_managed_access` - (Optional) When `true`, this resource supports user-managed access. Defaults to `false`.
- `scopes` - (Optional) A set of scope names that this resource uses.
- `type` - (Optional) The type of this resource (e.g., `urn:myapp:resources:default`).
- `attributes` - (Optional) A map of attributes for the resource. Values can be comma-separated lists.

### Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

- `id` - Resource ID representing the authorization resource.

## Import

Client authorization resources can be imported using the format: `{{realmId}}/{{resourceServerId}}/{{authorizationResourceId}}`.

Example:

```bash
$ terraform import keycloak_openid_client_authorization_resource.test my-realm/3bd4a686-1062-4b59-97b8-e4e3f10b99da/63b3cde8-987d-4cd9-9306-1955579281d9
```
