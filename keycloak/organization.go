package keycloak

import (
	"context"
	"fmt"
)

type Organization struct {
	Id          string               `json:"id,omitempty"`
	Realm       string               `json:"-"`
	Name        string               `json:"name"`
	Alias       string               `json:"alias,omitempty"`
	Enabled     bool                 `json:"enabled"`
	Description string               `json:"description,omitempty"`
	RedirectUrl string               `json:"redirectUrl,omitempty"`
	Domains     []OrganizationDomain `json:"domains"`
	Attributes  map[string][]string  `json:"attributes"`
}

type OrganizationDomain struct {
	Name                  string `json:"name"`
	Verified              bool   `json:"verified,omitempty"`
	IdentityProviderAlias string `json:"identityProviderAlias,omitempty"`
	AutoRedirect          bool   `json:"autoRedirect,omitempty"`
}

type OrganizationIdentityProvider struct {
	Alias          string
	AutoMembership bool
	MembershipType string
}

type organizationIdentityProviderRepresentation struct {
	Alias             string                             `json:"alias"`
	OrganizationLinks []IdentityProviderOrganizationLink `json:"organizationLinks"`
}

func (keycloakClient *KeycloakClient) NewOrganization(ctx context.Context, organization *Organization) error {
	_, location, err := keycloakClient.post(ctx, fmt.Sprintf("/realms/%s/organizations", organization.Realm), organization)
	if err != nil {
		return err
	}
	organization.Id = getIdFromLocationHeader(location)

	return nil
}

func (keycloakClient *KeycloakClient) GetOrganization(ctx context.Context, realm, id string) (*Organization, error) {
	var organization Organization

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/organizations/%s", realm, id), &organization, nil)
	if err != nil {
		return nil, err
	}

	organization.Realm = realm

	return &organization, nil
}

func (keycloakClient *KeycloakClient) GetOrganizationByName(ctx context.Context, realm string, name string) (*Organization, error) {
	var organizations []Organization
	var orgFound *Organization

	params := map[string]string{
		"search": name,
	}

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/organizations", realm), &organizations, params)
	if err != nil {
		return nil, err
	}

	for _, org := range organizations {
		if org.Name == name {
			if orgFound, err = keycloakClient.GetOrganization(ctx, realm, org.Id); err != nil {
				return nil, err
			}
			return orgFound, nil
		}
	}

	return nil, fmt.Errorf("organization with name %s not found", name)
}

func (keycloakClient *KeycloakClient) UpdateOrganization(ctx context.Context, organization *Organization) error {
	modern, err := keycloakClient.VersionIsGreaterThanOrEqualTo(ctx, Version_26_8)
	if err != nil {
		return err
	}
	if modern {
		return keycloakClient.updateOrganizationPreservingIdentityProviderRouting(ctx, organization)
	}
	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/organizations/%s", organization.Realm, organization.Id), organization)
}

// UpdateOrganizationWithDomainRouting writes the identity provider routing of each domain as given,
// where UpdateOrganization keeps the routing that the server has. Requires Keycloak 26.8 or later.
func (keycloakClient *KeycloakClient) UpdateOrganizationWithDomainRouting(ctx context.Context, organization *Organization) error {
	identityProviderOrganizationMutex.Lock()
	defer identityProviderOrganizationMutex.Unlock()
	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/organizations/%s", organization.Realm, organization.Id), organization)
}

