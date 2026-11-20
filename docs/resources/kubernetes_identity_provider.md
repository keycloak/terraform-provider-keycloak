---
page_title: "keycloak_kubernetes_identity_provider Resource"
---

# keycloak\_kubernetes\_identity\_provider Resource

Allows for creating and managing Kubernetes Identity Providers within Keycloak. Workloads inside a Kubernetes cluster can authenticate using service account tokens.

> **NOTICE:**
> This is part of a preview keycloak feature. You need to enable this feature to be able to use this resource.
> More information about enabling the preview feature can be found here: https://www.keycloak.org/docs/latest/server_admin/index.html#_identity_broker_kubernetes

## Example Usage with an OpenID client
```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_kubernetes_identity_provider" "kubernetes" {
  realm   = keycloak_realm.realm.id
  alias   = "my-k8s-idp"
  issuer  = "https://example.com/issuer/"
}

resource "keycloak_openid_client" "k8s_client" {
  realm_id  = keycloak_realm.realm.id
  client_id = "k8s-client"

  name    = "K8s Client"
  enabled = true

  access_type               = "CONFIDENTIAL"
  service_accounts_enabled  = true
  client_authenticator_type = "federated-jwt"
  extra_config = {
    "jwt.credential.issuer" = keycloak_kubernetes_identity_provider.kubernetes.alias
    "jwt.credential.sub"    = "system:serviceaccount:<namespace>:<service-account-name>"
  }
}
```

## Example Usage with a Kubernetes workload authentication

### Keycloak configuration

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_kubernetes_identity_provider" "kubernetes" {
  realm   = keycloak_realm.realm.id
  alias   = "my-k8s-idp"
  issuer  = "https://example.com/issuer/"
}

resource "keycloak_openid_client" "k8s_client" {
  realm_id  = keycloak_realm.realm.id
  client_id = "k8s-client"

  name    = "K8s Client"
  enabled = true

  access_type               = "CONFIDENTIAL"
  service_accounts_enabled  = true
  client_authenticator_type = "federated-jwt"
  extra_config = {
    "jwt.credential.issuer" = keycloak_kubernetes_identity_provider.kubernetes.alias
    "jwt.credential.sub"    = "system:serviceaccount:<namespace>:<service-account-name>"
  }
}

# You need to create a new `Client Authentication` flow. In this example there is only one authenticator in it, but more can be configured if needed

resource "keycloak_authentication_flow" "client_authentication" {
  realm_id    = keycloak_realm.realm.id
  alias       = "clients-federated-jwt"
  provider_id = "client-flow"
}

resource "keycloak_authentication_execution" "federated_jwt" {
  realm_id          = keycloak_realm.realm.id
  parent_flow_alias = keycloak_authentication_flow.client_authentication.alias
  authenticator     = "federated-jwt"
  requirement       = "ALTERNATIVE"
}

resource "keycloak_authentication_bindings" "auth_bindings" {
  realm_id                   = keycloak_realm.realm.id
  client_authentication_flow = keycloak_authentication_flow.client_authentication.alias
}
```

### Kubernetes workload

In your Kubernetes workload, you need to mount a service account token with the right audience pointing to your Keycloak instance
```yaml
---
apiVersion: v1
kind: Pod
...
spec:
  serviceAccountName: <serviceaccount>
  ...
      volumeMounts:
      - mountPath: /var/run/secrets
        name: aud-token
  ...
  volumes:
  - name: aud-token
    projected:
      defaultMode: 420
      sources:
      - serviceAccountToken:
          audience: https://example.com:8443/realms/test <1>
          expirationSeconds: 600 <2>
          path: keycloak
---
```

1. Issuer URL of the Keycloak realm.
2. Maximum time allowed by Kubernetes is 3600 seconds

###  In the Pod, use curl to authenticate to Keycloak:

```bash
curl -k https://example.com:8443/realms/<realm>/protocol/openid-connect/token \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  --data-urlencode grant_type=client_credentials \
  --data-urlencode
  client_assertion_type=urn:ietf:params:oauth:client-assertion-type:jwt-bearer \
  --data-urlencode client_assertion=$(cat /var/run/secrets/keycloak)
```

### And the response should look something like:
```
{
  "access_token": "ey...bw",
  "expires_in": 600,
  ....
}
```

> **NOTICE:**
> Changing authentication flow bindings in your Realm settings can break existing clients' ability to authenticate, if not configured properly!


## Argument Reference

- `realm` - (Required) The name of the realm. This is unique across Keycloak.
- `alias` - (Required) The alias uniquely identifies an identity provider, and it is also used to build the redirect uri.
- `issuer` - (Required) The Kubernetes issuer URL of service account tokens. The URL <ISSUER>/.well-known/openid-configuration must be available to Keycloak.

## Import

Identity providers can be imported using the format `{{realm_id}}/{{idp_alias}}`, where `idp_alias` is the identity provider alias.

Example:

```bash
$ terraform import keycloak_kubernetes_identity_provider.realm_identity_provider my-realm/my-idp
```
