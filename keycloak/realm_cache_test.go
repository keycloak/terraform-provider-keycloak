package keycloak

import (
	"testing"
	"time"
)

func TestRealmCacheRemembersBothDirections(t *testing.T) {
	var c realmCache
	c.remember(c.currentGeneration(), "uuid-1", "realm-a")

	if name, ok := c.name("uuid-1"); !ok || name != "realm-a" {
		t.Errorf("name by uuid = %q, %v", name, ok)
	}
	if id, ok := c.id("realm-a"); !ok || id != "uuid-1" {
		t.Errorf("id by name = %q, %v", id, ok)
	}
	if _, ok := c.name("other"); ok {
		t.Error("unknown uuid must not be found")
	}
}

// A realm whose id Keycloak hides is remembered with an empty id, so it is not looked up again, but
// it must not be reachable by the empty id.
func TestRealmCacheHiddenIds(t *testing.T) {
	var c realmCache
	c.remember(c.currentGeneration(), "", "realm-a")

	if id, ok := c.id("realm-a"); !ok || id != "" {
		t.Errorf("id by name = %q, %v", id, ok)
	}
	if _, ok := c.name(""); ok {
		t.Error("the empty id must not resolve to a realm")
	}
}

func TestRealmCacheEntriesExpire(t *testing.T) {
	var c realmCache
	c.remember(c.currentGeneration(), "uuid-1", "realm-a")

	c.mu.Lock()
	old := time.Now().Add(-2 * realmCacheTTL)
	c.names["uuid-1"] = realmCacheEntry{value: "realm-a", at: old}
	c.ids["realm-a"] = realmCacheEntry{value: "uuid-1", at: old}
	c.mu.Unlock()

	if _, ok := c.name("uuid-1"); ok {
		t.Error("expired name must not be returned")
	}
	if _, ok := c.id("realm-a"); ok {
		t.Error("expired id must not be returned")
	}
}

// A listing that started before a rename must not bring the old names back once the rename cleared the cache.
func TestRealmCacheForgetBeatsSlowListing(t *testing.T) {
	var c realmCache

	generation := c.currentGeneration() // a listing starts here...
	c.forget()                          // ...a rename completes meanwhile...
	c.remember(generation, "uuid-1", "old-name")

	if name, ok := c.name("uuid-1"); ok {
		t.Errorf("stale name %q was stored after the cache was cleared", name)
	}

	c.remember(c.currentGeneration(), "uuid-1", "new-name")
	if name, ok := c.name("uuid-1"); !ok || name != "new-name" {
		t.Errorf("name after rename = %q, %v", name, ok)
	}
}
