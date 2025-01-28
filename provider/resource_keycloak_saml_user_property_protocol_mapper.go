package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakSamlUserPropertyProtocolMapper() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakSamlUserPropertyProtocolMapperCreate,
		ReadContext:   resourceKeycloakSamlUserPropertyProtocolMapperRead,
		UpdateContext: resourceKeycloakSamlUserPropertyProtocolMapperUpdate,
		DeleteContext: resourceKeycloakSamlUserPropertyProtocolMapperDelete,
		Importer: &schema.ResourceImporter{
			// import a mapper tied to a client:
			// {{realmId}}/client/{{clientId}}/{{protocolMapperId}}
			// or a client scope:
			// {{realmId}}/client-scope/{{clientScopeId}}/{{protocolMapperId}}
			StateContext: genericProtocolMapperImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"client_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"client_scope_id"},
			},
			"client_scope_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"client_id"},
			},
			"user_property": {
				Type:     schema.TypeString,
				Required: true,
			},
			"friendly_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"saml_attribute_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"saml_attribute_name_format": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(keycloakSamlUserAttributeProtocolMapperNameFormats, false),
			},
		},
	}
}

func mapFromDataToSamlUserPropertyProtocolMapper(data *schema.ResourceData) *keycloak.SamlUserPropertyProtocolMapper {
	return &keycloak.SamlUserPropertyProtocolMapper{
		Id:            data.Id(),
		Name:          data.Get("name").(string),
		RealmId:       data.Get("realm_id").(string),
		ClientId:      data.Get("client_id").(string),
		ClientScopeId: data.Get("client_scope_id").(string),

		UserProperty:            data.Get("user_property").(string),
		FriendlyName:            data.Get("friendly_name").(string),
		SamlAttributeName:       data.Get("saml_attribute_name").(string),
		SamlAttributeNameFormat: data.Get("saml_attribute_name_format").(string),
	}
}

func mapFromSamlUserPropertyProtocolMapperToData(mapper *keycloak.SamlUserPropertyProtocolMapper, data *schema.ResourceData) {
	data.SetId(mapper.Id)
	data.Set("name", mapper.Name)
	data.Set("realm_id", mapper.RealmId)

	if mapper.ClientId != "" {
		data.Set("client_id", mapper.ClientId)
	} else {
		data.Set("client_scope_id", mapper.ClientScopeId)
	}

	data.Set("user_property", mapper.UserProperty)
	data.Set("friendly_name", mapper.FriendlyName)
	data.Set("saml_attribute_name", mapper.SamlAttributeName)
	data.Set("saml_attribute_name_format", mapper.SamlAttributeNameFormat)
}

func resourceKeycloakSamlUserPropertyProtocolMapperCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	samlUserPropertyMapper := mapFromDataToSamlUserPropertyProtocolMapper(data)

	err := keycloakClient.ValidateSamlUserPropertyProtocolMapper(ctx, samlUserPropertyMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.NewSamlUserPropertyProtocolMapper(ctx, samlUserPropertyMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	mapFromSamlUserPropertyProtocolMapperToData(samlUserPropertyMapper, data)

	return resourceKeycloakSamlUserPropertyProtocolMapperRead(ctx, data, meta)
}

func resourceKeycloakSamlUserPropertyProtocolMapperRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)
	clientScopeId := data.Get("client_scope_id").(string)

	samlUserPropertyMapper, err := keycloakClient.GetSamlUserPropertyProtocolMapper(ctx, realmId, clientId, clientScopeId, data.Id())
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	mapFromSamlUserPropertyProtocolMapperToData(samlUserPropertyMapper, data)

	return nil
}

func resourceKeycloakSamlUserPropertyProtocolMapperUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	samlUserPropertyMapper := mapFromDataToSamlUserPropertyProtocolMapper(data)

	err := keycloakClient.ValidateSamlUserPropertyProtocolMapper(ctx, samlUserPropertyMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.UpdateSamlUserPropertyProtocolMapper(ctx, samlUserPropertyMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceKeycloakSamlUserPropertyProtocolMapperRead(ctx, data, meta)
}

func resourceKeycloakSamlUserPropertyProtocolMapperDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)
	clientScopeId := data.Get("client_scope_id").(string)

	return diag.FromErr(keycloakClient.DeleteSamlUserPropertyProtocolMapper(ctx, realmId, clientId, clientScopeId, data.Id()))
}
