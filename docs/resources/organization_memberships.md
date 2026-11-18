---
page_title: "keycloak_organization_memberships Resource"
---

# keycloak\_organization\_memberships Resource

Allows for managing the members of a Keycloak organization.

Note that this resource attempts to be an **authoritative** source over organization members. When this resource takes
control over an organization's members, users that are manually added to the organization will be removed, and users that
are manually removed will be re-added upon the next run of `terraform apply`.

Managed members are the exception. A membership is managed (`MANAGED` in Keycloak) when, for example, the user joined
the organization by signing in through an identity provider linked to it (from Keycloak 26.8, a link with managed
membership), or when an administrator changed the membership type to managed (Keycloak 26.8 and later). Keycloak
deletes the user account when a managed membership is removed, so this resource never removes a managed member, listed
in `members` or not: create, update and destroy leave it in the organization. A managed member is in the resource's
state only while it is listed in `members`, so unlisted managed members never cause a diff, and removing a managed
member from `members` removes it from the state, not from the organization.

This resource paginates its data loading on refresh by 50 items.

## Example Usage

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_organization" "org" {
  realm = keycloak_realm.realm.id
  name  = "my-org"

  domain {
    name = "example.com"
  }
}

resource "keycloak_user" "alice" {
  realm_id = keycloak_realm.realm.id
  username = "alice"
  email    = "alice@example.com"
}

resource "keycloak_user" "bob" {
  realm_id = keycloak_realm.realm.id
  username = "bob"
  email    = "bob@example.com"
}

resource "keycloak_organization_memberships" "org_members" {
  realm_id        = keycloak_realm.realm.id
  organization_id = keycloak_organization.org.id

  members = [
    keycloak_user.alice.username,
    keycloak_user.bob.username,
  ]
}
```

## Argument Reference

- `realm_id` - (Required) The realm the organization belongs to.
- `organization_id` - (Required) The ID of the organization to manage memberships for.
- `members` - (Required) A set of usernames to assign as members of the organization. Managed members may be listed, but are never removed.

## Import

This resource does not support import. Instead of importing, feel free to create this resource
as if it did not already exist on the server.
