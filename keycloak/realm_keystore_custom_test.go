package keycloak

import (
	"reflect"
	"strings"
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

func TestValidateRealmKeystoreCustomConfig(t *testing.T) {
	properties := map[string]ComponentTypeProperty{
		"priority": {Name: "priority", Type: "String"},
		"hosts":    {Name: "hosts", Type: "MultivaluedString"},
	}

	cases := []struct {
		name       string
		properties map[string]ComponentTypeProperty
		config     map[string]string
		wantErr    string
	}{
		{
			name:       "declared keys are valid",
			properties: properties,
			config:     map[string]string{"priority": "100", "hosts": "a, b"},
		},
		{
			name:       "undeclared key",
			properties: properties,
			config:     map[string]string{"keySize": "2048"},
			wantErr:    `config key "keySize" is not declared by key provider provider, declared keys are: hosts, priority`,
		},
		{
			name:       "undeclared key of a provider without config keys",
			properties: map[string]ComponentTypeProperty{},
			config:     map[string]string{"priority": "100"},
			wantErr:    "which declares no config keys",
		},
		{
			name:       "empty value",
			properties: properties,
			config:     map[string]string{"priority": " "},
			wantErr:    `config key "priority" must not be empty`,
		},
		{
			name:       "trailing comma in a multi-valued key",
			properties: properties,
			config:     map[string]string{"hosts": "a,"},
			wantErr:    `config key "hosts" must not contain empty values`,
		},
		{
			name:       "empty element in a multi-valued key",
			properties: properties,
			config:     map[string]string{"hosts": "a, ,b"},
			wantErr:    `config key "hosts" must not contain empty values`,
		},
		{
			name:       "comma in a single-valued key",
			properties: map[string]ComponentTypeProperty{"label": {Name: "label", Type: "String"}},
			config:     map[string]string{"label": "a,"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateRealmKeystoreCustomConfig("provider", tc.config, tc.properties)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tc.wantErr, err)
			}
		})
	}
}
