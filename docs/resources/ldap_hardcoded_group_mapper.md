---
page_title: "keycloak_ldap_hardcoded_group_mapper Resource"
---

# keycloak\_ldap\_hardcoded\_group\_mapper Resource

Allows for creating and managing hardcoded group mappers for Keycloak users federated via LDAP.

The LDAP hardcoded group mapper will grant a specified Keycloak group to each Keycloak user linked with LDAP.

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

resource "keycloak_group" "realm_group" {
    realm_id    = keycloak_realm.realm.id
    name        = "my-group"
}

resource "keycloak_ldap_hardcoded_group_mapper" "assign_group_to_users" {
    realm_id                = keycloak_realm.realm.id
    ldap_user_federation_id = keycloak_ldap_user_federation.ldap_user_federation.id
    name                    = "assign-group-to-users"
    group                   = keycloak_group.realm_group.name
}
```

## Argument Reference

- `realm_id` - (Required) The realm that this LDAP mapper will exist in.
- `ldap_user_federation_id` - (Required) The ID of the LDAP user federation provider to attach this mapper to.
- `name` - (Required) Display name of this mapper when displayed in the console.
- `group` - (Required) The name of the group which should be assigned to the users.

## Import

LDAP mappers can be imported using the format `{{realm_id}}/{{ldap_user_federation_id}}/{{ldap_mapper_id}}`.
The ID of the LDAP user federation provider and the mapper can be found within the Keycloak GUI, and they are typically GUIDs.

Example:

```bash
$ terraform import keycloak_ldap_hardcoded_group_mapper.assign_group_to_users my-realm/af2a6ca3-e4d7-49c3-b08b-1b3c70b4b860/3d923ece-1a91-4bf7-adaf-3b82f2a12b67
```
