package keycloak

import (
	"encoding/base64"
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
	// when set, the account does not have the admin role of the master realm, so its realm list may be incomplete
	notAdmin bool

	requests atomic.Int32
}

// testAccessToken is an unsigned token with the given realm roles, which is all that is read from it.
func testAccessToken(roles ...string) string {
	encode := func(value interface{}) string {
		encoded, _ := json.Marshal(value)

		return base64.RawURLEncoding.EncodeToString(encoded)
	}

	return encode(map[string]string{"alg": "HS256", "typ": "JWT"}) + "." +
		encode(map[string]interface{}{"realm_access": map[string][]string{"roles": roles}}) + "." +
		base64.RawURLEncoding.EncodeToString([]byte("signature"))
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

	accessToken := testAccessToken("admin")
	if f.notAdmin {
		accessToken = testAccessToken("view-realm")
	}

	return &KeycloakClient{
		baseUrl: server.URL, authUrl: server.URL, realm: "master", httpClient: server.Client(), initialLogin: true,
		accessTokenProvided: !f.tokenFails,
		clientCredentials:   &ClientCredentials{AccessToken: accessToken, TokenType: "Bearer", GrantType: "client_credentials", ClientId: "c", ClientSecret: "s"},
	}
}

func TestGetRealmNameByInternalIdConfirmsTheKnownName(t *testing.T) {
	fake := &fakeRealms{realms: []fakeRealm{{Realm: "a", Id: "u1"}}}
	client := fake.client(t)

	name, err := client.GetRealmNameByInternalId(t.Context(), "u1", "a")
	if err != nil || name != "a" {
		t.Fatalf("name = %q, err = %v", name, err)
	}
	if n := fake.requests.Load(); n != 1 {
		t.Errorf("one request to confirm the name is enough, got %d", n)
	}
}

// After a rename the old name can belong to another realm. A name that Keycloak does not confirm is never used.
func TestGetRealmNameByInternalIdDoesNotTrustAReusedName(t *testing.T) {
	fake := &fakeRealms{realms: []fakeRealm{{Realm: "a", Id: "u2"}, {Realm: "b", Id: "u1"}}}
	client := fake.client(t)

	name, err := client.GetRealmNameByInternalId(t.Context(), "u1", "a")
	if err != nil || name != "b" {
		t.Fatalf("name = %q, err = %v; realm a is another realm now", name, err)
	}
}

// What this client remembers is a hint and has to be confirmed too: another client may have renamed the realm.
func TestGetRealmNameByInternalIdDoesNotReturnARememberedNameWithoutAsking(t *testing.T) {
	fake := &fakeRealms{realms: []fakeRealm{{Realm: "a", Id: "u1"}}}
	client := fake.client(t)

	if name, err := client.GetRealmNameByInternalId(t.Context(), "u1", ""); err != nil || name != "a" {
		t.Fatalf("name = %q, err = %v", name, err)
	}

	// someone else renames a to b and creates a new realm called a
	fake.realms = []fakeRealm{{Realm: "a", Id: "u2"}, {Realm: "b", Id: "u1"}}

	if name, err := client.GetRealmNameByInternalId(t.Context(), "u1", ""); err != nil || name != "b" {
		t.Fatalf("name = %q, err = %v; the remembered name a belongs to another realm now", name, err)
	}
}

func TestGetRealmNameByInternalIdKeepsTheAuthorizationErrorWhenTheListOmitsTheRealm(t *testing.T) {
	fake := &fakeRealms{
		realms:    []fakeRealm{{Realm: "a", Id: "u1"}, {Realm: "other", Id: "u9"}},
		forbidden: map[string]bool{"a": true},
	}
	client := fake.client(t)

	_, err := client.GetRealmNameByInternalId(t.Context(), "u1", "a")
	if err == nil || !ErrorIs403(err) || ErrorIs404(err) {
		t.Fatalf("a realm that may not be read is not a realm that is gone: err = %v", err)
	}
}

