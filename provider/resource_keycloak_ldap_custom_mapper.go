package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakLdapCustomMapper() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakLdapCustomMapperCreate,
		ReadContext:   resourceKeycloakLdapCustomMapperRead,
		UpdateContext: resourceKeycloakLdapCustomMapperUpdate,
		DeleteContext: resourceKeycloakLdapCustomMapperDelete,
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
			"provider_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the custom LDAP mapper.",
			},
			"provider_type": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Fully-qualified name of the Java class implementing the custom LDAP mapper.",
			},
			"config": {
				Type:     schema.TypeMap,
				Optional: true,
			},
		},
	}
}

func getLdapCustomMapperFromData(data *schema.ResourceData) *keycloak.LdapCustomMapper {
	config := make(map[string]string)
	if v, ok := data.GetOk("config"); ok {
		for key, value := range v.(map[string]interface{}) {
			config[key] = value.(string)
		}
	}
	return &keycloak.LdapCustomMapper{
		Id:                   data.Id(),
		Name:                 data.Get("name").(string),
		RealmId:              data.Get("realm_id").(string),
		LdapUserFederationId: data.Get("ldap_user_federation_id").(string),
		ProviderId:           data.Get("provider_id").(string),
		ProviderType:         data.Get("provider_type").(string),
		Config:               config,
	}
}

func setLdapCustomMapperData(data *schema.ResourceData, ldapCustomMapper *keycloak.LdapCustomMapper) {
	data.SetId(ldapCustomMapper.Id)

	data.Set("name", ldapCustomMapper.Name)
	data.Set("realm_id", ldapCustomMapper.RealmId)
	data.Set("ldap_user_federation_id", ldapCustomMapper.LdapUserFederationId)

	data.Set("provider_id", ldapCustomMapper.ProviderId)
	data.Set("provider_type", ldapCustomMapper.ProviderType)
	data.Set("config", ldapCustomMapper.Config)
}

func resourceKeycloakLdapCustomMapperCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	ldapCustomMapper := getLdapCustomMapperFromData(data)

	err := keycloakClient.NewLdapCustomMapper(ctx, ldapCustomMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	setLdapCustomMapperData(data, ldapCustomMapper)

	return resourceKeycloakLdapCustomMapperRead(ctx, data, meta)
}

func resourceKeycloakLdapCustomMapperRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	ldapCustomMapper, err := keycloakClient.GetLdapCustomMapper(ctx, realmId, id)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	setLdapCustomMapperData(data, ldapCustomMapper)

	return nil
}

func resourceKeycloakLdapCustomMapperUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	ldapCustomMapper := getLdapCustomMapperFromData(data)

	err := keycloakClient.UpdateLdapCustomMapper(ctx, ldapCustomMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	setLdapCustomMapperData(data, ldapCustomMapper)

	return nil
}

func resourceKeycloakLdapCustomMapperDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	return diag.FromErr(keycloakClient.DeleteLdapCustomMapper(ctx, realmId, id))
}
