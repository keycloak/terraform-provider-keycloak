package keycloak

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/keycloak/terraform-provider-keycloak/keycloak/types"
)

func legacyCompatClient(t *testing.T, serverVersion string, handler http.HandlerFunc) *KeycloakClient {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return &KeycloakClient{
		baseUrl: server.URL, httpClient: server.Client(), initialLogin: true,
		version:           version.Must(version.NewVersion(serverVersion)),
		clientCredentials: &ClientCredentials{AccessToken: "test", TokenType: "Bearer"},
	}
}

func legacyCompatDocument(t *testing.T, value string) legacyOrganizationDocument {
	t.Helper()
	var document legacyOrganizationDocument
	if err := json.Unmarshal([]byte(value), &document); err != nil {
		t.Fatal(err)
	}
	return document
}

func TestLegacyIdentityProviderRequestVersionGate(t *testing.T) {
	for _, serverVersion := range []string{"26.7.5", "26.8.0", "27.0.0"} {
		t.Run(serverVersion, func(t *testing.T) {
			var bodies []map[string]interface{}
			client := legacyCompatClient(t, serverVersion, func(w http.ResponseWriter, r *http.Request) {
				var body map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
				}
				bodies = append(bodies, body)
				w.WriteHeader(http.StatusNoContent)
			})
			idp := &IdentityProvider{Realm: "realm", Alias: "idp", OrganizationId: "org", Config: &IdentityProviderConfig{
				OrgDomain: "example.com", OrgRedirectModeEmailMatches: true,
				ClientSecret: "secret", ExtraConfig: map[string]interface{}{"custom": "value"},
			}}
			if err := client.NewIdentityProvider(context.Background(), idp); err != nil {
				t.Fatal(err)
			}
			if err := client.UpdateIdentityProvider(context.Background(), idp); err != nil {
				t.Fatal(err)
			}
			for _, body := range bodies {
				config := body["config"].(map[string]interface{})
				_, hasOrganization := body["organizationId"]
				_, hasDomain := config["kc.org.domain"]
				_, hasRedirect := config["kc.org.broker.redirect.mode.email-matches"]
				legacy := serverVersion == "26.7.5"
				if hasOrganization != legacy || hasDomain != legacy || hasRedirect != legacy {
					t.Fatalf("unexpected legacy fields in %s payload", serverVersion)
				}
				if config["clientSecret"] != "secret" || config["custom"] != "value" {
					t.Fatal("unrelated config was lost")
				}
			}
			if idp.OrganizationId != "org" || idp.Config.OrgDomain != "example.com" || !bool(idp.Config.OrgRedirectModeEmailMatches) {
				t.Fatal("request preparation mutated caller")
			}
		})
	}
}

