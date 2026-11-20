---
page_title: "keycloak_required_action Resource"
---

# keycloak\_required\_action Resource

Allows for creating and managing required actions within Keycloak.

[Required actions](https://www.keycloak.org/docs/latest/server_admin/#con-required-actions_server_administration_guide) specify actions required before the first login of all new users.


## Example Usage

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_required_action" "required_action" {
  realm_id = keycloak_realm.realm.realm
  alias    = "UPDATE_PASSWORD"
  enabled  = true
  name     = "Update Password"
  config = {
    max_auth_age = "600"
  }
}
```

## Argument Reference

- `realm_id` - (Required) The realm the required action exists in.
- `alias` - (Required) The alias of the action to attach as a required action. Case sensitive.
- `name` - (Optional) The name of the required action to use in the UI.
- `enabled` - (Optional) When `false`, the required action is not enabled for new users. Defaults to `false`.
- `default_action` - (Optional) When `true`, the required action is set as the default action for new users. Defaults to `false`.
- `priority`- (Optional) An integer to specify the running order of required actions with lower numbers meaning higher precedence.
- `config`- (Optional) The configuration. Keys are specific to each configurable required action and not checked when applying.

## Keycloak built-in required actions

| Alias                             | Description                                 | Class
|-----------------------------------|---------------------------------------------|-----------------------------
| `CONFIGURE_RECOVERY_AUTHN_CODES`  | Configure recovery authentication codes     | [RecoveryAuthnCodesAction](https://www.keycloak.org/docs-api/latest/javadocs/org/keycloak/authentication/requiredactions/RecoveryAuthnCodesAction.html)
| `CONFIGURE_TOTP`                  | Require user to configure 2FA (TOTP)        | [UpdateTotp](https://www.keycloak.org/docs-api/latest/javadocs/org/keycloak/authentication/requiredactions/UpdateTotp.html)
| `delete_account`                  | Allow user to delete their account          | [DeleteAccount](https://www.keycloak.org/docs-api/latest/javadocs/org/keycloak/authentication/requiredactions/DeleteAccount.html)
| `delete_credential`               | Allow user to delete a credential           |
| `idp_link`                        | Link account with identity provider         |
| `TERMS_AND_CONDITIONS`            | Require user to accept terms and conditions | [TermsAndConditions](https://www.keycloak.org/docs-api/latest/javadocs/org/keycloak/authentication/requiredactions/TermsAndConditions.html)
| `UPDATE_PASSWORD`                 | Prompt user to update their password        | [UpdatePassword](https://www.keycloak.org/docs-api/latest/javadocs/org/keycloak/authentication/requiredactions/UpdatePassword.html)
| `UPDATE_PROFILE`                  | Prompt user to update their profile         | [UpdateProfile](https://www.keycloak.org/docs-api/latest/javadocs/org/keycloak/authentication/requiredactions/UpdateProfile.html)
| `update_user_locale`              | Prompt user to set or update their locale   | [UpdateUserLocaleAction](https://www.keycloak.org/docs-api/21.0.2/javadocs/org/keycloak/authentication/requiredactions/UpdateUserLocaleAction.html)
| `VERIFY_EMAIL`                    | Require user to verify their email address  | [VerifyEmail](https://www.keycloak.org/docs-api/latest/javadocs/org/keycloak/authentication/requiredactions/VerifyEmail.html)
| `VERIFY_PROFILE`                  | Verify user profile information             | [VerifyUserProfile](https://www.keycloak.org/docs-api/latest/javadocs/org/keycloak/authentication/requiredactions/VerifyUserProfile.html)


## Import

Authentication executions can be imported using the formats: `{{realm}}/{{alias}}`.

Example:

```bash
$ terraform import keycloak_required_action.required_action my-realm/my-default-action-alias
```
