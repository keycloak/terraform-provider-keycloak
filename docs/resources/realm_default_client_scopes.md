---
page_title: "keycloak_realm_default_client_scopes Resource"
---

# keycloak\_realm\_default\_client\_scopes Resource

Allows you to manage the set of default client scopes for a Keycloak realm, which are used when new clients are created.

Note that this resource attempts to be an **authoritative** source over the default client scopes for a Keycloak realm,
so any Keycloak defaults and manual adjustments will be overwritten.

Additionally, if a scope is set as optional, this resource will remove and re-add it as a default scope.

Warning: Configuring the same scope in both this resource and `keycloak_realm_optional_client_scopes` for the same realm
is an error. When the realm and scope names are known during planning, the plan fails with
`scopes [...] are configured as both default and optional scopes`. Apply-time validation only sees resources that
reconcile during that apply. If one resource is unchanged and the other resource's scope names or realm identity are
unknown until apply, the changed resource can move a scope managed by the unchanged resource without reporting a
conflict. This also applies when applying a saved plan. See
[Validating default and optional scopes](#validating-default-and-optional-scopes) for a workaround.

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

resource "keycloak_realm_default_client_scopes" "default_scopes" {
  realm_id  = keycloak_realm.realm.id

  default_scopes = [
    "profile",
    "email",
    "roles",
    "web-origins",
    keycloak_openid_client_scope.client_scope.name,
  ]
}
```

## Validating default and optional scopes

When managing both scope types for the same realm, use shared desired scope sets and add the same lifecycle
precondition to **both** resources. This requires Terraform 1.2 or later. Terraform checks known conditions during
planning and defers unknown conditions until apply, before changing the scope assignments.

Use the following pattern instead of the standalone scope resource above. Replace the example lists with your desired
scope names or expressions, including values that are only known during apply:

```hcl
locals {
  default_scopes  = ["profile", "email", "roles", "web-origins", keycloak_openid_client_scope.client_scope.name]
  optional_scopes = ["address", "phone", "offline_access", "microprofile-jwt"]

  scopes_disjoint = length(setintersection(
    toset(local.default_scopes),
    toset(local.optional_scopes),
  )) == 0
}

resource "keycloak_realm_default_client_scopes" "default_scopes" {
  realm_id       = keycloak_realm.realm.id
  default_scopes = local.default_scopes

  lifecycle {
    precondition {
      condition     = local.scopes_disjoint
      error_message = "A client scope cannot be both default and optional for the same realm."
    }
  }
}

resource "keycloak_realm_optional_client_scopes" "optional_scopes" {
  realm_id        = keycloak_realm.realm.id
  optional_scopes = local.optional_scopes

  lifecycle {
    precondition {
      condition     = local.scopes_disjoint
      error_message = "A client scope cannot be both default and optional for the same realm."
    }
  }
}
```

The condition must compare the same desired values passed to the resource arguments, not the scope resources'
attributes. Referencing the resources from each other's preconditions can create a dependency cycle. An explicit
`depends_on` alone does not close the validation gap, because it does not make an unchanged resource reconcile.

## Argument Reference

- `realm_id` - (Required) The realm this client and scopes exists in.
- `default_scopes` - (Required) An array of default client scope names that should be used when creating new Keycloak clients.

## Import

This resource can be imported using the realm id, which adopts the realm's current
default client scopes into state. This is useful for reviewing the exact delta between
a realm's existing (e.g. Keycloak factory) default scopes and your declared set in the
plan, rather than starting from an opaque create.

Example:

```bash
terraform import keycloak_realm_default_client_scopes.default_scopes my-realm
```
