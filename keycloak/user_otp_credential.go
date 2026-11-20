package keycloak

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
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

// CreateUserOtpCredential adds an OTP credential to the user.
//
// Keycloak only supports creating an OTP credential from a secret through the user update endpoint. Only the
// credential is sent: Keycloak leaves every user field that is missing from the request as it is, so the user is
// not read first and nothing else on the user, such as its profile, attributes or required actions, is
// overwritten or can be overwritten by a concurrent change.
func (keycloakClient *KeycloakClient) CreateUserOtpCredential(ctx context.Context, credential *UserOtpCredential) error {
	userPath := fmt.Sprintf("/realms/%s/users/%s", credential.RealmId, credential.UserId)

	// The new credential's ID is found by diffing the credential list around the PUT, so concurrent
	// creations for the same user must not interleave.
	lockKey := fmt.Sprintf("CreateUserOtpCredential:%s:%s", credential.RealmId, credential.UserId)
	keycloakClient.Mutex.Lock(lockKey)
	defer keycloakClient.Mutex.Unlock(lockKey)

	existingCredentials, err := keycloakClient.getUserCredentials(ctx, credential.RealmId, credential.UserId)
	if err != nil {
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

	// Keycloak counts Base32 padding when it sizes the decoded key and pads it with zero bytes, which
	// changes the key, so the padding must not be sent. The case is normalised the way the provider validated it.
	secretData, err := json.Marshal(otpSecretData{Value: strings.ToUpper(strings.TrimRight(credential.Secret, "="))})
	if err != nil {
		return err
	}

	user := map[string]interface{}{
		"credentials": []map[string]interface{}{
			{
				"type":           "otp",
				"userLabel":      credential.UserLabel,
				"credentialData": string(credentialData),
				"secretData":     string(secretData),
			},
		},
	}

	// the request body contains the TOTP secret
	if err := keycloakClient.put(withRedactedRequestBody(ctx), userPath, user); err != nil {
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

	var createdIds []string
	for _, created := range credentials {
		if created.Type == "otp" && !knownIds[created.Id] {
			createdIds = append(createdIds, created.Id)
		}
	}

	switch len(createdIds) {
	case 1:
		credential.Id = createdIds[0]
		return nil
	case 0:
		return fmt.Errorf("created an OTP credential for user %s, but it was not found afterwards", credential.UserId)
	default:
		// Another client added an OTP credential to the same user at the same time, so there is no telling which one is ours.
		return fmt.Errorf("created an OTP credential for user %s, but %d new OTP credentials appeared and it is unclear which one belongs to this resource; remove the unwanted ones and apply again", credential.UserId, len(createdIds))
	}
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
