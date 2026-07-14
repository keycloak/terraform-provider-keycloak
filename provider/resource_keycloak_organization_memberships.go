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

func resourceKeycloakOrganizationMembershipsCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	organizationId := data.Get("organization_id").(string)
	members := data.Get("members").(*schema.Set).List()

	for _, username := range members {
		user, err := keycloakClient.GetUserByUsername(ctx, realmId, username.(string))
		if err != nil {
			return diag.FromErr(err)
		}
		if user == nil {
			return diag.FromErr(fmt.Errorf("user with username %s does not exist", username.(string)))
		}

		if err = keycloakClient.AddUserToOrganization(ctx, realmId, organizationId, user.Id); err != nil {
			return diag.FromErr(err)
		}
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
	for _, member := range keycloakMembers {
		members = append(members, member.Username)
	}

	data.Set("members", members)
	data.SetId(organizationMembershipsId(realmId, organizationId))

	return nil
}

func resourceKeycloakOrganizationMembershipsUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	organizationId := data.Get("organization_id").(string)
	tfMembers := data.Get("members").(*schema.Set)

	keycloakMembers, err := keycloakClient.GetOrganizationMembers(ctx, realmId, organizationId)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, keycloakMember := range keycloakMembers {
		if tfMembers.Contains(keycloakMember.Username) {
			// user exists in both Keycloak and Terraform state – nothing to do,
			// remove from set so we can identify users that need to be added later
			tfMembers.Remove(keycloakMember.Username)
		} else {
			// user exists in Keycloak but not in Terraform state – remove from org
			if err = keycloakClient.RemoveUserFromOrganization(ctx, realmId, organizationId, keycloakMember.Id); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	// tfMembers now contains only usernames that need to be added
	for _, username := range tfMembers.List() {
		user, err := keycloakClient.GetUserByUsername(ctx, realmId, username.(string))
		if err != nil {
			return diag.FromErr(err)
		}
		if user == nil {
			return diag.FromErr(fmt.Errorf("user with username %s does not exist", username.(string)))
		}

		if err = keycloakClient.AddUserToOrganization(ctx, realmId, organizationId, user.Id); err != nil {
			return diag.FromErr(err)
		}
	}

	data.SetId(organizationMembershipsId(realmId, organizationId))

	return resourceKeycloakOrganizationMembershipsRead(ctx, data, meta)
}

func resourceKeycloakOrganizationMembershipsDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	organizationId := data.Get("organization_id").(string)
	members := data.Get("members").(*schema.Set).List()

	for _, username := range members {
		user, err := keycloakClient.GetUserByUsername(ctx, realmId, username.(string))
		if err != nil {
			return handleNotFoundError(ctx, err, data)
		}
		if user == nil {
			// user no longer exists; skip
			continue
		}

		if err = keycloakClient.RemoveUserFromOrganization(ctx, realmId, organizationId, user.Id); err != nil {
			return handleNotFoundError(ctx, err, data)
		}
	}

	return nil
}
