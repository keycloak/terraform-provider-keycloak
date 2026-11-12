package keycloak

import (
	"encoding/json"
	"strings"
	"testing"
)

// Regression test for https://github.com/keycloak/terraform-provider-keycloak/issues/1740:
// an empty events_listeners must be sent as an empty list, because Keycloak leaves the
// listeners unchanged when "eventsListeners" is missing from the request.
func TestRealmEventsConfigMarshalsEmptyEventsListeners(t *testing.T) {
	realmEventsConfig := &RealmEventsConfig{
		EventsListeners: []string{},
	}

	data, err := json.Marshal(realmEventsConfig)
	if err != nil {
		t.Fatalf("failed to marshal realm events config: %s", err)
	}

	if body := string(data); !strings.Contains(body, `"eventsListeners":[]`) {
		t.Errorf("expected an empty eventsListeners list, got: %s", body)
	}
}
