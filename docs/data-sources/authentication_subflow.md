---
page_title: "keycloak_authentication_subflow Data Source"
---

# keycloak\_authentication\_subflow Data Source

This data source can be used to fetch the details of an authentication subflow within Keycloak.

An authentication subflow is a nested flow within a parent authentication flow that groups related authentication steps together.

## Example Usage

### Lookup by Alias (Human-readable)

```hcl
resource "keycloak_realm" "realm" {
    realm   = "my-realm"
    enabled = true
}

resource "keycloak_authentication_flow" "my_flow" {
  realm_id = keycloak_realm.realm.id
  alias    = "my-custom-flow"
}

resource "keycloak_authentication_subflow" "my_subflow" {
  realm_id          = keycloak_realm.realm.id
  parent_flow_alias = keycloak_authentication_flow.my_flow.alias
  alias             = "my-subflow"
  provider_id       = "basic-flow"
}

data "keycloak_authentication_subflow" "subflow" {
  realm_id          = keycloak_realm.realm.id
  parent_flow_alias = keycloak_authentication_flow.my_flow.alias
  alias             = "my-subflow"
}

output "subflow_id" {
  value = data.keycloak_authentication_subflow.subflow.id
}
```

### Lookup by ID (Direct)

```hcl
data "keycloak_authentication_subflow" "subflow" {
  realm_id          = "my-realm-id"
  parent_flow_alias = "browser"
  id                = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
}

output "subflow_alias" {
  value = data.keycloak_authentication_subflow.subflow.alias
}
```

## Argument Reference

The following arguments are supported:

- `realm_id` - (Required) The realm the authentication subflow exists in.
- `parent_flow_alias` - (Required) The alias of the parent authentication flow.
- `id` - (Optional) The unique ID of the authentication subflow. Either `id` or `alias` must be specified.
- `alias` - (Optional) The alias of the authentication subflow. Either `id` or `alias` must be specified.

~> **Note:** You must specify either `id` or `alias`, but not both. Use `id` for direct lookup by GUID, or `alias` for human-readable lookup by name.

## Attributes Reference

In addition to the arguments listed above, the following attributes are exported:

- `id` - The unique ID of the authentication subflow.
- `alias` - The alias of the subflow.
- `provider_id` - The provider ID for the subflow (e.g., `basic-flow`, `form-flow`, or `client-flow`).
- `description` - The description of the subflow.
- `requirement` - The requirement setting for the subflow. Can be one of `REQUIRED`, `ALTERNATIVE`, `OPTIONAL`, `CONDITIONAL`, or `DISABLED`.
- `priority` - (Keycloak 25+) The priority of the subflow within its parent flow.
