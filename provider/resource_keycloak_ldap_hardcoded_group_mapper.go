package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakLdapHardcodedGroupMapper() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakLdapHardcodedGroupMapperCreate,
		ReadContext:   resourceKeycloakLdapHardcodedGroupMapperRead,
		UpdateContext: resourceKeycloakLdapHardcodedGroupMapperUpdate,
		DeleteContext: resourceKeycloakLdapHardcodedGroupMapperDelete,
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
			"group": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Group to grant to user.",
			},
		},
	}
}

func getLdapHardcodedGroupMapperFromData(data *schema.ResourceData) *keycloak.LdapHardcodedGroupMapper {
	return &keycloak.LdapHardcodedGroupMapper{
		Id:                   data.Id(),
		Name:                 data.Get("name").(string),
		RealmId:              data.Get("realm_id").(string),
		LdapUserFederationId: data.Get("ldap_user_federation_id").(string),
		Group:                data.Get("group").(string),
	}
}

func setLdapHardcodedGroupMapperData(data *schema.ResourceData, ldapMapper *keycloak.LdapHardcodedGroupMapper) {
	data.SetId(ldapMapper.Id)
	data.Set("name", ldapMapper.Name)
	data.Set("realm_id", ldapMapper.RealmId)
	data.Set("ldap_user_federation_id", ldapMapper.LdapUserFederationId)
	data.Set("group", ldapMapper.Group)
}

func resourceKeycloakLdapHardcodedGroupMapperCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	ldapMapper := getLdapHardcodedGroupMapperFromData(data)

	err := keycloakClient.ValidateLdapHardcodedGroupMapper(ctx, ldapMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.NewLdapHardcodedGroupMapper(ctx, ldapMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	setLdapHardcodedGroupMapperData(data, ldapMapper)

	return resourceKeycloakLdapHardcodedGroupMapperRead(ctx, data, meta)
}

func resourceKeycloakLdapHardcodedGroupMapperRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	ldapMapper, err := keycloakClient.GetLdapHardcodedGroupMapper(ctx, realmId, id)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	setLdapHardcodedGroupMapperData(data, ldapMapper)

	return nil
}

func resourceKeycloakLdapHardcodedGroupMapperUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	ldapMapper := getLdapHardcodedGroupMapperFromData(data)

	err := keycloakClient.ValidateLdapHardcodedGroupMapper(ctx, ldapMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.UpdateLdapHardcodedGroupMapper(ctx, ldapMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	setLdapHardcodedGroupMapperData(data, ldapMapper)

	return nil
}

func resourceKeycloakLdapHardcodedGroupMapperDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	return diag.FromErr(keycloakClient.DeleteLdapHardcodedGroupMapper(ctx, realmId, id))
}
