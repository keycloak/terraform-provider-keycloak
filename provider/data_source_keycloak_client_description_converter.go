package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func dataSourceKeycloakClientDescriptionConverter() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKeycloakClientDescriptionConverterRead,

		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"body": {
				Type:     schema.TypeString,
				Required: true,
			},
			"access": {
				Type:     schema.TypeMap,
				Computed: true,
			},
			"admin_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"attributes": {
				Type:     schema.TypeMap,
				Computed: true,
			},
			"authentication_flow_binding_overrides": {
				Type:     schema.TypeMap,
				Computed: true,
			},
			"authorization_services_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"authorization_settings": {
				Type:     schema.TypeMap,
				Computed: true,
			},
			"base_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bearer_only": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"client_authenticator_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"client_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"consent_required": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_client_scopes": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"default_roles": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"direct_access_grants_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"frontchannel_logout": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"full_scope_allowed": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"implicit_flow_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"not_before": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"optional_client_scopes": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"origin": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"protocol": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"protocol_mappers": {
				Type: schema.TypeSet,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"protocol": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"protocol_mapper": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"config": {
							Type:     schema.TypeMap,
							Computed: true,
						},
					},
				},
				Computed: true,
			},
			"public_client": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"redirect_uris": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"registered_nodes": {
				Type:     schema.TypeMap,
				Computed: true,
			},
			"registration_access_token": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"root_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"secret": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_accounts_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"standard_flow_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"surrogate_auth_required": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"web_origins": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
		},
	}
}

func setClientDescriptionConverterData(data *schema.ResourceData, description *keycloak.GenericClientRepresentation) {
	data.SetId(description.ClientId)

	data.Set("access", description.Access)
	data.Set("admin_url", description.AdminUrl)
	data.Set("attributes", description.Attributes)
	data.Set("authentication_flow_binding_overrides", description.AuthenticationFlowBindingOverrides)
	data.Set("authorization_services_enabled", description.AuthorizationServicesEnabled)
	data.Set("authorization_settings", description.AuthorizationSettings)
	data.Set("base_url", description.BaseUrl)
	data.Set("bearer_only", description.BearerOnly)
	data.Set("client_authenticator_type", description.ClientAuthenticatorType)
	data.Set("client_id", description.ClientId)
	data.Set("consent_required", description.ConsentRequired)
	data.Set("default_client_scopes", description.DefaultClientScopes)
	data.Set("default_roles", description.DefaultRoles)
	data.Set("description", description.Description)
	data.Set("direct_access_grants_enabled", description.DirectAccessGrantsEnabled)
	data.Set("enabled", description.Enabled)
	data.Set("frontchannel_logout", description.FrontchannelLogout)
	data.Set("full_scope_allowed", description.FullScopeAllowed)
	data.Set("implicit_flow_enabled", description.ImplicitFlowEnabled)
	data.Set("name", description.Name)
	data.Set("not_before", description.NotBefore)
	data.Set("optional_client_scopes", description.OptionalClientScopes)
	data.Set("origin", description.Origin)
	data.Set("protocol", description.Protocol)
	data.Set("protocol_mappers", description.ProtocolMappers)
	data.Set("public_client", description.PublicClient)
	data.Set("redirect_uris", description.RedirectUris)
	data.Set("registered_nodes", description.RegisteredNodes)
	data.Set("registration_access_token", description.RegistrationAccessToken)
	data.Set("root_url", description.RootUrl)
	data.Set("secret", description.Secret)
	data.Set("service_accounts_enabled", description.ServiceAccountsEnabled)
	data.Set("standard_flow_enabled", description.StandardFlowEnabled)
	data.Set("surrogate_auth_required", description.SurrogateAuthRequired)
	data.Set("web_origins", description.WebOrigins)
}

func dataSourceKeycloakClientDescriptionConverterRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	body := data.Get("body").(string)

	description, err := keycloakClient.NewGenericClientDescription(ctx, realmId, body)

	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	setClientDescriptionConverterData(data, description)

	return nil
}
