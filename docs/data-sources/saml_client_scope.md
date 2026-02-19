---
page_title: "keycloak_saml_client_scope Data Source"
---

# keycloak_saml_client_scope Data Source

This data source can be used to fetch properties of a Keycloak SAML client scope for usage with other resources.

## Example Usage

```hcl
data "keycloak_saml_client_scope" "mysamlscope" {
  realm_id = "my-realm"
  name     = "mysamlscope"
}

resource keycloak_saml_client "saml_client" {
	realm_id = "my-realm"
	client_id = "saml-client"
}

# use the data source
resource keycloak_saml_client_default_scopes "default" {
	realm_id = "my-realm"
	client_id = keycloak_saml_client.saml_client.id
	default_scopes = [
		data.keycloak_saml_client_scope.mysamlscope.name
	]
}
```

## Argument Reference

- `realm_id` - (Required) The realm id.
- `name` - (Required) The name of the client scope.

## Attributes Reference

See the docs for the `keycloak_saml_client_scope` resource for details on the exported attributes.
