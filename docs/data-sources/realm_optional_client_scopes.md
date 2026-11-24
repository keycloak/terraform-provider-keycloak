---
page_title: "keycloak_realm_optional_client_scopes Data Source"
---

# keycloak\_realm\_optional\_client\_scopes Data Source

Use this data source to get the optional client scopes of a realm, the client scopes Keycloak assigns as optional scopes to every new client of that realm.

## Example Usage

```hcl
data "keycloak_realm_optional_client_scopes" "optional_scopes" {
  realm_id = "my-realm"
}
```

## Argument Reference

- `realm_id` - (Required) The realm id.

## Attributes Reference

- `optional_scopes` - (Computed) The names of the realm's optional client scopes.
