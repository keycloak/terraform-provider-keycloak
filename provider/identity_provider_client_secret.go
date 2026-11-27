package provider

import (
	"errors"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// identityProviderClientSecretSchema returns the client_secret, client_secret_wo and
// client_secret_wo_version arguments shared by the social identity provider resources.
// It returns new schemas on every call, because mergeSchemas keeps the pointers it is given.
func identityProviderClientSecretSchema(clientSecretDescription string) map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"client_secret": {
			Type:          schema.TypeString,
			Optional:      true,
			Sensitive:     true,
			ConflictsWith: []string{"client_secret_wo", "client_secret_wo_version"},
			AtLeastOneOf:  []string{"client_secret", "client_secret_wo"},
			Description:   clientSecretDescription,
		},
		"client_secret_wo": {
			Type:          schema.TypeString,
			Optional:      true,
			Sensitive:     true,
			WriteOnly:     true,
			ConflictsWith: []string{"client_secret"},
			RequiredWith:  []string{"client_secret_wo_version"},
			AtLeastOneOf:  []string{"client_secret", "client_secret_wo"},
			Description:   "Client Secret as write-only argument",
		},
		"client_secret_wo_version": {
			Type:          schema.TypeString,
			Optional:      true,
			ConflictsWith: []string{"client_secret"},
			RequiredWith:  []string{"client_secret_wo"},
			Description:   "Version of the Client secret write-only argument",
		},
	}
}

// getIdentityProviderClientSecretFromData returns the client secret to send to Keycloak.
// With client_secret_wo, the secret is only read from the configuration on create and when
// client_secret_wo_version changes. Otherwise it is empty, and the identity provider update
// keeps the secret already stored in Keycloak.
func getIdentityProviderClientSecretFromData(data *schema.ResourceData) (string, error) {
	if data.Get("client_secret_wo_version").(string) == "" {
		return data.Get("client_secret").(string), nil
	}

	if data.Id() != "" && !data.HasChange("client_secret_wo_version") {
		return "", nil
	}

	clientSecretWriteOnly, diags := data.GetRawConfigAt(cty.GetAttrPath("client_secret_wo"))
	if diags.HasError() {
		return "", errors.New("error reading 'client_secret_wo' argument")
	}
	if !clientSecretWriteOnly.IsKnown() || clientSecretWriteOnly.IsNull() || !clientSecretWriteOnly.Type().Equals(cty.String) {
		return "", errors.New("'client_secret_wo' must be a known, non-null string")
	}

	return clientSecretWriteOnly.AsString(), nil
}
