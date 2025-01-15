---
page_title: "keycloak_identity_provider_token_exchange_scope_permission Resource"
---

# keycloak\_identity\_provider\_token\_exchange\_scope\_permission Resource

Allows you to manage Identity Provider "Token exchange" Scope Based Permissions.

This is part of a preview keycloak feature. You need to enable this feature to be able to use this resource.
More information about enabling the preview feature can be found here: https://www.keycloak.org/docs/latest/securing_apps/index.html#_token-exchange

When enabling Identity Provider Permissions, Keycloak does several things automatically:
1. Enable Authorization on build-in realm-management client
1. Create a "token-exchange" scope
1. Create a resource representing the identity provider
1. Create a scope based permission for the "token-exchange" scope and identity provider resource

The only thing that is missing is a policy set on the permission.
As the policy lives within the context of the realm-management client, you cannot create a policy resource and link to from with your _.tf_ file. This would also cause an implicit cycle dependency.
Thus, the only way to manage this in terraform is to create and manage the policy internally from within this terraform resource itself.
At the moment only a client policy type is supported. The client policy will automatically be created for the `clients` parameter.

## Example Usage

```hcl
resource "keycloak_realm" "token_exchange_realm" {
  realm   = "token-exchange_destination_realm"
  enabled = true
}

resource "keycloak_oidc_identity_provider" "token_exchange_my_oidc_idp" {
  realm              = keycloak_realm.token_exchange_realm.id
  alias              = "myIdp"
  authorization_url  = "http://localhost:8080/auth/realms/someRealm/protocol/openid-connect/auth"
  token_url          = "http://localhost:8080/auth/realms/someRealm/protocol/openid-connect/token"
  client_id          = "clientId"
  client_secret      = "secret"
  default_scopes     = "openid"
}

resource "keycloak_openid_client" "token-exchange_webapp_client" {
  realm_id              = keycloak_realm.token_exchange_realm.id
  name                  = "webapp_client"
  client_id             = "webapp_client"
  client_secret         = "secret"
  description           = "a webapp client on the destination realm"
  access_type           = "CONFIDENTIAL"
  standard_flow_enabled = true
  valid_redirect_uris = [
    "http://localhost:8080/*",
  ]
}

//relevant part
resource "keycloak_identity_provider_token_exchange_scope_permission" "oidc_idp_permission" {
  realm_id       = keycloak_realm.token_exchange_realm.id
  provider_alias = keycloak_oidc_identity_provider.token_exchange_my_oidc_idp.alias
  policy_type    = "client"
  clients        = [
    keycloak_openid_client.token-exchange_webapp_client.id
  ]
}
```

## Argument Reference

- `realm_id` - (Required) The realm that the identity provider exists in.
- `provider_alias` - (Required) Alias of the identity provider.
- `policy_type` - (Optional) Defaults to "client" This is also the only value policy type supported by this provider.
- `clients` - (Required) A list of IDs of the clients for which a policy will be created and set on scope based token exchange permission.

## Attributes Reference

- `policy_id` - (Computed) Policy ID that will be set on the scope based token exchange permission automatically created by enabling permissions on the reference identity provider.
- `authorization_resource_server_id` - (Computed) Resource server ID representing the realm management client on which this permission is managed.
- `authorization_idp_resource_id` - (Computed) Resource ID representing the identity provider, this automatically created by keycloak.
- `authorization_token_exchange_scope_permission_id` - (Computed) Permission ID representing the Permission with scope 'Token Exchange' and the resource 'authorization_idp_resource_id', this automatically created by keycloak, the policy ID will be set on this permission.


## Import

This resource can be imported using the format `{{realm_id}}/{{provider_alias}}`, where `provider_alias` is the alias that
you assign to the identity provider upon creation.

Example:

```bash
$ terraform import keycloak_identity_provider_token_exchange_scope_permission.oidc_idp_permission my-realm/myIdp
```