// GetOrganizationIdentityProvider reads one link, and returns nil when the identity provider is not
// linked to the organization. It does not use the list of the organization's identity providers:
// on Keycloak 26.8.0 that list was seen to fail with 400 "Cannot parse the JSON" while two
// organizations shared an identity provider, because the server loads the other organization after
// the request's database session has closed.
func (keycloakClient *KeycloakClient) GetOrganizationIdentityProvider(ctx context.Context, realm, organizationId, alias string) (*OrganizationIdentityProvider, error) {
	var representation organizationIdentityProviderRepresentation

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/organizations/%s/identity-providers/%s", realm, organizationId, alias), &representation, nil)
	if ErrorIs404(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	for _, link := range representation.OrganizationLinks {
		if link.OrganizationId == organizationId {
			return &OrganizationIdentityProvider{
				Alias:          representation.Alias,
				AutoMembership: link.AutoMembership,
				MembershipType: link.MembershipType,
			}, nil
		}
	}

	return nil, nil
}

func (keycloakClient *KeycloakClient) UpdateOrganizationIdentityProviderLink(ctx context.Context, realm, organizationId string, identityProvider OrganizationIdentityProvider) error {
	link := IdentityProviderOrganizationLink{
		OrganizationId: organizationId,
		AutoMembership: identityProvider.AutoMembership,
		MembershipType: identityProvider.MembershipType,
	}

	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/organizations/%s/identity-providers/%s", realm, organizationId, identityProvider.Alias), link)
}

// LinkOrganizationIdentityProviders links each identity provider to the organization, or updates its
// link, so that a domain can route to it.
func (keycloakClient *KeycloakClient) LinkOrganizationIdentityProviders(ctx context.Context, organization *Organization, identityProviders []OrganizationIdentityProvider) error {
	identityProviderOrganizationMutex.Lock()
	defer identityProviderOrganizationMutex.Unlock()

	for _, identityProvider := range identityProviders {
		current, err := keycloakClient.GetOrganizationIdentityProvider(ctx, organization.Realm, organization.Id, identityProvider.Alias)
		if err != nil {
			return err
		}
		if current == nil {
			// 409: another client linked it after the read.
			err := keycloakClient.LinkIdentityProviderWithOrganization(ctx, organization.Realm, identityProvider.Alias, organization.Id)
			if err != nil && !ErrorIs409(err) {
				return err
			}
		}
		if current == nil || *current != identityProvider {
			if err := keycloakClient.UpdateOrganizationIdentityProviderLink(ctx, organization.Realm, organization.Id, identityProvider); err != nil {
				return err
			}
		}
	}

	return nil
}

// UnlinkOrganizationIdentityProviders removes the links of the given identity providers. A link that is
// already gone is not an error. When a removal fails, it returns the aliases that it did not remove: the
// one that failed and those after it.
func (keycloakClient *KeycloakClient) UnlinkOrganizationIdentityProviders(ctx context.Context, organization *Organization, aliases []string) ([]string, error) {
	identityProviderOrganizationMutex.Lock()
	defer identityProviderOrganizationMutex.Unlock()

	for i, alias := range aliases {
		err := keycloakClient.UnlinkIdentityProviderFromOrganization(ctx, organization.Realm, alias, organization.Id)
		if err != nil && !ErrorIs404(err) {
			return aliases[i:], err
		}
	}

	return nil, nil
}

func (keycloakClient *KeycloakClient) DeleteOrganization(ctx context.Context, realmId, id string) error {
	return keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/organizations/%s", realmId, id), nil)
}

type OrganizationMember struct {
	Id             string `json:"id"`
	Username       string `json:"username"`
	Email          string `json:"email"`
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
	MembershipType string `json:"membershipType"`
}

func (keycloakClient *KeycloakClient) GetOrganizationMembers(ctx context.Context, realm, orgId string) ([]*OrganizationMember, error) {
	var members []*OrganizationMember
	var first, pagination = 0, 50
	var iterationMembers []*OrganizationMember

	for ok := true; ok; ok = len(iterationMembers) > 0 {
		iterationMembers = nil
		err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/organizations/%s/members?max=%d&first=%d", realm, orgId, pagination, first), &iterationMembers, nil)
		if err != nil {
			return nil, err
		}
		members = append(members, iterationMembers...)
		first += pagination
	}

	return members, nil
}

func (keycloakClient *KeycloakClient) AddUserToOrganization(ctx context.Context, realm, orgId, userId string) error {
	_, err := keycloakClient.sendRaw(ctx, fmt.Sprintf("/realms/%s/organizations/%s/members", realm, orgId), []byte(userId))
	return err
}

func (keycloakClient *KeycloakClient) UpdateOrganizationMembershipType(ctx context.Context, realm, orgId, userId, membershipType string) error {
	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/organizations/%s/members/%s/membership-type", realm, orgId, userId), membershipType)
}

func (keycloakClient *KeycloakClient) RemoveUserFromOrganization(ctx context.Context, realm, orgId, userId string) error {
	return keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/organizations/%s/members/%s", realm, orgId, userId), nil)
}
