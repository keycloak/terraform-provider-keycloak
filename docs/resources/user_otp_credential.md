---
page_title: "keycloak_user_otp_credential Resource"
---

# keycloak\_user\_otp\_credential Resource

Allows for provisioning a time-based one-time password (TOTP) credential for a Keycloak user, so that a "require OTP"
authentication flow is satisfied without the user having to enroll an authenticator by hand.

Only the credential is created: the rest of the user is left as it is. In particular the `CONFIGURE_TOTP` required action is
not removed, so do not set it on the user if the OTP credential is provisioned here.

~> **Note:** Keycloak never returns the secret of an OTP credential, so changes to the secret made outside of Terraform
cannot be detected. With `secret` the secret is stored in the Terraform state, so make sure the state is stored securely.
Use `secret_wo` to keep it out of the state, which requires Terraform 1.11 or later.

~> **Note:** Do not add OTP credentials to the same user from several places at the same time: if more than one new
OTP credential shows up while this resource is created, the provider fails instead of guessing which one it created, and
the extra credentials have to be removed by hand.

~> **Note:** This resource requires Keycloak 22 or later.

## Example Usage

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_user" "user" {
  realm_id = keycloak_realm.realm.id
  username = "my-user"
}

# A Base32-encoded secret generated outside of Terraform, for example with `openssl rand 20 | base32`
variable "totp_secret" {
  type      = string
  sensitive = true
}

resource "keycloak_user_otp_credential" "otp" {
  realm_id   = keycloak_realm.realm.id
  user_id    = keycloak_user.user.id
  secret     = var.totp_secret
  user_label = "Provisioned by Terraform"
}
```

## Example Usage (write-only secret)

With `secret_wo` the secret is not stored in the state. Because Terraform does not keep the value, it cannot tell
whether it changed: change `secret_wo_version` to replace the credential with one made from the current `secret_wo`.

```hcl
resource "keycloak_user_otp_credential" "otp" {
  realm_id          = keycloak_realm.realm.id
  user_id           = keycloak_user.user.id
  secret_wo         = var.totp_secret
  secret_wo_version = "1"
}
```

## Argument Reference

All arguments force a new resource to be created when changed.

- `realm_id` - (Required) The realm the user exists in. Alternatively, set `realm_internal_id` to keep the resource when the realm is renamed, see [Renaming realms](../index.md#renaming-realms).
- `user_id` - (Required) The ID of the user to provision the OTP credential for.
- `secret` - (Optional) The Base32-encoded TOTP secret. Trailing `=` padding is allowed and is removed before the secret is sent to Keycloak. This value is sensitive and is stored in the state. Exactly one of `secret` and `secret_wo` must be set.
- `secret_wo` - (Optional) The same secret as a [write-only argument](https://developer.hashicorp.com/terraform/language/resources/ephemeral#write-only-arguments), which is not stored in the state. Requires `secret_wo_version`.
- `secret_wo_version` - (Optional) The version of `secret_wo`. Changing it replaces the credential. Required with `secret_wo`.
- `algorithm` - (Optional) The HMAC algorithm, one of `HmacSHA1`, `HmacSHA256` or `HmacSHA512`. Defaults to `HmacSHA1`.
- `digits` - (Optional) The number of digits of the one-time password, either `6` or `8`. Defaults to `6`.
- `period` - (Optional) The number of seconds a one-time password is valid for. Defaults to `30`.
- `user_label` - (Optional) The label of the credential, as shown to the user. It is only set when the credential is created and is not read back, so a user renaming the credential does not cause a change.

## Attributes Reference

- `id` - The ID of the resource, in the format `{{realm}}/{{userId}}/{{credentialId}}`.

## Import

This resource does not support import. Keycloak never returns the secret, so an imported credential could not be
compared to the configuration. Instead of importing, create this resource as if it did not already exist on the server.
