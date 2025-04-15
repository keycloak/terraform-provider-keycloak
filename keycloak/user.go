package keycloak

import (
	"context"
	"fmt"
)

type FederatedIdentity struct {
	IdentityProvider string `json:"identityProvider"`
	UserId           string `json:"userId"`
	UserName         string `json:"userName"`
}

type FederatedIdentities []*FederatedIdentity

type User struct {
	Id      string `json:"id,omitempty"`
	RealmId string `json:"-"`

	Username            string              `json:"username"`
	Email               string              `json:"email"`
	EmailVerified       bool                `json:"emailVerified"`
	FirstName           string              `json:"firstName"`
	LastName            string              `json:"lastName"`
	Enabled             bool                `json:"enabled"`
	Attributes          map[string][]string `json:"attributes"`
	FederatedIdentities FederatedIdentities `json:"federatedIdentities"`
	RequiredActions     []string            `json:"requiredActions"`
}

type PasswordCredentials struct {
	Value     string `json:"value"`
	Type      string `json:"type"`
	Temporary bool   `json:"temporary"`
}

func (keycloakClient *KeycloakClient) NewUser(ctx context.Context, user *User) error {
	newUser := User{
		Id:              user.Id,
		RealmId:         user.RealmId,
		Username:        user.Username,
		Email:           user.Email,
		EmailVerified:   user.EmailVerified,
		FirstName:       user.FirstName,
		LastName:        user.LastName,
		Enabled:         user.Enabled,
		Attributes:      user.Attributes,
		RequiredActions: user.RequiredActions,
	}
	_, location, err := keycloakClient.post(ctx, fmt.Sprintf("/realms/%s/users", user.RealmId), newUser)
	if err != nil {
		return err
	}

	user.Id = getIdFromLocationHeader(location)

	for _, federatedIdentity := range user.FederatedIdentities {
		_, _, err := keycloakClient.post(ctx, fmt.Sprintf("/realms/%s/users/%s/federated-identity/%s", user.RealmId, user.Id, federatedIdentity.IdentityProvider), federatedIdentity)
		if err != nil {
			return err
		}
	}

	return nil
}

func (keycloakClient *KeycloakClient) ResetUserPassword(ctx context.Context, realmId, userId string, newPassword string, isTemporary bool) error {
	resetCredentials := &PasswordCredentials{
		Value:     newPassword,
		Type:      "password",
		Temporary: isTemporary,
	}

	err := keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/users/%s/reset-password", realmId, userId), resetCredentials)
	if err != nil {
		return err
	}
	return nil
}

func (keycloakClient *KeycloakClient) GetUsers(ctx context.Context, realmId string) ([]*User, error) {
	var users []*User

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/users", realmId), &users, nil)
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		user.RealmId = realmId
	}

	return users, nil
}

func (keycloakClient *KeycloakClient) GetUser(ctx context.Context, realmId, id string) (*User, error) {
	var user User

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/users/%s", realmId, id), &user, nil)
	if err != nil {
		return nil, err
	}

	user.RealmId = realmId

	return &user, nil
}

func (keycloakClient *KeycloakClient) UpdateUser(ctx context.Context, user *User) error {
	err := keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/users/%s", user.RealmId, user.Id), user)
	if err != nil {
		return err
	}

	var federatedIdentities []*FederatedIdentity
	err = keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/users/%s/federated-identity", user.RealmId, user.Id), &federatedIdentities, nil)
	if err != nil {
		return err
	}

	for _, federatedIdentity := range federatedIdentities {
		keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/users/%s/federated-identity/%s", user.RealmId, user.Id, federatedIdentity.IdentityProvider), nil)
	}

	for _, federatedIdentity := range user.FederatedIdentities {
		_, _, err := keycloakClient.post(ctx, fmt.Sprintf("/realms/%s/users/%s/federated-identity/%s", user.RealmId, user.Id, federatedIdentity.IdentityProvider), federatedIdentity)
		if err != nil {
			return err
		}
	}

	return nil
}

func (keycloakClient *KeycloakClient) DeleteUser(ctx context.Context, realmId, id string) error {
	return keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/users/%s", realmId, id), nil)
}

func (keycloakClient *KeycloakClient) GetUserByUsername(ctx context.Context, realmId, username string) (*User, error) {
	var users []*User

	params := map[string]string{
		"username": escapeBackslashes(username),
	}

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/users", realmId), &users, params)
	if err != nil {
		return nil, err
	}

	// more than one user could be returned so we need to search through all results and return the correct one
	// ex: foo and foo-user could both exist, but searching for "foo" will return both
	for _, user := range users {
		if user.Username == username {
			user.RealmId = realmId

			return user, nil
		}
	}

	// the requested user does not exist
	// we shouldn't raise an error here since it will be difficult to differentiate between a non-existent user and a network error
	return nil, nil
}

func (keycloakClient *KeycloakClient) GetUserGroups(ctx context.Context, realmId, userId string) ([]*Group, error) {
	var groups []*Group
	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/users/%s/groups/", realmId, userId), &groups, nil)

	if err != nil {
		return nil, err
	}

	return groups, nil
}

func (keycloakClient *KeycloakClient) addUserToGroup(ctx context.Context, user *User, groupId string) error {
	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/users/%s/groups/%s", user.RealmId, user.Id, groupId), nil)
}

func (keycloakClient *KeycloakClient) AddUsersToGroup(ctx context.Context, realmId, groupId string, users []interface{}) error {
	for _, username := range users {
		user, err := keycloakClient.GetUserByUsername(ctx, realmId, username.(string)) // we need the user's id in order to add them to a group
		if err != nil {
			return err
		}
		if user == nil {
			return fmt.Errorf("user with username %s does not exist", username.(string))
		}

		err = keycloakClient.addUserToGroup(ctx, user, groupId)
		if err != nil {
			return err
		}
	}

	return nil
}

func (keycloakClient *KeycloakClient) RemoveUserFromGroup(ctx context.Context, user *User, groupId string) error {
	return keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/users/%s/groups/%s", user.RealmId, user.Id, groupId), nil)
}

func (keycloakClient *KeycloakClient) RemoveUsersFromGroup(ctx context.Context, realmId, groupId string, usernames []interface{}) error {
	for _, username := range usernames {
		user, err := keycloakClient.GetUserByUsername(ctx, realmId, username.(string)) // we need the user's id in order to remove them from a group
		if err != nil {
			return err
		}
		if user == nil {
			return fmt.Errorf("user with username %s does not exist", username.(string))
		}

		err = keycloakClient.RemoveUserFromGroup(ctx, user, groupId)
		if err != nil {
			return err
		}
	}

	return nil
}

func (keycloakClient *KeycloakClient) AddUserToGroups(ctx context.Context, groupIds []string, userId string, realmId string) error {
	for _, groupId := range groupIds {
		var user User
		user.Id = userId
		user.RealmId = realmId
		err := keycloakClient.addUserToGroup(ctx, &user, groupId)

		if err != nil {
			return err
		}
	}
	return nil
}

func (keycloakClient *KeycloakClient) RemoveUserFromGroups(ctx context.Context, groupIds []string, userId string, realmId string) error {
	for _, groupId := range groupIds {
		var user User
		user.Id = userId
		user.RealmId = realmId
		err := keycloakClient.RemoveUserFromGroup(ctx, &user, groupId)

		if err != nil {
			return err
		}
	}
	return nil
}
