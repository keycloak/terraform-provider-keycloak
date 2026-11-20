# custom-authz-policy-example

A minimal Keycloak provider JAR that deploys a JavaScript authorization policy via a
`META-INF/keycloak-scripts.json` descriptor. It exists so the acceptance tests for the
`keycloak_generic_client_authorization_policy` resource have a real JAR-deployed policy
provider to reference.

Modern Keycloak versions no longer allow uploading the JavaScript code of an authorization
policy through the API, so a deployed script is the only way to create such a policy.

## Layout

- `META-INF/keycloak-scripts.json` — descriptor declaring the deployed policy.
- `always-granting-policy.js` — the policy code.

After deployment the policy is available as a policy provider of type
`script-always-granting-policy.js` (the `script-` prefix plus the descriptor `fileName`).

## Building

A scripts provider is just a JAR (a ZIP). Build it with:

```sh
make authz-policy-example
```

This produces `build/keycloak-authz-policy-example.jar`, which the local Keycloak
(`docker-compose.yml`) and CI mount into `/opt/keycloak/providers/`. The `scripts`
feature must be enabled (it is part of `KC_FEATURES=preview`).
