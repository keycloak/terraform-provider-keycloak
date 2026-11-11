---
page_title: "keycloak_client_scope_assignments Resource"
---

# keycloak\_client\_scope\_assignments Resource

Authoritatively manages both default and optional client-scope assignments for one realm or one client. Omit `client_id`
to manage realm defaults; specify the Keycloak-generated client ID to manage a client's assignments. SAML clients support
both assignment categories through Keycloak's default and optional client-scope endpoints. Only `openid-connect` and
`saml` clients are supported, and client assignments must use scopes of the client's protocol.

Declare one resource per target and configure both sets together. A scope cannot appear in both sets; Terraform validates
this overlap during planning when both sets are known, and checks again before writing at apply time. Moving a scope
between sets detaches it from the old category before attaching it to the new one.

During that update, the scope is briefly unassigned because Keycloak exposes separate remove and add operations.
Tokens issued in that interval will not include the scope.

## Example Usage

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_openid_client" "web" {
  realm_id    = keycloak_realm.realm.id
  client_id   = "web"
  access_type = "CONFIDENTIAL"
}

resource "keycloak_client_scope" "groups" {
  realm_id = keycloak_realm.realm.id
  name     = "groups"

  openid_connect {}
}

resource "keycloak_client_scope_assignments" "web_client" {
  realm_id  = keycloak_realm.realm.id
  client_id = keycloak_openid_client.web.id

  default_scopes = ["profile", "email"]
  optional_scopes = [
    keycloak_client_scope.groups.name,
  ]
}
```

## Argument Reference

- `realm_id` - (Required) The target realm.
- `client_id` - (Optional, ForceNew) The Keycloak-generated ID of the client. If omitted, assignments apply to the realm.
- `default_scopes` - (Required) Set of client-scope names assigned as defaults.
- `optional_scopes` - (Required) Set of client-scope names assigned as optional.

The resource is authoritative for the target's assignments. Existing assignments not listed in either set are removed,
including Keycloak's built-in defaults. Include every assignment you want to retain; the provider does not inject built-in
or custom scopes into your configured sets. Realm assignments set defaults for newly created clients; they do not
retroactively update existing clients.

~> Creating this resource for an existing target removes any assignment not listed, and the plan does not show those
removals. Import the target first so the plan shows its current assignments before you apply.

Destroying or replacing this resource detaches the assignments it manages. To prevent accidental destruction or
replacement, use Terraform's built-in `lifecycle.prevent_destroy`:

```hcl
resource "keycloak_client_scope_assignments" "web_client" {
  # ...

  lifecycle {
    prevent_destroy = true
  }
}
```

Declare exactly one assignments resource per target, including across modules and Terraform states. Do not manage the
same target's assignments with a deprecated assignment resource, or manage realm assignments alongside
`keycloak_realm.default_default_client_scopes` or `keycloak_realm.default_optional_client_scopes`. These independent
resources cannot validate each other's ownership and will overwrite each other's assignments.

To migrate without detaching assignments, remove the old resource from configuration, run `terraform state rm` for its
address, then import the target into this resource using the import format below. Review the resulting plan before applying.

## Replacing scope definitions

Assignments refer to scope names, not IDs. Replacing a scope without changing its name (for example, by switching
between its `openid_connect` and `saml` blocks or using `terraform apply -replace=keycloak_client_scope.groups`) deletes
the old scope's assignments.
Terraform may see no change to the configured name sets, so the replacement scope can remain unassigned after that apply.
Referencing the scope's `name` or adding `depends_on` orders operations but does not force reassignment.

To reassign in the same apply, add Terraform's `lifecycle.replace_triggered_by` to each affected assignments resource,
referencing the IDs of the scope definitions it uses:

```hcl
resource "keycloak_client_scope_assignments" "web_client" {
  realm_id        = keycloak_realm.realm.id
  client_id       = keycloak_openid_client.web.id
  default_scopes  = ["profile", "email"]
  optional_scopes = [keycloak_client_scope.groups.name]

  lifecycle {
    replace_triggered_by = [keycloak_client_scope.groups.id]
  }
}
```

This is user-configured Terraform lifecycle behavior, not an automatic provider safeguard. Referencing `id` avoids
replacement for in-place scope edits that leave its ID unchanged. The assignments resource is destroyed and recreated,
briefly detaching all its managed assignments, not just the replaced scope. `prevent_destroy = true` blocks this
replacement; temporarily remove it if the replacement is intended. Any replacement scope must still match the
client's protocol.

If a same-name replacement has already left a scope unassigned, run a normal `terraform plan` and `terraform apply` again.
Refresh detects the missing assignment and the provider reattaches it.

## Import

Import a realm target with `realm/<realm_id>` or a client target with `client/<realm_id>/<client_id>`.

```shell
terraform import keycloak_client_scope_assignments.web_client client/my-realm/8e8f7fe1-df9b-40ed-bed3-4597aa0dac52
```
