---
page_title: "keycloak_workflow Data Source"
---

# keycloak\_workflow Data Source

This data source can be used to fetch properties of a Keycloak workflow for
usage with other resources.

## Example Usage

```hcl
data "keycloak_realm" "realm" {
  realm = "my-realm"
}

data "keycloak_workflow" "onboarding" {
  realm = data.keycloak_realm.realm.id
  name  = "onboarding-new-users"
}
```

## Argument Reference

- `realm` - (Required) The name of the realm this workflow exists within.
- `name` - (Required) The workflow name.

## Attributes Reference

See the docs for the [`keycloak_workflow` resource](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/workflow) for details on the exported attributes.
