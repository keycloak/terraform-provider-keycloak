package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/imdario/mergo"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakRole() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakRoleCreate,
		ReadContext:   resourceKeycloakRoleRead,
		DeleteContext: resourceKeycloakRoleDelete,
		UpdateContext: resourceKeycloakRoleUpdate,
		// This resource can be imported using {{realm}}/{{roleId}}. The role's ID (a GUID) can be found in the URL when viewing the role
		Importer: &schema.ResourceImporter{
			StateContext: resourceKeycloakRoleImport,
		},
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"client_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"composite_roles": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Optional: true,
				Computed: true,
			},
			// misc attributes
			"attributes": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
			},
			"import": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
		},
	}
}

func mapFromDataToRole(data *schema.ResourceData) *keycloak.Role {
	attributes := map[string][]string{}
	if v, ok := data.GetOk("attributes"); ok {
		for key, value := range v.(map[string]interface{}) {
			attributes[key] = strings.Split(value.(string), MULTIVALUE_ATTRIBUTE_SEPARATOR)
		}
	}

	role := &keycloak.Role{
		Id:          data.Id(),
		RealmId:     data.Get("realm_id").(string),
		ClientId:    data.Get("client_id").(string),
		Name:        data.Get("name").(string),
		Description: data.Get("description").(string),
		Attributes:  attributes,
	}

	return role
}

func mapFromRoleToData(data *schema.ResourceData, role *keycloak.Role) {
	attributes := map[string]string{}
	for k, v := range role.Attributes {
		attributes[k] = strings.Join(v, MULTIVALUE_ATTRIBUTE_SEPARATOR)
	}
	data.SetId(role.Id)

	data.Set("realm_id", role.RealmId)
	data.Set("client_id", role.ClientId)
	data.Set("name", role.Name)
	data.Set("description", role.Description)
	data.Set("attributes", attributes)
}

func resourceKeycloakRoleCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	role := mapFromDataToRole(data)

	if data.Get("import").(bool) {
		realmId := data.Get("realm_id").(string)
		name := data.Get("name").(string)
		clientId := data.Get("client_id").(string)
		compositeRolesValue, compositeRolesSpecified := data.GetOk("composite_roles")

		compositeRoles, err := mapCompositeRoleIdsToRoleObjects(ctx, keycloakClient, compositeRolesValue.(*schema.Set).List(), role.RealmId)
		if err != nil {
			return diag.FromErr(err)
		}

		if len(compositeRoles) != 0 {
			role.Composite = true
		}

		existingRole, err := keycloakClient.GetRoleByName(ctx, realmId, clientId, name)
		if err != nil {
			return diag.FromErr(err)
		}

		if err = mergo.Merge(role, existingRole); err != nil {
			return diag.FromErr(err)
		}
		if err = keycloakClient.UpdateRole(ctx, role); err != nil {
			return diag.FromErr(err)
		}

		existingCompositeRoles, err := keycloakClient.GetRoleComposites(ctx, role)
		if err != nil {
			return diag.FromErr(err)
		}

		if role.Composite && compositeRolesSpecified {
			if err = keycloakClient.RemoveCompositesFromRole(ctx, role, existingCompositeRoles); err != nil {
				return diag.FromErr(err)
			}
			if err = keycloakClient.AddCompositesToRole(ctx, role, compositeRoles); err != nil {
				return diag.FromErr(err)
			}
		}

	} else {
		compositeRoles, err := mapCompositeRoleIdsToRoleObjects(ctx, keycloakClient, data.Get("composite_roles").(*schema.Set).List(), role.RealmId)
		if err != nil {
			return diag.FromErr(err)
		}

		if len(compositeRoles) != 0 { // technically you can still specify composite_roles = [] in HCL
			role.Composite = true
		}

		if err = keycloakClient.CreateRole(ctx, role); err != nil {
			return diag.FromErr(err)
		}

		if role.Composite {
			if err = keycloakClient.AddCompositesToRole(ctx, role, compositeRoles); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	mapFromRoleToData(data, role)

	return resourceKeycloakRoleRead(ctx, data, meta)
}

func mapCompositeRoleIdsToRoleObjects(ctx context.Context, keycloakClient *keycloak.KeycloakClient, compositeRoleIds []interface{}, realmId string) ([]*keycloak.Role, error) {
	var compositeRoles []*keycloak.Role

	for _, compositeRoleId := range compositeRoleIds {
		compositeRoleToAdd, err := keycloakClient.GetRole(ctx, realmId, compositeRoleId.(string))
		if err != nil {
			return nil, err
		}

		compositeRoles = append(compositeRoles, compositeRoleToAdd)
	}
	return compositeRoles, nil
}

func resourceKeycloakRoleRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	role, err := keycloakClient.GetRole(ctx, realmId, id)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	mapFromRoleToData(data, role)

	if role.Composite {
		composites, err := keycloakClient.GetRoleComposites(ctx, role)
		if err != nil {
			return diag.FromErr(err)
		}

		var compositeRoleIds []string

		for _, composite := range composites {
			compositeRoleIds = append(compositeRoleIds, composite.Id)
		}

		data.Set("composite_roles", compositeRoleIds)
	}

	return nil
}

func resourceKeycloakRoleUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	role := mapFromDataToRole(data)

	err := keycloakClient.UpdateRole(ctx, role)
	if err != nil {
		return diag.FromErr(err)
	}

	keycloakComposites, err := keycloakClient.GetRoleComposites(ctx, role)
	if err != nil {
		return diag.FromErr(err)
	}

	if v, ok := data.GetOk("composite_roles"); ok {
		tfCompositeIds := v.(*schema.Set)
		var keycloakCompositesToRemove []*keycloak.Role

		// get a list of all composites to remove and all composites to add
		for _, keycloakComposite := range keycloakComposites {
			if tfCompositeIds.Contains(keycloakComposite.Id) {
				// if the composite exists in keycloak and tf state, we can remove them from the local list because this role does not need to be added
				tfCompositeIds.Remove(keycloakComposite.Id)
			} else {
				// if the composite exists in keycloak but not tf state, it needs to be removed on keycloak's side
				keycloakCompositesToRemove = append(keycloakCompositesToRemove, keycloakComposite)
			}
		}

		// at this point we have two slices:
		// `keycloakCompositesToRemove` should be removed from the role's list of composites
		// `tfCompositeIds` should be added to the role's list of composites. all of the roles that exist on both sides have already been removed

		if len(keycloakCompositesToRemove) != 0 {
			err = keycloakClient.RemoveCompositesFromRole(ctx, role, keycloakCompositesToRemove)
			if err != nil {
				return diag.FromErr(err)
			}
		}

		if tfCompositeIds.Len() != 0 {
			var compositesToAdd []*keycloak.Role
			for _, tfCompositeId := range tfCompositeIds.List() {
				compositeToAdd, err := keycloakClient.GetRole(ctx, role.RealmId, tfCompositeId.(string))
				if err != nil {
					return diag.FromErr(err)
				}

				compositesToAdd = append(compositesToAdd, compositeToAdd)
			}

			err = keycloakClient.AddCompositesToRole(ctx, role, compositesToAdd)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	} else {
		// the user wants this role to have zero composites. if there are composites attached, remove them
		if len(keycloakComposites) != 0 {
			err = keycloakClient.RemoveCompositesFromRole(ctx, role, keycloakComposites)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return resourceKeycloakRoleRead(ctx, data, meta)
}

func resourceKeycloakRoleDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if data.Get("import").(bool) {
		return nil
	}

	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	return diag.FromErr(keycloakClient.DeleteRole(ctx, realmId, id))
}

func resourceKeycloakRoleImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("Invalid import. Supported import format: {{realm}}/{{roleId}}.")
	}

	_, err := keycloakClient.GetRole(ctx, parts[0], parts[1])
	if err != nil {
		return nil, err
	}

	d.Set("realm_id", parts[0])
	d.Set("import", false)
	d.SetId(parts[1])

	diagnostics := resourceKeycloakRoleRead(ctx, d, meta)
	if diagnostics.HasError() {
		return nil, errors.New(diagnostics[0].Summary)
	}

	return []*schema.ResourceData{d}, nil
}
