package keycloak

import (
	"reflect"
	"testing"
)

func TestConvertFromComponentToRealmKeystoreCustomKeepsAllValues(t *testing.T) {
	c := &component{
		Id:         "id",
		ProviderId: "provider",
		Config: map[string][]string{
			"single": {"a"},
			"multi":  {"a", "b"},
			"empty":  {},
		},
	}

	got := convertFromComponentToRealmKeystoreCustom(c, "realm")

	want := map[string]string{
		"single": "a",
		"multi":  "a,b",
	}
	if !reflect.DeepEqual(got.Config, want) {
		t.Fatalf("expected keystore config %v, got %v", want, got.Config)
	}
}

func TestConvertRealmKeystoreCustomRoundTrip(t *testing.T) {
	multiValuedKeys := map[string]bool{"hosts": true}

	keystore := &RealmKeystoreCustom{
		Id:         "id",
		Name:       "name",
		RealmId:    "realm",
		ProviderId: "provider",
		Config: map[string]string{
			"priority": "100",
			"hosts":    "a, b,c",
			"label":    "x,y",
		},
	}

	c := convertFromRealmKeystoreCustomToComponent(keystore, multiValuedKeys, []string{"removed", "priority"})

	if c.ParentId != "" {
		t.Fatalf("expected parent id to be left to keycloak, got %q", c.ParentId)
	}
	if c.ProviderType != "org.keycloak.keys.KeyProvider" {
		t.Fatalf("unexpected provider type %q", c.ProviderType)
	}

	wantConfig := map[string][]string{
		"priority": {"100"},
		"hosts":    {"a", "b", "c"},
		"label":    {"x,y"},
		"removed":  {},
	}
	if !reflect.DeepEqual(c.Config, wantConfig) {
		t.Fatalf("expected component config %v, got %v", wantConfig, c.Config)
	}

	got := convertFromComponentToRealmKeystoreCustom(c, "realm")

	wantKeystoreConfig := map[string]string{
		"priority": "100",
		"hosts":    "a,b,c",
		"label":    "x,y",
	}
	if !reflect.DeepEqual(got.Config, wantKeystoreConfig) {
		t.Fatalf("expected keystore config %v, got %v", wantKeystoreConfig, got.Config)
	}
}
