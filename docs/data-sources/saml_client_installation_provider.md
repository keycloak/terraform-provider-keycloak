---
page_title: "keycloak_saml_client_installation_provider Data Source"
---

# keycloak\_saml\_client\_installation\_provider Data Source

This data source can be used to retrieve Installation Provider of a SAML Client.

## Example Usage

In the example below, we extract the SAML metadata IDPSSODescriptor to pass it to the AWS IAM SAML Provider.

```hcl
resource "keycloak_realm" "realm" {
    realm   = "my-realm"
    enabled = true
}

resource "keycloak_saml_client" "saml_client" {
    realm_id                = keycloak_realm.realm.id
    client_id               = "test-saml-client"
    name                    = "test-saml-client"

    sign_documents          = false
    sign_assertions         = true
    include_authn_statement = true

    signing_certificate = file("saml-cert.pem")
    signing_private_key = file("saml-key.pem")
}

data "keycloak_saml_client_installation_provider" "saml_idp_descriptor" {
  realm_id    = keycloak_realm.realm.id
  client_id   = keycloak_saml_client.saml_client.id
  provider_id = "saml-idp-descriptor"
}


resource "aws_iam_saml_provider" "default" {
  name                   = "myprovider"
  saml_metadata_document = data.keycloak_saml_client_installation_provider.saml_idp_descriptor.value
}
```

## Argument Reference

- `realm_id` - (Required) The realm that the SAML client exists within.
- `client_id` - (Required) The ID of the SAML client. The `id` attribute of a `keycloak_client` resource should be used here.
- `provider_id` - (Required) The ID of the SAML installation provider. Could be one of `saml-idp-descriptor`, `keycloak-saml`, `saml-sp-descriptor`, `keycloak-saml-subsystem`, `mod-auth-mellon`, etc.

## Attributes Reference

- `id` - (Computed) The hash of the value.
- `value` - (Computed) The returned document needed for SAML installation.
- `zip_files` - (Computed) A map of files if the returned document is a zip file. (ex: provider `mod-auth-mellon`)
