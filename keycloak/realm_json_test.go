package keycloak

import (
	"encoding/json"
	"testing"
)

// Keycloak keeps the stored value for a field that is missing from the request, so false must be sent explicitly.
func TestRealmSendsOfflineSessionMaxLifespanEnabledWhenFalse(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		body, err := json.Marshal(&Realm{OfflineSessionMaxLifespanEnabled: enabled})
		if err != nil {
			t.Fatal(err)
		}

		var fields map[string]interface{}
		if err := json.Unmarshal(body, &fields); err != nil {
			t.Fatal(err)
		}

		got, ok := fields["offlineSessionMaxLifespanEnabled"]
		if !ok {
			t.Fatalf("offlineSessionMaxLifespanEnabled = %v is missing from the request body: %s", enabled, body)
		}
		if got != enabled {
			t.Fatalf("offlineSessionMaxLifespanEnabled = %v in the request body, want %v", got, enabled)
		}
	}
}
