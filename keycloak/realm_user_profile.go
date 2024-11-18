package keycloak

import (
	"context"
	"encoding/json"
	"fmt"
)

type RealmUserProfilePermissions struct {
	Edit []string `json:"edit"`
	View []string `json:"view"`
}

type RealmUserProfileRequired struct {
	Roles  []string `json:"roles,omitempty"`
	Scopes []string `json:"scopes,omitempty"`
}

type RealmUserProfileSelector struct {
	Scopes []string `json:"scopes,omitempty"`
}

type RealmUserProfileValidationConfig map[string]interface{}

type RealmUserProfileAttribute struct {
	Annotations map[string]interface{}                      `json:"annotations,omitempty"`
	DisplayName string                                      `json:"displayName,omitempty"`
	Group       string                                      `json:"group,omitempty"`
	Name        string                                      `json:"name"`
	Permissions *RealmUserProfilePermissions                `json:"permissions,omitempty"`
	Required    *RealmUserProfileRequired                   `json:"required,omitempty"`
	Selector    *RealmUserProfileSelector                   `json:"selector,omitempty"`
	Validations map[string]RealmUserProfileValidationConfig `json:"validations,omitempty"`
}

type RealmUserProfileGroup struct {
	Annotations        map[string]interface{} `json:"annotations,omitempty"`
	DisplayDescription string                 `json:"displayDescription,omitempty"`
	DisplayHeader      string                 `json:"displayHeader,omitempty"`
	Name               string                 `json:"name"`
}

type RealmUserProfile struct {
	Attributes               []*RealmUserProfileAttribute `json:"attributes"`
	Groups                   []*RealmUserProfileGroup     `json:"groups,omitempty"`
	UnmanagedAttributePolicy string                       `json:"unmanagedAttributePolicy,omitempty"`
}

func (keycloakClient *KeycloakClient) UpdateRealmUserProfile(ctx context.Context, realmId string, realmUserProfile *RealmUserProfile) error {
	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/users/profile", realmId), realmUserProfile)
}

func (keycloakClient *KeycloakClient) GetRealmUserProfile(ctx context.Context, realmId string) (*RealmUserProfile, error) {
	var realmUserProfile RealmUserProfile
	body, err := keycloakClient.getRaw(ctx, fmt.Sprintf("/realms/%s/users/profile", realmId), nil)
	if err != nil {
		return nil, err
	}

	if string(body) == "" {
		return nil, fmt.Errorf("User Profile is disabled for the %s realm", realmId)
	}

	err = json.Unmarshal(body, &realmUserProfile)
	if err != nil {
		return nil, err
	}

	for _, attr := range realmUserProfile.Attributes {
		if attr.Validations != nil {
			for name, config := range attr.Validations {

				c := make(map[string]interface{})
				for k, v := range config {
					if _, ok := v.([]interface{}); ok {
						tmp, _ := json.Marshal(v)
						c[k] = string(tmp)
					} else {
						c[k] = v
					}
				}

				attr.Validations[name] = c
			}
		}
		if attr.Annotations != nil {
			for k, v := range attr.Annotations {

				if _, ok := v.(map[string]interface{}); ok {
					tmp, _ := json.Marshal(v)
					attr.Annotations[k] = string(tmp)
				} else {
					attr.Annotations[k] = v
				}

			}
		}
	}

	for _, attr := range realmUserProfile.Groups {
		if attr.Annotations != nil {
			for k, v := range attr.Annotations {

				if _, ok := v.(map[string]interface{}); ok {
					tmp, _ := json.Marshal(v)
					attr.Annotations[k] = string(tmp)
				} else {
					attr.Annotations[k] = v
				}

			}
		}
	}
	return &realmUserProfile, nil
}
