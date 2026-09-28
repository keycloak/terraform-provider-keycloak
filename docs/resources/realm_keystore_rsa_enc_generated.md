---
page_title: "keycloak_realm_keystore_rsa_enc_generated Resources"
---

# keycloak\_realm\_keystore\_rsa_enc_generated Resources

Allows for creating and managing `rsa-enc-generated` Realm keystores within Keycloak.

A realm keystore manages generated key pairs that are used by Keycloak to perform cryptographic signatures and encryption.

## Example Usage

```hcl
resource "keycloak_realm" "realm" {
	realm = "my-realm"
}

resource "keycloak_realm_keystore_rsa_enc_generated" "keystore_rsa_enc_generated" {
	name      = "my-rsa-enc-generated-key"
	realm_id  = keycloak_realm.realm.id

	enabled = true
	active  = true

	priority  = 100
	algorithm = "RSA-OAEP"
	key_size  = 2048
}
```

## Argument Reference

- `name` - (Required) Display name of provider when linked in admin console.
- `realm_id` - (Required) The realm this keystore exists in.
- `enabled` - (Optional) When `false`, key is not accessible in this realm. Defaults to `true`.
- `active` - (Optional) When `false`, key is not used for encryption. Defaults to `true`.
- `priority` - (Optional) Priority for the provider. Defaults to `0`
- `algorithm` - (Optional) Intended algorithm for the key. Defaults to `RSA-OAEP`
- `key_size` - (Optional) Size for the generated keys. Defaults to `2048`.

## Import

Realm keys can be imported using realm name and keystore id, you can find it in web UI.

Example:

```bash
$ terraform import keycloak_realm_keystore_rsa_enc_generated.keystore_rsa_enc_generated my-realm/618cfba7-49aa-4c09-9a19-2f699b576f0b
```
