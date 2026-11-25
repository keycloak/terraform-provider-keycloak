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

func TestLinkOrganizationIdentityProvidersRemovesNewLinkWhenSettingsFail(t *testing.T) {
	tests := []struct {
		name string
		// linked: the link is there before the call.
		linked bool
		// conflict: another client makes the link after the read, so the POST gives 409.
		conflict   bool
		wantLinked bool
	}{
		{name: "new link", wantLinked: false},
		{name: "link that was there", linked: true, wantLinked: true},
		{name: "link that another client made", conflict: true, wantLinked: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			linked := test.linked
			client := legacyCompatClient(t, "26.8.0", func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodGet:
					if !linked {
						http.Error(w, "not found", http.StatusNotFound)
						return
					}
					// Keycloak makes a link with these settings.
					w.Header().Set("Content-Type", "application/json")
					w.Write([]byte(`{"alias":"first","organizationLinks":[{"organizationId":"org","autoMembership":true,"membershipType":"UNMANAGED"}]}`))
				case http.MethodPost:
					linked = true
					if test.conflict {
						http.Error(w, "conflict", http.StatusConflict)
						return
					}
					w.WriteHeader(http.StatusNoContent)
				case http.MethodPut:
					http.Error(w, "refused", http.StatusBadRequest)
				case http.MethodDelete:
					linked = false
					w.WriteHeader(http.StatusNoContent)
				default:
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
				}
			})
			organization := &Organization{Id: "org", Realm: "realm"}
			identityProviders := []OrganizationIdentityProvider{{Alias: "first", AutoMembership: false, MembershipType: "UNMANAGED"}}

			if err := client.LinkOrganizationIdentityProviders(context.Background(), organization, identityProviders); err == nil {
				t.Fatal("expected the simulated failure")
			}
			// A link that this call made must not stay with auto membership, which the configuration
			// turned off. A link that this call did not make stays.
			if linked != test.wantLinked {
				t.Fatalf("linked after the failure: got %t, want %t", linked, test.wantLinked)
			}
		})
	}
}
