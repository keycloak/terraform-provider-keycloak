---
page_title: "keycloak_authentication_execution Resource"
---

# keycloak\_authentication\_execution Resource

Allows for creating and managing an authentication execution within Keycloak.

An authentication execution is an action that the user or service may or may not take when authenticating through an authentication
flow.

~> Following limitation affects Keycloak < 25:  Due to limitations in the Keycloak API, the ordering of authentication executions within a flow must be specified using `depends_on`. Authentication executions that are created first will appear first within the flow.

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

# first execution
resource "keycloak_authentication_execution" "execution_one" {
  realm_id          = "${keycloak_realm.realm.id}"
  parent_flow_alias = "${keycloak_authentication_flow.flow.alias}"
  authenticator     = "auth-cookie"
  requirement       = "ALTERNATIVE"
  priority			= 10
}

# second execution
resource "keycloak_authentication_execution" "execution_two" {
  realm_id          = "${keycloak_realm.realm.id}"
  parent_flow_alias = "${keycloak_authentication_flow.flow.alias}"
  authenticator     = "identity-provider-redirector"
  requirement       = "ALTERNATIVE"
  priority			= 20
}
```

## Argument Reference

- `realm_id` - (Required) The realm the authentication execution exists in.
- `parent_flow_alias` - (Required) The alias of the flow this execution is attached to.
- `authenticator` - (Required) The name of the authenticator. This can be found by experimenting with the GUI and looking at HTTP requests within the network tab of your browser's development tools.
- `requirement`- (Optional) The requirement setting, which can be one of `REQUIRED`, `ALTERNATIVE`, `OPTIONAL`, `CONDITIONAL`, or `DISABLED`. Defaults to `DISABLED`.
- `priority`- (Optional) The authenticator priority. Lower values will be executed prior higher values (Only supported by Keycloak >= 25).

## Import

Authentication executions can be imported using the formats: `{{realmId}}/{{parentFlowAlias}}/{{authenticationExecutionId}}`.

Example:

```bash
$ terraform import keycloak_authentication_execution.execution_one my-realm/my-flow-alias/30559fcf-6fb8-45ea-8c46-2b86f46ebc17
```
