---
page_title: "keycloak_realm_attributes Resource"
---

# keycloak\_realm_attributes Resource

Allows for creating and managing Realm attributes within Keycloak.  `attributes` must not be tracked by the `realm` resource if this is used.

## Example Usage

```hcl
resource "keycloak_realm" "realm_example" {
  realm             = "realm-example"
  enabled           = true
}

resource "keycloak_realm_attributes" "realm_attributes" {
    realm_id = keycloak_realm.realm_example.id
    attributes = {
        baz = "bat"
        qux = "quux"
    }
}
```