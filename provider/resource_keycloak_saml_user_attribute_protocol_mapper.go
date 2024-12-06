package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

var keycloakSamlUserAttributeProtocolMapperNameFormats = []string{"Basic", "URI Reference", "Unspecified"}

func resourceKeycloakSamlUserAttributeProtocolMapper() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakSamlUserAttributeProtocolMapperCreate,
		ReadContext:   resourceKeycloakSamlUserAttributeProtocolMapperRead,
		UpdateContext: resourceKeycloakSamlUserAttributeProtocolMapperUpdate,
		DeleteContext: resourceKeycloakSamlUserAttributeProtocolMapperDelete,
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
			"user_attribute": {
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

func mapFromDataToSamlUserAttributeProtocolMapper(data *schema.ResourceData) *keycloak.SamlUserAttributeProtocolMapper {
	return &keycloak.SamlUserAttributeProtocolMapper{
		Id:            data.Id(),
		Name:          data.Get("name").(string),
		RealmId:       data.Get("realm_id").(string),
		ClientId:      data.Get("client_id").(string),
		ClientScopeId: data.Get("client_scope_id").(string),

		UserAttribute:           data.Get("user_attribute").(string),
		FriendlyName:            data.Get("friendly_name").(string),
		SamlAttributeName:       data.Get("saml_attribute_name").(string),
		SamlAttributeNameFormat: data.Get("saml_attribute_name_format").(string),
	}
}

func mapFromSamlUserAttributeMapperToData(mapper *keycloak.SamlUserAttributeProtocolMapper, data *schema.ResourceData) {
	data.SetId(mapper.Id)
	data.Set("name", mapper.Name)
	data.Set("realm_id", mapper.RealmId)

	if mapper.ClientId != "" {
		data.Set("client_id", mapper.ClientId)
	} else {
		data.Set("client_scope_id", mapper.ClientScopeId)
	}

	data.Set("user_attribute", mapper.UserAttribute)
	data.Set("friendly_name", mapper.FriendlyName)
	data.Set("saml_attribute_name", mapper.SamlAttributeName)
	data.Set("saml_attribute_name_format", mapper.SamlAttributeNameFormat)
}

func resourceKeycloakSamlUserAttributeProtocolMapperCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	samlUserAttributeMapper := mapFromDataToSamlUserAttributeProtocolMapper(data)

	err := keycloakClient.ValidateSamlUserAttributeProtocolMapper(ctx, samlUserAttributeMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.NewSamlUserAttributeProtocolMapper(ctx, samlUserAttributeMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	mapFromSamlUserAttributeMapperToData(samlUserAttributeMapper, data)

	return resourceKeycloakSamlUserAttributeProtocolMapperRead(ctx, data, meta)
}

func resourceKeycloakSamlUserAttributeProtocolMapperRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)
	clientScopeId := data.Get("client_scope_id").(string)

	samlUserAttributeMapper, err := keycloakClient.GetSamlUserAttributeProtocolMapper(ctx, realmId, clientId, clientScopeId, data.Id())
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	mapFromSamlUserAttributeMapperToData(samlUserAttributeMapper, data)

	return nil
}

func resourceKeycloakSamlUserAttributeProtocolMapperUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	samlUserAttributeMapper := mapFromDataToSamlUserAttributeProtocolMapper(data)

	err := keycloakClient.ValidateSamlUserAttributeProtocolMapper(ctx, samlUserAttributeMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.UpdateSamlUserAttributeProtocolMapper(ctx, samlUserAttributeMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceKeycloakSamlUserAttributeProtocolMapperRead(ctx, data, meta)
}

func resourceKeycloakSamlUserAttributeProtocolMapperDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)
	clientScopeId := data.Get("client_scope_id").(string)

	return diag.FromErr(keycloakClient.DeleteSamlUserAttributeProtocolMapper(ctx, realmId, clientId, clientScopeId, data.Id()))
}
