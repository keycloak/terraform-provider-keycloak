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
	Name     string `json:"name"`
	Verified bool   `json:"verified,omitempty"`
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
	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/organizations/%s", organization.Realm, organization.Id), organization)
}

func (keycloakClient *KeycloakClient) DeleteOrganization(ctx context.Context, realmId, id string) error {
	return keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/organizations/%s", realmId, id), nil)
}