func TestLegacyIdentityProviderReadProjection(t *testing.T) {
	for _, tc := range []struct {
		name, links, domains, preferredOrg, preferredDomain, wantOrg, wantDomain string
		wantRedirect, wantError                                                  bool
	}{
		{name: "legacy", links: "[]", domains: "[]"},
		{name: "single", links: `[{"organizationId":"org"}]`, domains: `[{"name":"example.com","identityProviderAlias":"idp","autoRedirect":true}]`, wantOrg: "org", wantDomain: "example.com", wantRedirect: true},
		{name: "any retained for one domain", links: `[{"organizationId":"org"}]`, domains: `[{"name":"example.com","identityProviderAlias":"idp"}]`, preferredOrg: "org", preferredDomain: "ANY", wantOrg: "org", wantDomain: "ANY"},
		{name: "all domains import", links: `[{"organizationId":"org"}]`, domains: `[{"name":"a.com","identityProviderAlias":"idp"},{"name":"b.com","identityProviderAlias":"idp"}]`, wantOrg: "org", wantDomain: "ANY"},
		{name: "ambiguous import", links: `[{"organizationId":"org"},{"organizationId":"other"}]`, domains: "[]", wantError: true},
		{name: "known organization", links: `[{"organizationId":"org"},{"organizationId":"other"}]`, domains: "[]", preferredOrg: "org", wantOrg: "org"},
		{name: "removed association", links: `[{"organizationId":"other"}]`, domains: "[]", preferredOrg: "org"},
		{name: "mixed redirects", links: `[{"organizationId":"org"}]`, domains: `[{"name":"a.com","identityProviderAlias":"idp","autoRedirect":true},{"name":"b.com","identityProviderAlias":"idp"}]`, wantError: true},
		{name: "subset cannot be represented", links: `[{"organizationId":"org"}]`, domains: `[{"name":"a.com","identityProviderAlias":"idp"},{"name":"b.com","identityProviderAlias":"idp"},{"name":"c.com"}]`, wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := legacyCompatClient(t, "26.8.0", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if strings.Contains(r.URL.Path, "/identity-provider/") {
					fmt.Fprintf(w, `{"alias":"idp","config":{},"organizationLinks":%s}`, tc.links)
				} else {
					fmt.Fprintf(w, `{"domains":%s}`, tc.domains)
				}
			})
			idp, err := client.GetIdentityProviderForOrganization(context.Background(), "realm", "idp", tc.preferredOrg, tc.preferredDomain)
			if (err != nil) != tc.wantError {
				t.Fatalf("unexpected error: %v", err)
			}
			if err == nil && (idp.OrganizationId != tc.wantOrg || idp.Config.OrgDomain != tc.wantDomain || bool(idp.Config.OrgRedirectModeEmailMatches) != tc.wantRedirect) {
				t.Fatalf("wrong projection: org=%q domain=%q redirect=%v", idp.OrganizationId, idp.Config.OrgDomain, idp.Config.OrgRedirectModeEmailMatches)
			}
		})
	}
}

func TestLegacyOrganizationRoutingOwnership(t *testing.T) {
	document := legacyCompatDocument(t, `{"name":"Org","custom":"keep","domains":[{"name":"a.com","verified":true,"identityProviderAlias":"idp","autoRedirect":true,"custom":"keep"},{"name":"b.com"},{"name":"c.com","identityProviderAlias":"other","autoRedirect":true}]}`)
	changed, err := configureLegacyOrganizationRouting(document, "idp", "b.com", false)
	if err != nil || !changed {
		t.Fatalf("routing update: %v", err)
	}
	domains := legacyOrganizationDomains(document)
	if domains[0]["identityProviderAlias"] != nil || domains[1]["identityProviderAlias"] != "idp" || domains[2]["identityProviderAlias"] != "other" || domains[0]["custom"] != "keep" || domains[0]["verified"] != true {
		t.Fatal("routing ownership or metadata not preserved")
	}
	before, _ := json.Marshal(document)
	if _, err := configureLegacyOrganizationRouting(document, "idp", "ANY", true); err == nil {
		t.Fatal("expected conflict with another IdP")
	}
	after, _ := json.Marshal(document)
	if string(before) != string(after) {
		t.Fatal("validation modified organization")
	}
	if _, err := configureLegacyOrganizationRouting(document, "idp", "missing.com", false); err == nil {
		t.Fatal("expected missing domain error")
	}
}

