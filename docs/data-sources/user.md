---
page_title: "keycloak_user Data Source"
---

# keycloak\_user Data Source

This data source can be used to fetch properties of a user within Keycloak.

## Example Usage

```hcl
data "keycloak_realm" "master_realm" {
  realm = "master"
}

// use the keycloak_user data source to grab the admin user's ID
data "keycloak_user" "default_admin_user" {
  realm_id = data.keycloak_realm.master_realm.id
  username = "keycloak"
}

output "keycloak_user_id" {
  value = data.keycloak_user.default_admin_user.id
}
```

## Argument Reference

- `realm_id` - (Required) The realm this user belongs to.
- `username` - (Required) The unique username of this user.

## Attributes Reference

- `id` - (Computed) The unique ID of the user, which can be used as an argument to other resources supported by this provider.
- `enabled` - (Computed) When false, this user cannot log in. Defaults to `true`.
- `email` - (Computed) The user's email.
- `email_verified` - (Computed) Whether the email address was validated or not. Default to `false`.
- `first_name` - (Computed) The user's first name.
- `last_name` - (Computed) The user's last name.
- `attributes` - (Computed) A map representing attributes for the user
- `federated_identity` - (Computed) The user's federated identities, if applicable. This block has the following schema:
  - `identity_provider` - (Computed) The name of the identity provider
  - `user_id` - (Computed) The ID of the user defined in the identity provider
  - `user_name` - (Computed) The username of the user defined in the identity provider
