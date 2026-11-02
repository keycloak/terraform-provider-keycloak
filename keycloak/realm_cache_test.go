package keycloak

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestRealmCacheNameIsOnlyAHint(t *testing.T) {
	var c realmCache
	c.remember(c.currentGeneration(), "uuid-1", "realm-a")

	if name, ok := c.nameHint("uuid-1"); !ok || name != "realm-a" {
		t.Errorf("hint = %q, %v", name, ok)
	}
	if _, ok := c.nameHint("other"); ok {
		t.Error("unknown uuid must not have a hint")
	}

	// a name belongs to one realm: when another realm is seen under it, the first one's hint is dropped
	c.remember(c.currentGeneration(), "uuid-2", "realm-a")
	if _, ok := c.nameHint("uuid-1"); ok {
		t.Error("the hint of a realm that lost its name must be dropped")
	}
	if name, ok := c.nameHint("uuid-2"); !ok || name != "realm-a" {
		t.Errorf("hint = %q, %v", name, ok)
	}
}

func TestRealmCacheNeverStoresEmptyIds(t *testing.T) {
	var c realmCache
	c.remember(c.currentGeneration(), "", "realm-a")

	if _, ok := c.nameHint(""); ok {
		t.Error("the empty id must not resolve to a realm")
	}
}

func TestRealmCacheEntriesExpire(t *testing.T) {
	var c realmCache
	c.remember(c.currentGeneration(), "uuid-1", "realm-a")
	c.markUnreadable(c.currentGeneration(), "realm-b")

	c.mu.Lock()
	old := time.Now().Add(-2 * realmCacheTTL)
	c.names["uuid-1"] = realmCacheEntry{name: "realm-a", at: old}
	c.unreadable["realm-b"] = old
	c.mu.Unlock()

	if _, ok := c.nameHint("uuid-1"); ok {
		t.Error("an expired hint must not be returned")
	}
	if c.isUnreadable("realm-b") {
		t.Error("an expired unreadable mark must not be returned")
	}
}

// A lookup that started before a rename must not bring the old names back once the rename cleared the cache.
func TestRealmCacheForgetBeatsSlowLookup(t *testing.T) {
	var c realmCache

	generation := c.currentGeneration() // a lookup starts here...
	c.forget()                          // ...a rename completes meanwhile...
	c.remember(generation, "uuid-1", "old-name")
	c.markUnreadable(generation, "old-name")

	if name, ok := c.nameHint("uuid-1"); ok {
		t.Errorf("stale hint %q was stored after the cache was cleared", name)
	}
	if c.isUnreadable("old-name") {
		t.Error("a stale mark was stored after the cache was cleared")
	}
}

// fakeRealms is a minimal Keycloak that serves realms by name and as a list.
type fakeRealms struct {
	// realms are listed in this order; the id may be empty to simulate an account that may not see it
	realms []fakeRealm
	// forbidden names answer 403 when asked for directly and are left out of the list
	forbidden map[string]bool
	// when set, the token endpoint fails
	tokenFails bool

	requests atomic.Int32
}

type fakeRealm struct {
	Realm string `json:"realm"`
	Id    string `json:"id,omitempty"`
}

