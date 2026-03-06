---
page_title: "keycloak_workflow Resource"
---

# keycloak\_workflow Resource

Allow for creating and managing Workflows within Keycloak.

Workflows automate administrative tasks in response to realm events (e.g. disabling inactive users, sending notifications). This feature requires Keycloak 26.4 or later and must be enabled with `--features=workflows` at startup.

## Example Usage

### Disable inactive users

This example mirrors the [official Keycloak workflow guide](https://www.keycloak.org/docs/latest/server_admin/index.html#_understanding_workflow_definition_).
The workflow triggers on every `USER_LOGIN` event and resets on the same event.
Users who do not hold the `realm-management/realm-admin` role will first receive
a warning notification, and then have their account disabled — each step running
30 seconds after the previous one.

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_workflow" "disable_inactive" {
  realm              = keycloak_realm.realm.name
  name               = "disable inactive users"
  on                 = "USER_LOGIN"
  restart_in_progress = "USER_LOGIN"
  enabled            = true
  conditions         = "!has-role(\"realm-management/realm-admin\")"

  step {
    uses  = "notify-user"
    after = "30000"
    config = {
      custom_message = "Your account can be disabled due to inactivity!"
    }
  }

  step {
    uses  = "disable-user"
    after = "30000"
  }
}
```

### Notify then delete a user after account creation

```hcl
resource "keycloak_workflow" "offboard" {
  realm   = keycloak_realm.realm.name
  name    = "offboard-users"
  on      = "USER_ADD"
  enabled = true

  step {
    uses  = "notify-user"
    after = "2505600000"
    config = {
      emailTemplate = "offboarding-warning"
    }
  }

  step {
    uses  = "delete-user"
    after = "2592000000"
  }
}
```

## Argument Reference

- `realm` - (Required) The realm this workflow exists in. Changing this forces a new resource.
- `name` - (Required) The name of the workflow.
- `on` - (Required) The realm event that triggers the workflow. Supported values: `USER_LOGIN`, `USER_ADD`, `USER_GROUP_MEMBERSHIP_ADD`, `USER_ROLE_ADD`.
- `enabled` - (Optional) Whether the workflow is enabled. Defaults to `true`.
- `conditions` - (Optional) An expression that must evaluate to true for the workflow to run (e.g. `!has-role('some-role')`).
- `cancel_in_progress` - (Optional) Event that cancels an in-progress workflow execution.
- `restart_in_progress` - (Optional) Event that restarts an in-progress workflow execution.
- `step` - (Required) One or more [step blocks](#step-arguments) defining the actions to execute, in order.

### Step arguments

- `uses` - (Required) The step type. Built-in values: `disable-user`, `delete-user`, `notify-user`.
- `after` - (Optional) Delay in milliseconds before executing this step.
- `config` - (Optional) A map of key-value pairs configuring the step (e.g. `emailTemplate` for `notify-user`).

## Import

Workflows can be imported using the format `{{realm_id}}/{{workflow_id}}`, where `workflow_id` is the unique ID Keycloak assigns upon creation.

```bash
$ terraform import keycloak_workflow.disable_inactive my-realm/cec54914-b702-4c7b-9431-b407817d059a
```
