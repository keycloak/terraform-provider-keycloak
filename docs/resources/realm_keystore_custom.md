---
page_title: "keycloak_realm_keystore_custom Resources"
---

# keycloak\_realm\_keystore\_custom Resources

Allows for creating and managing Realm keystores of any key provider within Keycloak, including key providers from
custom extensions (for example a key provider that signs tokens with a key held in an HSM).

A realm keystore manages keys that are used by Keycloak to perform cryptographic signatures and encryption.

For built-in key providers that have their own resource, such as `keycloak_realm_keystore_rsa_generated`, prefer that
resource, as it validates the config at plan time and documents the available arguments.

## Example Usage

### Key provider from a custom extension

```hcl
resource "keycloak_realm" "realm" {
	realm = "my-realm"
}

resource "keycloak_realm_keystore_custom" "hsm" {
	name        = "hsm-token-signing"
	realm_id    = keycloak_realm.realm.id
	provider_id = "my-hsm-key-provider"

	config = {
		priority = "200"
		endpoint = "hsm.example.com:50051"
		keyId    = "token-signing-es256"
	}
}
```

### Built-in key provider without its own resource

```hcl
resource "keycloak_realm_keystore_custom" "eddsa" {
	name        = "eddsa-generated"
	realm_id    = keycloak_realm.realm.id
	provider_id = "eddsa-generated"

	config = {
		priority              = "100"
		enabled               = "true"
		active                = "true"
		eddsaEllipticCurveKey = "Ed25519"
	}
}
```

## Argument Reference

- `name` - (Required) Display name of provider when linked in admin console.
- `realm_id` - (Required) The realm this keystore exists in.
- `provider_id` - (Required) The id of the key provider, as returned by the `getId` implementation of its `KeyProviderFactory`. Changing this forces a new resource.
- `config` - (Optional) A map of config keys and values of the key provider. Common keys are `priority`, `enabled` and `active`, the other keys depend on the key provider.

### Config

- Only keys that the key provider declares as config properties are allowed, because Keycloak drops all other keys.
  The plan fails for an undeclared key and lists the declared keys. The declared keys are shown in the admin console
  and returned by the `/admin/serverinfo` endpoint under `componentTypes["org.keycloak.keys.KeyProvider"]`.
- All values are strings, including numbers and booleans, and must not be empty. Remove a key to unset it.
- Values of multi-valued properties are written as a comma-separated string. A different order of the elements does not cause a diff.
- Keycloak does not return the values of secret properties. Terraform keeps the value from the state for them, so a
  change made outside of Terraform is not detected. Pass secret values through a sensitive variable to keep them out of the plan output.
- Some key providers add keys to the config on their own, for example the `rsa` key provider generates a certificate
  if none is given. Keys that are not in `config` are ignored.
- The plan fails if the key provider is not installed on the server.

## Import

Realm keystores can be imported using one of the following formats:

- `{{realmId}}/{{keystoreId}}`, you can find the keystore id in the web UI.
- `{{realmId}}/{{name}}/{{providerId}}`, which lets you import keystores with static values in an `import` block.
  The import fails if more than one keystore of the realm has this name and provider id.

The import takes all config keys that Keycloak returns, with secret values masked as `**********`. Keys that are
missing from the `config` of the resource are removed on the next apply, and secret values are set again.

Example:

```bash
$ terraform import keycloak_realm_keystore_custom.hsm my-realm/618cfba7-49aa-4c09-9a19-2f699b576f0b
$ terraform import keycloak_realm_keystore_custom.hsm my-realm/hsm-token-signing/my-hsm-key-provider
```
