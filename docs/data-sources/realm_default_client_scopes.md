---
page_title: "keycloak_realm_default_client_scopes Data Source"
---

# keycloak\_realm\_default\_client\_scopes Data Source

Use this data source to get the default client scopes of a realm, the client scopes Keycloak assigns as default scopes to every new client of that realm.

## Example Usage

```hcl
data "keycloak_realm_default_client_scopes" "default_scopes" {
  realm_id = "my-realm"
}
```

## Argument Reference

- `realm_id` - (Required) The realm id.

## Attributes Reference

- `default_scopes` - (Computed) The names of the realm's default client scopes.
