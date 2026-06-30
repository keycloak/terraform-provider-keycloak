package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakOpenIdUserPropertyProtocolMapper() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakOpenIdUserPropertyProtocolMapperCreate,
		ReadContext:   resourceKeycloakOpenIdUserPropertyProtocolMapperRead,
		UpdateContext: resourceKeycloakOpenIdUserPropertyProtocolMapperUpdate,
		DeleteContext: resourceKeycloakOpenIdUserPropertyProtocolMapperDelete,
		Importer: &schema.ResourceImporter{
			// import a mapper tied to a client:
			// {{realmId}}/client/{{clientId}}/{{protocolMapperId}}
			// or a client scope:
			// {{realmId}}/client-scope/{{clientScopeId}}/{{protocolMapperId}}
			StateContext: genericProtocolMapperImportWithImportFlag,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A human-friendly name that will appear in the Keycloak console.",
			},
			"realm_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The realm id where the associated client or client scope exists.",
			},
			"client_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Description:  "The mapper's associated client. Cannot be used at the same time as client_scope_id.",
				ExactlyOneOf: []string{"client_id", "client_scope_id"},
			},
			"client_scope_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Description:  "The mapper's associated client scope. Cannot be used at the same time as client_id.",
				ExactlyOneOf: []string{"client_id", "client_scope_id"},
			},
			"add_to_id_token": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Indicates if the property should be a claim in the id token.",
			},
			"add_to_access_token": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Indicates if the property should be a claim in the access token.",
			},
			"add_to_userinfo": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Indicates if the property should appear in the userinfo response body.",
			},
			"user_property": {
				Type:     schema.TypeString,
				Required: true,
			},
			"claim_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"claim_value_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Claim type used when serializing tokens.",
				Default:      "String",
				ValidateFunc: validation.StringInSlice([]string{"JSON", "String", "long", "int", "boolean"}, true),
			},
			"import": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    false,
				Description: "When true, the provider looks up an existing protocol mapper by name and updates it with the new configuration, overwriting any existing values. After the first apply, the resource behaves identically to one that was created normally, including deletion on terraform destroy.",
			},
		},
	}
}

func mapFromDataToOpenIdUserPropertyProtocolMapper(data *schema.ResourceData) *keycloak.OpenIdUserPropertyProtocolMapper {
	return &keycloak.OpenIdUserPropertyProtocolMapper{
		Id:               data.Id(),
		Name:             data.Get("name").(string),
		RealmId:          data.Get("realm_id").(string),
		ClientId:         data.Get("client_id").(string),
		ClientScopeId:    data.Get("client_scope_id").(string),
		AddToIdToken:     data.Get("add_to_id_token").(bool),
		AddToAccessToken: data.Get("add_to_access_token").(bool),
		AddToUserInfo:    data.Get("add_to_userinfo").(bool),

		UserProperty:   data.Get("user_property").(string),
		ClaimName:      data.Get("claim_name").(string),
		ClaimValueType: data.Get("claim_value_type").(string),
	}
}

func mapFromOpenIdUserPropertyMapperToData(mapper *keycloak.OpenIdUserPropertyProtocolMapper, data *schema.ResourceData) {
	data.SetId(mapper.Id)
	data.Set("name", mapper.Name)
	data.Set("realm_id", mapper.RealmId)

	if mapper.ClientId != "" {
		data.Set("client_id", mapper.ClientId)
	} else {
		data.Set("client_scope_id", mapper.ClientScopeId)
	}

	data.Set("add_to_id_token", mapper.AddToIdToken)
	data.Set("add_to_access_token", mapper.AddToAccessToken)
	data.Set("add_to_userinfo", mapper.AddToUserInfo)
	data.Set("user_property", mapper.UserProperty)
	data.Set("claim_name", mapper.ClaimName)
	data.Set("claim_value_type", mapper.ClaimValueType)
}

func resourceKeycloakOpenIdUserPropertyProtocolMapperCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	importMode := data.Get("import").(bool)

	openIdUserPropertyMapper := mapFromDataToOpenIdUserPropertyProtocolMapper(data)

	if importMode {
		existingMapper, err := keycloakClient.GetOpenIdUserPropertyProtocolMapperByName(ctx, openIdUserPropertyMapper.RealmId, openIdUserPropertyMapper.ClientId, openIdUserPropertyMapper.ClientScopeId, openIdUserPropertyMapper.Name)
		if err != nil {
			return diag.FromErr(err)
		}

		if existingMapper == nil {
			return diag.Errorf("protocol mapper with name %q not found for import", openIdUserPropertyMapper.Name)
		}

		// We only preserve the ID of the existing mapper to ensure we update the correct resource.
		openIdUserPropertyMapper.Id = existingMapper.Id

		err = keycloakClient.UpdateOpenIdUserPropertyProtocolMapper(ctx, openIdUserPropertyMapper)
		if err != nil {
			return diag.FromErr(err)
		}
	} else {
		// Validation checks for existing protocol mappers with the same name, which is only
		// relevant when creating a new protocol mapper, not when importing an existing one
		err := openIdUserPropertyMapper.Validate(ctx, keycloakClient)
		if err != nil {
			return diag.FromErr(err)
		}

		err = keycloakClient.NewOpenIdUserPropertyProtocolMapper(ctx, openIdUserPropertyMapper)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	mapFromOpenIdUserPropertyMapperToData(openIdUserPropertyMapper, data)

	return resourceKeycloakOpenIdUserPropertyProtocolMapperRead(ctx, data, meta)
}

func resourceKeycloakOpenIdUserPropertyProtocolMapperRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)
	clientScopeId := data.Get("client_scope_id").(string)

	openIdUserPropertyMapper, err := keycloakClient.GetOpenIdUserPropertyProtocolMapper(ctx, realmId, clientId, clientScopeId, data.Id())
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	mapFromOpenIdUserPropertyMapperToData(openIdUserPropertyMapper, data)

	return nil
}

func resourceKeycloakOpenIdUserPropertyProtocolMapperUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	openIdUserPropertyMapper := mapFromDataToOpenIdUserPropertyProtocolMapper(data)
	err := keycloakClient.UpdateOpenIdUserPropertyProtocolMapper(ctx, openIdUserPropertyMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceKeycloakOpenIdUserPropertyProtocolMapperRead(ctx, data, meta)
}

func resourceKeycloakOpenIdUserPropertyProtocolMapperDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)
	clientScopeId := data.Get("client_scope_id").(string)

	return diag.FromErr(keycloakClient.DeleteOpenIdUserPropertyProtocolMapper(ctx, realmId, clientId, clientScopeId, data.Id()))
}
