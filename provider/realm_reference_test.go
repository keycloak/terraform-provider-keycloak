package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Every resource and data source that is scoped to a realm must accept realm_internal_id. Without this
// check a new resource that the wrapper does not recognise would silently miss the feature.
func TestRealmInternalIdIsOnEveryRealmScopedResource(t *testing.T) {
	p := KeycloakProvider(nil)

	check := func(kind, name string, r *schema.Resource, writable bool) {
		if name == "keycloak_realm" {
			if _, ok := r.Schema["realm_internal_id"]; ok {
				t.Errorf("%s %s is the realm itself and must not have realm_internal_id", kind, name)
			}
			return
		}

		realmField := "realm_id"
		if _, ok := r.Schema[realmField]; !ok {
			realmField = "realm"
		}
		realm, ok := r.Schema[realmField]
		if !ok {
			t.Errorf("%s %s has neither realm_id nor realm", kind, name)
			return
		}

		uuid, ok := r.Schema["realm_internal_id"]
		if !ok {
			t.Errorf("%s %s has no realm_internal_id", kind, name)
			return
		}

		if !realm.Optional || !realm.Computed || realm.Required {
			t.Errorf("%s %s: %s must be optional and computed", kind, name, realmField)
		}
		if !uuid.Optional || !uuid.Computed {
			t.Errorf("%s %s: realm_internal_id must be optional and computed", kind, name)
		}
		if len(realm.ExactlyOneOf) != 2 || len(uuid.ExactlyOneOf) != 2 {
			t.Errorf("%s %s: %s and realm_internal_id must be ExactlyOneOf each other", kind, name, realmField)
		}
		if writable && (!realm.ForceNew || !uuid.ForceNew) {
			t.Errorf("%s %s: %s and realm_internal_id must force a new resource, a resource cannot move to another realm", kind, name, realmField)
		}
	}

	for name, r := range p.ResourcesMap {
		check("resource", name, r, true)
	}
	for name, r := range p.DataSourcesMap {
		check("data source", name, r, false)
	}
}

func TestRealmRenameUsesDeclaredResourceIDFormat(t *testing.T) {
	for _, test := range []struct{ resource, original, want string }{
		{"keycloak_openid_client_service_account_role", "a/role", "a/role"},
		{"keycloak_openid_client_service_account_realm_role", "a/role", "a/role"},
		{"keycloak_realm_localization", "a/de", "b/de"},
		{"keycloak_generic_role_mapper", "a/client/c/scope-mappings/c/r", "b/client/c/scope-mappings/c/r"},
		{"keycloak_realm_events", "a", "b"},
	} {
		t.Run(test.resource, func(t *testing.T) {
			resource := &schema.Resource{}
			data := resource.Data(nil)
			data.SetId(test.original)
			renameRealmInId(data, "a", "b", resourceRealmIDFormats[test.resource])
			if data.Id() != test.want {
				t.Fatalf("got %q, want %q", data.Id(), test.want)
			}
		})
	}
}
