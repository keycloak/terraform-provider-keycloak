---
page_title: "keycloak_openid_client_initial_access_token Resource"
---

# keycloak\_openid\_client\_initial\_access\_token Resource

Allows for managing Keycloak clients initial access tokens within a given realm.

The initial access token is a token used in the Dynamic client registration process.

Dynamic client registration protocol provides a mechanism for a client to register new clients with the Identity Provider.


## Example Usage

In this example, we'll create a new Realm and a initial access token. We'll use the `keycloak_openid_client_initial_access_token` resource to create new initial access token
which can be used to register new two clients and it will expire in 4 days (expiration attribute is in seconds).

```hcl
resource "keycloak_realm" "realm" {
	realm   = "my-realm"
	enabled = true
}

resource "keycloak_openid_client_initial_access_token" "test_initial_access_token" {
	realm_id = keycloak_realm.realm.id
	token_count = 2
	expiration = 345600
}

```

## Argument Reference

- `realm_id` - (Required) The realm this initial access token exists within.
- `count` - (Optional) Specifies how many clients can be created using the token.
- `expiration` - (Optional) Specifies how long the token should be valid.

## Attributes Reference

- `remaining_count` - (Computed) Dictates how many times left the token can be used for creating new clients. Equals to count argument value after token is issued.
- `token_value` - (Computed) Is used to create new clients.
