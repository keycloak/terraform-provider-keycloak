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

### Onboard gold members only

This mirrors the [common use cases guide](https://www.keycloak.org/docs/latest/server_admin/index.html#_understanding_common_use_cases).
The `conditions` expression restricts the workflow to users that have the `membership=gold` attribute, then
sends a welcome message and grants the `gold-member` role.

```hcl
resource "keycloak_workflow" "onboard_gold_members" {
  realm      = keycloak_realm.realm.id
  name       = "onboarding-gold-members"
  on         = "user_created"
  enabled    = true
  conditions = "has-user-attribute(membership=gold)"

  step {
    uses = "notify-user"
    config = {
      message = "Welcome to the Gold Membership program!"
    }
  }

  step {
    uses = "grant-role"
    config = {
      role = "gold-member"
    }
  }
}
```

### Track inactive users on a schedule

This example mirrors the [scheduling workflows guide](https://www.keycloak.org/docs/latest/server_admin/index.html#_scheduling_workflows).
Instead of reacting to a single event, the workflow engine periodically scans realm resources (here, every
`30s`, up to `100` users per run) and progresses each matching user through the steps. `restart_in_progress`
restarts an in-progress execution when the user authenticates again, so the inactivity countdown resets.

```hcl
resource "keycloak_workflow" "track_inactive_users" {
  realm               = keycloak_realm.realm.id
  name                = "track-inactive-users"
  on                  = "user_authenticated"
  enabled             = true
  restart_in_progress = "true"

  schedule {
    after      = "30s"
    batch_size = 100
  }

  step {
    uses  = "notify-user"
    after = "180d"
    config = {
      message = "It has been a while since your last login. We miss you!"
    }
  }

  step {
    uses  = "notify-user"
    after = "60d"
    config = {
      message = "Your account will be disabled in $${workflow.daysUntilNextStep} days!"
    }
  }

  step {
    uses  = "disable-user"
    after = "7d"
  }
}
```

## Argument Reference

- `realm` - (Required) The realm this workflow exists in. Changing this forces a new resource.
- `name` - (Required) The name of the workflow.
- `on` - (Required) The realm event that triggers the workflow. Supported values: `user_created`, `user_removed`, `user_authenticated`, `user_federated_identity_added`, `user_federated_identity_removed`, `user_group_membership_added`, `user_group_membership_removed`, `user_role_granted`, `user_role_revoked`.
- `enabled` - (Optional) Whether the workflow is enabled. Defaults to `true`.
- `conditions` - (Optional) An expression that must evaluate to true for the workflow to run (e.g. `has-role('some-role')`).
- `cancel_in_progress` - (Optional) Whether to cancel an already in-progress execution when the workflow is re-triggered for the same resource. Set to `"true"` to enable.
- `restart_in_progress` - (Optional) Whether to restart an already in-progress execution (resetting it to the first step) when the workflow is re-triggered for the same resource. Set to `"true"` to enable.
- `schedule` - (Optional) A [schedule block](#schedule-arguments) that makes the workflow run periodically over matching realm resources instead of (or in addition to) reacting to a single event.
- `step` - (Required) One or more [step blocks](#step-arguments) defining the actions to execute, in order.

### Schedule arguments

- `after` - (Optional) Interval between successive scheduled runs, as a duration string (e.g. `30s`, `1d`).
- `batch_size` - (Optional) Maximum number of realm resources processed during each scheduled run.

### Step arguments

- `uses` - (Required) The step type. Built-in values: `disable-user`, `delete-user`, `notify-user`, `set-user-required-action`, `set-user-attribute`.
- `after` - (Optional) Delay before executing this step, as a duration string (e.g. `7d`) or milliseconds (e.g. `2592000000`).
- `priority` - (Optional) Execution priority used to order steps that would otherwise run at the same time.
- `config` - (Optional) A map of key-value pairs configuring the step (e.g. `emailTemplate` for `notify-user`).

## Attributes Reference

In addition to the arguments above, the following read-only attributes are exported:

- `state` - The runtime state of the workflow as reported by Keycloak. Contains:
  - `errors` - A list of error messages recorded for the workflow.
- Each `step` block additionally exports:
  - `scheduled_at` - Epoch timestamp in milliseconds at which the step is scheduled to execute.
  - `status` - The execution status of the step.

## Import

Workflows can be imported using the format `{{realm}}/{{workflow_id}}`, where `realm` is the realm name and `workflow_id` is the unique ID Keycloak assigns upon creation.

```bash
$ terraform import keycloak_workflow.disable_inactive my-realm/cec54914-b702-4c7b-9431-b407817d059a
```
