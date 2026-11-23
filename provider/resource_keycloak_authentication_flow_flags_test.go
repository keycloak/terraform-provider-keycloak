package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAuthenticationFlowFlagObservations(t *testing.T) {
	for _, builtIn := range []bool{true, false} {
		for _, topLevel := range []bool{true, false} {
			data := schema.TestResourceDataRaw(t, resourceKeycloakAuthenticationFlow().Schema, nil)
			mapFromAuthenticationFlowToData(data, &keycloak.AuthenticationFlow{
				Id: "flow", RealmId: "fixture", Alias: "form", ProviderId: "form-flow", BuiltIn: builtIn, TopLevel: topLevel,
			})
			if data.Get("built_in") != builtIn || data.Get("top_level") != topLevel || data.Get("provider_id") != "form-flow" {
				t.Fatalf("incorrect flag observations: %#v", data.State())
			}
		}
	}
}

func TestAuthenticationFlowRejectTopLevelFormCreation(t *testing.T) {
	data := schema.TestResourceDataRaw(t, resourceKeycloakAuthenticationFlow().Schema, map[string]interface{}{
		"realm_id": "fixture", "alias": "form", "provider_id": "form-flow",
	})
	// A nil client makes any accidental API operation fail the test.
	diagnostics := resourceKeycloakAuthenticationFlowCreate(context.Background(), data, (*keycloak.KeycloakClient)(nil))
	if !diagnostics.HasError() || data.Id() != "" {
		t.Fatalf("unsupported creation was not rejected: %v", diagnostics)
	}
}
