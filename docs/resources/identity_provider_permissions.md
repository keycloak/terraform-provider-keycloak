---
page_title: "keycloak_identity_provider_permissions Resource"
---

# keycloak_identity_provider_permissions

Allows you to manage fine-grained permissions for a Keycloak identity provider.

This resource is part of the Fine-Grained Admin Permissions feature (`admin-fine-grained-authz`). It must be enabled on the Keycloak server. See the [`docker-compose.yml`](https://github.com/keycloak/terraform-provider-keycloak/blob/master/docker-compose.yml) for an example.

When enabling permissions for an identity provider, Keycloak automatically:

1. Enables authorization on the built-in `realm-management` client (if not already enabled).
1. Creates a resource representing the identity provider.
1. Creates scopes `view`, `manage`, and `token-exchange`.
1. Creates scope-based permissions for each of those scopes.

Note: if you only need to manage the `token-exchange` scope and want Keycloak to automatically create a client policy for it, consider using [`keycloak_identity_provider_token_exchange_scope_permission`](identity_provider_token_exchange_scope_permission.md) instead.

## Example Usage

```hcl
resource "keycloak_realm" "realm" {
  realm = "my-realm"
}

data "keycloak_openid_client" "realm_management" {
  realm_id  = keycloak_realm.realm.id
  client_id = "realm-management"
}

resource "keycloak_openid_client_permissions" "realm_management_permission" {
  realm_id  = keycloak_realm.realm.id
  client_id = data.keycloak_openid_client.realm_management.id
}

resource "keycloak_oidc_identity_provider" "idp" {
  realm             = keycloak_realm.realm.id
  alias             = "my-idp"
  authorization_url = "https://idp.example.com/auth"
  token_url         = "https://idp.example.com/token"
  client_id         = "my-client-id"
  client_secret     = "my-client-secret"
}

resource "keycloak_group" "group" {
  realm_id = keycloak_realm.realm.id
  name     = "idp-admins"
}

resource "keycloak_openid_client_group_policy" "policy" {
  realm_id           = keycloak_realm.realm.id
  resource_server_id = data.keycloak_openid_client.realm_management.id
  name               = "idp-admins-policy"
  groups {
    id              = keycloak_group.group.id
    path            = keycloak_group.group.path
    extend_children = false
  }
  logic             = "POSITIVE"
  decision_strategy = "UNANIMOUS"
  depends_on = [
    keycloak_openid_client_permissions.realm_management_permission,
  ]
}

resource "keycloak_identity_provider_permissions" "idp_permissions" {
  realm_id       = keycloak_realm.realm.id
  provider_alias = keycloak_oidc_identity_provider.idp.alias

  manage_scope {
    policies          = [keycloak_openid_client_group_policy.policy.id]
    description       = "Allow idp-admins group to manage this IDP"
    decision_strategy = "UNANIMOUS"
  }
}
```

## Argument Reference

- `realm_id` - (Required) The realm in which to manage fine-grained identity provider permissions.
- `provider_alias` - (Required) The alias of the identity provider.

Each scope block is optional. When specified, the block configures the policy attached to that scope's permission.

- `view_scope` - (Optional) Policies that decide if an admin can view this identity provider.
- `manage_scope` - (Optional) Policies that decide if an admin can manage this identity provider.
- `token_exchange_scope` - (Optional) Policies that decide which clients can use this identity provider for token exchange.

Each scope block supports:

- `policies` - (Optional) List of policy IDs to attach to the permission.
- `description` - (Optional) Description of the permission.
- `decision_strategy` - (Optional) Decision strategy of the permission. Can be `UNANIMOUS`, `AFFIRMATIVE`, or `CONSENSUS`.

## Attributes Reference

In addition to the arguments above, the following computed attributes are exported:

- `enabled` - When `true`, fine-grained permissions are enabled for this identity provider. This will always be `true`.
- `authorization_resource_server_id` - The ID of the `realm-management` client, which acts as the resource server for these permissions.

## Import

Identity provider permissions can be imported using the format `{{realmId}}/{{providerAlias}}`:

```bash
terraform import keycloak_identity_provider_permissions.idp_permissions my-realm/my-idp
```
