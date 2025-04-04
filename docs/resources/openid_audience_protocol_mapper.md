---
page_title: "keycloak_openid_audience_protocol_mapper Resource"
---

# keycloak\_openid\_audience\_protocol\_mapper Resource

Allows for creating and managing audience protocol mappers within Keycloak.

Audience protocol mappers allow you to add audiences to the `aud` claim within issued tokens. The audience can be a custom
string, or it can be mapped to the ID of a pre-existing client.

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

resource "keycloak_openid_audience_protocol_mapper" "audience_mapper" {
  realm_id  = keycloak_realm.realm.id
  client_id = keycloak_openid_client.openid_client.id
  name      = "audience-mapper"

  included_custom_audience = "foo"
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
  name     = "test-client-scope"
}

resource "keycloak_openid_audience_protocol_mapper" "audience_mapper" {
  realm_id        = keycloak_realm.realm.id
  client_scope_id = keycloak_openid_client_scope.client_scope.id
  name            = "audience-mapper"

  included_custom_audience = "foo"
}
```

## Argument Reference

- `realm_id` - (Required) The realm this protocol mapper exists within.
- `name` - (Required) The display name of this protocol mapper in the GUI.
- `client_id` - (Optional) The client this protocol mapper should be attached to. Conflicts with `client_scope_id`. One of `client_id` or `client_scope_id` must be specified.
- `client_scope_id` - (Optional) The client scope this protocol mapper should be attached to. Conflicts with `client_id`. One of `client_id` or `client_scope_id` must be specified.
- `included_client_audience` - (Optional) A client ID to include within the token's `aud` claim. Conflicts with `included_custom_audience`. One of `included_client_audience` or `included_custom_audience` must be specified.
- `included_custom_audience` - (Optional) A custom audience to include within the token's `aud` claim. Conflicts with `included_client_audience`. One of `included_client_audience` or `included_custom_audience` must be specified.
- `add_to_id_token` - (Optional) Indicates if the audience should be included in the `aud` claim for the id token. Defaults to `true`.
- `add_to_access_token` - (Optional) Indicates if the audience should be included in the `aud` claim for the id token. Defaults to `true`.

## Import

Protocol mappers can be imported using one of the following formats:
- Client: `{{realm_id}}/client/{{client_keycloak_id}}/{{protocol_mapper_id}}`
- Client Scope: `{{realm_id}}/client-scope/{{client_scope_keycloak_id}}/{{protocol_mapper_id}}`

Example:

```bash
$ terraform import keycloak_openid_audience_protocol_mapper.audience_mapper my-realm/client/a7202154-8793-4656-b655-1dd18c181e14/71602afa-f7d1-4788-8c49-ef8fd00af0f4
$ terraform import keycloak_openid_audience_protocol_mapper.audience_mapper my-realm/client-scope/b799ea7e-73ee-4a73-990a-1eafebe8e20a/71602afa-f7d1-4788-8c49-ef8fd00af0f4
```
