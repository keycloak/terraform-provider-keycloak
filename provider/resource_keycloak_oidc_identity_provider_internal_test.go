package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func oidcProviderIdDiff(t *testing.T, stateProviderId, configProviderId string) *terraform.ResourceAttrDiff {
	t.Helper()

	state := &terraform.InstanceState{
		ID:         "my-idp",
		Attributes: map[string]string{"provider_id": stateProviderId},
	}
	diff, err := resourceKeycloakOidcIdentityProvider().Diff(context.Background(), state, terraform.NewResourceConfigRaw(map[string]interface{}{
		"provider_id": configProviderId,
	}), nil)
	if err != nil {
		t.Fatal(err)
	}
	if diff == nil {
		return nil
	}
	return diff.Attributes["provider_id"]
}

// Keycloak ignores providerId on update, so a changed provider_id has to plan as a replacement.
func TestOidcIdentityProviderProviderIdChangeRequiresReplacement(t *testing.T) {
	attr := oidcProviderIdDiff(t, "github", "oidc")
	if attr == nil {
		t.Fatal("changing provider_id should produce a diff")
	}
	if !attr.RequiresNew {
		t.Fatal("changing provider_id should require replacement")
	}
}

func TestOidcIdentityProviderUnchangedProviderIdProducesNoDiff(t *testing.T) {
	// The SDK may still report the attribute, but it must carry no change and no replacement flags.
	attr := oidcProviderIdDiff(t, "oidc", "oidc")
	if attr != nil && (attr.Old != attr.New || attr.RequiresNew || attr.NewComputed || attr.NewRemoved) {
		t.Fatalf("an unchanged provider_id should not produce a change, got %+v", attr)
	}
}
