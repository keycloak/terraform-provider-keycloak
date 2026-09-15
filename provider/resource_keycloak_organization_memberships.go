package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakOrganizationMemberships() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakOrganizationMembershipsCreate,
		ReadContext:   resourceKeycloakOrganizationMembershipsRead,
		DeleteContext: resourceKeycloakOrganizationMembershipsDelete,
		UpdateContext: resourceKeycloakOrganizationMembershipsUpdate,
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The realm this organization exists in.",
			},
			"organization_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the organization to manage members for.",
			},
			"members": {
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Description: "The set of usernames to assign to the organization.",
			},
		},
	}
}

func organizationMembershipsId(realmId, organizationId string) string {
	return fmt.Sprintf("%s/organization-memberships/%s", realmId, organizationId)
}

func reconcileOrganizationMemberships(ctx context.Context, realmId, organizationId string, targetMembers []interface{}, keycloakClient *keycloak.KeycloakClient) error {
	// 1. Resolve all desired usernames to user IDs and check existence
	desiredUsersMap := make(map[string]string) // userId -> username
	desiredUserIdsSet := make(map[string]bool) // userId -> true
	for _, rawUsername := range targetMembers {
		username := rawUsername.(string)
		user, err := keycloakClient.GetUserByUsername(ctx, realmId, username)
		if err != nil {
			return err
		}
		if user == nil {
			return fmt.Errorf("user with username %s does not exist", username)
		}
		desiredUsersMap[user.Id] = username
		desiredUserIdsSet[user.Id] = true
	}

	// 2. Fetch current members from Keycloak
	currentMembers, err := keycloakClient.GetOrganizationMembers(ctx, realmId, organizationId)
	if err != nil {
		return err
	}

	currentUsersMap := make(map[string]*keycloak.OrganizationMember) // userId -> Member
	currentUserIdsSet := make(map[string]bool)                       // userId -> true
	for _, member := range currentMembers {
		currentUsersMap[member.Id] = member
		currentUserIdsSet[member.Id] = true
	}

	// 3. Preflight check: verify no undesired members are MANAGED before issuing any delete requests
	var usersToRemove []*keycloak.OrganizationMember
	for currentUserId, member := range currentUsersMap {
		if !desiredUserIdsSet[currentUserId] {
			if member.MembershipType == "MANAGED" {
				return fmt.Errorf("cannot remove managed member %s from organization %s: Keycloak deletes the underlying user account for managed memberships. Please remove the membership through the Identity Provider federation or Keycloak admin console instead", member.Username, organizationId)
			}
			usersToRemove = append(usersToRemove, member)
		}
	}

	// 4. Remove users that are currently members but not desired
	for _, member := range usersToRemove {
		err = keycloakClient.RemoveUserFromOrganization(ctx, realmId, organizationId, member.Id)
		if err != nil && !keycloak.ErrorIs404(err) {
			return fmt.Errorf("error removing user %s from organization: %w", member.Username, err)
		}
	}

	// 5. Add users that are desired but not currently members
	for desiredUserId, desiredUsername := range desiredUsersMap {
		if !currentUserIdsSet[desiredUserId] {
			err = keycloakClient.AddUserToOrganization(ctx, realmId, organizationId, desiredUserId)
			if err != nil {
				return fmt.Errorf("error adding user %s to organization: %w", desiredUsername, err)
			}
		}
	}

	return nil
}

func resourceKeycloakOrganizationMembershipsCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	organizationId := data.Get("organization_id").(string)
	members := data.Get("members").(*schema.Set).List()

	if err := reconcileOrganizationMemberships(ctx, realmId, organizationId, members, keycloakClient); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(organizationMembershipsId(realmId, organizationId))

	return resourceKeycloakOrganizationMembershipsRead(ctx, data, meta)
}

func resourceKeycloakOrganizationMembershipsRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	organizationId := data.Get("organization_id").(string)

	keycloakMembers, err := keycloakClient.GetOrganizationMembers(ctx, realmId, organizationId)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	var members []string
	tfMembers := data.Get("members").(*schema.Set)
	for _, member := range keycloakMembers {
		if member.MembershipType == "UNMANAGED" || tfMembers.Contains(member.Username) {
			members = append(members, member.Username)
		}
	}

	data.Set("members", members)
	data.SetId(organizationMembershipsId(realmId, organizationId))

	return nil
}

func resourceKeycloakOrganizationMembershipsUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	organizationId := data.Get("organization_id").(string)
	members := data.Get("members").(*schema.Set).List()

	if err := reconcileOrganizationMemberships(ctx, realmId, organizationId, members, keycloakClient); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(organizationMembershipsId(realmId, organizationId))

	return resourceKeycloakOrganizationMembershipsRead(ctx, data, meta)
}

func resourceKeycloakOrganizationMembershipsDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	organizationId := data.Get("organization_id").(string)
	tfMembers := data.Get("members").(*schema.Set)

	// Fetch current members to retain their membership representations (including membershipType)
	currentMembers, err := keycloakClient.GetOrganizationMembers(ctx, realmId, organizationId)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	for _, member := range currentMembers {
		if !tfMembers.Contains(member.Username) {
			continue
		}

		// To prevent deleting the Keycloak user account, do not issue remove requests for MANAGED members.
		if member.MembershipType == "MANAGED" {
			continue
		}

		if err = keycloakClient.RemoveUserFromOrganization(ctx, realmId, organizationId, member.Id); err != nil && !keycloak.ErrorIs404(err) {
			return handleNotFoundError(ctx, err, data)
		}
	}

	return nil
}
