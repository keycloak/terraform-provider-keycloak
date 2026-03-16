resource "keycloak_realm" "workflows" {
  realm        = "workflows-example"
  enabled      = true
  display_name = "Workflows Example"
}

resource "keycloak_role" "active_user" {
  realm_id    = keycloak_realm.workflows.id
  name        = "active-user"
  description = "Assigned to users who are actively using the platform"
}

resource "keycloak_group" "workflows_bar" {
  realm_id = keycloak_realm.workflows.id
  name     = "bar"
}

# Onboarding workflow: welcome email on user creation, then prompt password change after 30 days
resource "keycloak_workflow" "onboarding" {
  realm   = keycloak_realm.workflows.id
  name    = "onboarding-new-users"
  on      = "user_created"
  enabled = true

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

# Set the attribute field1 when a user joins any group in this realm
resource "keycloak_workflow" "set_attribute_on_group_join" {
  realm   = keycloak_realm.workflows.id
  name    = "set-attribute-on-group-join"
  on      = "user_group_membership_added"
  enabled = true

  step {
    uses = "set-user-attribute"
    config = {
      field1     = "bar"
    }
  }
}

data "keycloak_workflow" "onboarding" {
  realm = keycloak_realm.workflows.id
  name  = keycloak_workflow.onboarding.name
}

output "onboarding_workflow_id" {
  value = data.keycloak_workflow.onboarding.id
}
