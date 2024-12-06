package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakAuthenticationBindings() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakAuthenticationBindingsCreate,
		ReadContext:   resourceKeycloakAuthenticationBindingsRead,
		DeleteContext: resourceKeycloakAuthenticationBindingsDelete,
		UpdateContext: resourceKeycloakAuthenticationBindingsUpdate,
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"browser_flow": {
				Type:        schema.TypeString,
				Description: "Which flow should be used for BrowserFlow",
				Optional:    true,
				Computed:    true,
			},
			"registration_flow": {
				Type:        schema.TypeString,
				Description: "Which flow should be used for RegistrationFlow",
				Optional:    true,
				Computed:    true,
			},
			"direct_grant_flow": {
				Type:        schema.TypeString,
				Description: "Which flow should be used for DirectGrantFlow",
				Optional:    true,
				Computed:    true,
			},
			"reset_credentials_flow": {
				Type:        schema.TypeString,
				Description: "Which flow should be used for ResetCredentialsFlow",
				Optional:    true,
				Computed:    true,
			},
			"client_authentication_flow": {
				Type:        schema.TypeString,
				Description: "Which flow should be used for ClientAuthenticationFlow",
				Optional:    true,
				Computed:    true,
			},
			"docker_authentication_flow": {
				Type:        schema.TypeString,
				Description: "Which flow should be used for DockerAuthenticationFlow",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func setAuthenticationBindingsData(data *schema.ResourceData, realm *keycloak.Realm) {
	data.SetId(realm.Realm)
	data.Set("browser_flow", realm.BrowserFlow)
	data.Set("registration_flow", realm.RegistrationFlow)
	data.Set("direct_grant_flow", realm.DirectGrantFlow)
	data.Set("reset_credentials_flow", realm.ResetCredentialsFlow)
	data.Set("client_authentication_flow", realm.ClientAuthenticationFlow)
	data.Set("docker_authentication_flow", realm.DockerAuthenticationFlow)
}

func resetAuthenticationBindingsForRealm(realm *keycloak.Realm) {
	realm.BrowserFlow = stringPointer("browser")
	realm.RegistrationFlow = stringPointer("registration")
	realm.DirectGrantFlow = stringPointer("direct grant")
	realm.ResetCredentialsFlow = stringPointer("reset credentials")
	realm.ClientAuthenticationFlow = stringPointer("clients")
	realm.DockerAuthenticationFlow = stringPointer("docker auth")
}

func resourceKeycloakAuthenticationBindingsCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realm, err := keycloakClient.GetRealm(ctx, data.Get("realm_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	setRealmFlowBindings(data, realm)

	err = keycloakClient.ValidateRealm(ctx, realm)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.UpdateRealm(ctx, realm)
	if err != nil {
		return diag.FromErr(err)
	}

	realm, err = keycloakClient.GetRealm(ctx, realm.Realm)
	if err != nil {
		return diag.FromErr(err)
	}

	setAuthenticationBindingsData(data, realm)

	return resourceKeycloakAuthenticationBindingsRead(ctx, data, meta)
}

func resourceKeycloakAuthenticationBindingsRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realm, err := keycloakClient.GetRealm(ctx, data.Id())
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	setAuthenticationBindingsData(data, realm)

	return nil
}

func resourceKeycloakAuthenticationBindingsDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realm, err := keycloakClient.GetRealm(ctx, data.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	resetAuthenticationBindingsForRealm(realm)

	err = keycloakClient.UpdateRealm(ctx, realm)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakAuthenticationBindingsUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realm, err := keycloakClient.GetRealm(ctx, data.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	setRealmFlowBindings(data, realm)

	err = keycloakClient.ValidateRealm(ctx, realm)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.UpdateRealm(ctx, realm)
	if err != nil {
		return diag.FromErr(err)
	}

	setAuthenticationBindingsData(data, realm)

	return resourceKeycloakAuthenticationBindingsRead(ctx, data, meta)
}
