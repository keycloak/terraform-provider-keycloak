package provider

import (
	"reflect"
	"testing"

	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestMergeRealmKeystoreCustomConfig(t *testing.T) {
	properties := map[string]keycloak.ComponentTypeProperty{
		"priority":   {Name: "priority", Type: "String"},
		"privateKey": {Name: "privateKey", Type: "File", Secret: true},
		"hosts":      {Name: "hosts", Type: "MultivaluedString"},
	}

	cases := []struct {
		name   string
		prior  map[string]string
		remote map[string]string
		want   map[string]string
	}{
		{
			name:   "keys added by keycloak are ignored",
			prior:  map[string]string{"priority": "100"},
			remote: map[string]string{"priority": "100", "certificate": "generated"},
			want:   map[string]string{"priority": "100"},
		},
		{
			name:   "changed value is taken from keycloak",
			prior:  map[string]string{"priority": "100"},
			remote: map[string]string{"priority": "200"},
			want:   map[string]string{"priority": "200"},
		},
		{
			name:   "removed key is dropped",
			prior:  map[string]string{"priority": "100", "hosts": "a"},
			remote: map[string]string{"priority": "100"},
			want:   map[string]string{"priority": "100"},
		},
		{
			name:   "masked secret keeps the prior value",
			prior:  map[string]string{"privateKey": "secret"},
			remote: map[string]string{"privateKey": keycloak.RealmKeystoreSecretValue},
			want:   map[string]string{"privateKey": "secret"},
		},
		{
			name:   "vault reference of a secret is taken from keycloak",
			prior:  map[string]string{"privateKey": "secret"},
			remote: map[string]string{"privateKey": "${vault.key}"},
			want:   map[string]string{"privateKey": "${vault.key}"},
		},
		{
			name:   "masked value of a non-secret key is taken from keycloak",
			prior:  map[string]string{"priority": "100"},
			remote: map[string]string{"priority": keycloak.RealmKeystoreSecretValue},
			want:   map[string]string{"priority": keycloak.RealmKeystoreSecretValue},
		},
		{
			name:   "reordered multi-valued key keeps the prior value",
			prior:  map[string]string{"hosts": "a,b,c"},
			remote: map[string]string{"hosts": "c,a,b"},
			want:   map[string]string{"hosts": "a,b,c"},
		},
		{
			name:   "changed multi-valued key is taken from keycloak",
			prior:  map[string]string{"hosts": "a,b"},
			remote: map[string]string{"hosts": "a,c"},
			want:   map[string]string{"hosts": "a,c"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeRealmKeystoreCustomConfig(tc.prior, tc.remote, properties)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}
