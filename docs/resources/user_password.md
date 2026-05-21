---
page_title: "keycloak_user_password Resource"
---

# keycloak\_user\_password Resource

Manages a user's password credential in Keycloak. This resource is separate from `keycloak_user` so that password
changes can be planned and applied independently of the user resource.

~> This resource is not compatible with `initial_password` on the `keycloak_user` resource. Use one or the other,
not both.

The password value is write-only: it is never stored in Terraform state and never appears in plan output. Change
detection is handled by comparing a SHA-256 hash of the configured value with a hash stored in state.

## Example Usage

```hcl
data "keycloak_realm" "realm" {
  realm = "my-realm"
}

resource "keycloak_user" "user" {
  realm_id = data.keycloak_realm.realm.id
  username = "alice"
}

resource "keycloak_user_password" "user_password" {
  realm_id = data.keycloak_realm.realm.id
  user_id  = keycloak_user.user.id
  value    = "some password"
}
```

## Argument Reference

- `realm_id` - (Required) The realm the user belongs to.
- `user_id` - (Required) The ID of the user to manage the password for.
- `value` - (Required, write-only) The password value. This value is sensitive and is not stored in state or shown in plan output.
- `temporary` - (Optional) If set to `true`, the password is set as temporary and the user will be prompted to change it on first login. Defaults to `false`.

## Attribute Reference

- `value_hash` - A SHA-256 hash of the currently applied password, used internally to detect when `value` changes across Terraform runs.
- `credential_id` - The UUID of the Keycloak password credential. Used internally to detect out-of-band password resets.

## Import

User passwords can be imported using the format `{{realm_id}}/{{user_id}}`. The `value` attribute cannot be recovered from the API and must be supplied in the configuration after import.

Example:

```bash
$ terraform import keycloak_user_password.user_password my-realm/60c3f971-b1d3-4b3a-9035-d16d7540a5e4
```
