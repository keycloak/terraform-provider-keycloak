---
page_title: "keycloak_hardcoded_attribute_mapper Resource"
---

# keycloak_hardcoded_attribute_mapper Resource

Allows for creating and managing hardcoded attribute mappers for Keycloak users federated via LDAP.

The user model hardcoded attribute mapper will set the specified value to the attribute.


## Example Usage

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_ldap_user_federation" "ldap_user_federation" {
  name     = "openldap"
  realm_id = keycloak_realm.realm.id

  username_ldap_attribute = "cn"
  rdn_ldap_attribute      = "cn"
  uuid_ldap_attribute     = "entryDN"
  user_object_classes     = [
    "simpleSecurityObject",
    "organizationalRole"
  ]

  connection_url  = "ldap://openldap"
  users_dn        = "dc=example,dc=org"
  bind_dn         = "cn=admin,dc=example,dc=org"
  bind_credential = "admin"

  sync_registrations = true
}

resource "keycloak_hardcoded_attribute_mapper" "email_verified" {
  realm_id                = keycloak_realm.realm.id
  ldap_user_federation_id = keycloak_ldap_user_federation.ldap_user_federation.id
  name                    = "email_verified"
  attribute_name          = "email_verified"
  attribute_value         = "true"
}
```

## Argument Reference

-   `realm_id` - (Required) The realm that this LDAP mapper will exist in.
-   `ldap_user_federation_id` - (Required) The ID of the LDAP user federation provider to attach this mapper to.
-   `name` - (Required) Display name of this mapper when displayed in the console.
-   `attribute_name` - (Required) The name of the user model attribute to set.
-   `attribute_value` - (Required) The value to set to model attribute. You can hardcode any value like 'foo'.

## Import

LDAP mappers can be imported using the format `{{realm_id}}/{{ldap_user_federation_id}}/{{attribute__mapper_id}}`.
The ID of the LDAP user federation provider and the mapper can be found within the Keycloak GUI, and they are typically GUIDs.

Example:

```bash
$ terraform import keycloak_hardcoded_attribute_mapper.email_verified my-realm/af2a6ca3-e4d7-49c3-b08b-1b3c70b4b860/3d923ece-1a91-4bf7-adaf-3b82f2a12b67
```