func TestGetRealmNameByInternalIdReportsHiddenIds(t *testing.T) {
	fake := &fakeRealms{realms: []fakeRealm{{Realm: "a"}}}
	client := fake.client(t)

	_, err := client.GetRealmNameByInternalId(t.Context(), "u1", "a")
	if err == nil || !ErrorIs403(err) || !strings.Contains(err.Error(), "view-realm") {
		t.Fatalf("err = %v", err)
	}
}

func TestGetRealmNameByInternalIdReportsARealmThatIsGone(t *testing.T) {
	fake := &fakeRealms{realms: []fakeRealm{{Realm: "other", Id: "u9"}}}
	client := fake.client(t)

	_, err := client.GetRealmNameByInternalId(t.Context(), "u1", "a")
	if err == nil || !ErrorIs404(err) {
		t.Fatalf("err = %v", err)
	}
}

// A failed token request says nothing about the realm and must not be reported as a missing realm.
func TestGetRealmNameByInternalIdReturnsAFailedTokenRefresh(t *testing.T) {
	fake := &fakeRealms{realms: []fakeRealm{{Realm: "other", Id: "u9"}}, tokenFails: true}
	client := fake.client(t)

	_, err := client.GetRealmNameByInternalId(t.Context(), "u1", "")
	if err == nil || ErrorIs404(err) || ErrorIs403(err) {
		t.Fatalf("err = %v", err)
	}
}

func TestGetRealmInternalIdDoesNotRememberAnId(t *testing.T) {
	fake := &fakeRealms{realms: []fakeRealm{{Realm: "a", Id: "u1"}}}
	client := fake.client(t)

	if id, err := client.GetRealmInternalId(t.Context(), "a"); err != nil || id != "u1" {
		t.Fatalf("id = %q, err = %v", id, err)
	}

	// the name is reused by another realm
	fake.realms = []fakeRealm{{Realm: "a", Id: "u2"}}

	if id, err := client.GetRealmInternalId(t.Context(), "a"); err != nil || id != "u2" {
		t.Fatalf("id = %q, err = %v; an id that was seen before must not be returned", id, err)
	}
}

func TestGetRealmInternalIdRemembersThatAnIdIsHidden(t *testing.T) {
	fake := &fakeRealms{realms: []fakeRealm{{Realm: "a"}}}
	client := fake.client(t)

	if id, err := client.GetRealmInternalId(t.Context(), "a"); err != nil || id != "" {
		t.Fatalf("id = %q, err = %v", id, err)
	}
	before := fake.requests.Load()

	if id, err := client.GetRealmInternalId(t.Context(), "a"); err != nil || id != "" {
		t.Fatalf("id = %q, err = %v", id, err)
	}
	if after := fake.requests.Load(); after != before {
		t.Errorf("the lookup was repeated: %d requests more", after-before)
	}
}

func TestGetRealmInternalIdRemembersThatARealmIsForbidden(t *testing.T) {
	fake := &fakeRealms{realms: []fakeRealm{{Realm: "a", Id: "u1"}}, forbidden: map[string]bool{"a": true}}
	client := fake.client(t)

	if _, err := client.GetRealmInternalId(t.Context(), "a"); err == nil || !ErrorIs403(err) {
		t.Fatalf("err = %v", err)
	}
	before := fake.requests.Load()

	if id, err := client.GetRealmInternalId(t.Context(), "a"); err != nil || id != "" {
		t.Fatalf("id = %q, err = %v", id, err)
	}
	if after := fake.requests.Load(); after != before {
		t.Errorf("the lookup was repeated: %d requests more", after-before)
	}
}

