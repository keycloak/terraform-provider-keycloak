---
page_title: "keycloak_organization Resource"
---

# keycloak\_organization Resource

Allow for creating and managing Organizations within Keycloak.

Attributes can also be defined on Groups.

Linkage with identity providers is managed with the identity provider resources.

## Example usage

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_organization" "this" {
  realm    = keycloak_realm.realm.name
  name     = "org"
  alias = "org"
  enabled = true
  
  domain {
    name = "example.com"
  }
}

resource "keycloak_oidc_identity_provider" "this" {
  realm             = keycloak_realm.realm.name
  alias             = "my-idp"
  authorization_url = "https://authorizationurl.com"
  client_id         = "clientID"
  client_secret     = "clientSecret"
  token_url         = "https://tokenurl.com"

  organization_id                 = keycloak_organization.this.id
  org_domain                      = "example.com"
  org_redirect_mode_email_matches = true
}

```

## Argument Reference

- `realm` - (Required) The realm this organization exists in.
- `name` - (Required) The name of the organization.
- `alias` - (Optional) The alias unique identifies the organization. Same as the name if not specified. The alias cannot be changed after the organization has been created.
- `description` - (Optional) The description of the organization.
- `redirect_url` - (Optional) The landing page after user completes registration or accepts an invitation to the organization. If left empty, the user will be redirected to the account console by default.
- `domain` - (Required) A list of [domains](#domain-arguments). At least one domain is required.
- `attributes` - (Optional) A map representing attributes for the group. In order to add multivalued attributes, use `##` to separate the values. Max length for each value is 255 chars.

### Domain arguments

- `name` - (Required) The domain name
- `verified` - (Optional) Whether domain is verified or not. Default is false.

## Import

Organizations can be imported using the format `{{realm_id}}/{{organization_id}}`, where `organization_id` is the unique ID that Keycloak
assigns to the organizations upon creation. This value can be found in the URI when editing this organization in the GUI, and is typically a GUID.

Example:

```bash
$ terraform import keycloak_organization.this my-realm/cec54914-b702-4c7b-9431-b407817d059a
```
