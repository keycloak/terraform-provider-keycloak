package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"strings"
)

func resourceKeycloakUserRoles() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakUserRolesReconcile,
		ReadContext:   resourceKeycloakUserRolesRead,
		UpdateContext: resourceKeycloakUserRolesReconcile,
		DeleteContext: resourceKeycloakUserRolesDelete,
		// This resource can be imported using {{realm}}/{{userId}}.
		Importer: &schema.ResourceImporter{
			StateContext: resourceKeycloakUserRolesImport,
		},
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"user_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"role_ids": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Required: true,
			},
			"exhaustive": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func userRolesId(realmId, userId string) string {
	return fmt.Sprintf("%s/%s", realmId, userId)
}

func addRolesToUser(ctx context.Context, keycloakClient *keycloak.KeycloakClient, clientRolesToAdd map[string][]*keycloak.Role, realmRolesToAdd []*keycloak.Role, user *keycloak.User) error {
	if len(realmRolesToAdd) != 0 {
		err := keycloakClient.AddRealmRolesToUser(ctx, user.RealmId, user.Id, realmRolesToAdd)
		if err != nil {
			return err
		}
	}

	for k, roles := range clientRolesToAdd {
		if len(roles) != 0 {
			err := keycloakClient.AddClientRolesToUser(ctx, user.RealmId, user.Id, k, roles)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func removeRolesFromUser(ctx context.Context, keycloakClient *keycloak.KeycloakClient, clientRolesToRemove map[string][]*keycloak.Role, realmRolesToRemove []*keycloak.Role, user *keycloak.User) error {
	if len(realmRolesToRemove) != 0 {
		err := keycloakClient.RemoveRealmRolesFromUser(ctx, user.RealmId, user.Id, realmRolesToRemove)
		if err != nil {
			return err
		}
	}

	for k, roles := range clientRolesToRemove {
		if len(roles) != 0 {
			err := keycloakClient.RemoveClientRolesFromUser(ctx, user.RealmId, user.Id, k, roles)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func resourceKeycloakUserRolesReconcile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	userId := data.Get("user_id").(string)
	roleIds := interfaceSliceToStringSlice(data.Get("role_ids").(*schema.Set).List())
	exhaustive := data.Get("exhaustive").(bool)

	user, err := keycloakClient.GetUser(ctx, realmId, userId)
	if err != nil {
		return diag.FromErr(err)
	}

	if data.HasChange("role_ids") && !data.IsNewResource() {
		o, n := data.GetChange("role_ids")
		os := o.(*schema.Set)
		ns := n.(*schema.Set)
		remove := interfaceSliceToStringSlice(os.Difference(ns).List())

		tfRolesToRemove, err := getExtendedRoleMapping(ctx, keycloakClient, realmId, remove)
		if err != nil {
			return diag.FromErr(err)
		}

		if err = removeRolesFromUser(ctx, keycloakClient, tfRolesToRemove.clientRoles, tfRolesToRemove.realmRoles, user); err != nil {
			return diag.FromErr(err)
		}
	}

	tfRoles, err := getExtendedRoleMapping(ctx, keycloakClient, realmId, roleIds)
	if err != nil {
		return diag.FromErr(err)
	}

	// get the list of currently assigned roles. Due to default realm and client roles
	// (e.g. roles of the account client) this is probably not empty upon resource creation
	roleMappings, err := keycloakClient.GetUserRoleMappings(ctx, realmId, userId)

	// sort into roles we need to add and roles we need to remove
	updates := calculateRoleMappingUpdates(tfRoles, intoRoleMapping(roleMappings))

	// add roles
	err = addRolesToUser(ctx, keycloakClient, updates.clientRolesToAdd, updates.realmRolesToAdd, user)
	if err != nil {
		return diag.FromErr(err)
	}

	// remove roles if exhaustive (authoritative)
	if exhaustive {
		err = removeRolesFromUser(ctx, keycloakClient, updates.clientRolesToRemove, updates.realmRolesToRemove, user)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	data.SetId(userRolesId(realmId, userId))
	return resourceKeycloakUserRolesRead(ctx, data, meta)
}

func resourceKeycloakUserRolesRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	userId := data.Get("user_id").(string)
	sortedRoleIds := interfaceSliceToStringSlice(data.Get("role_ids").(*schema.Set).List())
	exhaustive := data.Get("exhaustive").(bool)

	// check if user exists, remove from state if not found
	if _, err := keycloakClient.GetUser(ctx, realmId, userId); err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	roles, err := keycloakClient.GetUserRoleMappings(ctx, realmId, userId)
	if err != nil {
		return diag.FromErr(err)
	}

	var roleIds []string

	for _, realmRole := range roles.RealmMappings {
		if exhaustive || stringSliceContains(sortedRoleIds, realmRole.Id) {
			roleIds = append(roleIds, realmRole.Id)
		}
	}

	for _, clientRoleMapping := range roles.ClientMappings {
		for _, clientRole := range clientRoleMapping.Mappings {
			if exhaustive || stringSliceContains(sortedRoleIds, clientRole.Id) {
				roleIds = append(roleIds, clientRole.Id)
			}
		}
	}

	data.Set("role_ids", roleIds)
	data.SetId(userRolesId(realmId, userId))

	return nil
}

func resourceKeycloakUserRolesDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	userId := data.Get("user_id").(string)

	user, err := keycloakClient.GetUser(ctx, realmId, userId)

	roleIds := interfaceSliceToStringSlice(data.Get("role_ids").(*schema.Set).List())
	rolesToRemove, err := getExtendedRoleMapping(ctx, keycloakClient, realmId, roleIds)
	if err != nil {
		return diag.FromErr(err)
	}

	err = removeRolesFromUser(ctx, keycloakClient, rolesToRemove.clientRoles, rolesToRemove.realmRoles, user)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakUserRolesImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("Invalid import. Supported import format: {{realm}}/{{userId}}.")
	}

	realmId := parts[0]
	userId := parts[1]

	if _, err := keycloakClient.GetUser(ctx, realmId, userId); err != nil {
		return nil, err
	}

	_, err := keycloakClient.GetUserRoleMappings(ctx, realmId, userId)
	if err != nil {
		return nil, err
	}

	d.Set("realm_id", realmId)
	d.Set("user_id", userId)
	d.Set("exhaustive", true)

	d.SetId(userRolesId(realmId, userId))

	diagnostics := resourceKeycloakUserRolesRead(ctx, d, meta)
	if diagnostics.HasError() {
		return nil, errors.New(diagnostics[0].Summary)
	}

	return []*schema.ResourceData{d}, nil
}
