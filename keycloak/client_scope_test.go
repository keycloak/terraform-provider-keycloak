package keycloak

import (
	"encoding/json"
	"fmt"
	"testing"
)

// Keycloak treats a boolean client scope attribute that is not stored as true (ClientScopeModel),
// and returns "attributes": {} for a client scope without attributes.
func TestOpenidClientScopeAttributes_missingBooleansAreTrue(t *testing.T) {
	var clientScope OpenidClientScope
	if err := json.Unmarshal([]byte(`{"name":"offline_access","attributes":{"consent.screen.text":"text"}}`), &clientScope); err != nil {
		t.Fatal(err)
	}

	attributes := clientScope.Attributes
	if !attributes.DisplayOnConsentScreen || !attributes.IncludeInTokenScope || !attributes.IncludeInOpenidProviderMetadata {
		t.Fatalf("expected missing boolean attributes to be true, got %+v", attributes)
	}
}

func TestOpenidClientScopeAttributes_storedBooleansAreRead(t *testing.T) {
	for _, stored := range []string{"false", "", "no", "fAlSe"} {
		var clientScope OpenidClientScope
		data := fmt.Sprintf(`{"name":"phone","attributes":{"display.on.consent.screen":%[1]q,"include.in.token.scope":%[1]q,"include.in.openid.provider.metadata":%[1]q}}`, stored)
		if err := json.Unmarshal([]byte(data), &clientScope); err != nil {
			t.Fatal(err)
		}

		attributes := clientScope.Attributes
		if attributes.DisplayOnConsentScreen || attributes.IncludeInTokenScope || attributes.IncludeInOpenidProviderMetadata {
			t.Fatalf("expected boolean attributes stored as %q to be false, got %+v", stored, attributes)
		}
	}
}

func TestSamlClientScopeAttributes_missingDisplayOnConsentScreenIsTrue(t *testing.T) {
	var clientScope SamlClientScope
	if err := json.Unmarshal([]byte(`{"name":"role_list","attributes":{}}`), &clientScope); err != nil {
		t.Fatal(err)
	}

	if !clientScope.Attributes.DisplayOnConsentScreen {
		t.Fatalf("expected a missing display.on.consent.screen to be true, got %+v", clientScope.Attributes)
	}
}
