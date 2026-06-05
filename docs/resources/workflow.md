---
page_title: "keycloak_workflow Resource"
---

# keycloak\_workflow Resource

Allows creating and managing workflows within Keycloak.

Workflows automate administrative tasks in response to realm events (e.g. onboarding new users, sending notifications). This feature requires Keycloak 26.4 or later and must be enabled with `--features=workflows` at startup.

## Example Usage

### Onboard new users

This example mirrors the [official Keycloak workflow guide](https://www.keycloak.org/docs/latest/server_admin/index.html#_understanding_workflow_definition_).
The example workflow sends the user a welcome notification, then requires them to update their password after 30 days.
Finally, it restarts the workflow so the password update requirement repeats every 30 days.

The workflow triggers on every `user_created` event.

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_workflow" "onboarding" {
  realm              = keycloak_realm.realm.id
  name               = "onboarding-new-users"
  on                 = "user_created"
  enabled            = true

  step {
    uses = "notify-user"
    config = {
      message = <<-EOT
        <p>Dear $${user.firstName} $${user.lastName}, </p>
        <p>Welcome to $${realm.displayName}!</p>
        <p>
           Best regards,<br/>
           The Keycloak Team
        </p>
      EOT
    }
  }

  step {
    uses  = "set-user-required-action"
    after = "2592000000" # 30 days in milliseconds
    config = {
      action = "UPDATE_PASSWORD"
    }
  }

  # restart is a workflow control-flow step (loop back to step 1)
  step {
    uses = "restart"
    config = {
      position = "1"
    }
  }
}
```

### Notify then delete a user after account creation

```hcl
resource "keycloak_workflow" "offboard" {
  realm   = keycloak_realm.realm.id
  name    = "offboard-users"
  on      = "user_created"
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
- `on` - (Required) The realm event that triggers the workflow. Supported values: `user_created`, `user_removed`, `user_authenticated`, `user_federated_identity_added`, `user_federated_identity_removed`, `user_group_membership_added`, `user_group_membership_removed`, `user_role_granted`, `user_role_revoked`.
- `enabled` - (Optional) Whether the workflow is enabled. Defaults to `true`.
- `conditions` - (Optional) An expression that must evaluate to true for the workflow to run (e.g. `has-role('some-role')`).
- `cancel_in_progress` - (Optional) Event that cancels an in-progress workflow execution.
- `restart_in_progress` - (Optional) Event that restarts an in-progress workflow execution.
- `step` - (Required) One or more [step blocks](#step-arguments) defining the actions to execute, in order.

### Step arguments

- `uses` - (Required) The step type. Built-in values: `disable-user`, `delete-user`, `notify-user`, `set-user-required-action`, `set-user-attribute`.
- `after` - (Optional) Delay in milliseconds before executing this step.
- `config` - (Optional) A map of key-value pairs configuring the step (e.g. `emailTemplate` for `notify-user`).

## Import

Workflows can be imported using the format `{{realm}}/{{workflow_id}}`, where `realm` is the realm name and `workflow_id` is the unique ID Keycloak assigns upon creation.

```bash
$ terraform import keycloak_workflow.disable_inactive my-realm/cec54914-b702-4c7b-9431-b407817d059a
```
