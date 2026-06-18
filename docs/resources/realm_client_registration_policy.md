---
page_title: "keycloak_realm_client_registration_policy Resource"
---

# keycloak_realm_client_registration_policy Resource

Allows for managing Realm Client Registration Policies.

## Example Usage

```hcl
resource "keycloak_realm" "realm" {
  realm = "my-realm"
}

resource "keycloak_realm_client_registration_policy" "custom" {
  realm_id    = keycloak_realm.realm.id
  name        = "My Custom Policy"
  provider_id = "my-client-registration-policy"
  sub_type    = "anonymous"
}

# A built-in policy with multi-value config (trusted-hosts is an array in Keycloak)
resource "keycloak_realm_client_registration_policy" "trusted_hosts" {
  realm_id    = keycloak_realm.realm.id
  name        = "Trusted Hosts"
  provider_id = "trusted-hosts"
  sub_type    = "anonymous"

  config = {
    "host-sending-registration-request-must-match" = "true"
    "client-uris-must-match"                        = "true"
    "trusted-hosts"                                 = "localhost,auth.example.com"
  }
}
```

### Attribute Arguments

- `name` - (Required) Display name of the policy.
- `realm_id` - (Required) The realm this policy exists in.
- `provider_id` - (Required) The ID of the policy provider. NOTE! The provider needs to exist.
- `sub_type` - (Required) Whether this policy applies to `anonymous` or `authenticated` client registration.
- `config` - (Optional) A map of provider-specific configuration values.

#### Multi-value configuration

A few built-in policies store their configuration as a list rather than a single value
(currently `trusted-hosts` and `allowed-client-scopes`). Supply these as a single
comma-separated string, e.g. `"trusted-hosts" = "localhost,auth.example.com"`. The provider
splits the value into the array Keycloak expects on write and re-joins it on read. Because
Keycloak does not preserve element order, a pure reorder of these values is treated as a
no-op and does not produce a perpetual diff.

## Import

Client registration policies can be imported using the realm name and policy ID:

```bash
$ terraform import keycloak_realm_client_registration_policy.custom my-realm/618cfba7-49aa-4c09-9a19-2f699b576f0b
```

Note: Keycloak automatically creates default policies for every realm (e.g. "Trusted Hosts", "Max Clients Limit").
These can be managed by importing them first and then removing the resource block to delete them.
