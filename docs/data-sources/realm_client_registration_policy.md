---
page_title: "keycloak_realm_client_registration_policy Data Source"
---

# keycloak\_realm\_client\_registration\_policy Data Source

Use this data source to look up an existing client registration policy in a realm by name,
optionally narrowing the match by `provider_id` and/or `sub_type`. This is primarily useful
for discovering the server-generated ID of a default policy Keycloak auto-creates (e.g.
"Trusted Hosts", "Max Clients Limit") so it can be referenced elsewhere.

If no policy matches, or more than one policy matches, an error is returned.

## Example Usage

```hcl
resource "keycloak_realm" "realm" {
  realm = "my-realm"
}

data "keycloak_realm_client_registration_policy" "trusted_hosts" {
  realm_id    = keycloak_realm.realm.id
  name        = "Trusted Hosts"
  provider_id = "trusted-hosts"
  sub_type    = "anonymous"
}

output "trusted_hosts_policy_id" {
  value = data.keycloak_realm_client_registration_policy.trusted_hosts.id
}
```

## Argument Reference

- `realm_id` - (Required) The realm the policy belongs to.
- `name` - (Required) The name of the policy to find.
- `provider_id` - (Optional) Filter by provider ID (e.g. `trusted-hosts`, `max-clients`).
- `sub_type` - (Optional) Filter by sub-type (`anonymous` or `authenticated`).

## Attributes Reference

- `id` - The ID of the matched policy.
- `config` - A map of the policy's provider-specific configuration values.
