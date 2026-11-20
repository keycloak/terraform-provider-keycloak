---
page_title: "keycloak_realm_events Resource"
---

# keycloak_realm_events Resource

Allows for managing Realm Events settings within Keycloak.

## Example Usage

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_realm_events" "realm_events" {
  realm_id = keycloak_realm.realm.id

  events_enabled    = true
  events_expiration = 3600

  admin_events_enabled         = true
  admin_events_details_enabled = true

  # When omitted or left empty, keycloak will enable all event types
  enabled_event_types = [
    "LOGIN",
    "LOGOUT",
  ]

  events_listeners = [
    "jboss-logging", # keycloak enables the 'jboss-logging' event listener by default.
  ]
}
```

## Argument Reference

- `realm_id` - (Required) The name of the realm the event settings apply to.
- `admin_events_enabled` - (Optional) When `true`, admin events are saved to the database, making them available through the admin console. Defaults to `false`.
- `admin_events_details_enabled` - (Optional) When `true`, saved admin events will included detailed information for create/update requests. Defaults to `false`.
- `events_enabled` - (Optional) When `true`, events from `enabled_event_types` are saved to the database, making them available through the admin console. Defaults to `false`.
- `events_expiration` - (Optional) The amount of time in seconds events will be saved in the database. Defaults to `0` or never.
- `enabled_event_types` - (Optional) The event types that will be saved to the database. Omitting this field enables all event types. Defaults to `[]` or all event types.
- `events_listeners` - (Optional) The event listeners that events should be sent to. Defaults to `[]` or none. Note that new realms enable the `jboss-logging` listener by default, and this resource will remove that unless it is specified.

## Import

This resource can be imported using the realm ID.

Example:

```bash
$ terraform import keycloak_realm_events.realm_events my-realm
```

Note: `enabled_event_types` is not populated on import (the provider only tracks it when set in configuration). If you manage the list, declare it explicitly; the first apply records it in state.
