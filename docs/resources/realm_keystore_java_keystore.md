---
page_title: "keycloak_realm_keystore_java_keystore Resources"
---

# keycloak\_realm\_keystore\_java_keystore Resources

Allows for creating and managing `java-keystore` Realm keystores within Keycloak.

A realm keystore manages generated key pairs that are used by Keycloak to perform cryptographic signatures and encryption.

## Example Usage

```hcl
resource "keycloak_realm" "realm" {
	realm = "my-realm"
}

resource "keycloak_realm_keystore_java_keystore" "java_keystore" {
	name      = "my-java-keystore"
	realm_id  = keycloak_realm.realm.id

	enabled = true
	active  = true

	keystore          = "<path to your keystore>"
	keystore_password = "<password for keystore>"
	key_alias         = "<alias for the private key>"
	key_password      = "<password for the private key>"

	priority  = 100
	algorithm = "RS256"
}
```

## Argument Reference

- `name` - (Required) Display name of provider when linked in admin console.
- `realm_id` - (Required) The realm this keystore exists in.
- `internal_realm_id` - (Optional) The internal id for the realm, if the realm is imported into Terraform. This is not relevant for realms created through Terraform.
- `keystore` - (Required) Path to keys file on keycloak instance.
- `keystore_password` - (Required) Password for the keys.
- `key_alias` - (Required) Alias for the private key.
- `key_password` - (Required) Password for the private key.
- `enabled` - (Optional) When `false`, key is not accessible in this realm. Defaults to `true`.
- `active` - (Optional) When `false`, key in not used for signing. Defaults to `true`.
- `priority` - (Optional) Priority for the provider. Defaults to `0`
- `algorithm` - (Optional) Intended algorithm for the key. Defaults to `RS256`

## Import

Realm keys can be imported using realm name and keystore id, you can find it in web UI.

Example:

```bash
$ terraform import keycloak_realm_keystore_java_keystore.java_keystore my-realm/618cfba7-49aa-4c09-9a19-2f699b576f0b
```
