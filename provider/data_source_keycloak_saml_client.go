package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func dataSourceKeycloakSamlClient() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKeycloakSamlClientRead,

		Schema: map[string]*schema.Schema{
			"client_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"include_authn_statement": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"sign_documents": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"sign_assertions": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"encrypt_assertions": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"client_signature_required": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"signature_key_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"force_post_binding": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"consent_required": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"always_display_in_console" : {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"front_channel_logout": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"force_name_id_format": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"signature_algorithm": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"saml_signature_key_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"canonicalization_method": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name_id_format": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"root_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"valid_redirect_uris": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Computed: true,
			},
			"base_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"master_saml_processing_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"encryption_certificate": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"signing_certificate": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"signing_private_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"encryption_certificate_sha1": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"signing_certificate_sha1": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"signing_private_key_sha1": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"idp_initiated_sso_url_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"idp_initiated_sso_relay_state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"assertion_consumer_post_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"assertion_consumer_redirect_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"logout_service_post_binding_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"logout_service_redirect_binding_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"full_scope_allowed": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"authentication_flow_binding_overrides": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"browser_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"direct_grant_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"extra_config": {
				Type:     schema.TypeMap,
				Computed: true,
			},
			"login_theme": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceKeycloakSamlClientRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)

	client, err := keycloakClient.GetSamlClientByClientId(ctx, realmId, clientId)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	err = mapToDataFromSamlClient(ctx, data, client)
	if err != nil {
		return diag.FromErr(err)
	}
	data.Set("extra_config", client.Attributes.ExtraConfig)

	return nil
}
