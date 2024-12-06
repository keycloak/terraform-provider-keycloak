resource keycloak_realm test_authorization {
  realm                = "test_authorization"
  enabled              = true
  display_name         = "foo"
  account_theme        = "base"
  access_code_lifespan = "30m"
}

resource keycloak_openid_client test {
  client_id                = "test-openid-client"
  name                     = "test-openid-client"
  realm_id                 = keycloak_realm.test_authorization.id
  description              = "a test openid client"
  standard_flow_enabled    = true
  service_accounts_enabled = true
  access_type              = "CONFIDENTIAL"
  client_secret            = "secret"
  valid_redirect_uris = [
    "http://localhost:5555/callback",
  ]
  authorization {
    policy_enforcement_mode = "ENFORCING"
  }
}

#
# create aggregate_policy
#

resource keycloak_role test_authorization {
  realm_id = keycloak_realm.test_authorization.id
  name     = "aggregate_policy_role"
}

resource keycloak_openid_client_role_policy test {
  resource_server_id = keycloak_openid_client.test.resource_server_id
  realm_id           = keycloak_realm.test_authorization.id
  name               = "keycloak_openid_client_role_policy"
  decision_strategy  = "UNANIMOUS"
  logic              = "POSITIVE"
  type               = "role"
  role {
    id       = keycloak_role.test_authorization.id
    required = false
  }
}

resource keycloak_openid_client_aggregate_policy test {
  resource_server_id = keycloak_openid_client.test.resource_server_id
  realm_id           = keycloak_realm.test_authorization.id
  name               = "keycloak_openid_client_aggregate_policy"
  decision_strategy  = "UNANIMOUS"
  logic              = "POSITIVE"
  policies           = [keycloak_openid_client_role_policy.test.id]
}

#
# create client policy
#

resource keycloak_openid_client_client_policy test {
  resource_server_id = keycloak_openid_client.test.resource_server_id
  realm_id           = keycloak_realm.test_authorization.id
  name               = "keycloak_openid_client_client_policy"
  decision_strategy  = "AFFIRMATIVE"
  logic              = "POSITIVE"
  clients            = [keycloak_openid_client.test.resource_server_id]
}

#
# create group policy
#

resource keycloak_group test {
  realm_id = keycloak_realm.test_authorization.id
  name     = "foo"
}

resource keycloak_openid_client_group_policy test {
  resource_server_id = keycloak_openid_client.test.resource_server_id
  realm_id           = keycloak_realm.test_authorization.id
  name               = "client_group_policy_test"
  groups {
    id              = keycloak_group.test.id
    path            = keycloak_group.test.path
    extend_children = false
  }
  logic             = "POSITIVE"
  decision_strategy = "UNANIMOUS"
}


#
#  create role policy
#

resource keycloak_role test_authorization2 {
  realm_id = keycloak_realm.test_authorization.id
  name     = "new_role"
}

resource keycloak_openid_client_role_policy test1 {
  resource_server_id = keycloak_openid_client.test.resource_server_id
  realm_id           = keycloak_realm.test_authorization.id
  name               = "keycloak_openid_client_role_policy1"
  decision_strategy  = "AFFIRMATIVE"
  logic              = "POSITIVE"
  type               = "role"
  role {
    id       = keycloak_role.test_authorization2.id
    required = false
  }
}

#
# create time policy
#

resource keycloak_openid_client_time_policy test {
  resource_server_id = keycloak_openid_client.test.resource_server_id
  realm_id           = keycloak_realm.test_authorization.id
  name               = "%s"
  not_on_or_after    = "2500-12-12 01:01:11"
  not_before         = "2400-12-12 01:01:11"
  day_month          = "1"
  day_month_end      = "2"
  year               = "2500"
  year_end           = "2501"
  month              = "1"
  month_end          = "5"
  hour               = "1"
  hour_end           = "5"
  minute             = "10"
  minute_end         = "30"
  logic              = "POSITIVE"
  decision_strategy  = "UNANIMOUS"
}

#
# create user policy
#

resource keycloak_user test {
  realm_id = keycloak_realm.test_authorization.id
  username = "test-user"

  email      = "test-user@fakedomain.com"
  first_name = "Testy"
  last_name  = "Tester"
}

resource keycloak_openid_client_user_policy test {
  resource_server_id = keycloak_openid_client.test.resource_server_id
  realm_id           = keycloak_realm.test_authorization.id
  name               = "client_user_policy_test"
  users              = [keycloak_user.test.id]
  logic              = "POSITIVE"
  decision_strategy  = "UNANIMOUS"
}

# users permissions

resource "keycloak_users_permissions" "my_permission" {
  realm_id = keycloak_realm.test_authorization.id

  view_scope {
    policies          = [
      keycloak_openid_client_user_policy.test.id
    ]
    description       = "view_scope"
    decision_strategy = "CONSENSUS"
  }

  manage_scope {
    policies          = [
      keycloak_openid_client_user_policy.test.id
    ]
    description       = "manage_scope"
    decision_strategy = "UNANIMOUS"
  }
}

resource "keycloak_openid_client_permissions" "my_permission" {
  realm_id  = keycloak_realm.test_authorization.id
  client_id = keycloak_openid_client.test.id

  view_scope {
    policies          = [
      keycloak_openid_client_user_policy.test.id,
    ]
    description       = "my description"
    decision_strategy = "UNANIMOUS"
  }
  //needed because otherwise there is a conflict/race condition with the other permission
  depends_on = [
    keycloak_users_permissions.my_permission
  ]
}
