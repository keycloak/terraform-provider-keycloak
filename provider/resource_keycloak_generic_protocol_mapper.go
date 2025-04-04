package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakGenericProtocolMapper() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakGenericProtocolMapperCreate,
		ReadContext:   resourceKeycloakGenericProtocolMapperRead,
		DeleteContext: resourceKeycloakGenericProtocolMapperDelete,
		UpdateContext: resourceKeycloakGenericProtocolMapperUpdate,
		//  import a mapper tied to a client:
		// {{realmId}}/client/{{clientId}}/{{protocolMapperId}}
		// or a client scope:
		// {{realmId}}/client-scope/{{clientScopeId}}/{{protocolMapperId}}
		Importer: &schema.ResourceImporter{
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

func mapFromDataToGenericProtocolMapper(data *schema.ResourceData) *keycloak.GenericProtocolMapper {
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

func mapFromGenericProtocolMapperToData(data *schema.ResourceData, mapper *keycloak.GenericProtocolMapper) {
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

func resourceKeycloakGenericProtocolMapperCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	genericProtocolMapper := mapFromDataToGenericProtocolMapper(data)

	err := genericProtocolMapper.Validate(ctx, keycloakClient)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.NewGenericProtocolMapper(ctx, genericProtocolMapper)
	if err != nil {
		return diag.FromErr(err)
	}
	mapFromGenericProtocolMapperToData(data, genericProtocolMapper)

	return resourceKeycloakGenericProtocolMapperRead(ctx, data, meta)
}

func resourceKeycloakGenericProtocolMapperRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)
	clientScopeId := data.Get("client_scope_id").(string)
	id := data.Id()

	resource, err := keycloakClient.GetGenericProtocolMapper(ctx, realmId, clientId, clientScopeId, id)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	mapFromGenericProtocolMapperToData(data, resource)

	return nil
}

func resourceKeycloakGenericProtocolMapperUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	resource := mapFromDataToGenericProtocolMapper(data)

	err := keycloakClient.UpdateGenericProtocolMapper(ctx, resource)
	if err != nil {
		return diag.FromErr(err)
	}

	mapFromGenericProtocolMapperToData(data, resource)

	return nil
}

func resourceKeycloakGenericProtocolMapperDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)
	clientScopeId := data.Get("client_scope_id").(string)
	id := data.Id()

	return diag.FromErr(keycloakClient.DeleteGenericProtocolMapper(ctx, realmId, clientId, clientScopeId, id))
}
