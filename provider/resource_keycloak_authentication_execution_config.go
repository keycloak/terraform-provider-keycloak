package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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

// mergeMaskedAuthenticationExecutionConfig returns the config as read from Keycloak, except that every
// value Keycloak masked (secret config properties are returned as "**********" since Keycloak 26.6.7 / 26.7.3)
// is replaced by the value that is already known from the Terraform config or state. Keycloak interprets the
// masked value on update as "keep the stored value", so keeping the known value does not break updates and
// avoids a perpetual diff. A value that is not known at all (e.g. right after an import) is kept masked.
func mergeMaskedAuthenticationExecutionConfig(remote map[string]string, known map[string]interface{}) map[string]string {
	merged := make(map[string]string, len(remote))
	for key, value := range remote {
		if value == keycloakMaskedSecret {
			if knownValue, ok := known[key].(string); ok {
				value = knownValue
			}
		}
		merged[key] = value
	}
	return merged
}

func setAuthenticationExecutionConfigData(data *schema.ResourceData, config *keycloak.AuthenticationExecutionConfig) {
	data.SetId(config.Id)
	data.Set("realm_id", config.RealmId)
	data.Set("execution_id", config.ExecutionId)
	data.Set("alias", config.Alias)
	data.Set("config", mergeMaskedAuthenticationExecutionConfig(config.Config, data.Get("config").(map[string]interface{})))
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
