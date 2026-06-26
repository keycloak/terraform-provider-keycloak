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
- `provider_id` - (Required) The ID of the policy provider (e.g. `trusted-hosts`, `max-clients`). The referenced provider must already be registered in Keycloak.
- `sub_type` - (Required) Whether this policy applies to `anonymous` or `authenticated` client registration.
- `config` - (Optional) A map of provider-specific configuration values.

#### Multi-value configuration

Some policy config fields store a list of values rather than a single value (for the built-in
policies these are `trusted-hosts`, `allowed-client-scopes` and `allowed-protocol-mapper-types`).
Supply these as a single comma-separated string, e.g.
`"trusted-hosts" = "localhost,auth.example.com"`. The provider splits the value into the array
Keycloak expects on write and re-joins it on read. Because Keycloak does not preserve element
order, a pure reorder of these values is treated as a no-op and does not produce a perpetual
diff.

Which fields are multi-valued is detected automatically from Keycloak's policy provider
metadata (properties reported with a `type` of `MultivaluedString` or `MultivaluedList`), so
this works for any provider — including custom SPI policies — without a hardcoded list.

## Import

Client registration policies can be imported using either of the following formats.

By realm name and policy ID:

```bash
$ terraform import keycloak_realm_client_registration_policy.custom my-realm/618cfba7-49aa-4c09-9a19-2f699b576f0b
```

By realm name, policy name, provider ID and sub-type (`{realmId}/{name}/{providerId}/{subType}`).
This is useful for taking ownership of the default policies Keycloak auto-creates, whose
server-generated UUID is not known ahead of time:

```bash
$ terraform import keycloak_realm_client_registration_policy.trusted_hosts "my-realm/Trusted Hosts/trusted-hosts/anonymous"
```

Note: Keycloak automatically creates default policies for every realm (e.g. "Trusted Hosts", "Max Clients Limit").
These can be managed by importing them first and then removing the resource block to delete them. The
`keycloak_realm_client_registration_policy` data source can also be used to look one up dynamically.
