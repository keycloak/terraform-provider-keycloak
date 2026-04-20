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
```

### Attribute Arguments

- `name` - (Required) Display name of the policy.
- `realm_id` - (Required) The realm this policy exists in.
- `provider_id` - (Required) The ID of the policy provider. NOTE! The provider needs to exist.
- `sub_type` - (Required) Whether this policy applies to `anonymous` or `authenticated` client registration.
- `config` - (Optional) A map of provider-specific configuration values.

## Import

Client registration policies can be imported using the realm name and policy ID:

```bash
$ terraform import keycloak_realm_client_registration_policy.custom my-realm/618cfba7-49aa-4c09-9a19-2f699b576f0b
```

Note: Keycloak automatically creates default policies for every realm (e.g. "Trusted Hosts", "Max Clients Limit").
These can be managed by importing them first and then removing the resource block to delete them.
