---
page_title: "keycloak_saml_client Resource"
---

# keycloak\_saml\_client Resource

Allows for creating and managing Keycloak clients that use the SAML protocol.

Clients are entities that can use Keycloak for user authentication. Typically, clients are applications that redirect users
to Keycloak for authentication in order to take advantage of Keycloak's user sessions for SSO.

## Example Usage

```hcl
resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_saml_client" "saml_client" {
  realm_id  = keycloak_realm.realm.id
  client_id = "saml-client"
  name      = "saml-client"

  sign_documents          = false
  sign_assertions         = true
  include_authn_statement = true

  signing_certificate = file("saml-cert.pem")
  signing_private_key = file("saml-key.pem")
}
```

## Argument Reference

- `realm_id` - (Required) The realm this client is attached to.
- `client_id` - (Required) The unique ID of this client, referenced in the URI during authentication and in issued tokens.
- `name` - (Optional) The display name of this client in the GUI.
- `enabled` - (Optional) When false, this client will not be able to initiate a login or obtain access tokens. Defaults to `true`.
- `description` - (Optional) The description of this client in the GUI.
- `login_theme` - (Optional) The login theme of this client.
- `include_authn_statement` - (Optional) When `true`, an `AuthnStatement` will be included in the SAML response. Defaults to `true`.
- `sign_documents` - (Optional) When `true`, the SAML document will be signed by Keycloak using the realm's private key. Defaults to `true`.
- `sign_assertions` - (Optional) When `true`, the SAML assertions will be signed by Keycloak using the realm's private key, and embedded within the SAML XML Auth response. Defaults to `false`.
- `encrypt_assertions` - (Optional) When `true`, the SAML assertions will be encrypted by Keycloak using the client's public key. Defaults to `false`.
- `client_signature_required` - (Optional) When `true`, Keycloak will expect that documents originating from a client will be signed using the certificate and/or key configured via `signing_certificate` and `signing_private_key`. Defaults to `true`.
- `force_post_binding` - (Optional) When `true`, Keycloak will always respond to an authentication request via the SAML POST Binding. Defaults to `true`.
 `consent_required` - (Optional) When `true`, users have to consent to client access. Defaults to `false`.
- `always_display_in_console` - (Optional) When `True` Keycloak will always list this client in the Account UI, even if the user does not have an active session. Defaults to `false`.
- `front_channel_logout` - (Optional) When `true`, this client will require a browser redirect in order to perform a logout. Defaults to `true`.
- `name_id_format` - (Optional) Sets the Name ID format for the subject.
- `force_name_id_format` - (Optional) Ignore requested NameID subject format and use the one defined in `name_id_format` instead. Defaults to `false`.
- `signature_algorithm` - (Optional) The signature algorithm used to sign documents. Should be one of "RSA_SHA1", "RSA_SHA256", "RSA_SHA256_MGF1, "RSA_SHA512", "RSA_SHA512_MGF1" or "DSA_SHA1".
- `signature_key_name` - (Optional) The value of the `KeyName` element within the signed SAML document. Should be one of "NONE", "KEY_ID", or "CERT_SUBJECT". Defaults to "KEY_ID".
- `canonicalization_method` - (Optional) The Canonicalization Method for XML signatures. Should be one of "EXCLUSIVE", "EXCLUSIVE_WITH_COMMENTS", "INCLUSIVE", or "INCLUSIVE_WITH_COMMENTS". Defaults to "EXCLUSIVE".
- `root_url` - (Optional) When specified, this value is prepended to all relative URLs.
- `valid_redirect_uris` - (Optional) When specified, Keycloak will use this list to validate given Assertion Consumer URLs specified in the authentication request.
- `base_url` - (Optional) When specified, this URL will be used whenever Keycloak needs to link to this client.
- `master_saml_processing_url` - (Optional) When specified, this URL will be used for all SAML requests.
- `encryption_certificate` - (Optional) If assertions for the client are encrypted, this certificate will be used for encryption.
- `signing_certificate` - (Optional) If documents or assertions from the client are signed, this certificate will be used to verify the signature.
- `signing_private_key` - (Optional) If documents or assertions from the client are signed, this private key will be used to verify the signature.
- `idp_initiated_sso_url_name` - (Optional) URL fragment name to reference client when you want to do IDP Initiated SSO.
- `idp_initiated_sso_relay_state` - (Optional) Relay state you want to send with SAML request when you want to do IDP Initiated SSO.
- `assertion_consumer_post_url` - (Optional) SAML POST Binding URL for the client's assertion consumer service (login responses).
- `assertion_consumer_redirect_url` - (Optional) SAML Redirect Binding URL for the client's assertion consumer service (login responses).
- `logout_service_post_binding_url` - (Optional) SAML POST Binding URL for the client's single logout service.
- `logout_service_redirect_binding_url` - (Optional) SAML Redirect Binding URL for the client's single logout service.
- `full_scope_allowed` - (Optional) - Allow to include all roles mappings in the access token
- `authentication_flow_binding_overrides` - (Optional) Override realm authentication flow bindings
    - `browser_id` - (Optional) Browser flow id, (flow needs to exist)
    - `direct_grant_id` - (Optional) Direct grant flow id (flow needs to exist)
- `extra_config` - (Optional) A map of key/value pairs to add extra configuration attributes to this client. This can be used for custom attributes, or to add configuration attributes that is not yet supported by this Terraform provider. Use this attribute at your own risk, as s may conflict with top-level configuration attributes in future provider updates.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

- `encryption_certificate_sha1` - (Computed) The sha1sum fingerprint of the encryption certificate. If the encryption certificate is not in correct base64 format, this will be left empty.
- `signing_certificate_sha1` - (Computed) The sha1sum fingerprint of the signing certificate. If the signing certificate is not in correct base64 format, this will be left empty.
- `signing_private_key_sha1` - (Computed) The sha1sum fingerprint of the signing private key. If the signing private key is not in correct base64 format, this will be left empty.

## Import

Clients can be imported using the format `{{realm_id}}/{{client_keycloak_id}}`, where `client_keycloak_id` is the unique ID that Keycloak
assigns to the client upon creation. This value can be found in the URI when editing this client in the GUI, and is typically a GUID.

Example:

```bash
$ terraform import keycloak_saml_client.saml_client my-realm/dcbc4c73-e478-4928-ae2e-d5e420223352
```
