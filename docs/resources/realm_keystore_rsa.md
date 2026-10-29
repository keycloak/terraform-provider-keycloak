---
page_title: "keycloak_realm_keystore_rsa Resources"
---

# keycloak\_realm\_keystore\_rsa Resources

Allows for creating and managing `rsa` Realm keystores within Keycloak.

A realm keystore manages generated key pairs that are used by Keycloak to perform cryptographic signatures and encryption.

> **Write-only arguments:** `private_key_wo` and `certificate_wo` (with `private_key_wo_version` and
> `certificate_wo_version`) can be used instead of `private_key` and `certificate`. Write-only arguments are never
> stored in the plan or state files. Each write-only argument conflicts with its non-write-only counterpart.

## Example Usage

```hcl
resource "keycloak_realm" "realm" {
	realm = "my-realm"
}

resource "keycloak_realm_keystore_rsa" "keystore_rsa" {
	name      = "my-rsa-key"
	realm_id  = keycloak_realm.realm.id

	enabled = true
	active  = true

	private_key = "<your rsa private key>"
	certificate = "<your certificate>"

	priority  = 100
	algorithm = "RS256"
	keystore_size  = 2048
	provider_id = "rsa"

    extra_config = {
      kid = "my-key-id"
    }
}
```

## Example Usage with write-only arguments

```hcl
data "azurerm_key_vault" "my-vault" {
	name                = "my-vault"
	resource_group_name = "my-rg"
}

ephemeral "azurerm_key_vault_certificate" "my-cert" {
	name         = "my-cert"
	key_vault_id = data.azurerm_key_vault.my-vault.id
}

resource "keycloak_realm_keystore_rsa" "keystore_rsa" {
	name      = "my-rsa-key"
	realm_id  = keycloak_realm.realm.id

	private_key_wo         = ephemeral.azurerm_key_vault_certificate.my-cert.key
	private_key_wo_version = ephemeral.azurerm_key_vault_certificate.my-cert.expiration_date
	certificate_wo         = ephemeral.azurerm_key_vault_certificate.my-cert.pem
	certificate_wo_version = ephemeral.azurerm_key_vault_certificate.my-cert.expiration_date
}
```

## Argument Reference

- `name` - (Required) Display name of provider when linked in admin console.
- `realm_id` - (Required) The realm this keystore exists in.
- `private_key` - (Optional) Private RSA Key encoded in PEM format. Required without `private_key_wo` and `private_key_wo_version`.
- `private_key_wo` - (Optional, Write-Only) Private RSA Key encoded in PEM format. Not stored in state or plan files. Requires `private_key_wo_version`.
- `private_key_wo_version` - (Optional) Trigger for `private_key_wo`: the key is only sent to Keycloak when this value changes. Stored in state.
- `certificate` - (Optional) X509 Certificate encoded in PEM format. Required without `certificate_wo` and `certificate_wo_version`.
- `certificate_wo` - (Optional, Write-Only) X509 Certificate encoded in PEM format. Not stored in state or plan files. Requires `certificate_wo_version`. The certificate is sent on every update, since Keycloak requires it.
- `certificate_wo_version` - (Optional) Version of the `certificate_wo` argument. Stored in state.
- `enabled` - (Optional) When `false`, key is not accessible in this realm. Defaults to `true`.
- `active` - (Optional) When `false`, key in not used for signing. Defaults to `true`.
- `priority` - (Optional) Priority for the provider. Defaults to `0`
- `algorithm` - (Optional) Intended algorithm for the key. Defaults to `RS256`. Use `RSA-OAEP` for encryption keys
- `keystore_size` - (Optional) Size for the generated keys. Defaults to `2048`.
- `provider_id` - (Optional) Use `rsa` for signing keys, `rsa-enc` for encryption keys
- `extra_config` - (Optional) Map of additional provider configuration options passed through to the Keycloak component config. For RSA keystores this can include keys like `kid`.

## Import

Realm keys can be imported using realm name and keystore id, you can find it in web UI.

Example:

```bash
$ terraform import keycloak_realm_keystore_rsa.keystore_rsa my-realm/618cfba7-49aa-4c09-9a19-2f699b576f0b
```
