---
page_title: "keycloak_ldap_user_federation Resource"
---

# keycloak\_ldap\_user\_federation Resource

Allows for creating and managing LDAP user federation providers within Keycloak.

Keycloak can use an LDAP user federation provider to federate users to Keycloak
from a directory system such as LDAP or Active Directory. Federated users
will exist within the realm and will be able to log in to clients. Federated
users can have their attributes defined using mappers.

## Example Usage

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_ldap_user_federation" "ldap_user_federation" {
  name     = "openldap"
  realm_id = keycloak_realm.realm.id
  enabled  = true

  username_ldap_attribute = "cn"
  rdn_ldap_attribute      = "cn"
  uuid_ldap_attribute     = "entryDN"
  user_object_classes     = [
    "simpleSecurityObject",
    "organizationalRole"
  ]
  connection_url          = "ldap://openldap"
  users_dn                = "dc=example,dc=org"
  bind_dn                 = "cn=admin,dc=example,dc=org"
  bind_credential         = "admin"

  connection_timeout = "5s"
  read_timeout       = "10s"

  kerberos {
    kerberos_realm   = "FOO.LOCAL"
    server_principal = "HTTP/host.foo.com@FOO.LOCAL"
    key_tab          = "/etc/host.keytab"
  }
}
```

## Argument Reference

- `realm_id` - (Required) The realm that this provider will provide user federation for.
- `name` - (Required) Display name of the provider when displayed in the console.
- `enabled` - (Optional) When `false`, this provider will not be used when performing queries for users. Defaults to `true`.
- `priority` - (Optional) Priority of this provider when looking up users. Lower values are first. Defaults to `0`.
- `import_enabled` - (Optional) When `true`, LDAP users will be imported into the Keycloak database. Defaults to `true`.
- `edit_mode` - (Optional) Can be one of `READ_ONLY`, `WRITABLE`, or `UNSYNCED`. `UNSYNCED` allows user data to be imported but not synced back to LDAP. Defaults to `READ_ONLY`.
- `sync_registrations` - (Optional) When `true`, newly created users will be synced back to LDAP. Defaults to `false`.
- `vendor` - (Optional) Can be one of `OTHER`, `EDIRECTORY`, `AD`, `RHDS`, or `TIVOLI`. When this is selected in the GUI, it provides reasonable defaults for other fields. When used with the Keycloak API, this attribute does nothing, but is still required. Defaults to `OTHER`.
- `username_ldap_attribute` - (Required) Name of the LDAP attribute to use as the Keycloak username.
- `rdn_ldap_attribute` - (Required) Name of the LDAP attribute to use as the relative distinguished name.
- `uuid_ldap_attribute` - (Required) Name of the LDAP attribute to use as a unique object identifier for objects in LDAP.
- `user_object_classes` - (Required) Array of all values of LDAP objectClass attribute for users in LDAP. Must contain at least one.
- `connection_url` - (Required) Connection URL to the LDAP server.
- `users_dn` - (Required) Full DN of LDAP tree where your users are.
- `bind_dn` - (Optional) DN of LDAP admin, which will be used by Keycloak to access LDAP server. This attribute must be set if `bind_credential` is set.
- `bind_credential` - (Optional) Password of LDAP admin. This attribute must be set if `bind_dn` is set.
- `custom_user_search_filter` - (Optional) Additional LDAP filter for filtering searched users. Must begin with `(` and end with `)`.
- `search_scope` - (Optional) Can be one of `ONE_LEVEL` or `SUBTREE`:
    - `ONE_LEVEL`: Only search for users in the DN specified by `user_dn`.
    - `SUBTREE`: Search entire LDAP subtree.
- `start_tls` - (Optional) When `true`, Keycloak will encrypt the connection to LDAP using STARTTLS, which will disable connection pooling.
- `connection_pooling` - (Optional) When `true`, LDAP connection pooling is enabled. Defaults to `false`.
- `use_password_modify_extended_op` - (Optional) When `true`, use the LDAPv3 Password Modify Extended Operation (RFC-3062).
- `validate_password_policy` - (Optional) When `true`, Keycloak will validate passwords using the realm policy before updating it.
- `trust_email` - (Optional) If enabled, email provided by this provider is not verified even if verification is enabled for the realm.
- `use_truststore_spi` - (Optional) Can be one of `ALWAYS`, `ONLY_FOR_LDAPS`, or `NEVER`:
    - `ALWAYS` - Always use the truststore SPI for LDAP connections.
    - `NEVER` - Never use the truststore SPI for LDAP connections.
    - `ONLY_FOR_LDAPS` - Only use the truststore SPI if your LDAP connection uses the ldaps protocol.
- `connection_timeout` - (Optional) LDAP connection timeout in the format of a [Go duration string](https://golang.org/pkg/time/#Duration.String).
- `read_timeout` - (Optional) LDAP read timeout in the format of a [Go duration string](https://golang.org/pkg/time/#Duration.String).
- `pagination` - (Optional) When true, Keycloak assumes the LDAP server supports pagination. Defaults to `true`.
- `batch_size_for_sync` - (Optional) The number of users to sync within a single transaction. Defaults to `1000`.
- `full_sync_period` - (Optional) How frequently Keycloak should sync all LDAP users, in seconds. Omit this property to disable periodic full sync.
- `changed_sync_period` - (Optional) How frequently Keycloak should sync changed LDAP users, in seconds. Omit this property to disable periodic changed users sync.
- `cache_policy` - (Optional) **Deprecated** Can be one of `DEFAULT`, `EVICT_DAILY`, `EVICT_WEEKLY`, `MAX_LIFESPAN`, or `NO_CACHE`. Defaults to `DEFAULT`.
- `cache` - (Optional) A block containing the cache settings.
  - `policy` - (Optional) Can be one of `DEFAULT`, `EVICT_DAILY`, `EVICT_WEEKLY`, `MAX_LIFESPAN`, or `NO_CACHE`. Defaults to `DEFAULT`.
  - `max_lifespan` - (Optional) Max lifespan of cache entry (duration string).
  - `eviction_day` - (Optional) Day of the week the entry will become invalid on
  - `eviction_hour` - (Optional) Hour of day the entry will become invalid on.
  - `eviction_minute` - (Optional) Minute of day the entry will become invalid on.
- `kerberos` - (Optional) A block containing the kerberos settings.
  - `kerberos_realm` - (Required) The name of the kerberos realm, e.g. FOO.LOCAL.
  - `server_principal` - (Required) The kerberos server principal, e.g. 'HTTP/host.foo.com@FOO.LOCAL'.
  - `key_tab` - (Required) Path to the kerberos keytab file on the server with credentials of the service principal.
  - `use_kerberos_for_password_authentication` - (Optional) Use kerberos login module instead of ldap service api. Defaults to `false`.
- `delete_default_mappers` - (Optional) When true, the provider will delete the default mappers which are normally created by Keycloak when creating an LDAP user federation provider. Defaults to `false`.
## Import

LDAP user federation providers can be imported using the format `{{realm_id}}/{{ldap_user_federation_id}}`.
The ID of the LDAP user federation provider can be found within the Keycloak GUI and is typically a GUID:

```bash
$ terraform import keycloak_ldap_user_federation.ldap_user_federation my-realm/af2a6ca3-e4d7-49c3-b08b-1b3c70b4b860
```
