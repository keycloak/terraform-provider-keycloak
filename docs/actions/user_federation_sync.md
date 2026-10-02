---
page_title: "keycloak_user_federation_sync Action"
---

# keycloak\_user\_federation\_sync Action

Triggers a synchronization of users for a user federation provider, e.g. an LDAP user federation.

This is the equivalent of the "Sync all users" and "Sync changed users" actions in the Keycloak admin console.

~> Actions require Terraform 1.14 or later.

## Example Usage

The following example triggers a full sync whenever the LDAP user federation is created or updated:

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_ldap_user_federation" "ldap_user_federation" {
  name     = "openldap"
  realm_id = keycloak_realm.realm.id
  enabled  = true

  username_ldap_attribute = "cn"
  rdn_ldap_attribute      = "cn"
  uuid_ldap_attribute     = "entryDN"
  user_object_classes     = [
    "simpleSecurityObject",
    "organizationalRole"
  ]
  connection_url          = "ldap://openldap"
  users_dn                = "dc=example,dc=org"
  bind_dn                 = "cn=admin,dc=example,dc=org"
  bind_credential         = "admin"

  lifecycle {
    action_trigger {
      events  = [after_create, after_update]
      actions = [action.keycloak_user_federation_sync.ldap]
    }
  }
}

action "keycloak_user_federation_sync" "ldap" {
  config {
    realm_id           = keycloak_realm.realm.id
    user_federation_id = keycloak_ldap_user_federation.ldap_user_federation.id
    mode               = "full"
  }
}
```

-> Terraform 1.16 and later warn about references to the triggering resource. There the triggering resource should be accessed via the `caller` symbol instead, e.g. `realm_id = caller.realm_id` and `user_federation_id = caller.id`. Note that `caller` is not available in earlier Terraform versions.

An action can also be invoked on demand. Note that this requires a dedicated action, which is not used within an `action_trigger` of the resource it refers to:

```hcl
action "keycloak_user_federation_sync" "ldap_on_demand" {
  config {
    realm_id           = keycloak_realm.realm.id
    user_federation_id = keycloak_ldap_user_federation.ldap_user_federation.id
    mode               = "changed_users"
  }
}
```

```shell
terraform apply -invoke=action.keycloak_user_federation_sync.ldap_on_demand
```

## Argument Reference

- `realm_id` - (Required) The realm the user federation provider exists in.
- `user_federation_id` - (Required) The ID of the user federation provider to synchronize, e.g. the ID of a `keycloak_ldap_user_federation` resource.
- `mode` - (Optional) The kind of synchronization to trigger. Can be one of `full` or `changed_users`. Defaults to `full`.
- `client_timeout` - (Optional) Timeout (in seconds) of the sync request. Defaults to the `client_timeout` of the provider.

~> The synchronization runs within a single request to Keycloak. For large directories you might need to increase the `client_timeout` of the action, since the `client_timeout` of the provider defaults to 15 seconds.
