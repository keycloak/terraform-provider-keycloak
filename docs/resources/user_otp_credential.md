---
page_title: "keycloak_user_otp_credential Resource"
---

# keycloak\_user\_otp\_credential Resource

Allows for provisioning a time-based one-time password (TOTP) credential for a Keycloak user, so that a "require OTP"
authentication flow is satisfied without the user having to enroll an authenticator by hand.

Creating this resource also removes the `CONFIGURE_TOTP` required action from the user. Because of that, a `keycloak_user`
that sets `required_actions = ["CONFIGURE_TOTP"]` should ignore changes to `required_actions`, otherwise Terraform will
try to add the required action back.

~> **Note:** Keycloak never returns the secret of an OTP credential, so changes to the secret made outside of Terraform
cannot be detected. The secret is stored in the Terraform state, so make sure the state is stored securely.

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

  required_actions = ["CONFIGURE_TOTP"]

  lifecycle {
    ignore_changes = [required_actions]
  }
}

resource "random_id" "totp_secret" {
  byte_length = 20
}

resource "keycloak_user_otp_credential" "otp" {
  realm_id   = keycloak_realm.realm.id
  user_id    = keycloak_user.user.id
  secret     = random_id.totp_secret.b32_std
  user_label = "Provisioned by Terraform"
}
```

## Argument Reference

All arguments force a new resource to be created when changed.

- `realm_id` - (Required) The realm the user exists in.
- `user_id` - (Required) The ID of the user to provision the OTP credential for.
- `secret` - (Required) The Base32-encoded TOTP secret. This value is sensitive.
- `algorithm` - (Optional) The HMAC algorithm, one of `HmacSHA1`, `HmacSHA256` or `HmacSHA512`. Defaults to `HmacSHA1`.
- `digits` - (Optional) The number of digits of the one-time password, either `6` or `8`. Defaults to `6`.
- `period` - (Optional) The number of seconds a one-time password is valid for. Defaults to `30`.
- `user_label` - (Optional) The label of the credential, as shown to the user.

## Attributes Reference

- `id` - The ID of the resource, in the format `{{realm}}/{{userId}}/{{credentialId}}`.

## Import

This resource does not support import. Keycloak never returns the secret, so an imported credential could not be
compared to the configuration. Instead of importing, create this resource as if it did not already exist on the server.
