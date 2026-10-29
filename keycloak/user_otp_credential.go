package keycloak

import (
	"context"
	"encoding/json"
	"fmt"
)

type UserOtpCredential struct {
	Id        string
	RealmId   string
	UserId    string
	Secret    string
	Algorithm string
	Digits    int
	Period    int
	UserLabel string
}

type userCredentialRepresentation struct {
	Id        string `json:"id"`
	Type      string `json:"type"`
	UserLabel string `json:"userLabel"`
}

// otpCredentialData and otpSecretData are sent as JSON-encoded strings on the wire,
// which is how Keycloak's CredentialRepresentation models them.
type otpCredentialData struct {
	SubType        string `json:"subType"`
	Digits         int    `json:"digits"`
	Counter        int    `json:"counter"`
	Period         int    `json:"period"`
	Algorithm      string `json:"algorithm"`
	SecretEncoding string `json:"secretEncoding"`
}

type otpSecretData struct {
	Value string `json:"value"`
}

func (keycloakClient *KeycloakClient) getUserCredentials(ctx context.Context, realmId, userId string) ([]*userCredentialRepresentation, error) {
	var credentials []*userCredentialRepresentation

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/users/%s/credentials", realmId, userId), &credentials, nil)
	if err != nil {
		return nil, err
	}

	return credentials, nil
}

// CreateUserOtpCredential adds an OTP credential to the user and removes the CONFIGURE_TOTP required action.
//
// Keycloak only supports creating an OTP credential from a secret through the user update endpoint.
// Since that endpoint replaces the user with the given representation, the full representation is
// fetched and sent back, so that fields not known to this provider are left untouched.
func (keycloakClient *KeycloakClient) CreateUserOtpCredential(ctx context.Context, credential *UserOtpCredential) error {
	userPath := fmt.Sprintf("/realms/%s/users/%s", credential.RealmId, credential.UserId)

	existingCredentials, err := keycloakClient.getUserCredentials(ctx, credential.RealmId, credential.UserId)
	if err != nil {
		return err
	}

	user := map[string]interface{}{}
	if err := keycloakClient.get(ctx, userPath, &user, nil); err != nil {
		return err
	}

	credentialData, err := json.Marshal(otpCredentialData{
		SubType:        "totp",
		Digits:         credential.Digits,
		Counter:        0,
		Period:         credential.Period,
		Algorithm:      credential.Algorithm,
		SecretEncoding: "BASE32",
	})
	if err != nil {
		return err
	}

	secretData, err := json.Marshal(otpSecretData{Value: credential.Secret})
	if err != nil {
		return err
	}

	user["credentials"] = []map[string]interface{}{
		{
			"type":           "otp",
			"userLabel":      credential.UserLabel,
			"credentialData": string(credentialData),
			"secretData":     string(secretData),
		},
	}

	requiredActions := []interface{}{}
	if existing, ok := user["requiredActions"].([]interface{}); ok {
		for _, action := range existing {
			if action != "CONFIGURE_TOTP" {
				requiredActions = append(requiredActions, action)
			}
		}
	}
	user["requiredActions"] = requiredActions

	if err := keycloakClient.put(ctx, userPath, user); err != nil {
		return err
	}

	knownIds := map[string]bool{}
	for _, existing := range existingCredentials {
		knownIds[existing.Id] = true
	}

	credentials, err := keycloakClient.getUserCredentials(ctx, credential.RealmId, credential.UserId)
	if err != nil {
		return err
	}

	for _, created := range credentials {
		if created.Type == "otp" && !knownIds[created.Id] {
			credential.Id = created.Id
			return nil
		}
	}

	return fmt.Errorf("created an OTP credential for user %s, but it was not found afterwards", credential.UserId)
}

func (keycloakClient *KeycloakClient) GetUserOtpCredential(ctx context.Context, realmId, userId, id string) (*UserOtpCredential, error) {
	credentials, err := keycloakClient.getUserCredentials(ctx, realmId, userId)
	if err != nil {
		return nil, err
	}

	for _, credential := range credentials {
		if credential.Type == "otp" && credential.Id == id {
			return &UserOtpCredential{
				Id:        credential.Id,
				RealmId:   realmId,
				UserId:    userId,
				UserLabel: credential.UserLabel,
			}, nil
		}
	}

	return nil, &ApiError{
		Code:    404,
		Message: fmt.Sprintf("OTP credential %s not found for user %s", id, userId),
	}
}

func (keycloakClient *KeycloakClient) DeleteUserOtpCredential(ctx context.Context, realmId, userId, id string) error {
	return keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/users/%s/credentials/%s", realmId, userId, id), nil)
}
