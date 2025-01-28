---
page_title: "keycloak_realm Data Source"
---

# keycloak\_realm Data Source

This data source can be used to fetch properties of a Keycloak realm for
usage with other resources.

## Example Usage

```hcl
data "keycloak_realm" "realm" {
    realm = "my-realm"
}

# use the data source

resource "keycloak_role" "group" {
    realm_id = data.keycloak_realm.realm.id
    name     = "group"
}

```

## Argument Reference

- `realm` - (Required) The realm name.

## Attributes Reference

See the docs for the [`keycloak_realm` resource](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/realm) for details on the exported attributes.
