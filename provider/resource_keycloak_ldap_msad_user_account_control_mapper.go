package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakLdapMsadUserAccountControlMapper() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakLdapMsadUserAccountControlMapperCreate,
		ReadContext:   resourceKeycloakLdapMsadUserAccountControlMapperRead,
		UpdateContext: resourceKeycloakLdapMsadUserAccountControlMapperUpdate,
		DeleteContext: resourceKeycloakLdapMsadUserAccountControlMapperDelete,
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
			"ldap_password_policy_hints_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func getLdapMsadUserAccountControlMapperFromData(data *schema.ResourceData) *keycloak.LdapMsadUserAccountControlMapper {
	return &keycloak.LdapMsadUserAccountControlMapper{
		Id:                   data.Id(),
		Name:                 data.Get("name").(string),
		RealmId:              data.Get("realm_id").(string),
		LdapUserFederationId: data.Get("ldap_user_federation_id").(string),

		LdapPasswordPolicyHintsEnabled: data.Get("ldap_password_policy_hints_enabled").(bool),
	}
}

func setLdapMsadUserAccountControlMapperData(data *schema.ResourceData, ldapMsadUserAccountControlMapper *keycloak.LdapMsadUserAccountControlMapper) {
	data.SetId(ldapMsadUserAccountControlMapper.Id)

	data.Set("name", ldapMsadUserAccountControlMapper.Name)
	data.Set("realm_id", ldapMsadUserAccountControlMapper.RealmId)
	data.Set("ldap_user_federation_id", ldapMsadUserAccountControlMapper.LdapUserFederationId)

	data.Set("ldap_password_policy_hints_enabled", ldapMsadUserAccountControlMapper.LdapPasswordPolicyHintsEnabled)
}

func resourceKeycloakLdapMsadUserAccountControlMapperCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	ldapMsadUserAccountControlMapper := getLdapMsadUserAccountControlMapperFromData(data)

	err := keycloakClient.NewLdapMsadUserAccountControlMapper(ctx, ldapMsadUserAccountControlMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	setLdapMsadUserAccountControlMapperData(data, ldapMsadUserAccountControlMapper)

	return resourceKeycloakLdapMsadUserAccountControlMapperRead(ctx, data, meta)
}

func resourceKeycloakLdapMsadUserAccountControlMapperRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	ldapMsadUserAccountControlMapper, err := keycloakClient.GetLdapMsadUserAccountControlMapper(ctx, realmId, id)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	setLdapMsadUserAccountControlMapperData(data, ldapMsadUserAccountControlMapper)

	return nil
}

func resourceKeycloakLdapMsadUserAccountControlMapperUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	ldapMsadUserAccountControlMapper := getLdapMsadUserAccountControlMapperFromData(data)

	err := keycloakClient.UpdateLdapMsadUserAccountControlMapper(ctx, ldapMsadUserAccountControlMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	setLdapMsadUserAccountControlMapperData(data, ldapMsadUserAccountControlMapper)

	return nil
}

func resourceKeycloakLdapMsadUserAccountControlMapperDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	return diag.FromErr(keycloakClient.DeleteLdapMsadUserAccountControlMapper(ctx, realmId, id))
}
