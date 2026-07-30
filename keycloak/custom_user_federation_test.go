package keycloak

import (
	"reflect"
	"testing"
)

func TestConvertCustomUserFederationPreservesMultivaluedConfig(t *testing.T) {
	values := []string{"a=1", "b=2", "c=3"}

	custom := &CustomUserFederation{
		Id:      "id",
		Name:    "name",
		RealmId: "realm",
		Config: map[string][]string{
			"user-extension-mappings": values,
		},
	}

	component := convertFromCustomUserFederationToComponent(custom)
	if got := component.Config["user-extension-mappings"]; !reflect.DeepEqual(got, values) {
		t.Fatalf("write dropped config values: expected %v, got %v", values, got)
	}

	roundTripped, err := convertFromComponentToCustomUserFederation(component, "realm")
	if err != nil {
		t.Fatal(err)
	}
	if got := roundTripped.Config["user-extension-mappings"]; !reflect.DeepEqual(got, values) {
		t.Fatalf("read dropped config values: expected %v, got %v", values, got)
	}
}