func (f *fakeRealms) client(t *testing.T) *KeycloakClient {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/protocol/openid-connect/token"):
			if f.tokenFails {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			_, _ = w.Write([]byte(`{"access_token":"new","token_type":"Bearer"}`))
		case r.URL.Path == "/admin/realms":
			f.requests.Add(1)
			list := []fakeRealm{}
			for _, realm := range f.realms {
				if !f.forbidden[realm.Realm] {
					list = append(list, realm)
				}
			}
			_ = json.NewEncoder(w).Encode(list)
		case strings.HasPrefix(r.URL.Path, "/admin/realms/"):
			f.requests.Add(1)
			name := strings.TrimPrefix(r.URL.Path, "/admin/realms/")
			if f.forbidden[name] {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			for _, realm := range f.realms {
				if realm.Realm == name {
					_ = json.NewEncoder(w).Encode(realm)
					return
				}
			}
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)

	return &KeycloakClient{
		baseUrl: server.URL, authUrl: server.URL, realm: "master", httpClient: server.Client(), initialLogin: true,
		accessTokenProvided: !f.tokenFails,
		clientCredentials:   &ClientCredentials{AccessToken: "test", TokenType: "Bearer", GrantType: "client_credentials", ClientId: "c", ClientSecret: "s"},
	}
}

func TestGetRealmNameByUUIDConfirmsTheKnownName(t *testing.T) {
	fake := &fakeRealms{realms: []fakeRealm{{Realm: "a", Id: "u1"}}}
	client := fake.client(t)

	name, err := client.GetRealmNameByUUID(t.Context(), "u1", "a")
	if err != nil || name != "a" {
		t.Fatalf("name = %q, err = %v", name, err)
	}
	if n := fake.requests.Load(); n != 1 {
		t.Errorf("one request to confirm the name is enough, got %d", n)
	}
}

// After a rename the old name can belong to another realm. A name that Keycloak does not confirm is never used.
func TestGetRealmNameByUUIDDoesNotTrustAReusedName(t *testing.T) {
	fake := &fakeRealms{realms: []fakeRealm{{Realm: "a", Id: "u2"}, {Realm: "b", Id: "u1"}}}
	client := fake.client(t)

	name, err := client.GetRealmNameByUUID(t.Context(), "u1", "a")
	if err != nil || name != "b" {
		t.Fatalf("name = %q, err = %v; realm a is another realm now", name, err)
	}
}

// What this client remembers is a hint and has to be confirmed too: another client may have renamed the realm.
func TestGetRealmNameByUUIDDoesNotReturnARememberedNameWithoutAsking(t *testing.T) {
	fake := &fakeRealms{realms: []fakeRealm{{Realm: "a", Id: "u1"}}}
	client := fake.client(t)

	if name, err := client.GetRealmNameByUUID(t.Context(), "u1", ""); err != nil || name != "a" {
		t.Fatalf("name = %q, err = %v", name, err)
	}

	// someone else renames a to b and creates a new realm called a
	fake.realms = []fakeRealm{{Realm: "a", Id: "u2"}, {Realm: "b", Id: "u1"}}

	if name, err := client.GetRealmNameByUUID(t.Context(), "u1", ""); err != nil || name != "b" {
		t.Fatalf("name = %q, err = %v; the remembered name a belongs to another realm now", name, err)
	}
}

func TestGetRealmNameByUUIDKeepsTheAuthorizationErrorWhenTheListOmitsTheRealm(t *testing.T) {
	fake := &fakeRealms{
		realms:    []fakeRealm{{Realm: "a", Id: "u1"}, {Realm: "other", Id: "u9"}},
		forbidden: map[string]bool{"a": true},
	}
	client := fake.client(t)

	_, err := client.GetRealmNameByUUID(t.Context(), "u1", "a")
	if err == nil || !ErrorIs403(err) || ErrorIs404(err) {
		t.Fatalf("a realm that may not be read is not a realm that is gone: err = %v", err)
	}
}

func TestGetRealmNameByUUIDReportsHiddenIds(t *testing.T) {
	fake := &fakeRealms{realms: []fakeRealm{{Realm: "a"}}}
	client := fake.client(t)

	_, err := client.GetRealmNameByUUID(t.Context(), "u1", "a")
	if err == nil || !ErrorIs403(err) || !strings.Contains(err.Error(), "view-realm") {
		t.Fatalf("err = %v", err)
	}
}

func TestGetRealmNameByUUIDReportsARealmThatIsGone(t *testing.T) {
	fake := &fakeRealms{realms: []fakeRealm{{Realm: "other", Id: "u9"}}}
	client := fake.client(t)

	_, err := client.GetRealmNameByUUID(t.Context(), "u1", "a")
	if err == nil || !ErrorIs404(err) {
		t.Fatalf("err = %v", err)
	}
}

// A failed token request says nothing about the realm and must not be reported as a missing realm.
func TestGetRealmNameByUUIDReturnsAFailedTokenRefresh(t *testing.T) {
	fake := &fakeRealms{realms: []fakeRealm{{Realm: "other", Id: "u9"}}, tokenFails: true}
	client := fake.client(t)

	_, err := client.GetRealmNameByUUID(t.Context(), "u1", "")
	if err == nil || ErrorIs404(err) || ErrorIs403(err) {
		t.Fatalf("err = %v", err)
	}
}

func TestGetRealmUUIDDoesNotRememberAnId(t *testing.T) {
	fake := &fakeRealms{realms: []fakeRealm{{Realm: "a", Id: "u1"}}}
	client := fake.client(t)

	if id, err := client.GetRealmUUID(t.Context(), "a"); err != nil || id != "u1" {
		t.Fatalf("id = %q, err = %v", id, err)
	}

	// the name is reused by another realm
	fake.realms = []fakeRealm{{Realm: "a", Id: "u2"}}

	if id, err := client.GetRealmUUID(t.Context(), "a"); err != nil || id != "u2" {
		t.Fatalf("id = %q, err = %v; an id that was seen before must not be returned", id, err)
	}
}

func TestGetRealmUUIDRemembersThatAnIdIsHidden(t *testing.T) {
	fake := &fakeRealms{realms: []fakeRealm{{Realm: "a"}}}
	client := fake.client(t)

	if id, err := client.GetRealmUUID(t.Context(), "a"); err != nil || id != "" {
		t.Fatalf("id = %q, err = %v", id, err)
	}
	before := fake.requests.Load()

	if id, err := client.GetRealmUUID(t.Context(), "a"); err != nil || id != "" {
		t.Fatalf("id = %q, err = %v", id, err)
	}
	if after := fake.requests.Load(); after != before {
		t.Errorf("the lookup was repeated: %d requests more", after-before)
	}
}

func TestGetRealmUUIDRemembersThatARealmIsForbidden(t *testing.T) {
	fake := &fakeRealms{realms: []fakeRealm{{Realm: "a", Id: "u1"}}, forbidden: map[string]bool{"a": true}}
	client := fake.client(t)

	if _, err := client.GetRealmUUID(t.Context(), "a"); err == nil || !ErrorIs403(err) {
		t.Fatalf("err = %v", err)
	}
	before := fake.requests.Load()

	if id, err := client.GetRealmUUID(t.Context(), "a"); err != nil || id != "" {
		t.Fatalf("id = %q, err = %v", id, err)
	}
	if after := fake.requests.Load(); after != before {
		t.Errorf("the lookup was repeated: %d requests more", after-before)
	}
}
