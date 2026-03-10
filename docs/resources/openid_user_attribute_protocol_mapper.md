---
page_title: "keycloak_openid_user_attribute_protocol_mapper Resource"
---

# keycloak\_openid\_user\_attribute\_protocol\_mapper Resource

Allows for creating and managing user attribute protocol mappers within Keycloak.

User attribute protocol mappers allow you to map custom attributes defined for a user within Keycloak to a claim in a token.

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

resource "keycloak_openid_user_attribute_protocol_mapper" "user_attribute_mapper" {
  realm_id       = keycloak_realm.realm.id
  client_id      = keycloak_openid_client.openid_client.id
  name           = "user-attribute-mapper"

  user_attribute = "foo"
  claim_name     = "bar"
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

resource "keycloak_openid_user_attribute_protocol_mapper" "user_attribute_mapper" {
  realm_id        = keycloak_realm.realm.id
  client_scope_id = keycloak_openid_client_scope.client_scope.id
  name            = "user-attribute-mapper"

  user_attribute  = "foo"
  claim_name      = "bar"
}
```

## Argument Reference

- `realm_id` - (Required) The realm this protocol mapper exists within.
- `name` - (Required) The display name of this protocol mapper in the GUI.
- `user_attribute` - (Required) The custom user attribute to map a claim for.
- `claim_name` - (Required) The name of the claim to insert into a token.
- `client_id` - (Optional) The client this protocol mapper should be attached to. Conflicts with `client_scope_id`. One of `client_id` or `client_scope_id` must be specified.
- `client_scope_id` - (Optional) The client scope this protocol mapper should be attached to. Conflicts with `client_id`. One of `client_id` or `client_scope_id` must be specified.
- `claim_value_type` - (Optional) The claim type used when serializing JSON tokens. Can be one of `String`, `JSON`, `long`, `int`, or `boolean`. Defaults to `String`.
- `multivalued` - (Optional) Indicates whether this attribute is a single value or an array of values. Defaults to `false`.
- `add_to_id_token` - (Optional) Indicates if the attribute should be added as a claim to the id token. Defaults to `true`.
- `add_to_access_token` - (Optional) Indicates if the attribute should be added as a claim to the access token. Defaults to `true`.
- `add_to_userinfo` - (Optional) Indicates if the attribute should be added as a claim to the UserInfo response body. Defaults to `true`.
- `add_to_token_introspection` - (Optional) Indicates if the attribute should be added as a claim to the token introspection response. Defaults to `true`.
- `aggregate_attributes`- (Optional) Indicates whether this attribute is a single value or an array of values. Defaults to `false`.

## Import

Protocol mappers can be imported using one of the following formats:
- Client: `{{realm_id}}/client/{{client_keycloak_id}}/{{protocol_mapper_id}}`
- Client Scope: `{{realm_id}}/client-scope/{{client_scope_keycloak_id}}/{{protocol_mapper_id}}`

Example:

```bash
$ terraform import keycloak_openid_user_attribute_protocol_mapper.user_attribute_mapper my-realm/client/a7202154-8793-4656-b655-1dd18c181e14/71602afa-f7d1-4788-8c49-ef8fd00af0f4
$ terraform import keycloak_openid_user_attribute_protocol_mapper.user_attribute_mapper my-realm/client-scope/b799ea7e-73ee-4a73-990a-1eafebe8e20a/71602afa-f7d1-4788-8c49-ef8fd00af0f4
```