// A token endpoint that answers with something that is not JSON must be recognisable in the error.
func TestRefreshExplainsAnUnreadableTokenResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"error" e}`))
	}))
	t.Cleanup(server.Close)

	client := &KeycloakClient{
		authUrl: server.URL, realm: "master", httpClient: server.Client(), initialLogin: true,
		clientCredentials: &ClientCredentials{GrantType: "client_credentials", ClientId: "c", ClientSecret: "s"},
	}

	err := client.Refresh(t.Context())
	if err == nil {
		t.Fatal("expected an error")
	}
	for _, expected := range []string{"refreshing credentials", "503", `{"error" e}`} {
		if !strings.Contains(err.Error(), expected) {
			t.Errorf("the error %q does not mention %q", err, expected)
		}
	}
}

// A realm that is not in the list may have been renamed to a name this account cannot read.
func TestGetRealmNameByInternalIdDoesNotCallARealmGoneWhenTheListMayBeIncomplete(t *testing.T) {
	fake := &fakeRealms{realms: []fakeRealm{{Realm: "other", Id: "u2"}}, notAdmin: true}
	client := fake.client(t)

	_, err := client.GetRealmNameByInternalId(t.Context(), "u1", "a")
	if err == nil || ErrorIs404(err) {
		t.Fatalf("err = %v, want an error that is not a 404", err)
	}
	if !strings.Contains(err.Error(), "deleted or renamed") {
		t.Errorf("the error does not explain itself: %v", err)
	}
}

// A token endpoint error that is valid JSON must not replace a working token with an empty one.
func TestRefreshKeepsTheTokenWhenTheTokenEndpointFails(t *testing.T) {
	for name, response := range map[string]struct {
		status int
		body   string
	}{
		"error answer":        {http.StatusServiceUnavailable, `{"error":"temporarily_unavailable"}`},
		"unauthorized answer": {http.StatusUnauthorized, `{"error":"invalid_client"}`},
		"no access token":     {http.StatusOK, `{"token_type":"Bearer"}`},
	} {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(response.status)
				_, _ = w.Write([]byte(response.body))
			}))
			t.Cleanup(server.Close)

			client := &KeycloakClient{
				authUrl: server.URL, realm: "master", httpClient: server.Client(), initialLogin: true,
				clientCredentials: &ClientCredentials{GrantType: "client_credentials", ClientId: "c", ClientSecret: "s", AccessToken: "working", TokenType: "Bearer"},
			}

			if err := client.Refresh(t.Context()); err == nil {
				t.Fatal("expected an error")
			}
			if _, accessToken := client.tokenTypeAndAccessToken(); accessToken != "working" {
				t.Errorf("the access token was replaced by %q", accessToken)
			}
		})
	}
}

// What a failed read says about the response must not contain the response, which can be a secret.
func TestGetErrorLeavesTheResponseOut(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"id":42,"value":"supersecret"}`))
	}))
	t.Cleanup(server.Close)

	client := &KeycloakClient{
		baseUrl: server.URL, realm: "master", httpClient: server.Client(), initialLogin: true, accessTokenProvided: true,
		clientCredentials: &ClientCredentials{AccessToken: "test", TokenType: "Bearer"},
	}

	_, err := client.GetRealm(t.Context(), "a")
	if err == nil {
		t.Fatal("expected a decoding error")
	}
	if strings.Contains(err.Error(), "supersecret") {
		t.Errorf("the response ended up in the error: %v", err)
	}
	if !strings.Contains(err.Error(), "/realms/a") {
		t.Errorf("the error does not name the path: %v", err)
	}
}

// A name is claimed by one realm: the reverse index must follow a rename, or a name that is free again would
// drop the hint of a realm that does not have it.
func TestRealmCacheIndexFollowsARename(t *testing.T) {
	var c realmCache
	c.remember(c.currentGeneration(), "uuid-1", "a")
	c.remember(c.currentGeneration(), "uuid-1", "b") // uuid-1 was renamed from a to b
	c.remember(c.currentGeneration(), "uuid-2", "a") // another realm took the old name

	if name, ok := c.nameHint("uuid-1"); !ok || name != "b" {
		t.Errorf("the hint of the renamed realm = %q, %v", name, ok)
	}
	if name, ok := c.nameHint("uuid-2"); !ok || name != "a" {
		t.Errorf("hint = %q, %v", name, ok)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.idsByName) != len(c.names) {
		t.Errorf("the index has %d entries for %d realms", len(c.idsByName), len(c.names))
	}
}

func TestResponseSnippetLeavesTokensOut(t *testing.T) {
	if snippet := responseSnippet([]byte(`{"access_token":"secret`)); strings.Contains(snippet, "secret") {
		t.Errorf("a token ended up in %q", snippet)
	}
	if snippet := responseSnippet([]byte(strings.Repeat("x", 500))); len(snippet) > 130 {
		t.Errorf("the snippet is too long: %d", len(snippet))
	}
}
