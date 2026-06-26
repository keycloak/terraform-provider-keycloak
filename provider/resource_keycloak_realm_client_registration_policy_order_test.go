package provider

import "testing"

// TestEqualCommaSeparatedSet verifies the order-insensitive comparison used to suppress
// spurious diffs on multi-value client registration policy config fields (e.g. trusted-hosts,
// allowed-client-scopes, allowed-protocol-mapper-types). Keycloak returns these arrays in an
// order that differs from the written order, which would otherwise cause perpetual drift on
// every plan/apply, while genuine add/remove/duplicate changes must still diff.
func TestEqualCommaSeparatedSet(t *testing.T) {
	cases := []struct {
		name  string
		old   string
		new   string
		equal bool
	}{
		{
			name:  "reordered by keycloak",
			old:   "localhost,127.0.0.1,example.org,auth.example.com",
			new:   "auth.example.com,localhost,127.0.0.1,example.org",
			equal: true,
		},
		{
			name:  "genuine change is not equal",
			old:   "localhost,example.org",
			new:   "localhost,other.org",
			equal: false,
		},
		{
			name:  "added element is not equal",
			old:   "localhost",
			new:   "localhost,example.org",
			equal: false,
		},
		{
			name:  "whitespace around elements is tolerated",
			old:   "a, b ,c",
			new:   "c,a,b",
			equal: true,
		},
		{
			name:  "duplicate element is a genuine difference",
			old:   "openid,openid,basic",
			new:   "openid,basic",
			equal: false,
		},
		{
			name:  "single value unchanged",
			old:   "localhost",
			new:   "localhost",
			equal: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := equalCommaSeparatedSet(tc.old, tc.new)
			if got != tc.equal {
				t.Fatalf("equalCommaSeparatedSet(old=%q, new=%q) = %v, want %v", tc.old, tc.new, got, tc.equal)
			}
		})
	}
}
