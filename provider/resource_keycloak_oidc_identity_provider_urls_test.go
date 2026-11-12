package provider

import (
	"strings"
	"testing"

	"github.com/hashicorp/go-cty/cty"
)

func TestOidcIdentityProviderRequiredUrls(t *testing.T) {
	url := cty.StringVal("https://example.com")
	unknown := cty.UnknownVal(cty.String)
	null := cty.NullVal(cty.String)

	config := func(authorizationUrl, tokenUrl cty.Value) cty.Value {
		return cty.ObjectVal(map[string]cty.Value{
			"authorization_url": authorizationUrl,
			"token_url":         tokenUrl,
		})
	}

	tests := []struct {
		name            string
		providerId      string
		providerIdKnown bool
		rawConfig       cty.Value
		wantErrors      []string
	}{
		{
			name:            "oidc with urls",
			providerId:      "oidc",
			providerIdKnown: true,
			rawConfig:       config(url, url),
		},
		{
			name:            "oidc without urls",
			providerId:      "oidc",
			providerIdKnown: true,
			rawConfig:       config(null, null),
			wantErrors: []string{
				`"authorization_url" is required when provider_id is "oidc"`,
				`"token_url" is required when provider_id is "oidc"`,
			},
		},
		{
			name:            "oidc without token url",
			providerId:      "oidc",
			providerIdKnown: true,
			rawConfig:       config(url, null),
			wantErrors:      []string{`"token_url" is required when provider_id is "oidc"`},
		},
		{
			name:            "keycloak-oidc without authorization url",
			providerId:      "keycloak-oidc",
			providerIdKnown: true,
			rawConfig:       config(null, url),
			wantErrors:      []string{`"authorization_url" is required when provider_id is "keycloak-oidc"`},
		},
		{
			// An empty string was accepted while the urls were required, and Keycloak does not get it.
			name:            "oidc with empty urls",
			providerId:      "oidc",
			providerIdKnown: true,
			rawConfig:       config(cty.StringVal(""), cty.StringVal("")),
		},
		{
			name:            "oidc with unknown urls",
			providerId:      "oidc",
			providerIdKnown: true,
			rawConfig:       config(unknown, unknown),
		},
		{
			name:            "oidc with unknown config",
			providerId:      "oidc",
			providerIdKnown: true,
			rawConfig:       cty.UnknownVal(config(url, url).Type()),
		},
		{
			name:            "custom provider id without urls",
			providerId:      "my-custom-oidc",
			providerIdKnown: true,
			rawConfig:       config(null, null),
		},
		{
			name:            "social provider id without urls",
			providerId:      "github",
			providerIdKnown: true,
			rawConfig:       config(null, null),
		},
		{
			name:            "unknown provider id without urls",
			providerId:      "",
			providerIdKnown: false,
			rawConfig:       config(null, null),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateOidcIdentityProviderRequiredUrls(test.providerId, test.providerIdKnown, test.rawConfig)

			if len(test.wantErrors) == 0 {
				if err != nil {
					t.Fatalf("expected no error, got: %v", err)
				}
				return
			}

			if err == nil {
				t.Fatalf("expected errors %q, got none", test.wantErrors)
			}
			for _, want := range test.wantErrors {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("expected error to contain %q, got: %v", want, err)
				}
			}
			if got := strings.Count(err.Error(), "is required when provider_id is"); got != len(test.wantErrors) {
				t.Errorf("expected %d errors, got %d: %v", len(test.wantErrors), got, err)
			}
		})
	}
}
