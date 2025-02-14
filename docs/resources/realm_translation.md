---
page_title: "keycloak_realm_translation Resource"
---

# keycloak_realm_tranlsation Resource

Allows for managing Realm Translations overrides within Keycloak.

A translation defines a schema for representing a locale with a map of key/value pairs and how they are managed within a realm.

Note: whilst you can provide translations for unsupported locales, they will not take effect until they are defined within the realm resource.

## Example Usage

```hcl
resource "keycloak_realm" "realm" {
  realm = "my-realm"
}

resource "keycloak_realm_translation" "realm_translation" {
  realm_id = keycloak_realm.my_realm.id
  locale = "de"
  translations = {
    "Hello" : "Hallo"
  }
}
```

## Argument Reference

- `realm_id` - (Required) The ID of the realm the user profile applies to.
- `locale` - (Required) The locale (language code) the translations apply to.
- `translations` - (Optional) A map of translation keys to values.


## Import

This resource does not currently support importing.
