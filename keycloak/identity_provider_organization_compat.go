package keycloak

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/keycloak/terraform-provider-keycloak/keycloak/types"
)

// Serialize organization read/modify/write operations within this provider process.
// UpdateOrganization must use the same lock (see the accompanying patch).
var identityProviderOrganizationMutex sync.Mutex

type IdentityProviderOrganizationLink struct {
	OrganizationId string `json:"organizationId"`
	AutoMembership bool   `json:"autoMembership"`
	MembershipType string `json:"membershipType"`
}

// Use a separate wire representation so this compatibility layer does not change
// the Terraform organization schema or discard fields unknown to the provider.
type legacyOrganizationDocument map[string]interface{}

func legacyOrganizationPath(realm, organizationId string) string {
	return fmt.Sprintf("/realms/%s/organizations/%s", url.PathEscape(realm), url.PathEscape(organizationId))
}

func (keycloakClient *KeycloakClient) getLegacyOrganizationDocument(ctx context.Context, realm, organizationId string) (legacyOrganizationDocument, error) {
	var organization legacyOrganizationDocument
	if err := keycloakClient.get(ctx, legacyOrganizationPath(realm, organizationId), &organization, nil); err != nil {
		return nil, err
	}
	return organization, nil
}

func legacyOrganizationDomains(organization legacyOrganizationDocument) []map[string]interface{} {
	var domains []map[string]interface{}
	entries, _ := organization["domains"].([]interface{})
	for _, entry := range entries {
		if domain, ok := entry.(map[string]interface{}); ok {
			domains = append(domains, domain)
		}
	}
	return domains
}

