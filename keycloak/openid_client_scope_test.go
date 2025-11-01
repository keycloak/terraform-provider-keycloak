package keycloak

import (
	"testing"
)

func TestIsDynamicScope(t *testing.T) {
	tests := []struct {
		name     string
		scope    *OpenidClientScope
		expected bool
	}{
		{
			name: "static scope",
			scope: &OpenidClientScope{
				Name:               "profile",
				Dynamic:            false,
				DynamicScopeRegexp: "",
			},
			expected: false,
		},
		{
			name: "dynamic scope with flag",
			scope: &OpenidClientScope{
				Name:               "resource:read",
				Dynamic:            true,
				DynamicScopeRegexp: "",
			},
			expected: true,
		},
		{
			name: "dynamic scope with regexp only",
			scope: &OpenidClientScope{
				Name:               "group:admin",
				Dynamic:            false,
				DynamicScopeRegexp: "group:.*",
			},
			expected: true,
		},
		{
			name: "dynamic scope with both flag and regexp",
			scope: &OpenidClientScope{
				Name:               "custom:scope",
				Dynamic:            true,
				DynamicScopeRegexp: "custom:.*",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.scope.IsDynamicScope()
			if result != tt.expected {
				t.Errorf("IsDynamicScope() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestValidateDynamicScopeName(t *testing.T) {
	tests := []struct {
		name      string
		scope     *OpenidClientScope
		scopeName string
		expected  bool
	}{
		{
			name: "static scope always valid",
			scope: &OpenidClientScope{
				Dynamic:            false,
				DynamicScopeRegexp: "",
			},
			scopeName: "any-name",
			expected:  true,
		},
		{
			name: "dynamic scope with default pattern - valid",
			scope: &OpenidClientScope{
				Dynamic:            true,
				DynamicScopeRegexp: "",
			},
			scopeName: "resource:read",
			expected:  true,
		},
		{
			name: "dynamic scope with default pattern - valid with underscores",
			scope: &OpenidClientScope{
				Dynamic:            true,
				DynamicScopeRegexp: "",
			},
			scopeName: "user_profile:read",
			expected:  true,
		},
		{
			name: "dynamic scope with default pattern - valid with dashes",
			scope: &OpenidClientScope{
				Dynamic:            true,
				DynamicScopeRegexp: "",
			},
			scopeName: "api-resource:write",
			expected:  true,
		},
		{
			name: "dynamic scope with default pattern - invalid (no colon)",
			scope: &OpenidClientScope{
				Dynamic:            true,
				DynamicScopeRegexp: "",
			},
			scopeName: "resource",
			expected:  false,
		},
		{
			name: "dynamic scope with default pattern - invalid (empty after colon)",
			scope: &OpenidClientScope{
				Dynamic:            true,
				DynamicScopeRegexp: "",
			},
			scopeName: "resource:",
			expected:  false,
		},
		{
			name: "dynamic scope with default pattern - invalid (special chars)",
			scope: &OpenidClientScope{
				Dynamic:            true,
				DynamicScopeRegexp: "",
			},
			scopeName: "resource:read@write",
			expected:  false,
		},
		{
			name: "dynamic scope with custom pattern - valid",
			scope: &OpenidClientScope{
				Dynamic:            true,
				DynamicScopeRegexp: "group:.*",
			},
			scopeName: "group:admin",
			expected:  true,
		},
		{
			name: "dynamic scope with custom pattern - valid (wildcard)",
			scope: &OpenidClientScope{
				Dynamic:            true,
				DynamicScopeRegexp: "group:.*",
			},
			scopeName: "group:users",
			expected:  true,
		},
		{
			name: "dynamic scope with custom pattern - invalid",
			scope: &OpenidClientScope{
				Dynamic:            true,
				DynamicScopeRegexp: "group:.*",
			},
			scopeName: "resource:read",
			expected:  false,
		},
		{
			name: "dynamic scope with complex pattern - valid",
			scope: &OpenidClientScope{
				Dynamic:            true,
				DynamicScopeRegexp: "^[a-z]+:[0-9]+$",
			},
			scopeName: "resource:123",
			expected:  true,
		},
		{
			name: "dynamic scope with complex pattern - invalid",
			scope: &OpenidClientScope{
				Dynamic:            true,
				DynamicScopeRegexp: "^[a-z]+:[0-9]+$",
			},
			scopeName: "resource:abc",
			expected:  false,
		},
		{
			name: "dynamic scope with invalid regex pattern - returns false",
			scope: &OpenidClientScope{
				Dynamic:            true,
				DynamicScopeRegexp: "[invalid(",
			},
			scopeName: "anything",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.scope.ValidateDynamicScopeName(tt.scopeName)
			if result != tt.expected {
				t.Errorf("ValidateDynamicScopeName(%q) = %v, want %v", tt.scopeName, result, tt.expected)
			}
		})
	}
}

func TestValidateDynamicScopeName_CustomPatterns(t *testing.T) {
	tests := []struct {
		name      string
		regexp    string
		scopeName string
		expected  bool
	}{
		{
			name:      "OAuth2 resource scopes",
			regexp:    "^(read|write|delete):[a-z-]+$",
			scopeName: "read:users",
			expected:  true,
		},
		{
			name:      "OAuth2 resource scopes - invalid verb",
			regexp:    "^(read|write|delete):[a-z-]+$",
			scopeName: "execute:users",
			expected:  false,
		},
		{
			name:      "Hierarchical permissions",
			regexp:    "^[a-z0-9]+(/[a-z0-9]+)*:[a-z]+$",
			scopeName: "api/v1/users:read",
			expected:  true,
		},
		{
			name:      "Tenant-based scopes",
			regexp:    "^tenant:[a-zA-Z0-9-]+:[a-z]+$",
			scopeName: "tenant:acme-corp:admin",
			expected:  true,
		},
		{
			name:      "Tenant-based scopes - invalid format",
			regexp:    "^tenant:[a-zA-Z0-9-]+:[a-z]+$",
			scopeName: "tenant:admin",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scope := &OpenidClientScope{
				Dynamic:            true,
				DynamicScopeRegexp: tt.regexp,
			}
			result := scope.ValidateDynamicScopeName(tt.scopeName)
			if result != tt.expected {
				t.Errorf("ValidateDynamicScopeName(%q) with pattern %q = %v, want %v",
					tt.scopeName, tt.regexp, result, tt.expected)
			}
		})
	}
}
