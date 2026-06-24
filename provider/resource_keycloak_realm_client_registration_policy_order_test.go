package provider

import "testing"

// TestSuppressMultiValueClientRegistrationConfigOrder verifies that reordering the
// elements of multi-value config fields (e.g. trusted-hosts, allowed-client-scopes) does
// not produce spurious diffs, while genuine changes still do. Keycloak returns these
// arrays in an order that differs from the written order, which previously caused
// perpetual drift on every plan/apply.
func TestSuppressMultiValueClientRegistrationConfigOrder(t *testing.T) {
	cases := []struct {
		name     string
		key      string
		old      string
		new      string
		suppress bool
	}{
		{
			name:     "trusted-hosts reordered by keycloak",
			key:      "config.trusted-hosts",
			old:      "localhost,127.0.0.1,example.org,auth.example.com",
			new:      "auth.example.com,localhost,127.0.0.1,example.org",
			suppress: true,
		},
		{
			name:     "allowed-client-scopes reordered by keycloak",
			key:      "config.allowed-client-scopes",
			old:      "address,organization,basic,openid",
			new:      "openid,basic,address,organization",
			suppress: true,
		},
		{
			name:     "trusted-hosts genuine change is not suppressed",
			key:      "config.trusted-hosts",
			old:      "localhost,example.org",
			new:      "localhost,other.org",
			suppress: false,
		},
		{
			name:     "trusted-hosts added element is not suppressed",
			key:      "config.trusted-hosts",
			old:      "localhost",
			new:      "localhost,example.org",
			suppress: false,
		},
		{
			name:     "non multi-value key is never suppressed",
			key:      "config.client-uris-must-match",
			old:      "true",
			new:      "false",
			suppress: false,
		},
		{
			name:     "map element-count pseudo key is not suppressed",
			key:      "config.%",
			old:      "3",
			new:      "4",
			suppress: false,
		},
		{
			name:     "whitespace around elements is tolerated",
			key:      "config.trusted-hosts",
			old:      "a, b ,c",
			new:      "c,a,b",
			suppress: true,
		},
		{
			name:     "duplicate element is a genuine difference",
			key:      "config.allowed-client-scopes",
			old:      "openid,openid,basic",
			new:      "openid,basic",
			suppress: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := suppressMultiValueClientRegistrationConfigOrder(tc.key, tc.old, tc.new, nil)
			if got != tc.suppress {
				t.Fatalf("suppress(key=%q, old=%q, new=%q) = %v, want %v", tc.key, tc.old, tc.new, got, tc.suppress)
			}
		})
	}
}
