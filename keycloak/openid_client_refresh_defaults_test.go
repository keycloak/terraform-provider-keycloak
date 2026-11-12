package keycloak

import (
	"encoding/json"
	"testing"
)

func TestOpenidClientRefreshTokenDefault(t *testing.T) {
	for _, test := range []struct {
		name, body string
		want       bool
	}{
		{"omitted", `{}`, true},
		{"explicit_false", `{"use.refresh.tokens":"false"}`, false},
		{"explicit_true", `{"use.refresh.tokens":"true"}`, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			// Reuse an object with the opposite value to check that absent attributes
			// receive the server default rather than retaining a previous response.
			attrs := OpenidClientAttributes{}
			attrs.UseRefreshTokens = true
			if test.want {
				attrs.UseRefreshTokens = false
			}
			if err := json.Unmarshal([]byte(test.body), &attrs); err != nil {
				t.Fatal(err)
			}
			if bool(attrs.UseRefreshTokens) != test.want {
				t.Fatalf("refresh tokens = %v, want %v", attrs.UseRefreshTokens, test.want)
			}
			if attrs.BackchannelLogoutSessionRequired {
				t.Fatal("unrelated omitted logout attribute changed")
			}
		})
	}
}
