package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakOrganization() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakOrganizationCreate,
		ReadContext:   resourceKeycloakOrganizationRead,
		DeleteContext: resourceKeycloakOrganizationDelete,
		UpdateContext: resourceKeycloakOrganizationUpdate,
		// This resource can be imported using {{realm}}/{{client_id}}. The Client ID is displayed in the GUI
		Importer: &schema.ResourceImporter{
			StateContext: resourceKeycloakOrganizationImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the organization.",
			},
			"realm": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Realm ID.",
			},
			"alias": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The alias unique identifies the organization. Same as the name if not specified.",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Enable/disable this organization.",
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"redirect_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Landing page after successful login.",
			},
			"domain": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"verified": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"attributes": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func getOrganizationFromData(data *schema.ResourceData) (*keycloak.Organization, error) {
	organization := &keycloak.Organization{
		Id:          data.Id(),
		Realm:       data.Get("realm").(string),
		Name:        data.Get("name").(string),
		Alias:       data.Get("alias").(string),
		Enabled:     data.Get("enabled").(bool),
		Description: data.Get("description").(string),
		RedirectUrl: data.Get("redirect_url").(string),
	}

	domains := make([]keycloak.OrganizationDomain, 0)
	if v, ok := data.GetOk("domain"); ok {
		for _, domain := range v.(*schema.Set).List() {
			domainMap := domain.(map[string]interface{})
			orgDomain := keycloak.OrganizationDomain{
				Name:     domainMap["name"].(string),
				Verified: domainMap["verified"].(bool),
			}
			domains = append(domains, orgDomain)
		}
	}

	if len(domains) == 0 {
		return nil, fmt.Errorf("at least one domain is required")
	}
	organization.Domains = domains

	attributes := map[string][]string{}
	if v, ok := data.GetOk("attributes"); ok {
		for key, value := range v.(map[string]interface{}) {
			attributes[key] = strings.Split(value.(string), MULTIVALUE_ATTRIBUTE_SEPARATOR)
		}
	}
	organization.Attributes = attributes

	return organization, nil
}

func setOrganizationData(data *schema.ResourceData, organization *keycloak.Organization) error {
	attributes := map[string]string{}
	for k, v := range organization.Attributes {
		attributes[k] = strings.Join(v, MULTIVALUE_ATTRIBUTE_SEPARATOR)
	}

	domains := make([]map[string]interface{}, 0, len(organization.Domains))
	for _, domain := range organization.Domains {
		domainMap := make(map[string]interface{})
		domainMap["name"] = domain.Name
		domainMap["verified"] = domain.Verified
		domains = append(domains, domainMap)
	}

	data.SetId(organization.Id)
	data.Set("name", organization.Name)
	data.Set("realm", organization.Realm)
	data.Set("alias", organization.Alias)
	data.Set("enabled", organization.Enabled)
	data.Set("description", organization.Description)
	data.Set("redirect_url", organization.RedirectUrl)
	data.Set("domain", domains)
	data.Set("attributes", attributes)

	return nil
}

func resourceKeycloakOrganizationImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	parts := strings.Split(d.Id(), "/")

	if len(parts) != 2 {
		return nil, fmt.Errorf("Invalid import. Supported import formats: {{realm}}/{{OrganizationId}}")
	}

	_, err := keycloakClient.GetOrganization(ctx, parts[0], parts[1])
	if err != nil {
		return nil, err
	}

	d.Set("realm", parts[0])
	d.SetId(parts[1])

	diagnostics := resourceKeycloakOrganizationRead(ctx, d, meta)
	if diagnostics.HasError() {
		return nil, fmt.Errorf("Error reading Organization: %s", diagnostics[0].Summary)
	}

	return []*schema.ResourceData{d}, nil
}

func resourceKeycloakOrganizationCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	Organization, err := getOrganizationFromData(data)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = keycloakClient.NewOrganization(ctx, Organization); err != nil {
		return diag.FromErr(err)
	}
	if err = setOrganizationData(data, Organization); err != nil {
		return diag.FromErr(err)
	}
	return resourceKeycloakOrganizationRead(ctx, data, meta)
}

func resourceKeycloakOrganizationRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	realm := data.Get("realm").(string)
	id := data.Id()
	Organization, err := keycloakClient.GetOrganization(ctx, realm, id)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	return diag.FromErr(setOrganizationData(data, Organization))
}

func resourceKeycloakOrganizationUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	Organization, err := getOrganizationFromData(data)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.UpdateOrganization(ctx, Organization)
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(setOrganizationData(data, Organization))
}

func resourceKeycloakOrganizationDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realm := data.Get("realm").(string)
	id := data.Id()

	return diag.FromErr(keycloakClient.DeleteOrganization(ctx, realm, id))
}
