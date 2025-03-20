package keycloak

import (
	"context"
	"fmt"
)

type RealmClientPolicyProfileExecutor struct {
	Name          string                 `json:"executor"`
	Configuration map[string]interface{} `json:"configuration"`
}

type RealmClientPolicyProfile struct {
	Name        string                             `json:"name"`
	RealmId     string                             `json:"-"`
	Description string                             `json:"description"`
	Executors   []RealmClientPolicyProfileExecutor `json:"executors"`
}

type RealmClientPolicyProfiles struct {
	Profiles       []RealmClientPolicyProfile `json:"profiles"`
	GlobalProfiles []RealmClientPolicyProfile `json:"globalProfiles"`
}

func (keycloakClient *KeycloakClient) UpdateRealmClientPolicyProfiles(ctx context.Context, realmId string, profiles *RealmClientPolicyProfiles) error {
	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/client-policies/profiles", realmId), profiles)
}

func (keycloakClient *KeycloakClient) GetAllRealmClientPolicyProfiles(ctx context.Context, realmId string) (*RealmClientPolicyProfiles, error) {
	var realmClientPolicyProfiles *RealmClientPolicyProfiles
	params := map[string]string{"include-global-profiles": "true"}

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/client-policies/profiles", realmId), &realmClientPolicyProfiles, params)
	if err != nil {
		return nil, err
	}

	return realmClientPolicyProfiles, nil
}

func (keycloakClient *KeycloakClient) GetRealmClientPolicyProfileByName(ctx context.Context, realmId string, name string) (*RealmClientPolicyProfile, error) {
	realmClientPolicyProfiles, err := keycloakClient.GetAllRealmClientPolicyProfiles(ctx, realmId)
	if err != nil {
		return nil, err
	}

	for _, profile := range realmClientPolicyProfiles.Profiles {
		if profile.Name == name {
			return &profile, nil
		}
	}

	return nil, fmt.Errorf("profile with name: %s not found", name)

}
