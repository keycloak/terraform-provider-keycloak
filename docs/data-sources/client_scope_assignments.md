---
page_title: "keycloak_client_scope_assignments Data Source"
---

# keycloak\_client\_scope\_assignments Data Source

Use this data source to get the default and optional client-scope assignments of a realm or of a client. Omit
`client_id` to read the realm's assignments, the client scopes Keycloak assigns to every new client of that realm.

## Example Usage

This example assumes both realm assignment sets contain only scopes compatible with the OpenID Connect client.
If either set includes SAML scopes (for example, `role_list`), filter those scopes out before passing the sets to the client's assignments resource; otherwise, the apply fails protocol validation.

```hcl
data "keycloak_client_scope_assignments" "realm" {
  realm_id = "my-realm"
}

resource "keycloak_openid_client" "web" {
  realm_id    = "my-realm"
  client_id   = "web"
  access_type = "CONFIDENTIAL"
}

resource "keycloak_client_scope_assignments" "web" {
  realm_id        = "my-realm"
  client_id       = keycloak_openid_client.web.id
  default_scopes  = setunion(data.keycloak_client_scope_assignments.realm.default_scopes, ["my-scope"])
  optional_scopes = data.keycloak_client_scope_assignments.realm.optional_scopes
}
```

## Argument Reference

- `realm_id` - (Required) The realm id.
- `client_id` - (Optional) The Keycloak-generated id of an `openid-connect` or `saml` client. Omit it to read the realm's assignments.

## Attributes Reference

- `default_scopes` - (Computed) The names of the default client scopes.
- `optional_scopes` - (Computed) The names of the optional client scopes.
