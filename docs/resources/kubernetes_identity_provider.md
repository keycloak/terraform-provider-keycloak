---
page_title: "keycloak_kubernetes_identity_provider Resource"
---

# keycloak\_kubernetes_\_identity\_provider Resource

Allows for creating and managing Kubernetes Identity Providers within Keycloak. Workloads inside a Kubernetes cluster can authenticate using service account tokens.

> **NOTICE:**
> This is part of a preview keycloak feature. You need to enable this feature to be able to use this resource.
> More information about enabling the preview feature can be found here: https://www.keycloak.org/docs/latest/server_admin/index.html#_identity_broker_kubernetes

## Example Usage

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_kubernetes_identity_provider" "kubernetes" {
  realm   = keycloak_realm.realm.id
  alias   = "my-k8s-idp"
  issuer  = "https://example.com/issuer"
}

## Argument Reference

- `realm` - (Required) The name of the realm. This is unique across Keycloak.
- `alias` - (Required) The alias uniquely identifies an identity provider, and it is also used to build the redirect uri.
- `issuer` - (Required) The Kubernetes issuer URL of service account tokens. The URL <ISSUER>.well-known/openid-configuration must be available to Keycloak.
- `hide_on_login_page` - (Optional) When `true`, this provider will be hidden on the login page, and is only accessible when requested explicitly. Defaults to `true`.
