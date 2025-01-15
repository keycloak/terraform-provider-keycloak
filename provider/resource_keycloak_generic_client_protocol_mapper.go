package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakGenericClientProtocolMapper() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakGenericClientProtocolMapperCreate,
		ReadContext:   resourceKeycloakGenericClientProtocolMapperRead,
		DeleteContext: resourceKeycloakGenericClientProtocolMapperDelete,
		UpdateContext: resourceKeycloakGenericClientProtocolMapperUpdate,
		//  import a mapper tied to a client:
		// {{realmId}}/client/{{clientId}}/{{protocolMapperId}}
		// or a client scope:
		// {{realmId}}/client-scope/{{clientScopeId}}/{{protocolMapperId}}
		Importer: &schema.ResourceImporter{
			StateContext: genericProtocolMapperImport,
		},
		DeprecationMessage: "please use keycloak_generic_protocol_mapper instead",
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
			"protocol": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "The protocol of the client (openid-connect / saml).",
				ValidateFunc: validation.StringInSlice([]string{"openid-connect", "saml"}, false),
			},
			"protocol_mapper": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The type of the protocol mapper.",
			},
			"config": {
				Type:     schema.TypeMap,
				Required: true,
			},
		},
	}
}

func mapFromDataToGenericClientProtocolMapper(data *schema.ResourceData) *keycloak.GenericProtocolMapper {
	config := make(map[string]string)
	if v, ok := data.GetOk("config"); ok {
		for key, value := range v.(map[string]interface{}) {
			config[key] = value.(string)
		}
	}

	return &keycloak.GenericProtocolMapper{
		ClientId:       data.Get("client_id").(string),
		ClientScopeId:  data.Get("client_scope_id").(string),
		Config:         config,
		Id:             data.Id(),
		Name:           data.Get("name").(string),
		Protocol:       data.Get("protocol").(string),
		ProtocolMapper: data.Get("protocol_mapper").(string),
		RealmId:        data.Get("realm_id").(string),
	}
}

func mapFromGenericClientProtocolMapperToData(data *schema.ResourceData, mapper *keycloak.GenericProtocolMapper) {
	data.SetId(mapper.Id)
	if mapper.ClientId != "" {
		data.Set("client_id", mapper.ClientId)
	} else {
		data.Set("client_scope_id", mapper.ClientScopeId)
	}
	data.Set("config", mapper.Config)
	data.Set("name", mapper.Name)
	data.Set("protocol", mapper.Protocol)
	data.Set("protocol_mapper", mapper.ProtocolMapper)
	data.Set("realm_id", mapper.RealmId)
}

func resourceKeycloakGenericClientProtocolMapperCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	genericClientProtocolMapper := mapFromDataToGenericClientProtocolMapper(data)

	err := genericClientProtocolMapper.Validate(ctx, keycloakClient)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.NewGenericProtocolMapper(ctx, genericClientProtocolMapper)
	if err != nil {
		return diag.FromErr(err)
	}
	mapFromGenericClientProtocolMapperToData(data, genericClientProtocolMapper)

	return resourceKeycloakGenericClientProtocolMapperRead(ctx, data, meta)
}

func resourceKeycloakGenericClientProtocolMapperRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)
	clientScopeId := data.Get("client_scope_id").(string)
	id := data.Id()

	resource, err := keycloakClient.GetGenericProtocolMapper(ctx, realmId, clientId, clientScopeId, id)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	mapFromGenericClientProtocolMapperToData(data, resource)

	return nil
}

func resourceKeycloakGenericClientProtocolMapperUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	resource := mapFromDataToGenericClientProtocolMapper(data)

	err := keycloakClient.UpdateGenericProtocolMapper(ctx, resource)
	if err != nil {
		return diag.FromErr(err)
	}

	mapFromGenericClientProtocolMapperToData(data, resource)

	return nil
}

func resourceKeycloakGenericClientProtocolMapperDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)
	clientScopeId := data.Get("client_scope_id").(string)
	id := data.Id()

	return diag.FromErr(keycloakClient.DeleteGenericProtocolMapper(ctx, realmId, clientId, clientScopeId, id))
}