// Strip the legacy fields only from the request copy. Callers still need them
// to reconcile the association and to populate Terraform state.
func (keycloakClient *KeycloakClient) identityProviderRequest(ctx context.Context, idp *IdentityProvider) (interface{}, error) {
	modern, err := keycloakClient.VersionIsGreaterThanOrEqualTo(ctx, Version_26_8)
	if err != nil || !modern {
		return idp, err
	}
	request := *idp
	request.OrganizationId = ""
	request.OrganizationLinks = nil
	if idp.Config != nil {
		config := *idp.Config
		config.OrgDomain = ""
		config.OrgRedirectModeEmailMatches = false
		config.ExtraConfig = make(map[string]interface{}, len(idp.Config.ExtraConfig))
		for key, value := range idp.Config.ExtraConfig {
			switch key {
			case "kc.org.domain", "kc.org.broker.redirect.mode.email-matches", "kc.org.excluded.domains":
				return nil, fmt.Errorf("identity provider %q: legacy organization setting %q in extra_config is not supported on Keycloak 26.8+; use org_domain and org_redirect_mode_email_matches", idp.Alias, key)
			default:
				config.ExtraConfig[key] = value
			}
		}
		request.Config = &config
	}
	// IdentityProviderConfig.MarshalJSON does not honor omitempty. Remove the
	// legacy keys after marshaling so even empty values are not sent to 26.8+.
	encoded, err := json.Marshal(&request)
	if err != nil {
		return nil, err
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(encoded, &payload); err != nil {
		return nil, err
	}
	if config, ok := payload["config"].(map[string]interface{}); ok {
		delete(config, "kc.org.domain")
		delete(config, "kc.org.broker.redirect.mode.email-matches")
		delete(config, "kc.org.excluded.domains")
	}
	return payload, nil
}

// GetIdentityProviderForOrganization preserves the organization selected in state.
// With no preferred organization (e.g. import), only an unambiguous link can be
// represented by the legacy single-organization schema.
func (keycloakClient *KeycloakClient) GetIdentityProviderForOrganization(ctx context.Context, realm, alias, preferredOrganizationId, preferredDomain string) (*IdentityProvider, error) {
	idp, err := keycloakClient.getIdentityProvider(ctx, realm, alias)
	if err != nil {
		return nil, err
	}
	modern, err := keycloakClient.VersionIsGreaterThanOrEqualTo(ctx, Version_26_8)
	if err != nil || !modern {
		return idp, err
	}
	if err := keycloakClient.readLegacyIdentityProviderOrganization(ctx, idp, preferredOrganizationId, preferredDomain); err != nil {
		// Do not expose an organization's 404 as an IdP 404 to Terraform.
		return nil, fmt.Errorf("reading organization settings for identity provider %q: %v", alias, err)
	}
	return idp, nil
}

func (keycloakClient *KeycloakClient) readLegacyIdentityProviderOrganization(ctx context.Context, idp *IdentityProvider, preferredOrganizationId, preferredDomain string) error {
	idp.OrganizationId = ""
	if idp.Config == nil {
		idp.Config = &IdentityProviderConfig{}
	}
	idp.Config.OrgDomain = ""
	idp.Config.OrgRedirectModeEmailMatches = false
	if preferredOrganizationId != "" {
		for _, link := range idp.OrganizationLinks {
			if link.OrganizationId == preferredOrganizationId {
				idp.OrganizationId = preferredOrganizationId
				break
			}
		}
	} else {
		switch len(idp.OrganizationLinks) {
		case 0:
		case 1:
			idp.OrganizationId = idp.OrganizationLinks[0].OrganizationId
		default:
			return fmt.Errorf("multiple organizations are linked; the legacy organization_id schema cannot import this identity provider unambiguously")
		}
	}
	if idp.OrganizationId == "" {
		return nil
	}
	organization, err := keycloakClient.getLegacyOrganizationDocument(ctx, idp.Realm, idp.OrganizationId)
	if err != nil {
		return err
	}
	domains := legacyOrganizationDomains(organization)
	var routed []string
	var redirect bool
	for _, domain := range domains {
		if domain["identityProviderAlias"] != idp.Alias {
			continue
		}
		name, _ := domain["name"].(string)
		autoRedirect, _ := domain["autoRedirect"].(bool)
		if len(routed) > 0 && redirect != autoRedirect {
			return fmt.Errorf("domains routed to this identity provider have different autoRedirect values; the legacy schema cannot represent this configuration")
		}
		routed = append(routed, name)
		redirect = autoRedirect
	}
	switch {
	case len(routed) == 0:
		// ANY on a domain-less organization has no observable routing state.
		if len(domains) == 0 && preferredDomain == "ANY" {
			idp.Config.OrgDomain = "ANY"
		}
	case preferredDomain == "ANY" && len(routed) == len(domains):
		idp.Config.OrgDomain = "ANY"
	case len(routed) == 1:
		idp.Config.OrgDomain = routed[0]
		if strings.EqualFold(preferredDomain, routed[0]) {
			idp.Config.OrgDomain = preferredDomain
		}
	case len(routed) == len(domains):
		idp.Config.OrgDomain = "ANY"
	default:
		return fmt.Errorf("multiple but not all organization domains route to this identity provider; the legacy org_domain schema cannot represent this configuration")
	}
	idp.Config.OrgRedirectModeEmailMatches = types.KeycloakBoolQuoted(redirect)
	return nil
}

// Validate before creating an IdP or unlinking its old organization.
func (keycloakClient *KeycloakClient) ValidateLegacyIdentityProviderOrganization(ctx context.Context, idp *IdentityProvider) error {
	if _, err := keycloakClient.identityProviderRequest(ctx, idp); err != nil {
		return err
	}
	if idp.Config == nil {
		return fmt.Errorf("identity provider config is required")
	}
	if idp.OrganizationId == "" {
		if idp.Config.OrgDomain != "" || bool(idp.Config.OrgRedirectModeEmailMatches) {
			return fmt.Errorf("organization_id is required for organization domain routing on Keycloak 26.8+")
		}
		return nil
	}
	if bool(idp.Config.OrgRedirectModeEmailMatches) && idp.Config.OrgDomain == "" {
		return fmt.Errorf("org_domain is required when org_redirect_mode_email_matches is true on Keycloak 26.8+")
	}
	organization, err := keycloakClient.getLegacyOrganizationDocument(ctx, idp.Realm, idp.OrganizationId)
	if err != nil {
		return err
	}
	_, err = configureLegacyOrganizationRouting(organization, idp.Alias, idp.Config.OrgDomain, bool(idp.Config.OrgRedirectModeEmailMatches))
	return err
}

// Only routing owned by this IdP is cleared; routing to another IdP is never
// silently overwritten. Unknown organization and domain fields are retained.
func configureLegacyOrganizationRouting(organization legacyOrganizationDocument, alias, selectedDomain string, redirect bool) (bool, error) {
	domains := legacyOrganizationDomains(organization)
	found := selectedDomain == "" || selectedDomain == "ANY"
	for _, domain := range domains {
		name, _ := domain["name"].(string)
		selected := selectedDomain == "ANY" || (selectedDomain != "" && strings.EqualFold(name, selectedDomain))
		if selected {
			found = true
			existing, _ := domain["identityProviderAlias"].(string)
			if existing != "" && existing != alias {
				return false, fmt.Errorf("organization domain %q already routes to identity provider %q", name, existing)
			}
		}
	}
	if !found {
		return false, fmt.Errorf("org_domain %q does not belong to the selected organization", selectedDomain)
	}
	if len(domains) == 0 && redirect {
		return false, fmt.Errorf("auto redirect requires at least one organization domain on Keycloak 26.8+")
	}
	changed := false
	for _, domain := range domains {
		name, _ := domain["name"].(string)
		existing, _ := domain["identityProviderAlias"].(string)
		autoRedirect, _ := domain["autoRedirect"].(bool)
		selected := selectedDomain == "ANY" || (selectedDomain != "" && strings.EqualFold(name, selectedDomain))
		if selected && (existing != alias || autoRedirect != redirect) {
			domain["identityProviderAlias"] = alias
			domain["autoRedirect"] = redirect
			changed = true
		} else if !selected && existing == alias {
			domain["identityProviderAlias"] = nil
			domain["autoRedirect"] = false
			changed = true
		}
	}
	return changed, nil
}

func (keycloakClient *KeycloakClient) setLegacyOrganizationRouting(ctx context.Context, realm, organizationId, alias, domain string, redirect bool) error {
	if organizationId == "" {
		return nil
	}
	organization, err := keycloakClient.getLegacyOrganizationDocument(ctx, realm, organizationId)
	if err != nil {
		return err
	}
	changed, err := configureLegacyOrganizationRouting(organization, alias, domain, redirect)
	if err != nil || !changed {
		return err
	}
	return keycloakClient.put(ctx, legacyOrganizationPath(realm, organizationId), organization)
}

// Called only for Keycloak 26.8+. It is safe to retry after a partial failure:
// the actual links are read again instead of blindly POSTing the association.
func (keycloakClient *KeycloakClient) ReconcileLegacyIdentityProviderOrganization(ctx context.Context, desired *IdentityProvider, oldOrganizationId string) error {
	identityProviderOrganizationMutex.Lock()
	defer identityProviderOrganizationMutex.Unlock()
	if err := keycloakClient.ValidateLegacyIdentityProviderOrganization(ctx, desired); err != nil {
		return err
	}
	current, err := keycloakClient.getIdentityProvider(ctx, desired.Realm, desired.Alias)
	if err != nil {
		return err
	}
	linked := make(map[string]IdentityProviderOrganizationLink)
	for _, link := range current.OrganizationLinks {
		linked[link.OrganizationId] = link
		if desired.OrganizationId != "" && link.OrganizationId != desired.OrganizationId && link.OrganizationId != oldOrganizationId && link.MembershipType == "MANAGED" {
			return fmt.Errorf("identity provider %q already has a managed link to organization %q", desired.Alias, link.OrganizationId)
		}
	}
	if oldOrganizationId != "" && oldOrganizationId != desired.OrganizationId {
		if _, exists := linked[oldOrganizationId]; exists {
			if err := keycloakClient.setLegacyOrganizationRouting(ctx, desired.Realm, oldOrganizationId, desired.Alias, "", false); err != nil {
				return err
			}
			if err := keycloakClient.UnlinkIdentityProviderFromOrganization(ctx, desired.Realm, desired.Alias, oldOrganizationId); err != nil {
				return err
			}
		}
	}
	if desired.OrganizationId == "" {
		return nil
	}
	link, exists := linked[desired.OrganizationId]
	path := legacyOrganizationPath(desired.Realm, desired.OrganizationId) + "/identity-providers"
	if !exists {
		if _, _, err := keycloakClient.post(ctx, path, desired.Alias); err != nil {
			return err
		}
	}
	if !exists || !link.AutoMembership || link.MembershipType != "MANAGED" {
		link = IdentityProviderOrganizationLink{OrganizationId: desired.OrganizationId, AutoMembership: true, MembershipType: "MANAGED"}
		if err := keycloakClient.put(ctx, path+"/"+url.PathEscape(desired.Alias), link); err != nil {
			return err
		}
	}
	return keycloakClient.setLegacyOrganizationRouting(ctx, desired.Realm, desired.OrganizationId, desired.Alias, desired.Config.OrgDomain, bool(desired.Config.OrgRedirectModeEmailMatches))
}

// Preserve domain routing on ordinary organization updates. This is required
// because the existing organization schema knows only name and verified.
func (keycloakClient *KeycloakClient) updateOrganizationPreservingIdentityProviderRouting(ctx context.Context, organization *Organization) error {
	identityProviderOrganizationMutex.Lock()
	defer identityProviderOrganizationMutex.Unlock()
	current, err := keycloakClient.getLegacyOrganizationDocument(ctx, organization.Realm, organization.Id)
	if err != nil {
		return err
	}
	previous := make(map[string]map[string]interface{})
	for _, domain := range legacyOrganizationDomains(current) {
		name, _ := domain["name"].(string)
		previous[strings.ToLower(name)] = domain
	}
	var domains []map[string]interface{}
	if organization.Domains != nil {
		domains = make([]map[string]interface{}, 0, len(organization.Domains))
		for _, domain := range organization.Domains {
			wire := previous[strings.ToLower(domain.Name)]
			if wire == nil {
				wire = make(map[string]interface{})
			}
			wire["name"] = domain.Name
			wire["verified"] = domain.Verified
			domains = append(domains, wire)
		}
	}
	// Embed the existing representation but override domains with their complete
	// wire values. Other organization fields retain their normal update behavior.
	request := struct {
		*Organization
		Domains []map[string]interface{} `json:"domains"`
	}{Organization: organization, Domains: domains}
	return keycloakClient.put(ctx, legacyOrganizationPath(organization.Realm, organization.Id), request)
}
