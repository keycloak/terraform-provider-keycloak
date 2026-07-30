---
page_title: "keycloak_spiffe_identity_provider Resource"
---

# keycloak\_spiffe\_identity\_provider Resource

Allows for creating and managing SPIFFE Identity Providers within Keycloak. A SPIFFE identity provider supports authenticating clients with SPIFFE JWT SVIDs.

> **NOTICE:**
> This is part of a preview keycloak feature. You need to enable this feature to be able to use this resource.
> More information about enabling the preview feature can be found here: https://www.keycloak.org/docs/latest/server_admin/index.html#_identity_broker_spiffe

## Example Usage
```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_spiffe_identity_provider" "example" {
  realm           = keycloak_realm.realm.id
  alias           = "my-spiffe-idp"
  trust_domain    = "spiffe://my-trust-domain"
  bundle_endpoint = "https://example.com/spiffe/bundle"
}

resource "keycloak_openid_client" "spiffe_client" {
  realm_id  = keycloak_realm.realm.id
  client_id = "spiffe-client"

  name    = "SPIFFE Client"
  enabled = true

  access_type               = "CONFIDENTIAL"
  service_accounts_enabled  = true
  client_authenticator_type = "federated-jwt"
  extra_config = {
    "jwt.credential.issuer" = keycloak_spiffe_identity_provider.example.alias
    "jwt.credential.sub"    = "spiffe://my-trust-domain/workload"
  }
}
```

## Argument Reference

- `realm` - (Required) The name of the realm. This is unique across Keycloak.
- `alias` - (Required) The alias uniquely identifies an identity provider, and it is also used to build the redirect uri.
- `trust_domain` - (Required) The SPIFFE trust domain. This must use the `spiffe://` scheme.
- `bundle_endpoint` - (Required) The SPIFFE bundle endpoint or OpenID Connect JWKS endpoint exposing SPIFFE public keys. Depending on your Keycloak Realm `ssl_required` setting, this may need to be an HTTPS URL.

## Import

Identity providers can be imported using the format `{{realm_id}}/{{idp_alias}}`, where `idp_alias` is the identity provider alias.

Example:

```bash
$ terraform import keycloak_spiffe_identity_provider.realm_identity_provider my-realm/my-idp
```
