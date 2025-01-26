---
page_title: "keycloak_authentication_bindings Resource"
---

# keycloak\_authentication\_bindings Resource

Allows for creating and managing realm authentication flow bindings within Keycloak.

[Authentication flows](https://www.keycloak.org/docs/latest/server_admin/index.html#_authentication-flows) describe a sequence
of actions that a user or service must perform in order to be authenticated to Keycloak. The authentication flow itself
is a container for these actions, which are otherwise known as executions.

Realms assign authentication flows to supported user flows such as `registration` and `browser`. This resource allows the
updating of realm authentication flow bindings to custom authentication flows created by `keycloak_authentication_flow`.

Note that you can also use the `keycloak_realm` resource to assign authentication flow bindings at the realm level. This
resource is useful if you would like to create a realm and an authentication flow, and assign this flow to the realm within
a single run of `terraform apply`. In any case, do not attempt to use both the arguments within the `keycloak_realm` resource
and this resource to manage authentication flow bindings, you should choose one or the other.

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
}

# second execution
resource "keycloak_authentication_execution" "execution_two" {
  realm_id          = "${keycloak_realm.realm.id}"
  parent_flow_alias = "${keycloak_authentication_flow.flow.alias}"
  authenticator     = "identity-provider-redirector"
  requirement       = "ALTERNATIVE"

  depends_on = [
    keycloak_authentication_execution.execution_one
  ]
}

resource "keycloak_authentication_bindings" "browser_authentication_binding" {
  realm_id	    = keycloak_realm.realm.id
  browser_flow  = keycloak_authentication_flow.flow.alias
}
```

## Argument Reference

- `realm_id` - (Required) The realm the authentication flow binding exists in.
- `browser_flow` - (Optional) The alias of the flow to assign to the realm BrowserFlow.
- `registration_flow` - (Optional) The alias of the flow to assign to the realm RegistrationFlow.
- `direct_grant_flow` - (Optional) The alias of the flow to assign to the realm DirectGrantFlow.
- `reset_credentials_flow` - (Optional) The alias of the flow to assign to the realm ResetCredentialsFlow.
- `client_authentication_flow` - (Optional) The alias of the flow to assign to the realm ClientAuthenticationFlow.
- `docker_authentication_flow` - (Optional) The alias of the flow to assign to the realm DockerAuthenticationFlow.
- `first_broker_login_flow` - (Optional) The alias of the flow to assign to the realm FirstBrokerLoginFlow (since Keycloak 24).
