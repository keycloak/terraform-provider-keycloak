package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Every resource and data source that is scoped to a realm must accept realm_uuid. Without this
// check a new resource that the wrapper does not recognise would silently miss the feature.
func TestRealmUuidIsOnEveryRealmScopedResource(t *testing.T) {
	p := KeycloakProvider(nil)

	check := func(kind, name string, r *schema.Resource, writable bool) {
		if name == "keycloak_realm" {
			if _, ok := r.Schema["realm_uuid"]; ok {
				t.Errorf("%s %s is the realm itself and must not have realm_uuid", kind, name)
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

		uuid, ok := r.Schema["realm_uuid"]
		if !ok {
			t.Errorf("%s %s has no realm_uuid", kind, name)
			return
		}

		if !realm.Optional || !realm.Computed || realm.Required {
			t.Errorf("%s %s: %s must be optional and computed", kind, name, realmField)
		}
		if !uuid.Optional || !uuid.Computed {
			t.Errorf("%s %s: realm_uuid must be optional and computed", kind, name)
		}
		if len(realm.ExactlyOneOf) != 2 || len(uuid.ExactlyOneOf) != 2 {
			t.Errorf("%s %s: %s and realm_uuid must be ExactlyOneOf each other", kind, name, realmField)
		}
		if writable && (!realm.ForceNew || !uuid.ForceNew) {
			t.Errorf("%s %s: %s and realm_uuid must force a new resource, a resource cannot move to another realm", kind, name, realmField)
		}
	}

	for name, r := range p.ResourcesMap {
		check("resource", name, r, true)
	}
	for name, r := range p.DataSourcesMap {
		check("data source", name, r, false)
	}
}
