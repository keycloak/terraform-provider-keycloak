---
page_title: "keycloak_realm_keystore_ecdh_generated Resources"
---

# keycloak\_realm\_keystore\_ecdh_generated Resources

Allows for creating and managing `ecdh-generated` Realm keystores within Keycloak.

A realm keystore manages generated elliptic curve key pairs that are used by Keycloak to perform encryption with ECDH-ES based JWE algorithms.

## Example Usage

```hcl
resource "keycloak_realm" "realm" {
	realm = "my-realm"
}

resource "keycloak_realm_keystore_ecdh_generated" "keystore_ecdh_generated" {
	name     = "my-ecdh-generated-key"
	realm_id = keycloak_realm.realm.id

	enabled = true
	active  = true

	priority           = 100
	elliptic_curve_key = "P-256"
	algorithm          = "ECDH-ES"
}
```

## Argument Reference

- `name` - (Required) Display name of provider when linked in admin console.
- `realm_id` - (Required) The realm this keystore exists in.
- `enabled` - (Optional) When `false`, key is not accessible in this realm. Defaults to `true`.
- `active` - (Optional) When `false`, key is not used for encryption. Defaults to `true`.
- `priority` - (Optional) Priority for the provider. Defaults to `0`.
- `elliptic_curve_key` - (Optional) Elliptic Curve used in ECDH. Possible values are `P-256`, `P-384` and `P-521`. Defaults to `P-256`.
- `algorithm` - (Optional) Algorithm for processing the Content Encryption Key. Possible values are `ECDH-ES`, `ECDH-ES+A128KW`, `ECDH-ES+A192KW` and `ECDH-ES+A256KW`. Defaults to `ECDH-ES`.

~> Changing `elliptic_curve_key` makes Keycloak generate a new key pair, which gets a new key ID (`kid`). Changes to the other arguments keep the existing key pair.

## Import

Realm keys can be imported using realm name and keystore id, you can find it in web UI.

Example:

```bash
$ terraform import keycloak_realm_keystore_ecdh_generated.keystore_ecdh_generated my-realm/618cfba7-49aa-4c09-9a19-2f699b576f0b
```
