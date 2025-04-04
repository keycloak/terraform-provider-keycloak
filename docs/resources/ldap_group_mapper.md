---
page_title: "keycloak_ldap_group_mapper Resource"
---

# keycloak\_ldap\_group\_mapper

Allows for creating and managing group mappers for Keycloak users federated via LDAP.

The LDAP group mapper can be used to map an LDAP user's groups from some DN to Keycloak groups. This group mapper will also
create the groups within Keycloak if they do not already exist.

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

resource "keycloak_ldap_group_mapper" "ldap_group_mapper" {
  realm_id                = keycloak_realm.realm.id
  ldap_user_federation_id = keycloak_ldap_user_federation.ldap_user_federation.id
  name                    = "group-mapper"

  ldap_groups_dn                 = "dc=example,dc=org"
  group_name_ldap_attribute      = "cn"
  group_object_classes           = [
    "groupOfNames"
  ]
  membership_attribute_type      = "DN"
  membership_ldap_attribute      = "member"
  membership_user_ldap_attribute = "cn"
  memberof_ldap_attribute        = "memberOf"
}
```

## Argument Reference

- `realm_id` - (Required) The realm that this LDAP mapper will exist in.
- `ldap_user_federation_id` - (Required) The ID of the LDAP user federation provider to attach this mapper to.
- `name` - (Required) Display name of this mapper when displayed in the console.
- `ldap_groups_dn` - (Required) The LDAP DN where groups can be found.
- `group_name_ldap_attribute` - (Required) The name of the LDAP attribute that is used in group objects for the name and RDN of the group. Typically `cn`.
- `group_object_classes` - (Required) List of strings representing the object classes for the group. Must contain at least one.
- `preserve_group_inheritance` - (Optional) When `true`, group inheritance will be propagated from LDAP to Keycloak. When `false`, all LDAP groups will be propagated as top level groups within Keycloak.
- `ignore_missing_groups` - (Optional) When `true`, missing groups in the hierarchy will be ignored.
- `membership_ldap_attribute` - (Required) The name of the LDAP attribute that is used for membership mappings.
- `membership_attribute_type` - (Optional) Can be one of `DN` or `UID`. Defaults to `DN`.
- `membership_user_ldap_attribute` - (Required) The name of the LDAP attribute on a user that is used for membership mappings.
- `groups_ldap_filter` - (Optional) When specified, adds a custom filter to be used when querying for groups. Must start with `(` and end with `)`.
- `mode` - (Optional) Can be one of `READ_ONLY`, `LDAP_ONLY` or `IMPORT`. Defaults to `READ_ONLY`.
- `user_roles_retrieve_strategy` - (Optional) Can be one of `LOAD_GROUPS_BY_MEMBER_ATTRIBUTE`, `GET_GROUPS_FROM_USER_MEMBEROF_ATTRIBUTE`, or `LOAD_GROUPS_BY_MEMBER_ATTRIBUTE_RECURSIVELY`. Defaults to `LOAD_GROUPS_BY_MEMBER_ATTRIBUTE`.
- `memberof_ldap_attribute` - (Optional) Specifies the name of the LDAP attribute on the LDAP user that contains the groups the user is a member of. Defaults to `memberOf`.
- `mapped_group_attributes` - (Optional) Array of strings representing attributes on the LDAP group which will be mapped to attributes on the Keycloak group.
- `drop_non_existing_groups_during_sync` - (Optional) When `true`, groups that no longer exist within LDAP will be dropped in Keycloak during sync. Defaults to `false`.
- `groups_path` - (Optional) Keycloak group path the LDAP groups are added to. For example if value `/Applications/App1` is used, then LDAP groups will be available in Keycloak under group `App1`, which is the child of top level group `Applications`. The configured group path must already exist in Keycloak when creating this mapper.

## Import

LDAP mappers can be imported using the format `{{realm_id}}/{{ldap_user_federation_id}}/{{ldap_mapper_id}}`.
The ID of the LDAP user federation provider and the mapper can be found within the Keycloak GUI, and they are typically GUIDs.

Example:

```bash
$ terraform import keycloak_ldap_group_mapper.ldap_group_mapper my-realm/af2a6ca3-e4d7-49c3-b08b-1b3c70b4b860/3d923ece-1a91-4bf7-adaf-3b82f2a12b67
```
