package provider

import (
	"context"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

// a struct that represents the "desired" state configured via terraform
// the key for 'clientRoles' is keycloak's client-id (the uuid, not to be confused with the OAuth Client Id)
type roleMapping struct {
	clientRoles map[string][]*keycloak.Role
	realmRoles  []*keycloak.Role
}

// transform the keycloak response RoleMapping into the internal roleMapping
// so we can compare the current state (keycloak) against the desired state (terraform)
func intoRoleMapping(keycloakRoleMapping *keycloak.RoleMapping) *roleMapping {
	clientRoles := make(map[string][]*keycloak.Role)
	for _, clientRoleMapping := range keycloakRoleMapping.ClientMappings {
		clientRoles[clientRoleMapping.Id] = clientRoleMapping.Mappings
	}

	mapping := roleMapping{
		clientRoles: clientRoles,
		realmRoles:  keycloakRoleMapping.RealmMappings,
	}

	return &mapping
}

// given a list of roleIds, query keycloak for role details to find out if a role is a client role or a
// realm role (which is required to POST the role assignment to the correct API endpoint)
func getExtendedRoleMapping(ctx context.Context, keycloakClient *keycloak.KeycloakClient, realmId string, roleIds []string) (*roleMapping, error) {
	clientRoles := make(map[string][]*keycloak.Role)
	var realmRoles []*keycloak.Role

	for _, roleId := range roleIds {
		role, err := keycloakClient.GetRole(ctx, realmId, roleId)
		if err != nil {
			// if the role doesn't exist anymore, skip it
			if keycloak.ErrorIs404(err) {
				continue
			}

			return nil, err
		}

		if role.ClientRole {
			clientRoles[role.ClientId] = append(clientRoles[role.ClientId], role)
		} else {
			realmRoles = append(realmRoles, role)
		}
	}

	mapping := roleMapping{
		clientRoles: clientRoles,
		realmRoles:  realmRoles,
	}

	return &mapping, nil
}

type terraformRoleMappingUpdates struct {
	realmRolesToRemove  []*keycloak.Role
	realmRolesToAdd     []*keycloak.Role
	clientRolesToRemove map[string][]*keycloak.Role
	clientRolesToAdd    map[string][]*keycloak.Role
}

// given the existing roles (queried from keycloak) and the requested roles (via tf)
// calculate the required updates, i.e. roles to remove and roles to add
func calculateRoleMappingUpdates(requestedRoles *roleMapping, existingRoles *roleMapping) *terraformRoleMappingUpdates {
	clientRolesToRemove := make(map[string][]*keycloak.Role)
	clientRolesToAdd := make(map[string][]*keycloak.Role)

	realmRolesToRemove := minusRoles(existingRoles.realmRoles, requestedRoles.realmRoles)
	realmRolesToAdd := minusRoles(requestedRoles.realmRoles, existingRoles.realmRoles)

	for clientId, requestedClientRoles := range requestedRoles.clientRoles {
		if existingClientRoles, ok := existingRoles.clientRoles[clientId]; ok {
			clientRolesToAdd[clientId] = minusRoles(requestedClientRoles, existingClientRoles)
			clientRolesToRemove[clientId] = minusRoles(existingClientRoles, requestedClientRoles)
		} else {
			// if no roles for this client exist yet, then of course, all requested roles need to be created
			clientRolesToAdd[clientId] = requestedClientRoles
		}
	}

	// now check all existing roles, if there are even any roles configured for each client
	for clientId, existingClientRoles := range existingRoles.clientRoles {
		if _, ok := requestedRoles.clientRoles[clientId]; !ok {
			// no role requested for this client? -> remove all existing client-roles
			clientRolesToRemove[clientId] = existingClientRoles
		}
	}

	updates := terraformRoleMappingUpdates{
		realmRolesToRemove:  realmRolesToRemove,
		realmRolesToAdd:     realmRolesToAdd,
		clientRolesToRemove: clientRolesToRemove,
		clientRolesToAdd:    clientRolesToAdd,
	}

	return &updates
}

// check if given role exists in a list of roles
func roleExists(roles []*keycloak.Role, role *keycloak.Role) bool {
	for _, r := range roles {
		if r.Id == role.Id {
			return true
		}
	}

	return false
}

// calculate the set difference: returns `a \ b`, i.e. every role that exist in a, but not in b
func minusRoles(a, b []*keycloak.Role) []*keycloak.Role {
	var aWithoutB []*keycloak.Role

	for _, role := range a {
		if !roleExists(b, role) {
			aWithoutB = append(aWithoutB, role)
		}
	}

	return aWithoutB
}
