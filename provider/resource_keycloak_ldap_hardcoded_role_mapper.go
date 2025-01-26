package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakLdapHardcodedRoleMapper() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakLdapHardcodedRoleMapperCreate,
		ReadContext:   resourceKeycloakLdapHardcodedRoleMapperRead,
		UpdateContext: resourceKeycloakLdapHardcodedRoleMapperUpdate,
		DeleteContext: resourceKeycloakLdapHardcodedRoleMapperDelete,
		// This resource can be imported using {{realm}}/{{provider_id}}/{{mapper_id}}. The Provider and Mapper IDs are displayed in the GUI
		Importer: &schema.ResourceImporter{
			StateContext: resourceKeycloakLdapGenericMapperImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Display name of the mapper when displayed in the console.",
			},
			"realm_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The realm in which the ldap user federation provider exists.",
			},
			"ldap_user_federation_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ldap user federation provider to attach this mapper to.",
			},
			"role": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Role to grant to user.",
			},
		},
	}
}

func getLdapHardcodedRoleMapperFromData(data *schema.ResourceData) *keycloak.LdapHardcodedRoleMapper {
	return &keycloak.LdapHardcodedRoleMapper{
		Id:                   data.Id(),
		Name:                 data.Get("name").(string),
		RealmId:              data.Get("realm_id").(string),
		LdapUserFederationId: data.Get("ldap_user_federation_id").(string),
		Role:                 data.Get("role").(string),
	}
}

func setLdapHardcodedRoleMapperData(data *schema.ResourceData, ldapMapper *keycloak.LdapHardcodedRoleMapper) {
	data.SetId(ldapMapper.Id)
	data.Set("name", ldapMapper.Name)
	data.Set("realm_id", ldapMapper.RealmId)
	data.Set("ldap_user_federation_id", ldapMapper.LdapUserFederationId)
	data.Set("role", ldapMapper.Role)
}

func resourceKeycloakLdapHardcodedRoleMapperCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	ldapMapper := getLdapHardcodedRoleMapperFromData(data)

	err := keycloakClient.ValidateLdapHardcodedRoleMapper(ctx, ldapMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.NewLdapHardcodedRoleMapper(ctx, ldapMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	setLdapHardcodedRoleMapperData(data, ldapMapper)

	return resourceKeycloakLdapHardcodedRoleMapperRead(ctx, data, meta)
}

func resourceKeycloakLdapHardcodedRoleMapperRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	ldapMapper, err := keycloakClient.GetLdapHardcodedRoleMapper(ctx, realmId, id)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	setLdapHardcodedRoleMapperData(data, ldapMapper)

	return nil
}

func resourceKeycloakLdapHardcodedRoleMapperUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	ldapMapper := getLdapHardcodedRoleMapperFromData(data)

	err := keycloakClient.ValidateLdapHardcodedRoleMapper(ctx, ldapMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.UpdateLdapHardcodedRoleMapper(ctx, ldapMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	setLdapHardcodedRoleMapperData(data, ldapMapper)

	return nil
}

func resourceKeycloakLdapHardcodedRoleMapperDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	return diag.FromErr(keycloakClient.DeleteLdapHardcodedRoleMapper(ctx, realmId, id))
}
