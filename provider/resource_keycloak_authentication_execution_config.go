package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakAuthenticationExecutionConfig() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakAuthenticationExecutionConfigCreate,
		ReadContext:   resourceKeycloakAuthenticationExecutionConfigRead,
		DeleteContext: resourceKeycloakAuthenticationExecutionConfigDelete,
		UpdateContext: resourceKeycloakAuthenticationExecutionConfigUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: resourceKeycloakAuthenticationExecutionConfigImport,
		},
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"execution_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"alias": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"config": {
				Type:     schema.TypeMap,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Required: true,
			},
		},
	}
}

func getAuthenticationExecutionConfigFromData(data *schema.ResourceData) *keycloak.AuthenticationExecutionConfig {
	config := make(map[string]string)
	for key, value := range data.Get("config").(map[string]interface{}) {
		config[key] = value.(string)
	}
	return &keycloak.AuthenticationExecutionConfig{
		Id:          data.Id(),
		RealmId:     data.Get("realm_id").(string),
		ExecutionId: data.Get("execution_id").(string),
		Alias:       data.Get("alias").(string),
		Config:      config,
	}
}

func setAuthenticationExecutionConfigData(data *schema.ResourceData, config *keycloak.AuthenticationExecutionConfig) {
	data.SetId(config.Id)
	data.Set("realm_id", config.RealmId)
	data.Set("execution_id", config.ExecutionId)
	data.Set("alias", config.Alias)
	data.Set("config", config.Config)
}

func resourceKeycloakAuthenticationExecutionConfigCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	config := getAuthenticationExecutionConfigFromData(data)

	id, err := keycloakClient.NewAuthenticationExecutionConfig(ctx, config)
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(id)

	return resourceKeycloakAuthenticationExecutionConfigRead(ctx, data, meta)
}

func resourceKeycloakAuthenticationExecutionConfigRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	config := &keycloak.AuthenticationExecutionConfig{
		RealmId:     data.Get("realm_id").(string),
		ExecutionId: data.Get("execution_id").(string),
		Id:          data.Id(),
	}

	err := keycloakClient.GetAuthenticationExecutionConfig(ctx, config)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	setAuthenticationExecutionConfigData(data, config)

	return nil
}

func resourceKeycloakAuthenticationExecutionConfigUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	config := getAuthenticationExecutionConfigFromData(data)

	err := keycloakClient.UpdateAuthenticationExecutionConfig(ctx, config)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceKeycloakAuthenticationExecutionConfigRead(ctx, data, meta)
}

func resourceKeycloakAuthenticationExecutionConfigDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	config := &keycloak.AuthenticationExecutionConfig{
		RealmId: data.Get("realm_id").(string),
		Id:      data.Id(),
	}

	return diag.FromErr(keycloakClient.DeleteAuthenticationExecutionConfig(ctx, config))
}

func resourceKeycloakAuthenticationExecutionConfigImport(ctx context.Context, data *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	parts := strings.Split(data.Id(), "/")

	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return nil, fmt.Errorf("invalid import. Supported import formats: {{realm}}/{{authenticationExecutionId}}/{{authenticationExecutionConfigId}}")
	}

	err := keycloakClient.GetAuthenticationExecutionConfig(ctx, &keycloak.AuthenticationExecutionConfig{
		RealmId:     parts[0],
		ExecutionId: parts[1],
		Id:          parts[2],
	})
	if err != nil {
		return nil, err
	}

	data.Set("realm_id", parts[0])
	data.Set("execution_id", parts[1])
	data.SetId(parts[2])

	diagnostics := resourceKeycloakAuthenticationExecutionConfigRead(ctx, data, meta)
	if diagnostics.HasError() {
		return nil, errors.New(diagnostics[0].Summary)
	}

	return []*schema.ResourceData{data}, nil
}
