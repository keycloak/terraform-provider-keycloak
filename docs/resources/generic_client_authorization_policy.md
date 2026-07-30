---
page_title: "keycloak_generic_client_authorization_policy Resource"
---

# keycloak\_generic\_client\_authorization\_policy Resource

Allows you to manage a Keycloak client authorization policy of any provider type by referencing that
provider by its `type`. This is the policy equivalent of `keycloak_generic_role_mapper` /
`keycloak_generic_protocol_mapper`: it works with any policy provider registered in Keycloak, including
custom providers you implement yourself.

This is intended for custom policy providers that are deployed to Keycloak as a JAR. The `type` is simply
the provider id that the policy's
[`PolicyProviderFactory`](https://www.keycloak.org/docs-api/latest/javadocs/org/keycloak/authorization/policy/provider/PolicyProviderFactory.html)
exposes via `getId()`:

- For a policy implemented in Java as an SPI, `type` is whatever your `PolicyProviderFactory.getId()` returns.
- For a [JavaScript policy](https://www.keycloak.org/docs/latest/authorization_services/index.html#_policy_js)
  deployed as a script (in a JAR with a `META-INF/keycloak-scripts.json` descriptor), Keycloak generates the
  provider id for you as `script-` followed by the `fileName` from the descriptor (if the `fileName`
  contains a path, it is kept). The `script-` prefix is an implementation detail of the scripting SPI, not a
  requirement of this resource. For example, `"fileName": "my-policy.js"` is referenced as
  `type = "script-my-policy.js"`.
  (Modern Keycloak no longer allows uploading the JavaScript code through the API, so the script must be
  deployed alongside Keycloak and only referenced here.)

For the policy types that ship with Keycloak (role, group, user, client, time, regex, ...) prefer the
dedicated resources, e.g. `keycloak_openid_client_role_policy`.

## Example Usage

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_openid_client" "test" {
  client_id                = "client_id"
  realm_id                 = keycloak_realm.realm.id
  access_type              = "CONFIDENTIAL"
  service_accounts_enabled = true

  authorization {
    policy_enforcement_mode = "ENFORCING"
  }
}

# References a custom policy provider implemented in Java and deployed as a JAR.
# "type" is the id returned by your PolicyProviderFactory.getId().
resource "keycloak_generic_client_authorization_policy" "custom" {
  resource_server_id = keycloak_openid_client.test.resource_server_id
  realm_id           = keycloak_realm.realm.id
  name               = "my-custom-policy"
  type               = "my-custom-policy-provider"
  decision_strategy  = "UNANIMOUS"
  logic              = "POSITIVE"
  description        = "Authorization policy backed by a custom Java SPI provider"
}

# References a JavaScript policy deployed as a script. Here Keycloak generates the
# type as "script-" + the fileName from META-INF/keycloak-scripts.json.
resource "keycloak_generic_client_authorization_policy" "deployed_js" {
  resource_server_id = keycloak_openid_client.test.resource_server_id
  realm_id           = keycloak_realm.realm.id
  name               = "deployed-js-policy"
  type               = "script-my-policy.js"
  decision_strategy  = "UNANIMOUS"
  logic              = "POSITIVE"
  description        = "Authorization policy backed by a deployed JavaScript script"
}
```

## Argument Reference

The following arguments are supported:

- `realm_id` - (Required) The realm this policy exists in.
- `resource_server_id` - (Required) The ID of the resource server this policy is attached to.
- `name` - (Required) The name of the policy.
- `type` - (Required) The provider type of the policy, i.e. the id returned by the policy's
  `PolicyProviderFactory.getId()`. For a custom Java SPI this is whatever id your factory exposes; for a
  JavaScript policy deployed as a script it is `script-` followed by the `fileName` declared in
  `META-INF/keycloak-scripts.json`, e.g. `script-my-policy.js`.
- `decision_strategy` - (Required) The decision strategy, can be one of `UNANIMOUS`, `AFFIRMATIVE`, or `CONSENSUS`.
- `logic` - (Optional) The logic, can be one of `POSITIVE` or `NEGATIVE`. Defaults to `POSITIVE`.
- `description` - (Optional) A description for the authorization policy.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

- `id` - Policy ID representing the policy.

## Import

This resource can be imported using the format: `{{realmId}}/{{resourceServerId}}/{{policyId}}`.

Example:

```bash
$ terraform import keycloak_generic_client_authorization_policy.deployed_js my-realm/3bd4a686-1062-4b59-97b8-e4e3f10b99da/63b3cde8-987d-4cd9-9306-1955579281d9
```
