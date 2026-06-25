---
page_title: "keycloak_openid_audience_resolve_protocol_mapper Resource"
---

# keycloak\_openid\_audience\_resolve\_protocol\_mapper Resource

Allows for creating the "Audience Resolve" OIDC protocol mapper within Keycloak.

This protocol mapper is useful to avoid manual management of audiences, instead relying on the presence of client roles
to imply which audiences are appropriate for the token. See the
[Keycloak docs](https://www.keycloak.org/docs/latest/server_admin/#_audience_resolve) for more details.

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

resource "keycloak_openid_audience_resolve_protocol_mapper" "audience_mapper" {
  realm_id  = keycloak_realm.realm.id
  client_id = keycloak_openid_client.openid_client.id
  name      = "my-audience-resolve-mapper"
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

resource "keycloak_openid_audience_resolve_protocol_mapper" "audience_mapper" {
  realm_id        = keycloak_realm.realm.id
  client_scope_id = keycloak_openid_client_scope.client_scope.id
}
```

## Argument Reference

- `realm_id` - (Required) The realm this protocol mapper exists within.
- `name` - (Optional) The display name of this protocol mapper in the GUI. Defaults to "audience resolve".
- `client_id` - (Optional) The client this protocol mapper should be attached to. Conflicts with `client_scope_id`. One of `client_id` or `client_scope_id` must be specified.
- `client_scope_id` - (Optional) The client scope this protocol mapper should be attached to. Conflicts with `client_id`. One of `client_id` or `client_scope_id` must be specified.

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
$ terraform import keycloak_openid_audience_resolve_protocol_mapper.audience_mapper my-realm/client/a7202154-8793-4656-b655-1dd18c181e14/71602afa-f7d1-4788-8c49-ef8fd00af0f4
$ terraform import keycloak_openid_audience_resolve_protocol_mapper.audience_mapper my-realm/client-scope/b799ea7e-73ee-4a73-990a-1eafebe8e20a/71602afa-f7d1-4788-8c49-ef8fd00af0f4
$ terraform import keycloak_openid_audience_resolve_protocol_mapper.audience_mapper my-realm/client/a7202154-8793-4656-b655-1dd18c181e14/name/my%20protocol%20mapper
$ terraform import keycloak_openid_audience_resolve_protocol_mapper.audience_mapper my-realm/client-scope/b799ea7e-73ee-4a73-990a-1eafebe8e20a/name/my%20protocol%20mapper
```
