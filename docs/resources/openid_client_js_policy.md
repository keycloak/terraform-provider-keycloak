---
page_title: "keycloak_openid_client_js_policy Resource"
---

# keycloak\_openid\_client\_js\_policy Resource

~> **Note:** Consider [`keycloak_generic_client_authorization_policy`](generic_client_authorization_policy.md) instead — it references deployed (and custom Java SPI) policy providers by `type` and does not suffer from the `code` drift described below.

Allows you to manage JavaScript authorization policies.

Modern Keycloak no longer allows uploading the JavaScript code of a policy through the API — the inline `upload_scripts` feature was removed in Keycloak 18, which is no longer supported by this provider. JavaScript policies must therefore be [deployed as a script](https://www.keycloak.org/docs/latest/authorization_services/index.html#_policy_js) (a JAR with a `META-INF/keycloak-scripts.json` descriptor); this resource can then reference the deployed policy. Accordingly, `code` must be set to the deployed provider id that Keycloak assigns to the script, which is `script-` followed by the `fileName` from the descriptor (e.g. `script-my-policy.js`).

~> **Note:** On read, Keycloak returns the script's source in `code` rather than the provider id it was created with, which produces a permanent diff. Use `lifecycle { ignore_changes = [code] }` to suppress it (see the example below). The replacement resource does not have this problem.

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

resource "keycloak_openid_client_js_policy" "test" {
  resource_server_id = keycloak_openid_client.test.resource_server_id
  realm_id           = keycloak_realm.realm.id
  name               = "js_policy"
  decision_strategy  = "UNANIMOUS"
  logic              = "POSITIVE"
  # The deployed provider id: "script-" + the fileName from META-INF/keycloak-scripts.json.
  code = "script-my-policy.js"

  lifecycle {
    # Keycloak returns the script source in "code" on read; ignore that drift.
    ignore_changes = [code]
  }
}
```

### Argument Reference

The following arguments are supported:

- `realm_id` - (Required) The realm this policy exists in.
- `resource_server_id` - (Required) The ID of the resource server.
- `name` - (Required) The name of the policy.
- `decision_strategy` - (Required) The decision strategy, can be one of `UNANIMOUS`, `AFFIRMATIVE`, or `CONSENSUS`.
- `code` - (Required) The deployed JavaScript policy provider id, i.e. `script-` followed by the `fileName` declared in `META-INF/keycloak-scripts.json` (e.g. `script-my-policy.js`). Combine with `lifecycle { ignore_changes = [code] }` since Keycloak returns the script source on read.
- `logic` - (Optional) The logic, can be one of `POSITIVE` or `NEGATIVE`. Defaults to `POSITIVE`.
- `type` - (Optional) The type of the policy. Defaults to `js`.
- `description` - (Optional) A description for the authorization policy.

### Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

- `id` - Policy ID representing the JavaScript policy.

## Import

JavaScript policies can be imported using the format: `{{realmId}}/{{resourceServerId}}/{{policyId}}`.

Example:

```bash
$ terraform import keycloak_openid_client_js_policy.test my-realm/3bd4a686-1062-4b59-97b8-e4e3f10b99da/63b3cde8-987d-4cd9-9306-1955579281d9
```
