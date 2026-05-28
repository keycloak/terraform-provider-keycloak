package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

// checkFGAPv2NotEnabled returns a descriptive error when Fine-Grained Admin
// Permissions v2 is active. resourceType is the Terraform resource name,
// v2Alternative (if non-empty) is the replacement resource to point users to.
func checkFGAPv2NotEnabled(ctx context.Context, keycloakClient *keycloak.KeycloakClient, resourceType, v2Alternative string) (diag.Diagnostics, bool) {
	enabled, err := keycloakClient.FGAPv2IsEnabled(ctx)
	if err != nil {
		return diag.FromErr(err), false
	}
	if !enabled {
		return nil, true
	}

	detail := fmt.Sprintf(
		"%s only works with Fine-Grained Admin Permissions v1 (admin-fine-grained-authz:v1). "+
			"Your Keycloak instance has v2 enabled (ADMIN_FINE_GRAINED_AUTHZ_V2), which uses a different API that is incompatible with this resource.",
		resourceType,
	)
	if v2Alternative != "" {
		detail += fmt.Sprintf(
			"\n\nMigrate to %s, which supports v2. Remove this resource from your state first:\n\n  terraform state rm <resource_address>",
			v2Alternative,
		)
	} else {
		detail += "\n\nNo direct v2 replacement exists yet. Remove this resource from your state with:\n\n  terraform state rm <resource_address>"
	}

	return diag.Diagnostics{{
		Severity: diag.Error,
		Summary:  "Fine-Grained Admin Permissions v2 is not supported by this resource",
		Detail:   detail,
	}}, false
}

func checkFGAPv2NotEnabledForImport(ctx context.Context, keycloakClient *keycloak.KeycloakClient, resourceType, v2Alternative string) error {
	diags, ok := checkFGAPv2NotEnabled(ctx, keycloakClient, resourceType, v2Alternative)
	if !ok {
		return fmt.Errorf("%s: %s", diags[0].Summary, diags[0].Detail)
	}
	return nil
}

// --- v1 helpers (realm-management, pre-created permissions) ---

// scopePermissionsSchema returns the TypeSet schema for scope-permission blocks
// used by v1 resources. This matches the original schema and must not change to
// avoid breaking existing state for v1 users.
func scopePermissionsSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"policies": {
					Type:     schema.TypeSet,
					Elem:     &schema.Schema{Type: schema.TypeString},
					Optional: true,
				},
				"description": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"decision_strategy": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringInSlice(keycloakOpenidClientResourcePermissionDecisionStrategies, false),
				},
			},
		},
	}
}

func setOpenidClientScopePermissionPolicy(ctx context.Context, keycloakClient *keycloak.KeycloakClient, realmId string, realmManagementClientId string, authorizationPermissionId string, scopeDataSet *schema.Set) error {
	var policies []string

	scopePermission := scopeDataSet.List()[0].(map[string]interface{})

	if v, ok := scopePermission["policies"]; ok {
		for _, policy := range v.(*schema.Set).List() {
			policies = append(policies, policy.(string))
		}
	}

	permission, err := keycloakClient.GetOpenidClientAuthorizationPermission(ctx, realmId, realmManagementClientId, authorizationPermissionId)
	if err != nil {
		return err
	}

	if v, ok := scopePermission["description"]; ok {
		permission.Description = v.(string)
	}

	if v, ok := scopePermission["decision_strategy"]; ok {
		permission.DecisionStrategy = v.(string)
	}

	permission.Policies = policies

	return keycloakClient.UpdateOpenidClientAuthorizationPermission(ctx, permission)
}

func getOpenidClientScopePermissionPolicy(ctx context.Context, keycloakClient *keycloak.KeycloakClient, realmId, realmManagementClientId, permissionId string) (map[string]interface{}, error) {
	permission, err := keycloakClient.GetOpenidClientAuthorizationPermission(ctx, realmId, realmManagementClientId, permissionId)
	if err != nil {
		return nil, err
	}

	if permission.Description == "" && permission.DecisionStrategy == "UNANIMOUS" && len(permission.Policies) == 0 {
		return nil, nil
	}

	permissionViewSettings := make(map[string]interface{})

	if permission.Description != "" {
		permissionViewSettings["description"] = permission.Description
	}

	if permission.DecisionStrategy != "" {
		permissionViewSettings["decision_strategy"] = permission.DecisionStrategy
	}

	if len(permission.Policies) > 0 {
		permissionViewSettings["policies"] = permission.Policies
	}

	return permissionViewSettings, nil
}

// --- v2 helpers (admin-permissions client, provider-managed permissions) ---

// createOrAdoptFGAPv2Permission creates a new scope permission, or adopts an
// existing one with the same name (useful after import or if the permission was
// created outside Terraform). The perm.Id field is set on the struct after return.
func createOrAdoptFGAPv2Permission(ctx context.Context, kc *keycloak.KeycloakClient,
	perm *keycloak.OpenidClientAuthorizationPermission) (string, error) {

	existing, err := kc.FindFGAPv2PermissionByName(ctx, perm.RealmId, perm.ResourceServerId, perm.Name)
	if err != nil {
		return "", err
	}
	if existing != nil {
		perm.Id = existing.Id
		if err := kc.UpdateOpenidClientAuthorizationPermission(ctx, perm); err != nil {
			return "", err
		}
		return existing.Id, nil
	}
	if err := kc.NewOpenidClientAuthorizationPermission(ctx, perm); err != nil {
		return "", err
	}
	return perm.Id, nil
}

// readFGAPv2ScopePermission reads a v2 permission by its stored ID using the
// FGAPv2-specific read path (scope names, no resource UUID round-trip).
// Returns nil (not an error) if the permission no longer exists.
func readFGAPv2ScopePermission(ctx context.Context, keycloakClient *keycloak.KeycloakClient,
	realmId, apClientId, permissionId string) (*keycloak.OpenidClientAuthorizationPermission, error) {

	if permissionId == "" {
		return nil, nil
	}
	perm, err := keycloakClient.GetFGAPv2Permission(ctx, realmId, apClientId, permissionId)
	if err != nil {
		if keycloak.ErrorIs404(err) {
			return nil, nil
		}
		return nil, err
	}
	return perm, nil
}

// deleteFGAPv2Permission deletes a permission by its stored ID.
// A missing permission is not treated as an error.
func deleteFGAPv2Permission(ctx context.Context, keycloakClient *keycloak.KeycloakClient,
	realmId, apClientId, permissionId string) error {

	if permissionId == "" {
		return nil
	}
	err := keycloakClient.DeleteOpenidClientAuthorizationPermission(ctx, realmId, apClientId, permissionId)
	if err != nil && keycloak.ErrorIs404(err) {
		return nil
	}
	return err
}

// setToStringSlice converts a TypeSet value to a []string, always returning a
// non-nil slice so JSON serialisation produces [] rather than null.
func setToStringSlice(s *schema.Set) []string {
	result := make([]string, 0, s.Len())
	for _, v := range s.List() {
		result = append(result, v.(string))
	}
	return result
}

// stringsToInterfaces converts []string to []interface{} for use with schema.NewSet.
func stringsToInterfaces(ss []string) []interface{} {
	result := make([]interface{}, len(ss))
	for i, s := range ss {
		result[i] = s
	}
	return result
}
