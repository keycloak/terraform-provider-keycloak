package keycloak

import (
	"strings"
	"testing"
)

func TestScopeClaims_conflictingTypesError(t *testing.T) {
	c := (&KeycloakClient{}).ForNewRun()
	owner := OpenidClientScopeOwnerKey("realm", "client")

	if err := c.ClaimScopes(owner, "default", []string{"email", "profile"}); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	err := c.ClaimScopes(owner, "optional", []string{"phone", "profile", "email"})
	if err == nil {
		t.Fatal("expected a conflict error")
	}
	if !strings.Contains(err.Error(), "[email, profile]") {
		t.Fatalf("error should list the conflicting scopes, got: %s", err)
	}
}

func TestScopeClaims_sameTypeReplacesClaim(t *testing.T) {
	c := (&KeycloakClient{}).ForNewRun()
	owner := RealmScopeOwnerKey("realm")

	if err := c.ClaimScopes(owner, "default", []string{"email"}); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	// the default claim is replaced, so email is no longer claimed as default
	if err := c.ClaimScopes(owner, "default", []string{"profile"}); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if err := c.ClaimScopes(owner, "optional", []string{"email"}); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestScopeClaims_releaseAndOwnerIsolation(t *testing.T) {
	c := (&KeycloakClient{}).ForNewRun()
	owner := SamlClientScopeOwnerKey("realm", "client")

	if err := c.ClaimScopes(owner, "default", []string{"email"}); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	// a different owner may use the same scope as optional
	if err := c.ClaimScopes(SamlClientScopeOwnerKey("realm", "other"), "optional", []string{"email"}); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	c.ReleaseScopes(owner, "default")
	if err := c.ClaimScopes(owner, "optional", []string{"email"}); err != nil {
		t.Fatalf("unexpected error after release: %s", err)
	}
}

func TestScopeClaims_failedClaimIsNotRecorded(t *testing.T) {
	c := (&KeycloakClient{}).ForNewRun()
	owner := RealmScopeOwnerKey("realm")

	_ = c.ClaimScopes(owner, "default", []string{"email"})
	if err := c.ClaimScopes(owner, "optional", []string{"email"}); err == nil {
		t.Fatal("expected a conflict error")
	}
	// the rejected optional claim must not block a later default claim
	c.ReleaseScopes(owner, "default")
	if err := c.ClaimScopes(owner, "default", []string{"email"}); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestScopeClaims_newRunStartsWithoutClaims(t *testing.T) {
	c := (&KeycloakClient{}).ForNewRun()
	owner := RealmScopeOwnerKey("realm")

	_ = c.ClaimScopes(owner, "default", []string{"email"})
	if err := c.ForNewRun().ClaimScopes(owner, "optional", []string{"email"}); err != nil {
		t.Fatalf("a new run must not see the claims of an earlier run: %s", err)
	}
}
