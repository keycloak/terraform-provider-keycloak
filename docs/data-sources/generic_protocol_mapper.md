---
page_title: "keycloak_generic_protocol_mapper Data Source"
---

# keycloak_generic_protocol_mapper Data Source

This data source can be used to fetch properties of a protocol mapper attached to a client or client scope.

This is useful for importing auto-created protocol mappers (such as the organization membership mapper)
into Terraform state by looking up their ID by name.

## Example Usage

```hcl
data "keycloak_openid_client_scope" "organization" {
  realm_id = "my-realm"
  name     = "organization"
}

data "keycloak_generic_protocol_mapper" "organization_membership" {
  realm_id        = "my-realm"
  client_scope_id = data.keycloak_openid_client_scope.organization.id
  name            = "organization"
}

# Use in an import block
import {
  to = keycloak_generic_protocol_mapper.organization_membership
  id = "my-realm/client-scope/${data.keycloak_openid_client_scope.organization.id}/${data.keycloak_generic_protocol_mapper.organization_membership.id}"
}
```

## Argument Reference

- `realm_id` - (Required) The realm this protocol mapper exists within.
- `name` - (Required) The name of the protocol mapper to look up.
- `client_id` - (Optional) The client this protocol mapper is attached to. Conflicts with `client_scope_id`.
- `client_scope_id` - (Optional) The client scope this protocol mapper is attached to. Conflicts with `client_id`.

One of `client_id` or `client_scope_id` must be specified.

## Attribute Reference

- `id` - The unique ID of the protocol mapper.
- `protocol` - The protocol of the mapper (e.g., `openid-connect` or `saml`).
- `protocol_mapper` - The type of the protocol mapper (e.g., `oidc-organization-membership-mapper`).
- `config` - A map of configuration values for the protocol mapper.
