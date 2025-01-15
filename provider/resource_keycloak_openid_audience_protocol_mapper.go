package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakOpenIdAudienceProtocolMapper() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakOpenIdAudienceProtocolMapperCreate,
		ReadContext:   resourceKeycloakOpenIdAudienceProtocolMapperRead,
		UpdateContext: resourceKeycloakOpenIdAudienceProtocolMapperUpdate,
		DeleteContext: resourceKeycloakOpenIdAudienceProtocolMapperDelete,
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
			"included_client_audience": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "A client ID to include within the token's `aud` claim. Cannot be used with included_custom_audience",
				ConflictsWith: []string{"included_custom_audience"},
			},
			"included_custom_audience": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "A custom audience to include within the token's `aud` claim.  Cannot be used with included_custom_audience",
				ConflictsWith: []string{"included_client_audience"},
			},
			"add_to_id_token": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Indicates if this claim should be added to the id token.",
			},
			"add_to_access_token": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Indicates if this claim should be added to the access token.",
			},
		},
	}
}

func mapFromDataToOpenIdAudienceProtocolMapper(data *schema.ResourceData) *keycloak.OpenIdAudienceProtocolMapper {
	return &keycloak.OpenIdAudienceProtocolMapper{
		Id:            data.Id(),
		Name:          data.Get("name").(string),
		RealmId:       data.Get("realm_id").(string),
		ClientId:      data.Get("client_id").(string),
		ClientScopeId: data.Get("client_scope_id").(string),

		AddToIdToken:     data.Get("add_to_id_token").(bool),
		AddToAccessToken: data.Get("add_to_access_token").(bool),

		IncludedClientAudience: data.Get("included_client_audience").(string),
		IncludedCustomAudience: data.Get("included_custom_audience").(string),
	}
}

func mapFromOpenIdAudienceMapperToData(mapper *keycloak.OpenIdAudienceProtocolMapper, data *schema.ResourceData) {
	data.SetId(mapper.Id)
	data.Set("name", mapper.Name)
	data.Set("realm_id", mapper.RealmId)

	if mapper.ClientId != "" {
		data.Set("client_id", mapper.ClientId)
	} else {
		data.Set("client_scope_id", mapper.ClientScopeId)
	}

	if mapper.IncludedClientAudience != "" {
		data.Set("included_client_audience", mapper.IncludedClientAudience)
	} else {
		data.Set("included_custom_audience", mapper.IncludedCustomAudience)
	}

	data.Set("add_to_id_token", mapper.AddToIdToken)
	data.Set("add_to_access_token", mapper.AddToAccessToken)
}

func resourceKeycloakOpenIdAudienceProtocolMapperCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	openIdAudienceMapper := mapFromDataToOpenIdAudienceProtocolMapper(data)

	err := keycloakClient.ValidateOpenIdAudienceProtocolMapper(ctx, openIdAudienceMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.NewOpenIdAudienceProtocolMapper(ctx, openIdAudienceMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	mapFromOpenIdAudienceMapperToData(openIdAudienceMapper, data)

	return resourceKeycloakOpenIdAudienceProtocolMapperRead(ctx, data, meta)
}

func resourceKeycloakOpenIdAudienceProtocolMapperRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)
	clientScopeId := data.Get("client_scope_id").(string)

	openIdAudienceMapper, err := keycloakClient.GetOpenIdAudienceProtocolMapper(ctx, realmId, clientId, clientScopeId, data.Id())
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	mapFromOpenIdAudienceMapperToData(openIdAudienceMapper, data)

	return nil
}

func resourceKeycloakOpenIdAudienceProtocolMapperUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	openIdAudienceMapper := mapFromDataToOpenIdAudienceProtocolMapper(data)

	err := keycloakClient.ValidateOpenIdAudienceProtocolMapper(ctx, openIdAudienceMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.UpdateOpenIdAudienceProtocolMapper(ctx, openIdAudienceMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceKeycloakOpenIdAudienceProtocolMapperRead(ctx, data, meta)
}

func resourceKeycloakOpenIdAudienceProtocolMapperDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)
	clientScopeId := data.Get("client_scope_id").(string)

	return diag.FromErr(keycloakClient.DeleteOpenIdAudienceProtocolMapper(ctx, realmId, clientId, clientScopeId, data.Id()))
}
