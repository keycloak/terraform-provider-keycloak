---
page_title: "keycloak_ldap_custom_mapper Resource"
---

# keycloak\_ldap\_custom\_mapper Resource

Allows for creating and managing custom attribute mappers for Keycloak users federated via LDAP.

The LDAP custom mapper is implemented and deployed into Keycloak as a custom provider. This resource allows to
specify the custom id and custom implementation class of the self-implemented attribute mapper as well as additional
properties via config map.

The custom mapper should already be deployed into keycloak in order to be correctly configured.

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

resource "keycloak_ldap_custom_mapper" "custom_mapper" {
	name                    = "custom-mapper"
	realm_id                = keycloak_ldap_user_federation.openldap.realm_id
	ldap_user_federation_id = keycloak_ldap_user_federation.openldap.id

	provider_id             = "custom-provider-registered-in-keycloak"
	provider_type           = "com.example.custom.ldap.mappers.CustomMapper"

	config = {
		"attribute.name"       = "name"
		"attribute.value"      = "value"
	}
}
```

## Argument Reference

- `realm_id` - (Required) The realm that this LDAP mapper will exist in.
- `ldap_user_federation_id` - (Required) The ID of the LDAP user federation provider to attach this mapper to.
- `name` - (Required) Display name of this mapper when displayed in the console.
- `provider_id` - (Required) The id of the LDAP mapper implemented in MapperFactory.
- `provider_type` - (Required) The fully-qualified Java class name of the custom LDAP mapper.
- `config` - (Optional) A map with key / value pairs for configuring the LDAP mapper. The supported keys depend on the protocol mapper.

## Import

LDAP mappers can be imported using the format `{{realm_id}}/{{ldap_user_federation_id}}/{{ldap_mapper_id}}`.
The ID of the LDAP user federation provider and the mapper can be found within the Keycloak GUI, and they are typically GUIDs.

Example:

```bash
$ terraform import keycloak_ldap_custom_mapper.custom_mapper my-realm/af2a6ca3-e4d7-49c3-b08b-1b3c70b4b860/3d923ece-1a91-4bf7-adaf-3b82f2a12b67
```
