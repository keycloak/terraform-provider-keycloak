package keycloak

import (
	"context"
	"net/http"
	"path"
	"reflect"
	"testing"
)

func TestUnlinkOrganizationIdentityProvidersReturnsRemaining(t *testing.T) {
	links := map[string]bool{"first": true, "second": true, "third": true}
	refuseSecond := true
	client := legacyCompatClient(t, "26.8.0", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			return
		}
		alias := path.Base(r.URL.Path)
		if alias == "second" && refuseSecond {
			refuseSecond = false
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if !links[alias] {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		delete(links, alias)
		w.WriteHeader(http.StatusNoContent)
	})
	organization := &Organization{Id: "org", Realm: "realm"}

	remaining, err := client.UnlinkOrganizationIdentityProviders(context.Background(), organization, []string{"first", "second", "third", "gone"})
	if err == nil {
		t.Fatal("expected the simulated failure")
	}
	if want := []string{"second", "third", "gone"}; !reflect.DeepEqual(remaining, want) {
		t.Fatalf("remaining: got %v, want %v", remaining, want)
	}

	remaining, err = client.UnlinkOrganizationIdentityProviders(context.Background(), organization, remaining)
	if err != nil {
		t.Fatal(err)
	}
	if len(remaining) != 0 || len(links) != 0 {
		t.Fatalf("retry left links: returned %v, server has %v", remaining, links)
	}
}
