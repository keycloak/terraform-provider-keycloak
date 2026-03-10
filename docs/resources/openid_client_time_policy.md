---
page_title: "keycloak_openid_client_time_policy Resource"
---

# keycloak\_openid\_client\_time\_policy Resource

Allows you to manage time policies.

Time policies allow you to define conditions based on time ranges. You can specify when access should be granted using various time constraints including date, month, year, hour, and minute ranges.

## Example Usage

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_openid_client" "test" {
  client_id                = "client_id"
  realm_id                 = keycloak_realm.realm.id
  access_type              = "CONFIDENTIAL"
  service_accounts_enabled = true
  authorization {
    policy_enforcement_mode = "ENFORCING"
  }
}

# Policy for business hours only (9 AM - 5 PM)
resource "keycloak_openid_client_time_policy" "business_hours" {
  resource_server_id = keycloak_openid_client.test.resource_server_id
  realm_id           = keycloak_realm.realm.id
  name               = "business_hours_policy"
  decision_strategy  = "UNANIMOUS"
  logic              = "POSITIVE"

  hour     = "09"
  hour_end = "17"
}

# Policy for specific date range
resource "keycloak_openid_client_time_policy" "date_range" {
  resource_server_id = keycloak_openid_client.test.resource_server_id
  realm_id           = keycloak_realm.realm.id
  name               = "date_range_policy"
  decision_strategy  = "UNANIMOUS"
  logic              = "POSITIVE"

  not_before     = "2024-01-01 00:00:00"
  not_on_or_after = "2024-12-31 23:59:59"
}

# Policy for specific months (January to March)
resource "keycloak_openid_client_time_policy" "quarter1" {
  resource_server_id = keycloak_openid_client.test.resource_server_id
  realm_id           = keycloak_realm.realm.id
  name               = "q1_policy"
  decision_strategy  = "UNANIMOUS"
  logic              = "POSITIVE"

  month     = "1"
  month_end = "3"
}
```

### Argument Reference

The following arguments are supported:

- `realm_id` - (Required) The realm this policy exists in.
- `resource_server_id` - (Required) The ID of the resource server.
- `name` - (Required) The name of the policy.
- `decision_strategy` - (Required) The decision strategy, can be one of `UNANIMOUS`, `AFFIRMATIVE`, or `CONSENSUS`.
- `logic` - (Optional) The logic, can be one of `POSITIVE` or `NEGATIVE`. Defaults to `POSITIVE`.
- `not_before` - (Optional) The policy is valid only after this date/time (format: `YYYY-MM-DD HH:MM:SS`).
- `not_on_or_after` - (Optional) The policy is valid only before this date/time (format: `YYYY-MM-DD HH:MM:SS`).
- `day_month` - (Optional) Starting day of the month (1-31).
- `day_month_end` - (Optional) Ending day of the month (1-31).
- `month` - (Optional) Starting month (1-12).
- `month_end` - (Optional) Ending month (1-12).
- `year` - (Optional) Starting year.
- `year_end` - (Optional) Ending year.
- `hour` - (Optional) Starting hour (0-23).
- `hour_end` - (Optional) Ending hour (0-23).
- `minute` - (Optional) Starting minute (0-59).
- `minute_end` - (Optional) Ending minute (0-59).
- `description` - (Optional) A description for the authorization policy.

### Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

- `id` - Policy ID representing the time policy.

## Import

Time policies can be imported using the format: `{{realmId}}/{{resourceServerId}}/{{policyId}}`.

Example:

```bash
$ terraform import keycloak_openid_client_time_policy.test my-realm/3bd4a686-1062-4b59-97b8-e4e3f10b99da/63b3cde8-987d-4cd9-9306-1955579281d9
```