func TestLegacyOrganizationUpdatePreservesRouting(t *testing.T) {
	var updated legacyOrganizationDocument
	client := legacyCompatClient(t, "26.8.0", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			fmt.Fprint(w, `{"name":"Before","domains":[{"name":"example.com","verified":true,"identityProviderAlias":"idp","autoRedirect":true,"custom":"keep"},{"name":"removed.com","identityProviderAlias":"idp"}]}`)
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
			t.Error(err)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	err := client.UpdateOrganization(context.Background(), &Organization{Id: "org", Realm: "realm", Name: "After", Enabled: true, Domains: []OrganizationDomain{{Name: "example.com", Verified: false}, {Name: "new.com"}}})
	if err != nil {
		t.Fatal(err)
	}
	domains := legacyOrganizationDomains(updated)
	if updated["name"] != "After" || len(domains) != 2 || domains[0]["identityProviderAlias"] != "idp" || domains[0]["autoRedirect"] != true || domains[0]["verified"] != false || domains[0]["custom"] != "keep" || domains[1]["identityProviderAlias"] != nil {
		t.Fatalf("unexpected organization update: %#v", updated)
	}
}

func TestLegacyIdentityProviderReconcileRetrySwitchAndDetach(t *testing.T) {
	links := map[string]IdentityProviderOrganizationLink{
		"old":      {OrganizationId: "old", AutoMembership: true, MembershipType: "MANAGED"},
		"external": {OrganizationId: "external", AutoMembership: false, MembershipType: "UNMANAGED"},
	}
	organizations := map[string]legacyOrganizationDocument{
		"old": legacyCompatDocument(t, `{"name":"Old","domains":[{"name":"old.com","identityProviderAlias":"idp","autoRedirect":true}]}`),
		"new": legacyCompatDocument(t, `{"name":"New","domains":[{"name":"new.com"}]}`),
	}
	posts, failFirstLinkUpdate := 0, true
	client := legacyCompatClient(t, "26.8.0", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/identity-provider/instances/") {
			var items []IdentityProviderOrganizationLink
			for _, link := range links {
				items = append(items, link)
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"alias": "idp", "config": map[string]string{}, "organizationLinks": items})
			return
		}
		parts := strings.Split(r.URL.Path, "/")
		org := parts[5]
		if len(parts) == 6 {
			if r.Method == http.MethodGet {
				json.NewEncoder(w).Encode(organizations[org])
				return
			}
			var document legacyOrganizationDocument
			if err := json.NewDecoder(r.Body).Decode(&document); err != nil {
				t.Error(err)
			}
			organizations[org] = document
		} else {
			switch r.Method {
			case http.MethodPost:
				posts++
				if _, exists := links[org]; exists {
					http.Error(w, "duplicate", 409)
					return
				}
				links[org] = IdentityProviderOrganizationLink{OrganizationId: org, AutoMembership: true, MembershipType: "UNMANAGED"}
			case http.MethodPut:
				if failFirstLinkUpdate {
					failFirstLinkUpdate = false
					http.Error(w, "temporary failure", 400)
					return
				}
				var link IdentityProviderOrganizationLink
				if err := json.NewDecoder(r.Body).Decode(&link); err != nil {
					t.Error(err)
				}
				links[org] = link
			case http.MethodDelete:
				delete(links, org)
			}
		}
		w.WriteHeader(http.StatusNoContent)
	})
	desired := &IdentityProvider{Realm: "realm", Alias: "idp", OrganizationId: "new", Config: &IdentityProviderConfig{OrgDomain: "new.com", OrgRedirectModeEmailMatches: types.KeycloakBoolQuoted(true)}}
	if err := client.ReconcileLegacyIdentityProviderOrganization(context.Background(), desired, "old"); err == nil {
		t.Fatal("expected simulated partial failure")
	}
	if err := client.ReconcileLegacyIdentityProviderOrganization(context.Background(), desired, "old"); err != nil {
		t.Fatal(err)
	}
	if posts != 1 || links["new"].MembershipType != "MANAGED" || !links["new"].AutoMembership {
		t.Fatal("retry duplicated or failed to configure link")
	}
	if _, exists := links["old"]; exists {
		t.Fatal("old link retained")
	}
	if legacyOrganizationDomains(organizations["old"])[0]["identityProviderAlias"] != nil {
		t.Fatal("old routing retained")
	}
	if legacyOrganizationDomains(organizations["new"])[0]["identityProviderAlias"] != "idp" {
		t.Fatal("new routing missing")
	}
	desired.OrganizationId = ""
	desired.Config.OrgDomain = ""
	desired.Config.OrgRedirectModeEmailMatches = false
	if err := client.ReconcileLegacyIdentityProviderOrganization(context.Background(), desired, "new"); err != nil {
		t.Fatal(err)
	}
	want := map[string]IdentityProviderOrganizationLink{"external": {OrganizationId: "external", AutoMembership: false, MembershipType: "UNMANAGED"}}
	if !reflect.DeepEqual(links, want) {
		t.Fatalf("external links changed: %#v", links)
	}
	if legacyOrganizationDomains(organizations["new"])[0]["identityProviderAlias"] != nil {
		t.Fatal("routing retained after detach")
	}
}
