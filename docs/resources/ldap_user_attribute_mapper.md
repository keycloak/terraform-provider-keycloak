---
page_title: "keycloak_ldap_user_attribute_mapper Resource"
---

# keycloak\_ldap\_user\_attribute\_mapper Resource

Allows for creating and managing user attribute mappers for Keycloak users
federated via LDAP.

The LDAP user attribute mapper can be used to map a single LDAP attribute
to an attribute on the Keycloak user model.

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
}

resource "keycloak_ldap_user_attribute_mapper" "ldap_user_attribute_mapper" {
  realm_id                = keycloak_realm.realm.id
  ldap_user_federation_id = keycloak_ldap_user_federation.ldap_user_federation.id
  name                    = "user-attribute-mapper"

  user_model_attribute    = "foo"
  ldap_attribute          = "bar"
}
```

## Argument Reference

- `realm_id` - (Required) The realm that this LDAP mapper will exist in.
- `ldap_user_federation_id` - (Required) The ID of the LDAP user federation provider to attach this mapper to.
- `name` - (Required) Display name of this mapper when displayed in the console.
- `user_model_attribute` - (Required) Name of the user property or attribute you want to map the LDAP attribute into.
- `ldap_attribute` - (Required) Name of the mapped attribute on the LDAP object.
- `read_only` - (Optional) When `true`, this attribute is not saved back to LDAP when the user attribute is updated in Keycloak. Defaults to `false`.
- `always_read_value_from_ldap` - (Optional) When `true`, the value fetched from LDAP will override the value stored in Keycloak. Defaults to `false`.
- `is_mandatory_in_ldap` - (Optional) When `true`, this attribute must exist in LDAP. Defaults to `false`.
- `attribute_force_default` - (Optional) When `true`, an empty default value is forced for mandatory attributes even when a default value is not specified. Defaults to `true`.
- `attribute_default_value` - (Optional) Default value to set in LDAP if `is_mandatory_in_ldap` is true and the value is empty.
- `is_binary_attribute` - (Optional) Should be true for binary LDAP attributes.

## Import

LDAP mappers can be imported using the format `{{realm_id}}/{{ldap_user_federation_id}}/{{ldap_mapper_id}}`.
The ID of the LDAP user federation provider and the mapper can be found within the Keycloak GUI, and they are typically GUIDs.

Example:

```bash
$ terraform import keycloak_ldap_user_attribute_mapper.ldap_user_attribute_mapper my-realm/af2a6ca3-e4d7-49c3-b08b-1b3c70b4b860/3d923ece-1a91-4bf7-adaf-3b82f2a12b67
```
