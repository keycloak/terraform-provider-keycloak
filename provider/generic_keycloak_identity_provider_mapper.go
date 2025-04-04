package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

type identityProviderMapperDataGetterFunc func(ctx context.Context, data *schema.ResourceData, meta interface{}) (*keycloak.IdentityProviderMapper, error)
type identityProviderMapperDataSetterFunc func(data *schema.ResourceData, identityProviderMapper *keycloak.IdentityProviderMapper) error

func resourceKeycloakIdentityProviderMapper() *schema.Resource {
	return &schema.Resource{
		DeleteContext: resourceKeycloakIdentityProviderMapperDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceKeycloakIdentityProviderMapperImport,
		},
		Schema: map[string]*schema.Schema{
			"realm": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Realm Name",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "IDP Mapper Name",
			},
			"identity_provider_alias": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "IDP Alias",
			},
			"extra_config": {
				Type:     schema.TypeMap,
				Optional: true,
			},
		},
	}
}

func getIdentityProviderMapperFromData(data *schema.ResourceData) (*keycloak.IdentityProviderMapper, error) {
	rec := &keycloak.IdentityProviderMapper{
		Id:                    data.Id(),
		Realm:                 data.Get("realm").(string),
		Name:                  data.Get("name").(string),
		IdentityProviderAlias: data.Get("identity_provider_alias").(string),
		Config: &keycloak.IdentityProviderMapperConfig{
			ExtraConfig: getExtraConfigFromData(data),
		},
	}
	return rec, nil
}

func setIdentityProviderMapperData(data *schema.ResourceData, identityProviderMapper *keycloak.IdentityProviderMapper) error {
	data.SetId(identityProviderMapper.Id)
	data.Set("realm", identityProviderMapper.Realm)
	data.Set("name", identityProviderMapper.Name)
	data.Set("identity_provider_alias", identityProviderMapper.IdentityProviderAlias)
	setExtraConfigData(data, identityProviderMapper.Config.ExtraConfig)

	return nil
}

func resourceKeycloakIdentityProviderMapperDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realm := data.Get("realm").(string)
	alias := data.Get("identity_provider_alias").(string)
	id := data.Id()

	return diag.FromErr(keycloakClient.DeleteIdentityProviderMapper(ctx, realm, alias, id))
}

func resourceKeycloakIdentityProviderMapperImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")

	if len(parts) != 3 {
		return nil, fmt.Errorf("Invalid import. Supported import formats: {{realm}}/{{identityProviderAlias}}/{{identityProviderMapperId}}")
	}

	d.Set("realm", parts[0])
	d.Set("identity_provider_alias", parts[1])
	d.SetId(parts[2])

	return []*schema.ResourceData{d}, nil
}

func resourceKeycloakIdentityProviderMapperCreate(getIdentityProviderMapperFromData identityProviderMapperDataGetterFunc, setDataFromIdentityProviderMapper identityProviderMapperDataSetterFunc) func(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return func(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
		keycloakClient := meta.(*keycloak.KeycloakClient)

		identityProvider, err := getIdentityProviderMapperFromData(ctx, data, meta)
		if err != nil {
			return handleNotFoundError(ctx, err, data)
		}

		if identityProvider == nil {
			return diag.Errorf("identity provider with alias %s not found", data.Get("identity_provider_alias").(string))
		}

		if err = keycloakClient.NewIdentityProviderMapper(ctx, identityProvider); err != nil {
			return diag.FromErr(err)
		}

		if err = setDataFromIdentityProviderMapper(data, identityProvider); err != nil {
			return diag.FromErr(err)
		}

		return resourceKeycloakIdentityProviderMapperRead(setDataFromIdentityProviderMapper)(ctx, data, meta)
	}
}

func resourceKeycloakIdentityProviderMapperRead(setDataFromIdentityProviderMapper identityProviderMapperDataSetterFunc) func(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return func(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
		keycloakClient := meta.(*keycloak.KeycloakClient)

		realm := data.Get("realm").(string)
		alias := data.Get("identity_provider_alias").(string)
		id := data.Id()

		identityProvider, err := keycloakClient.GetIdentityProviderMapper(ctx, realm, alias, id)
		if err != nil {
			return handleNotFoundError(ctx, err, data)
		}

		if err = setDataFromIdentityProviderMapper(data, identityProvider); err != nil {
			return diag.FromErr(err)
		}

		return nil
	}
}

func resourceKeycloakIdentityProviderMapperUpdate(getIdentityProviderMapperFromData identityProviderMapperDataGetterFunc, setDataFromIdentityProviderMapper identityProviderMapperDataSetterFunc) func(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return func(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
		keycloakClient := meta.(*keycloak.KeycloakClient)

		identityProvider, err := getIdentityProviderMapperFromData(ctx, data, meta)
		if err != nil {
			return handleNotFoundError(ctx, err, data)
		}

		if err = keycloakClient.UpdateIdentityProviderMapper(ctx, identityProvider); err != nil {
			return diag.FromErr(err)
		}

		if err = setDataFromIdentityProviderMapper(data, identityProvider); err != nil {
			return diag.FromErr(err)
		}

		return nil
	}
}
