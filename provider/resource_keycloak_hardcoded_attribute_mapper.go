package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakHardcodedAttributeMapper() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakHardcodedAttributeMapperCreate,
		ReadContext:   resourceKeycloakHardcodedAttributeMapperRead,
		UpdateContext: resourceKeycloakHardcodedAttributeMapperUpdate,
		DeleteContext: resourceKeycloakHardcodedAttributeMapperDelete,
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
			"attribute_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the user schema attribute",
			},
			"attribute_value": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Value of the attribute. You can hardcode any value like 'foo'",
			},
		},
	}
}

func getHardcodedAttributeMapperFromData(data *schema.ResourceData) *keycloak.HardcodedAttributeMapper {
	return &keycloak.HardcodedAttributeMapper{
		Id:                   data.Id(),
		Name:                 data.Get("name").(string),
		RealmId:              data.Get("realm_id").(string),
		LdapUserFederationId: data.Get("ldap_user_federation_id").(string),
		AttributeName:        data.Get("attribute_name").(string),
		AttributeValue:       data.Get("attribute_value").(string),
	}
}

func setHardcodedAttributeMapperData(data *schema.ResourceData, ldapMapper *keycloak.HardcodedAttributeMapper) {
	data.SetId(ldapMapper.Id)
	data.Set("name", ldapMapper.Name)
	data.Set("realm_id", ldapMapper.RealmId)
	data.Set("ldap_user_federation_id", ldapMapper.LdapUserFederationId)
	data.Set("attribute_name", ldapMapper.AttributeName)
	data.Set("attribute_value", ldapMapper.AttributeValue)
}

func resourceKeycloakHardcodedAttributeMapperCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	ldapMapper := getHardcodedAttributeMapperFromData(data)

	err := keycloakClient.NewHardcodedAttributeMapper(ctx, ldapMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	setHardcodedAttributeMapperData(data, ldapMapper)

	return resourceKeycloakHardcodedAttributeMapperRead(ctx, data, meta)
}

func resourceKeycloakHardcodedAttributeMapperRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	ldapMapper, err := keycloakClient.GetHardcodedAttributeMapper(ctx, realmId, id)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	setHardcodedAttributeMapperData(data, ldapMapper)

	return diag.FromErr(nil)
}

func resourceKeycloakHardcodedAttributeMapperUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	ldapMapper := getHardcodedAttributeMapperFromData(data)

	err := keycloakClient.UpdateHardcodedAttributeMapper(ctx, ldapMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	setHardcodedAttributeMapperData(data, ldapMapper)

	return diag.FromErr(nil)
}

func resourceKeycloakHardcodedAttributeMapperDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	err := keycloakClient.DeleteHardcodedAttributeMapper(ctx, realmId, id)

	return diag.FromErr(err)
}
