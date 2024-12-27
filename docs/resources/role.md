---
page_title: "keycloak_role Resource"
---

# keycloak\_role Resource

Allows for creating and managing roles within Keycloak.

Roles allow you define privileges within Keycloak and map them to users and groups.

## Example Usage (Realm role)

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_role" "realm_role" {
  realm_id    = keycloak_realm.realm.id
  name        = "my-realm-role"
  description = "My Realm Role"
  attributes = {
    key = "value"
    multivalue = "value1##value2"
  }
}
```

## Example Usage (Client role)

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_openid_client" "openid_client" {
  realm_id  = keycloak_realm.realm.id
  client_id = "client"

  name    = "client"
  enabled = true

  access_type         = "CONFIDENTIAL"
  valid_redirect_uris = [
    "http://localhost:8080/openid-callback"
  ]
}

resource "keycloak_role" "client_role" {
  realm_id    = keycloak_realm.realm.id
  client_id   = keycloak_client.openid_client.id
  name        = "my-client-role"
  description = "My Client Role"
  attributes = {
    key = "value"
  }
}
```

## Example Usage (Composite role)

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

# realm roles

resource "keycloak_role" "create_role" {
  realm_id = keycloak_realm.realm.id
  name     = "create"
  attributes = {
    key = "value"
  }
}

resource "keycloak_role" "read_role" {
  realm_id = keycloak_realm.realm.id
  name     = "read"
  attributes = {
    key = "value"
  }
}

resource "keycloak_role" "update_role" {
  realm_id = keycloak_realm.realm.id
  name     = "update"
  attributes = {
    key = "value"
  }
}

resource "keycloak_role" "delete_role" {
  realm_id = keycloak_realm.realm.id
  name     = "delete"
  attributes = {
    key = "value"
  }
}

# client role

resource "keycloak_openid_client" "openid_client" {
  realm_id  = keycloak_realm.realm.id
  client_id = "client"

  name    = "client"
  enabled = true

  access_type         = "CONFIDENTIAL"
  valid_redirect_uris = [
    "http://localhost:8080/openid-callback"
  ]
}

resource "keycloak_role" "client_role" {
  realm_id    = keycloak_realm.realm.id
  client_id   = keycloak_client.openid_client.id
  name        = "my-client-role"
  description = "My Client Role"

  attributes = {
    key = "value"
  }
}

resource "keycloak_role" "admin_role" {
  realm_id        = keycloak_realm.realm.id
  name            = "admin"
  composite_roles = [
    keycloak_role.create_role.id,
    keycloak_role.read_role.id,
    keycloak_role.update_role.id,
    keycloak_role.delete_role.id,
    keycloak_role.client_role.id,
  ]

   attributes = {
    key = "value"
  }
}
```

## Argument Reference

- `realm_id` - (Required) The realm this role exists within.
- `name` - (Required) The name of the role
- `client_id` - (Optional) When specified, this role will be created as a client role attached to the client with the provided ID
- `description` - (Optional) The description of the role
- `composite_roles` - (Optional) When specified, this role will be a composite role, composed of all roles that have an ID present within this list.
- `attributes` - (Optional) A map representing attributes for the role. In order to add multivalue attributes, use `##` to seperate the values. Max length for each value is 255 chars
- `import` - (Optional) When `true`, the role with the specified `name` is assumed to already exist, and it will be imported into state instead of being created. This attribute is useful when dealing with roles that Keycloak creates automatically during realm creation, such as the client roles `create-client`, `view-realm`, ... for the client `realm-management` created per realm. Note, that the role will not be removed during destruction if `import` is `true`.


## Import

Roles can be imported using the format `{{realm_id}}/{{role_id}}`, where `role_id` is the unique ID that Keycloak assigns
to the role. The ID is not easy to find in the GUI, but it appears in the URL when editing the role.

Example:

```bash
$ terraform import keycloak_role.role my-realm/7e8cf32a-8acb-4d34-89c4-04fb1d10ccad
```
