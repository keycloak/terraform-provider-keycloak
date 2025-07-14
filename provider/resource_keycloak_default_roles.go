package provider

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakDefaultRoles() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakDefaultRolesReconcile,
		ReadContext:   resourceKeycloakDefaultRolesRead,
		DeleteContext: resourceKeycloakDefaultRolesDelete,
		UpdateContext: resourceKeycloakDefaultRolesReconcile,
		Importer: &schema.ResourceImporter{
			StateContext: resourceKeycloakDefaultRolesImport,
		},
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"default_roles": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Realm level roles (name) assigned to new users.",
				Required:    true,
			},
		},
	}
}

func getDefaultRolesFromData(data *schema.ResourceData) *keycloak.DefaultRoles {
	var defaultRolesList []string
	if v, ok := data.GetOk("default_roles"); ok {
		for _, defaultRole := range v.(*schema.Set).List() {
			defaultRolesList = append(defaultRolesList, defaultRole.(string))
		}
	}
	defaultRoles := &keycloak.DefaultRoles{
		Id:           data.Id(),
		RealmId:      data.Get("realm_id").(string),
		DefaultRoles: defaultRolesList,
	}
	return defaultRoles
}

func resourceKeycloakDefaultRolesRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	composites, err := keycloakClient.GetDefaultRoles(ctx, realmId, id)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	var defaultRoleNames []string
	for _, composite := range composites {
		name, err := keycloakClient.GetQualifiedRoleName(ctx, realmId, composite)
		if err != nil {
			return handleNotFoundError(ctx, err, data)
		}
		defaultRoleNames = append(defaultRoleNames, name)
	}

	data.Set("default_roles", defaultRoleNames)
	return nil
}

func resourceKeycloakDefaultRolesReconcile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	if ok, err := keycloakClient.VersionIsGreaterThanOrEqualTo(ctx, keycloak.Version_13); !ok && err == nil {
		return diag.Diagnostics{{
			Severity: diag.Error,
			Summary:  "this resource requires Keycloak v13 or higher",
		}}
	} else if err != nil {
		return diag.FromErr(err)
	}

	local := getDefaultRolesFromData(data)
	localDefaultRoles := make(map[string]struct{}, len(local.DefaultRoles))
	for _, defaultRole := range local.DefaultRoles {
		localDefaultRoles[defaultRole] = struct{}{}
	}

	realm, err := keycloakClient.GetRealm(ctx, local.RealmId)
	if err != nil {
		return diag.FromErr(err)
	}
	if realm == nil {
		return diag.Diagnostics{{
			Severity: diag.Error,
			Summary:  "realm not found: " + local.RealmId,
		}}
	}
	if realm.DefaultRole == nil || realm.DefaultRole.Id == "" {
		return diag.Diagnostics{{
			Severity: diag.Error,
			Summary:  "realm does not have a default role",
		}}
	}

	data.SetId(realm.DefaultRole.Id)

	composites, err := keycloakClient.GetDefaultRoles(ctx, local.RealmId, realm.DefaultRole.Id)
	if err != nil {
		return diag.FromErr(err)
	}

	currentDefaultRoles := make([]string, len(composites))
	defaultRolesMap := make(map[string]*keycloak.Role, len(composites))

	for i, composite := range composites {
		name, err := keycloakClient.GetQualifiedRoleName(ctx, local.RealmId, composite)
		if err != nil {
			return diag.FromErr(err)
		}
		currentDefaultRoles[i] = name
		defaultRolesMap[name] = composite
	}

	// skip if actual default roles in keycloak same as we want
	if reflect.DeepEqual(currentDefaultRoles, local.DefaultRoles) {
		return nil
	}

	getRole := func(roleName string) (*keycloak.Role, error) {
		var clientId string
		if parts := strings.Split(roleName, "/"); len(parts) == 2 {
			client, err := keycloakClient.GetGenericClientByClientId(ctx, local.RealmId, parts[0])
			if err != nil {
				return nil, err
			}
			clientId, roleName = client.Id, parts[1]
		}
		return keycloakClient.GetRoleByName(ctx, local.RealmId, clientId, roleName)
	}

	var putList, deleteList []*keycloak.Role

	for _, roleName := range local.DefaultRoles {
		if _, ok := defaultRolesMap[roleName]; !ok {
			role, err := getRole(roleName)
			if err != nil {
				return diag.FromErr(err)
			}
			putList = append(putList, role)
		}
	}

	for _, roleName := range currentDefaultRoles {
		if _, ok := localDefaultRoles[roleName]; !ok {
			deleteList = append(deleteList, defaultRolesMap[roleName])
		}
	}

	// apply if not empty
	if len(putList) > 0 {
		role := &keycloak.Role{
			RealmId: local.RealmId,
			Id:      realm.DefaultRole.Id,
		}
		err := keycloakClient.AddCompositesToRole(ctx, role, putList)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if len(deleteList) > 0 {
		role := &keycloak.Role{
			RealmId: local.RealmId,
			Id:      realm.DefaultRole.Id,
		}
		err := keycloakClient.RemoveCompositesFromRole(ctx, role, deleteList)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceKeycloakDefaultRolesRead(ctx, data, meta)
}

// remove all roles from default
func resourceKeycloakDefaultRolesDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	realm, err := keycloakClient.GetRealm(ctx, realmId)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	defaultRoles, err := keycloakClient.GetDefaultRoles(ctx, realmId, realm.DefaultRole.Id)
	if err != nil {
		return diag.FromErr(err)
	}

	if len(defaultRoles) > 0 {
		role := &keycloak.Role{
			RealmId: realmId,
			Id:      realm.DefaultRole.Id,
		}
		err := keycloakClient.RemoveCompositesFromRole(ctx, role, defaultRoles)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceKeycloakDefaultRolesImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid import, supported import format: {{realm}}/{{defaultRoleId}}")
	}

	_, err := keycloakClient.GetDefaultRoles(ctx, parts[0], parts[1])
	if err != nil {
		return nil, err
	}

	d.Set("realm_id", parts[0])
	d.SetId(parts[1])

	diagnostics := resourceKeycloakDefaultRolesRead(ctx, d, meta)
	if diagnostics.HasError() {
		return nil, errors.New(diagnostics[0].Summary)
	}

	return []*schema.ResourceData{d}, nil
}
