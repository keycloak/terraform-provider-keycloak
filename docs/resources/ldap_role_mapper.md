---
page_title: "keycloak_ldap_role_mapper Resource"
---

# keycloak\_ldap\_role\_mapper

Allows for creating and managing role mappers for Keycloak users federated via LDAP.

The LDAP group mapper can be used to map an LDAP user's roles from some DN to Keycloak roles.

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

resource "keycloak_ldap_role_mapper" "ldap_role_mapper" {
  realm_id                = keycloak_realm.realm.id
  ldap_user_federation_id = keycloak_ldap_user_federation.ldap_user_federation.id
  name                    = "role-mapper"

  ldap_roles_dn                 = "dc=example,dc=org"
  role_name_ldap_attribute      = "cn"
  role_object_classes           = [
    "groupOfNames"
  ]
  membership_attribute_type      = "DN"
  membership_ldap_attribute      = "member"
  membership_user_ldap_attribute = "cn"
  user_roles_retrieve_strategy   = "GET_ROLES_FROM_USER_MEMBEROF_ATTRIBUTE"
  memberof_ldap_attribute        = "memberOf"
}
```

## Argument Reference

- `realm_id` - (Required) The realm that this LDAP mapper will exist in.
- `ldap_user_federation_id` - (Required) The ID of the LDAP user federation provider to attach this mapper to.
- `name` - (Required) Display name of this mapper when displayed in the console.
- `ldap_roles_dn` - (Required) The LDAP DN where roles can be found.
- `role_name_ldap_attribute` - (Required) The name of the LDAP attribute that is used in role objects for the name and RDN of the role. Typically `cn`.
- `role_object_classes` - (Required) List of strings representing the object classes for the role. Must contain at least one.
- `membership_ldap_attribute` - (Required) The name of the LDAP attribute that is used for membership mappings.
- `membership_attribute_type` - (Optional) Can be one of `DN` or `UID`. Defaults to `DN`.
- `membership_user_ldap_attribute` - (Required) The name of the LDAP attribute on a user that is used for membership mappings.
- `roles_ldap_filter` - (Optional) When specified, adds a custom filter to be used when querying for roles. Must start with `(` and end with `)`.
- `mode` - (Optional) Can be one of `READ_ONLY`, `LDAP_ONLY` or `IMPORT`. Defaults to `READ_ONLY`.
- `user_roles_retrieve_strategy` - (Optional) Can be one of `LOAD_ROLES_BY_MEMBER_ATTRIBUTE`, `GET_ROLES_FROM_USER_MEMBEROF_ATTRIBUTE`, or `LOAD_ROLES_BY_MEMBER_ATTRIBUTE_RECURSIVELY`. Defaults to `LOAD_ROLES_BY_MEMBER_ATTRIBUTE`.
- `memberof_ldap_attribute` - (Optional) Specifies the name of the LDAP attribute on the LDAP user that contains the roles the user has. Defaults to `memberOf`. This is only used when
- `use_realm_roles_mapping` - (Optional) When `true`, LDAP role mappings will be mapped to realm roles within Keycloak. Defaults to `true`.
- `client_id` - (Optional) When specified, LDAP role mappings will be mapped to client role mappings tied to this client ID. Can only be set if `use_realm_roles_mapping` is `false`.

## Import

LDAP mappers can be imported using the format `{{realm_id}}/{{ldap_user_federation_id}}/{{ldap_mapper_id}}`.
The ID of the LDAP user federation provider and the mapper can be found within the Keycloak GUI, and they are typically GUIDs.

Example:

```bash
$ terraform import keycloak_ldap_role_mapper.ldap_role_mapper my-realm/af2a6ca3-e4d7-49c3-b08b-1b3c70b4b860/3d923ece-1a91-4bf7-adaf-3b82f2a12b67
```
