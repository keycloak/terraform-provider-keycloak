---
page_title: "keycloak_ldap_msad_lds_user_account_control_mapper Resource"
---

# keycloak\_ldap\_msad\_lds\_user\_account\_control\_mapper Resource

Allows for creating and managing MSAD-LDS user account control mappers for Keycloak
users federated via LDAP.

The MSAD-LDS (Microsoft Active Directory Lightweight Directory Service) user account control mapper is specific
to LDAP user federation providers that are pulling from AD-LDS, and it can propagate
AD-LDS user state to Keycloak in order to enforce settings like expired passwords
or disabled accounts.

## Example Usage

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_ldap_user_federation" "ldap_user_federation" {
  name     = "ad"
  realm_id = keycloak_realm.realm.id

  username_ldap_attribute = "cn"
  rdn_ldap_attribute      = "cn"
  uuid_ldap_attribute     = "objectGUID"
  user_object_classes     = [
    "person",
    "organizationalPerson",
    "user"
  ]
  connection_url          = "ldap://my-ad-server"
  users_dn                = "dc=example,dc=org"
  bind_dn                 = "cn=admin,dc=example,dc=org"
  bind_credential         = "admin"
    referral                     = "ignore"
}

resource "keycloak_ldap_msad_lds_user_account_control_mapper" "msad_lds_user_account_control_mapper" {
  realm_id                 = keycloak_realm.realm.id
  ldap_user_federation_id  = keycloak_ldap_user_federation.ldap_user_federation.id
  name                     = "msad-lds-user-account-control-mapper"
}
```

## Argument Reference

- `realm_id` - (Required) The realm that this LDAP mapper will exist in.
- `ldap_user_federation_id` - (Required) The ID of the LDAP user federation provider to attach this mapper to.
- `name` - (Required) Display name of this mapper when displayed in the console.

## Import

LDAP mappers can be imported using the format `{{realm_id}}/{{ldap_user_federation_id}}/{{ldap_mapper_id}}`.
The ID of the LDAP user federation provider and the mapper can be found within the Keycloak GUI, and they are typically GUIDs.

Example:

```bash
$ terraform import keycloak_ldap_msad_lds_user_account_control_mapper.msad_lds_user_account_control_mapper my-realm/af2a6ca3-e4d7-49c3-b08b-1b3c70b4b860/3d923ece-1a91-4bf7-adaf-3b82f2a12b67
```
