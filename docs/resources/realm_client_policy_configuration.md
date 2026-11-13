---
page_title: "keycloak_realm_client_policy_configuration Resource"
---

# keycloak_realm_client_policy_configuration Resource

Owns the complete realm-specific client-policy profile and policy collections,
including an explicitly empty configuration. Do not combine this singleton with
individual profile/policy resources for the same realm. Global templates are
read-only observations and are never sent in updates.

```hcl
resource "keycloak_realm_client_policy_configuration" "realm" {
  realm_id = keycloak_realm.realm.id
  profiles = jsonencode([])
  policies = jsonencode([])
  lifecycle { prevent_destroy = true }
}
```

- `realm_id`: Required realm ID; changing it requires replacement.
- `profiles`: Required JSON array of complete profile objects. Use `jsonencode`
  with HCL objects. Names must be nonempty and unique; executor arrays and all
  configuration JSON types retain their representation.
- `policies`: Required JSON array of complete policy objects, including enabled,
  conditions, profiles and optional mode. Names must be nonempty and unique.
- `global_profiles`, `global_policies`: Computed JSON arrays of server templates.
  They cannot be configured or overridden by local entries.

Arrays are ordered. Server defaults/omissions must match the captured desired
representation; a clean import does not certify future arbitrary configuration.
Updates retain old profile targets until policies switch references, then remove
obsolete profiles. Only changed collections write. The two endpoints are not
atomic: an error can leave staged profiles; refresh/retry completes from actual
server state. Deletion is rejected; explicitly clear arrays, or remove Terraform
ownership without destroying the remote singleton.

## Import

Import by realm ID. Import calls GET only and includes existing empty collections.

```sh
terraform import keycloak_realm_client_policy_configuration.realm gryphon
```

## Recovering from a failed initial apply

A failed update can be retried after refresh. A failed initial creation is
marked **tainted** by Terraform, even when the provider retains the realm ID.
An ordinary retry then attempts replacement, which this singleton rejects.
After correcting the cause of the failed request, back up the state and remove
only this resource's ownership, then reimport the current server collections:

```sh
terraform state pull > terraform-state-backup.json
terraform state rm keycloak_realm_client_policy_configuration.realm
terraform import keycloak_realm_client_policy_configuration.realm gryphon
terraform plan
terraform apply
```

State removal and import do not change Keycloak. Keep the desired profiles and
policies in configuration; review the plan before applying to complete the
remaining stages. If the resource is still present in state with the correct
realm ID, `terraform untaint keycloak_realm_client_policy_configuration.realm`
is an alternative to removal and reimport; refresh and review the plan before
retrying. Do not attempt replacement or destroy to recover this singleton.
