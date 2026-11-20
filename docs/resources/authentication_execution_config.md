---
page_title: "keycloak_authentication_execution_config Resource"
---

# keycloak\_authentication\_execution\_config Resource

Allows for managing an authentication execution's configuration. If a particular authentication execution supports additional
configuration (such as with the `identity-provider-redirector` execution), this can be managed with this resource.

## Example Usage

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_authentication_flow" "flow" {
  realm_id = keycloak_realm.realm.id
  alias    = "my-flow-alias"
}

resource "keycloak_authentication_execution" "execution" {
  realm_id          = keycloak_realm.realm.id
  parent_flow_alias = keycloak_authentication_flow.flow.alias
  authenticator     = "identity-provider-redirector"
}

resource "keycloak_authentication_execution_config" "config" {
  realm_id     = keycloak_realm.realm.id
  execution_id = keycloak_authentication_execution.execution.id
  alias        = "my-config-alias"
  config = {
    defaultProvider = "my-config-default-idp"
  }
}
```

## Argument Reference

- `realm_id` - (Required) The realm the authentication execution exists in.
- `execution_id` - (Required) The authentication execution this configuration is attached to.
- `alias` - (Required) The name of the configuration.
- `config` - (Optional) The configuration. Keys are specific to each configurable authentication execution and not checked when applying.

~> Since Keycloak 26.6.7 / 26.7.3, the Keycloak API returns secret config properties (for example the reCAPTCHA `secret.key`) masked
as `**********`. Keycloak treats a property as secret unless some authenticator declares it as a non-secret config property, so
properties of custom authenticators that are not declared in the factory are masked as well. For masked properties the provider keeps
the value from your Terraform configuration in state, so changing a secret in your configuration is still detected and applied.
Changes to masked properties made outside of Terraform are not detected and are overwritten on the next apply of this resource.
After an import, the masked value is stored in state until the next apply.

## Import

Configurations can be imported using the format `{{realm}}/{{authenticationExecutionId}}/{{authenticationExecutionConfigId}}`.
If the `authenticationExecutionId` is incorrect, the import will still be successful.
A subsequent apply will change the `authenticationExecutionId` to the correct one, which causes the configuration to be replaced.

Example:

```bash
$ terraform import keycloak_authentication_execution_config.config my-realm/be081463-ddbf-4b42-9eff-9c97886f24ff/30559fcf-6fb8-45ea-8c46-2b86f46ebc17
```
