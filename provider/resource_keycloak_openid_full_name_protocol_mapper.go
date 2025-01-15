package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakOpenIdFullNameProtocolMapper() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakOpenIdFullNameProtocolMapperCreate,
		ReadContext:   resourceKeycloakOpenIdFullNameProtocolMapperRead,
		UpdateContext: resourceKeycloakOpenIdFullNameProtocolMapperUpdate,
		DeleteContext: resourceKeycloakOpenIdFullNameProtocolMapperDelete,
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
				ForceNew:    true,
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
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"add_to_access_token": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"add_to_userinfo": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func mapFromDataToOpenIdFullNameProtocolMapper(data *schema.ResourceData) *keycloak.OpenIdFullNameProtocolMapper {
	return &keycloak.OpenIdFullNameProtocolMapper{
		Id:            data.Id(),
		Name:          data.Get("name").(string),
		RealmId:       data.Get("realm_id").(string),
		ClientId:      data.Get("client_id").(string),
		ClientScopeId: data.Get("client_scope_id").(string),

		AddToIdToken:     data.Get("add_to_id_token").(bool),
		AddToAccessToken: data.Get("add_to_access_token").(bool),
		AddToUserInfo:    data.Get("add_to_userinfo").(bool),
	}
}

func mapFromOpenIdFullNameMapperToData(mapper *keycloak.OpenIdFullNameProtocolMapper, data *schema.ResourceData) {
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
}

func resourceKeycloakOpenIdFullNameProtocolMapperCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	openIdFullNameMapper := mapFromDataToOpenIdFullNameProtocolMapper(data)

	err := keycloakClient.ValidateOpenIdFullNameProtocolMapper(ctx, openIdFullNameMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.NewOpenIdFullNameProtocolMapper(ctx, openIdFullNameMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	mapFromOpenIdFullNameMapperToData(openIdFullNameMapper, data)

	return resourceKeycloakOpenIdFullNameProtocolMapperRead(ctx, data, meta)
}

func resourceKeycloakOpenIdFullNameProtocolMapperRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)
	clientScopeId := data.Get("client_scope_id").(string)

	openIdFullNameMapper, err := keycloakClient.GetOpenIdFullNameProtocolMapper(ctx, realmId, clientId, clientScopeId, data.Id())
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	mapFromOpenIdFullNameMapperToData(openIdFullNameMapper, data)

	return nil
}

func resourceKeycloakOpenIdFullNameProtocolMapperUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	openIdFullNameMapper := mapFromDataToOpenIdFullNameProtocolMapper(data)

	err := keycloakClient.ValidateOpenIdFullNameProtocolMapper(ctx, openIdFullNameMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.UpdateOpenIdFullNameProtocolMapper(ctx, openIdFullNameMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceKeycloakOpenIdFullNameProtocolMapperRead(ctx, data, meta)
}

func resourceKeycloakOpenIdFullNameProtocolMapperDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)
	clientScopeId := data.Get("client_scope_id").(string)

	return diag.FromErr(keycloakClient.DeleteOpenIdFullNameProtocolMapper(ctx, realmId, clientId, clientScopeId, data.Id()))
}
