---
page_title: "keycloak_organization Data Source"
---

# keycloak\_organization Data Source

This data source can be used to fetch properties of a Keycloak organization for
usage with other resources.

## Example Usage

```hcl
data "keycloak_realm" "realm" {
    realm = "my-realm"
}

data "keycloak_organization" "organization" {
    realm = data.keycloak_realm.realm.id
    name  = "my-org"
}

# use the data source

resource "keycloak_oidc_identity_provider" "realm_identity_provider" {
  realm             = data.keycloak_realm.realm.id
  alias             = "my-idp"
  authorization_url = "https://authorizationurl.com"
  client_id         = "clientID"
  client_secret     = "clientSecret"
  token_url         = "https://tokenurl.com"

  organization_id = data.keycloak_organization.organization.id
}

```

## Argument Reference

- `realm` - (Required) The name of the realm this organization exists within.
- `name` - (Required) The organization name.

## Attributes Reference

See the docs for the [`keycloak_organization` resource](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/organization) for details on the exported attributes.
