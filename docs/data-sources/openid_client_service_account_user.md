---
page_title: "keycloak_openid_client_service_account_user Data Source"
---

# keycloak\_openid\_client\_service\_account\_user Data Source

This data source can be used to fetch information about the service account user that is associated with an OpenID client
that has service accounts enabled.

## Example Usage

In this example, we'll create an OpenID client with service accounts enabled. This causes Keycloak to create a special user
that represents the service account. We'll use this data source to grab this user's ID in order to assign some roles to this
user, using the `keycloak_user_roles` resource.

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_openid_client" "client" {
  realm_id  = keycloak_realm.realm.id
  client_id = "client"
  name      = "client"

  access_type              = "CONFIDENTIAL"
  service_accounts_enabled = true
}

data "keycloak_openid_client_service_account_user" "service_account_user" {
  realm_id  = keycloak_realm.realm.id
  client_id = keycloak_openid_client.client.id
}

data "keycloak_role" "offline_access" {
  realm_id = keycloak_realm.realm.id
  name     = "offline_access"
}

resource "keycloak_user_roles" "service_account_user_roles" {
  realm_id = keycloak_realm.realm.id
  user_id  = data.keycloak_openid_client_service_account_user.service_account_user.id

  role_ids = [
    data.keycloak_role.offline_access.id
  ]
}
```

## Argument Reference

- `realm_id` - (Required) The realm that the OpenID client exists within.
- `client_id` - (Required) The ID of the OpenID client with service accounts enabled.

## Attributes Reference

- `username` - (Computed) The service account user's username.  
- `email` - (Computed) The service account user's email.  
- `first_name` - (Computed) The service account user's first name.  
- `last_name` - (Computed) The service account user's last name.  
- `enabled` - (Computed) Whether the service account user is enabled.  
- `attributes` - (Computed) The service account user's attributes.  
- `federated_identity` - (Computed) This attribute exists in order to adhere to the spec of a Keycloak user, but a service account user will never have a federated identity, so this will always be `null`.  
