---
page_title: "keycloak_openid_client Data Source"
---

# keycloak\_openid\_client Data Source

This data source can be used to fetch properties of a Keycloak OpenID client for usage with other resources.

## Example Usage

```hcl
data "keycloak_openid_client" "realm_management" {
  realm_id  = "my-realm"
  client_id = "realm-management"
}

# use the data source
data "keycloak_role" "admin" {
  realm_id  = "my-realm"
  client_id = data.keycloak_openid_client.realm_management.id
  name      = "realm-admin"
}
```

## Argument Reference

- `realm_id` - (Required) The realm id.
- `client_id` - (Required) The client id (not its unique ID).

## Attributes Reference

See the docs for the `keycloak_openid_client` resource for details on the exported attributes.

`client_secret` is empty for Keycloak's internal clients, whose secret the server does not serve: `realm-management`, the master realm's `-realm` clients, and the realm's `admin-permissions` client (Keycloak 26.8.2 and later). The data source still serves such a client's `id`, which fine-grained admin permission policies need as their `resource_server_id`.
