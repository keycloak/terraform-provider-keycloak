---
page_title: "keycloak_authentication_execution Data Source"
---

# keycloak\_authentication\_execution Data Source

This data source can be used to fetch the ID of an authentication execution within Keycloak.

## Example Usage

```hcl
resource "keycloak_realm" "realm" {
    realm   = "my-realm"
    enabled = true
}

data "keycloak_authentication_execution" "browser_auth_cookie" {
  realm_id          = keycloak_realm.realm.id
  parent_flow_alias = "browser"
  provider_id       = "auth-cookie"
}
```

## Argument Reference

- `realm_id` - (Required) The realm the authentication execution exists in.
- `parent_flow_alias` - (Required) The alias of the flow this execution is attached to.
- `provider_id` - (Required) The name of the provider. This can be found by experimenting with the GUI and looking at HTTP requests within the network tab of your browser's development tools. This was previously known as the "authenticator".

## Attributes Reference

- `id` - (Computed) The unique ID of the authentication execution, which can be used as an argument to other resources supported by this provider.
- `priority` - (Computed) The authenticator priority.
