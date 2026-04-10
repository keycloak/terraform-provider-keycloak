---
page_title: "keycloak_realm_optional_client_scopes Resource"
---

# keycloak\_realm\_optional\_client\_scopes Resource

Allows you to manage the set of optional client scopes for a Keycloak realm, which are used when new clients are created.

Note that this resource attempts to be an **authoritative** source over the optional client scopes for a Keycloak realm,
so any Keycloak defaults and manual adjustments will be overwritten.

Additionally, if a scope is set as default, this resource will remove and re-add it as an optional scope.

Warning: Configuring the same scope in both this resource and `keycloak_realm_default_client_scopes` for the same realm
is an error. When the realm and scope names are known during planning, the plan fails with
`scopes [...] are configured as both default and optional scopes`. Apply-time validation only sees resources that
reconcile during that apply. If one resource is unchanged and the other resource's scope names or realm identity are
unknown until apply, the changed resource can move a scope managed by the unchanged resource without reporting a
conflict. This also applies when applying a saved plan.

To prevent this, use shared desired scope sets and add the same lifecycle precondition to **both** resources
(Terraform 1.2 or later). See the
[paired-resource example](realm_default_client_scopes.md#validating-default-and-optional-scopes).
Terraform evaluates the condition during planning when possible, or during apply before changing the scope assignments
when the values become known.

## Example Usage

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_openid_client_scope" "client_scope" {
  realm_id = keycloak_realm.realm.id
  name     = "test-client-scope"
}

resource "keycloak_realm_optional_client_scopes" "optional_scopes" {
  realm_id  = keycloak_realm.realm.id

  optional_scopes = [
    "address",
    "phone",
    "offline_access",
    "microprofile-jwt",
    keycloak_openid_client_scope.client_scope.name
  ]
}
```

## Argument Reference

- `realm_id` - (Required) The realm this client and scopes exists in.
- `optional_scopes` - (Required) An array of optional client scope names that should be used when creating new Keycloak clients.

## Import

This resource can be imported using the realm id, which adopts the realm's current
optional client scopes into state. This is useful for reviewing the exact delta between
a realm's existing (e.g. Keycloak factory) optional scopes and your declared set in the
plan, rather than starting from an opaque create.

Example:

```bash
terraform import keycloak_realm_optional_client_scopes.optional_scopes my-realm
```
