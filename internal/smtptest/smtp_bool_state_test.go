package smtptest

import (
	"testing"

	keycloakTypes "github.com/keycloak/terraform-provider-keycloak/keycloak/types"
)

// TestSmtpBoolStateRoundTrip demonstrates that KeycloakBoolQuoted must be
// converted to plain bool before being stored in the smtp state map.
//
// setRealmData builds a map[string]interface{} for smtp_server and passes it
// to data.Set. The schema declares starttls/ssl/allow_utf8 as schema.TypeBool,
// so the values must be plain Go bool — not the named type KeycloakBoolQuoted.
//
// If KeycloakBoolQuoted is stored directly, the .(bool) assertion inside
// getRealmFromData panics on the next Apply:
//
//	interface conversion: interface {} is types.KeycloakBoolQuoted, not bool
func TestSmtpBoolStateRoundTrip(t *testing.T) {
	const field = "allow_utf8"

	t.Run("BEFORE: KeycloakBoolQuoted stored, .(bool) asserted → panic", func(t *testing.T) {
		panicked := false
		func() {
			defer func() {
				if r := recover(); r != nil {
					panicked = true
					t.Logf("panic: %v", r)
				}
			}()
			smtpSettings := map[string]interface{}{
				field: keycloakTypes.KeycloakBoolQuoted(true), // named type stored directly
			}
			_ = keycloakTypes.KeycloakBoolQuoted(smtpSettings[field].(bool)) // panics here
		}()
		if !panicked {
			t.Error("expected a panic but got none")
		}
	})

	t.Run("AFTER: bool stored, .(bool) asserted → ok", func(t *testing.T) {
		raw := keycloakTypes.KeycloakBoolQuoted(true)
		smtpSettings := map[string]interface{}{
			field: bool(raw), // explicit cast to plain bool
		}
		result := keycloakTypes.KeycloakBoolQuoted(smtpSettings[field].(bool))
		if result != raw {
			t.Errorf("round-trip mismatch: got %v, want %v", result, raw)
		}
		t.Logf("round-trip OK: KeycloakBoolQuoted(%v) → bool(%v) → KeycloakBoolQuoted(%v)", raw, bool(raw), result)
	})
}
