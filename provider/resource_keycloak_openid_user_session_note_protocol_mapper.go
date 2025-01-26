package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakOpenIdUserSessionNoteProtocolMapper() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakOpenIdUserSessionNoteProtocolMapperCreate,
		ReadContext:   resourceKeycloakOpenIdUserSessionNoteProtocolMapperRead,
		UpdateContext: resourceKeycloakOpenIdUserSessionNoteProtocolMapperUpdate,
		DeleteContext: resourceKeycloakOpenIdUserSessionNoteProtocolMapperDelete,
		Importer: &schema.ResourceImporter{
			// import a mapper tied to a client:
			// {{realmId}}/client/{{clientId}}/{{protocolMapperId}}
			// or a client scope:
			// {{realmId}}/client-scope/{{clientScopeId}}/{{protocolMapperId}}
			StateContext: genericProtocolMapperImport,
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
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Description:   "The mapper's associated client. Cannot be used at the same time as client_scope_id.",
				ConflictsWith: []string{"client_scope_id"},
			},
			"client_scope_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Description:   "The mapper's associated client scope. Cannot be used at the same time as client_id.",
				ConflictsWith: []string{"client_id"},
			},
			"add_to_id_token": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Indicates if the attribute should be a claim in the id token.",
			},
			"add_to_access_token": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Indicates if the attribute should be a claim in the access token.",
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
			"session_note": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "String value being the name of stored user session note within the UserSessionModel.note map.",
			},
		},
	}
}

func mapFromDataToOpenIdUserSessionNoteProtocolMapper(data *schema.ResourceData) *keycloak.OpenIdUserSessionNoteProtocolMapper {
	return &keycloak.OpenIdUserSessionNoteProtocolMapper{
		Id:               data.Id(),
		Name:             data.Get("name").(string),
		RealmId:          data.Get("realm_id").(string),
		ClientId:         data.Get("client_id").(string),
		ClientScopeId:    data.Get("client_scope_id").(string),
		AddToIdToken:     data.Get("add_to_id_token").(bool),
		AddToAccessToken: data.Get("add_to_access_token").(bool),

		ClaimName:       data.Get("claim_name").(string),
		ClaimValueType:  data.Get("claim_value_type").(string),
		UserSessionNote: data.Get("session_note").(string),
	}
}

func mapFromOpenIdUserSessionNoteMapperToData(mapper *keycloak.OpenIdUserSessionNoteProtocolMapper, data *schema.ResourceData) {
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
	data.Set("claim_name", mapper.ClaimName)
	data.Set("claim_value_type", mapper.ClaimValueType)
	data.Set("session_note", mapper.UserSessionNote)
}

func resourceKeycloakOpenIdUserSessionNoteProtocolMapperCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	openIdUserSessionNoteMapper := mapFromDataToOpenIdUserSessionNoteProtocolMapper(data)

	err := keycloakClient.ValidateOpenIdUserSessionNoteProtocolMapper(ctx, openIdUserSessionNoteMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.NewOpenIdUserSessionNoteProtocolMapper(ctx, openIdUserSessionNoteMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	mapFromOpenIdUserSessionNoteMapperToData(openIdUserSessionNoteMapper, data)

	return resourceKeycloakOpenIdUserSessionNoteProtocolMapperRead(ctx, data, meta)
}

func resourceKeycloakOpenIdUserSessionNoteProtocolMapperRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)
	clientScopeId := data.Get("client_scope_id").(string)

	openIdUserSessionNoteMapper, err := keycloakClient.GetOpenIdUserSessionNoteProtocolMapper(ctx, realmId, clientId, clientScopeId, data.Id())
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	mapFromOpenIdUserSessionNoteMapperToData(openIdUserSessionNoteMapper, data)

	return nil
}

func resourceKeycloakOpenIdUserSessionNoteProtocolMapperUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	openIdUserSessionNoteMapper := mapFromDataToOpenIdUserSessionNoteProtocolMapper(data)

	err := keycloakClient.ValidateOpenIdUserSessionNoteProtocolMapper(ctx, openIdUserSessionNoteMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.UpdateOpenIdUserSessionNoteProtocolMapper(ctx, openIdUserSessionNoteMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceKeycloakOpenIdUserSessionNoteProtocolMapperRead(ctx, data, meta)
}

func resourceKeycloakOpenIdUserSessionNoteProtocolMapperDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)
	clientScopeId := data.Get("client_scope_id").(string)

	return diag.FromErr(keycloakClient.DeleteOpenIdUserSessionNoteProtocolMapper(ctx, realmId, clientId, clientScopeId, data.Id()))
}
