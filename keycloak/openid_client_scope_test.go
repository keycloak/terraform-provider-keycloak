package keycloak

import (
	"testing"
)

func TestIsDynamicScope(t *testing.T) {
	tests := []struct {
		name     string
		dynamic  bool
		expected bool
	}{
		{"static scope", false, false},
		{"dynamic scope", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scope := &OpenidClientScope{Dynamic: tt.dynamic}
			result := scope.IsDynamicScope()
			if result != tt.expected {
				t.Errorf("IsDynamicScope() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestValidateWildcardPattern(t *testing.T) {
	tests := []struct {
		name        string
		scope       *OpenidClientScope
		expectError bool
		errorMsg    string
	}{
		// Valid cases
		{
			name: "static scope - no validation",
			scope: &OpenidClientScope{
				Dynamic:            false,
				DynamicScopeRegexp: "",
			},
			expectError: false,
		},
		{
			name: "dynamic scope with wildcard at end",
			scope: &OpenidClientScope{
				Dynamic:            true,
				DynamicScopeRegexp: "resource:*",
			},
			expectError: false,
		},
		{
			name: "dynamic scope with wildcard in middle",
			scope: &OpenidClientScope{
				Dynamic:            true,
				DynamicScopeRegexp: "api:*:read",
			},
			expectError: false,
		},
		{
			name: "just asterisk",
			scope: &OpenidClientScope{
				Dynamic:            true,
				DynamicScopeRegexp: "*",
			},
			expectError: false,
		},
		{
			name: "static scope with valid pattern",
			scope: &OpenidClientScope{
				Dynamic:            false,
				DynamicScopeRegexp: "resource:*",
			},
			expectError: false,
		},
		// Error cases
		{
			name: "dynamic without pattern - required",
			scope: &OpenidClientScope{
				Dynamic:            true,
				DynamicScopeRegexp: "",
			},
			expectError: true,
			errorMsg:    "requires a wildcard pattern",
		},
		{
			name: "no asterisk",
			scope: &OpenidClientScope{
				Dynamic:            true,
				DynamicScopeRegexp: "resource:read",
			},
			expectError: true,
			errorMsg:    "exactly one asterisk",
		},
		{
			name: "multiple asterisks",
			scope: &OpenidClientScope{
				Dynamic:            true,
				DynamicScopeRegexp: "resource:*:*",
			},
			expectError: true,
			errorMsg:    "exactly one asterisk",
		},
		{
			name: "whitespace not allowed",
			scope: &OpenidClientScope{
				Dynamic:            true,
				DynamicScopeRegexp: "resource: *",
			},
			expectError: true,
			errorMsg:    "no whitespace",
		},
		{
			name: "pattern with dot and asterisk - invalid (regex metacharacter)",
			scope: &OpenidClientScope{
				Dynamic:            true,
				DynamicScopeRegexp: "resource:.*",
			},
			expectError: true,
			errorMsg:    "regex metacharacters",
		},
		{
			name: "pattern with question mark - invalid (regex metacharacter)",
			scope: &OpenidClientScope{
				Dynamic:            true,
				DynamicScopeRegexp: "resource:?*",
			},
			expectError: true,
			errorMsg:    "regex metacharacters",
		},
		{
			name: "pattern with only question mark - invalid",
			scope: &OpenidClientScope{
				Dynamic:            true,
				DynamicScopeRegexp: "resource:?",
			},
			expectError: true,
			errorMsg:    "exactly one asterisk",
		},
		{
			name: "static with regex metacharacter pattern - invalid",
			scope: &OpenidClientScope{
				Dynamic:            false,
				DynamicScopeRegexp: "resource:.*",
			},
			expectError: true,
			errorMsg:    "regex metacharacters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.scope.ValidateWildcardPattern()
			if tt.expectError {
				if err == nil {
					t.Errorf("ValidateWildcardPattern() expected error but got none")
				} else if tt.errorMsg != "" && !stringContains(err.Error(), tt.errorMsg) {
					t.Errorf("ValidateWildcardPattern() error = %v, want error containing %q", err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateWildcardPattern() unexpected error: %v", err)
				}
			}
		})
	}
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
