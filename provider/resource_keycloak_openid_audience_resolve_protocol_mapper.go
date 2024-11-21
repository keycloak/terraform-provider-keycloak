package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakOpenIdAudienceResolveProtocolMapper() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakOpenIdAudienceResolveProtocolMapperCreate,
		ReadContext:   resourceKeycloakOpenIdAudienceResolveProtocolMapperRead,
		//UpdateContext: resourceKeycloakOpenIdAudienceResolveProtocolMapperUpdate,
		DeleteContext: resourceKeycloakOpenIdAudienceResolveProtocolMapperDelete,
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
				Optional:    true,
				ForceNew:    true,
				Description: "A human-friendly name that will appear in the Keycloak console.",
				Default:     "audience resolve",
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
		},
	}
}

func mapFromDataToOpenIdAudienceResolveProtocolMapper(data *schema.ResourceData) *keycloak.OpenIdAudienceResolveProtocolMapper {
	return &keycloak.OpenIdAudienceResolveProtocolMapper{
		Id:            data.Id(),
		Name:          data.Get("name").(string),
		RealmId:       data.Get("realm_id").(string),
		ClientId:      data.Get("client_id").(string),
		ClientScopeId: data.Get("client_scope_id").(string),
	}
}

func mapFromOpenIdAudienceResolveMapperToData(mapper *keycloak.OpenIdAudienceResolveProtocolMapper, data *schema.ResourceData) {
	data.SetId(mapper.Id)
	data.Set("name", mapper.Name)
	data.Set("realm_id", mapper.RealmId)

	if mapper.ClientId != "" {
		data.Set("client_id", mapper.ClientId)
	} else {
		data.Set("client_scope_id", mapper.ClientScopeId)
	}
}

func resourceKeycloakOpenIdAudienceResolveProtocolMapperCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	openIdAudienceResolveMapper := mapFromDataToOpenIdAudienceResolveProtocolMapper(data)

	err := keycloakClient.ValidateOpenIdAudienceResolveProtocolMapper(ctx, openIdAudienceResolveMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.NewOpenIdAudienceResolveProtocolMapper(ctx, openIdAudienceResolveMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	mapFromOpenIdAudienceResolveMapperToData(openIdAudienceResolveMapper, data)

	return resourceKeycloakOpenIdAudienceResolveProtocolMapperRead(ctx, data, meta)
}

func resourceKeycloakOpenIdAudienceResolveProtocolMapperRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)
	clientScopeId := data.Get("client_scope_id").(string)

	openIdAudienceResolveMapper, err := keycloakClient.GetOpenIdAudienceResolveProtocolMapper(ctx, realmId, clientId, clientScopeId, data.Id())
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	mapFromOpenIdAudienceResolveMapperToData(openIdAudienceResolveMapper, data)

	return nil
}

func resourceKeycloakOpenIdAudienceResolveProtocolMapperDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)
	clientScopeId := data.Get("client_scope_id").(string)

	return diag.FromErr(keycloakClient.DeleteOpenIdAudienceResolveProtocolMapper(ctx, realmId, clientId, clientScopeId, data.Id()))
}
