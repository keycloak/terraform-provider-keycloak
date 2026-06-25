---
page_title: "keycloak_openid_user_client_role_protocol_mapper Resource"
---

# keycloak\_openid\_user\_client\_role\_protocol\_mapper Resource

Allows for creating and managing user client role protocol mappers within Keycloak.

User client role protocol mappers allow you to define a claim containing the list of a client roles.

Protocol mappers can be defined for a single client, or they can be defined for a client scope which can be shared between
multiple different clients.

## Example Usage (Client)

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_openid_client" "openid_client" {
  realm_id  = keycloak_realm.realm.id
  client_id = "client"

  name    = "client"
  enabled = true

  access_type         = "CONFIDENTIAL"
  valid_redirect_uris = [
    "http://localhost:8080/openid-callback"
  ]
}

resource "keycloak_openid_user_client_role_protocol_mapper" "user_client_role_mapper" {
  realm_id   = keycloak_realm.realm.id
  client_id  = keycloak_openid_client.openid_client.id
  name       = "user-client-role-mapper"
  claim_name = "foo"
}
```

## Example Usage (Client Scope)

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_openid_client_scope" "client_scope" {
  realm_id = keycloak_realm.realm.id
  name     = "client-scope"
}

resource "keycloak_openid_user_client_role_protocol_mapper" "user_client_role_mapper" {
  realm_id        = keycloak_realm.realm.id
  client_scope_id = keycloak_openid_client_scope.client_scope.id
  name            = "user-client-role-mapper"
  claim_name      = "foo"
}
```

## Argument Reference

- `realm_id` - (Required) The realm this protocol mapper exists within.
- `name` - (Required) The display name of this protocol mapper in the GUI.
- `claim_name` - (Required) The name of the claim to insert into a token.
- `client_id` - (Optional) The client this protocol mapper should be attached to. Conflicts with `client_scope_id`. One of `client_id` or `client_scope_id` must be specified.
- `client_scope_id` - (Optional) The client scope this protocol mapper should be attached to. Conflicts with `client_id`. One of `client_id` or `client_scope_id` must be specified.
- `claim_value_type` - (Optional) The claim type used when serializing JSON tokens. Can be one of `String`, `JSON`, `long`, `int`, or `boolean`. Defaults to `String`.
- `multivalued` - (Optional) Indicates if attribute supports multiple values. If true, then the list of all values of this attribute will be set as claim. If false, then just first value will be set as claim. Defaults to `false`.
- `client_id_for_role_mappings` - (Optional) The Client ID for role mappings. Just client roles of this client will be added to the token. If this is unset, client roles of all clients will be added to the token.
- `client_role_prefix` - (Optional) A prefix for each Client Role.
- `add_to_id_token` - (Optional) Indicates if the property should be added as a claim to the id token. Defaults to `true`.
- `add_to_access_token` - (Optional) Indicates if the property should be added as a claim to the access token. Defaults to `true`.
- `add_to_userinfo` - (Optional) Indicates if the property should be added as a claim to the UserInfo response body. Defaults to `true`.

## Import

Protocol mappers can be imported using one of the following formats:
- Client: `{{realm_id}}/client/{{client_keycloak_id}}/{{protocol_mapper_id}}`
- Client Scope: `{{realm_id}}/client-scope/{{client_scope_keycloak_id}}/{{protocol_mapper_id}}`

As an alternative to importing by Keycloak protocol mapper ID, you can import by protocol mapper name.

When using name-based import:
- Use the literal `name` segment in the import path.
- URL-encode the mapper name using path encoding (for example, `my protocol mapper` becomes `my%20protocol%20mapper`).

Supported name-based formats:
- Client: `{{realm_id}}/client/{{client_keycloak_id}}/name/{{url_encoded_protocol_mapper_name}}`
- Client Scope: `{{realm_id}}/client-scope/{{client_scope_keycloak_id}}/name/{{url_encoded_protocol_mapper_name}}`

Example:

```bash
$ terraform import keycloak_openid_user_client_role_protocol_mapper.user_client_role_mapper my-realm/client/a7202154-8793-4656-b655-1dd18c181e14/71602afa-f7d1-4788-8c49-ef8fd00af0f4
$ terraform import keycloak_openid_user_client_role_protocol_mapper.user_client_role_mapper my-realm/client-scope/b799ea7e-73ee-4a73-990a-1eafebe8e20a/71602afa-f7d1-4788-8c49-ef8fd00af0f4
$ terraform import keycloak_openid_user_client_role_protocol_mapper.user_client_role_mapper my-realm/client/a7202154-8793-4656-b655-1dd18c181e14/name/my%20protocol%20mapper
$ terraform import keycloak_openid_user_client_role_protocol_mapper.user_client_role_mapper my-realm/client-scope/b799ea7e-73ee-4a73-990a-1eafebe8e20a/name/my%20protocol%20mapper
```
