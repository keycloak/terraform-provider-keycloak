package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakOpenIdSubProtocolMapper() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakOpenIdSubProtocolMapperCreate,
		ReadContext:   resourceKeycloakOpenIdSubProtocolMapperRead,
		UpdateContext: resourceKeycloakOpenIdSubProtocolMapperUpdate,
		DeleteContext: resourceKeycloakOpenIdSubProtocolMapperDelete,
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
			"add_to_access_token": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Indicates if the attribute should be a claim in the access token.",
			},
			"add_to_token_introspection": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Indicates if the attribute should be a claim in the token introspection response body.",
			},
		},
	}
}

func mapFromDataToOpenIdSubProtocolMapper(data *schema.ResourceData) *keycloak.OpenIdSubProtocolMapper {
	return &keycloak.OpenIdSubProtocolMapper{
		Id:                      data.Id(),
		Name:                    data.Get("name").(string),
		RealmId:                 data.Get("realm_id").(string),
		ClientId:                data.Get("client_id").(string),
		ClientScopeId:           data.Get("client_scope_id").(string),
		AddToAccessToken:        data.Get("add_to_access_token").(bool),
		AddToTokenIntrospection: data.Get("add_to_token_introspection").(bool),
	}
}

func mapFromOpenIdSubMapperToData(mapper *keycloak.OpenIdSubProtocolMapper, data *schema.ResourceData) {
	data.SetId(mapper.Id)
	data.Set("name", mapper.Name)
	data.Set("realm_id", mapper.RealmId)

	if mapper.ClientId != "" {
		data.Set("client_id", mapper.ClientId)
	} else {
		data.Set("client_scope_id", mapper.ClientScopeId)
	}

	data.Set("add_to_access_token", mapper.AddToAccessToken)
	data.Set("add_to_token_introspection", mapper.AddToTokenIntrospection)
}

func resourceKeycloakOpenIdSubProtocolMapperCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	openIdSubMapper := mapFromDataToOpenIdSubProtocolMapper(data)

	err := keycloakClient.ValidateOpenIdSubProtocolMapper(ctx, openIdSubMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.NewOpenIdSubProtocolMapper(ctx, openIdSubMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	mapFromOpenIdSubMapperToData(openIdSubMapper, data)

	return resourceKeycloakOpenIdSubProtocolMapperRead(ctx, data, meta)
}

func resourceKeycloakOpenIdSubProtocolMapperRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)
	clientScopeId := data.Get("client_scope_id").(string)

	openIdSubMapper, err := keycloakClient.GetOpenIdSubProtocolMapper(ctx, realmId, clientId, clientScopeId, data.Id())
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	mapFromOpenIdSubMapperToData(openIdSubMapper, data)

	return nil
}

func resourceKeycloakOpenIdSubProtocolMapperUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	openIdSubMapper := mapFromDataToOpenIdSubProtocolMapper(data)

	err := keycloakClient.ValidateOpenIdSubProtocolMapper(ctx, openIdSubMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.UpdateOpenIdSubProtocolMapper(ctx, openIdSubMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceKeycloakOpenIdSubProtocolMapperRead(ctx, data, meta)
}

func resourceKeycloakOpenIdSubProtocolMapperDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)
	clientScopeId := data.Get("client_scope_id").(string)

	return diag.FromErr(keycloakClient.DeleteOpenIdSubProtocolMapper(ctx, realmId, clientId, clientScopeId, data.Id()))
}
