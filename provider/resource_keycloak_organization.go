package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
		CustomizeDiff: validateKeycloakOrganizationIdentityProviders(),
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
				Optional: true,
				MinItems: 0,
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
						"identity_provider_alias": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The alias of an identity provider in an identity_provider block that users whose email matches this domain authenticate through.",
						},
						"auto_redirect": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Redirect users whose email matches this domain to the identity provider without showing the login form.",
						},
					},
				},
			},
			"identity_provider": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "The identity providers that this resource links to the organization. Links made elsewhere are kept. When set, the resource also manages the routing of the domains. Requires Keycloak 26.8 or later.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"alias": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The alias of the identity provider.",
						},
						"auto_membership": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Add users who authenticate through the identity provider as members of the organization.",
						},
						"membership_type": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "UNMANAGED",
							ValidateFunc: validation.StringInSlice([]string{"MANAGED", "UNMANAGED"}, false),
							Description:  "The membership type of users added by auto_membership.",
						},
					},
				},
			},
			"attributes": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
				// ignore ordering of multi-valued attributes
				DiffSuppressFunc: suppressDiffForMultivalueAttributeOrder(),
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
				Name:                  domainMap["name"].(string),
				Verified:              domainMap["verified"].(bool),
				IdentityProviderAlias: domainMap["identity_provider_alias"].(string),
				AutoRedirect:          domainMap["auto_redirect"].(bool),
			}
			domains = append(domains, orgDomain)
		}
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

// validateKeycloakOrganizationIdentityProviders refuses at plan time what Keycloak refuses at apply
// time: a domain routed to an identity provider that is not linked to the organization, and a
// MANAGED link without auto membership. It also refuses an alias in two identity_provider blocks: the
// set keeps both when their settings differ, but Keycloak has one link per alias, so the resource would
// never converge.
func validateKeycloakOrganizationIdentityProviders() schema.CustomizeDiffFunc {
	return func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
		// An alias that is not known yet, such as a computed one, skips the checks. Keycloak then
		// refuses the same mistakes at apply time.
		if !d.NewValueKnown("identity_provider") || !d.NewValueKnown("domain") {
			return nil
		}

		identityProviderAliases := make(map[string]bool)
		for _, identityProvider := range getOrganizationIdentityProvidersFromSet(d.Get("identity_provider").(*schema.Set)) {
			if identityProviderAliases[identityProvider.Alias] {
				return fmt.Errorf("identity provider \"%s\" is in more than one identity_provider block", identityProvider.Alias)
			}
			if identityProvider.MembershipType == "MANAGED" && !identityProvider.AutoMembership {
				return fmt.Errorf("identity provider \"%s\": membership_type MANAGED requires auto_membership", identityProvider.Alias)
			}
			identityProviderAliases[identityProvider.Alias] = true
		}

		for _, domain := range d.Get("domain").(*schema.Set).List() {
			domainMap := domain.(map[string]interface{})
			name := domainMap["name"].(string)
			identityProviderAlias := domainMap["identity_provider_alias"].(string)
			if identityProviderAlias != "" && !identityProviderAliases[identityProviderAlias] {
				return fmt.Errorf("domain \"%s\" routes to identity provider \"%s\", which is not in an identity_provider block", name, identityProviderAlias)
			}
			if domainMap["auto_redirect"].(bool) && identityProviderAlias == "" {
				return fmt.Errorf("auto_redirect on domain \"%s\" requires identity_provider_alias", name)
			}
		}

		return nil
	}
}

func getOrganizationIdentityProvidersFromData(data *schema.ResourceData) []keycloak.OrganizationIdentityProvider {
	return getOrganizationIdentityProvidersFromSet(data.Get("identity_provider").(*schema.Set))
}

func getOrganizationIdentityProvidersFromSet(set *schema.Set) []keycloak.OrganizationIdentityProvider {
	identityProviders := make([]keycloak.OrganizationIdentityProvider, 0, set.Len())
	for _, identityProvider := range set.List() {
		identityProviderMap := identityProvider.(map[string]interface{})
		identityProviders = append(identityProviders, keycloak.OrganizationIdentityProvider{
			Alias:          identityProviderMap["alias"].(string),
			AutoMembership: identityProviderMap["auto_membership"].(bool),
			MembershipType: identityProviderMap["membership_type"].(string),
		})
	}

	return identityProviders
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
		domainMap["identity_provider_alias"] = domain.IdentityProviderAlias
		domainMap["auto_redirect"] = domain.AutoRedirect
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

func setOrganizationIdentityProvidersData(data *schema.ResourceData, identityProviders []keycloak.OrganizationIdentityProvider) error {
	identityProviderMaps := make([]map[string]interface{}, 0, len(identityProviders))
	for _, identityProvider := range identityProviders {
		identityProviderMaps = append(identityProviderMaps, map[string]interface{}{
			"alias":           identityProvider.Alias,
			"auto_membership": identityProvider.AutoMembership,
			"membership_type": identityProvider.MembershipType,
		})
	}

	return data.Set("identity_provider", identityProviderMaps)
}

func withoutDomainRouting(domains []keycloak.OrganizationDomain) []keycloak.OrganizationDomain {
	unrouted := make([]keycloak.OrganizationDomain, 0, len(domains))
	for _, domain := range domains {
		unrouted = append(unrouted, keycloak.OrganizationDomain{
			Name:     domain.Name,
			Verified: domain.Verified,
		})
	}

	return unrouted
}

func checkOrganizationIdentityProvidersSupported(ctx context.Context, keycloakClient *keycloak.KeycloakClient, organization *keycloak.Organization) error {
	ok, err := keycloakClient.VersionIsLessThan(ctx, keycloak.Version_26_8)
	if err != nil {
		return err
	}
	if ok {
		return fmt.Errorf("identity_provider in organization \"%s\" is not supported by your Keycloak version (requires >= %s)", organization.Name, keycloak.Version_26_8)
	}

	return nil
}

// removedOrganizationIdentityProviderAliases lists the identity providers that the resource had and no
// longer has. Links made elsewhere, such as by an identity provider resource with organization_id,
// are not in the list, so they are kept.
func removedOrganizationIdentityProviderAliases(identityProviders, oldIdentityProviders []keycloak.OrganizationIdentityProvider) []string {
	wanted := make(map[string]bool, len(identityProviders))
	for _, identityProvider := range identityProviders {
		wanted[identityProvider.Alias] = true
	}

	removed := make([]string, 0)
	for _, identityProvider := range oldIdentityProviders {
		if !wanted[identityProvider.Alias] {
			removed = append(removed, identityProvider.Alias)
		}
	}

	return removed
}

// organizationIdentityProvidersAfterFailedUnlink is the identity_provider value to keep in state when an
// update could not remove some links: the planned links, and the old ones that were not removed. The SDK
// otherwise saves the planned value, and a refresh reads only the links in state, so a link that failed
// to be removed would never be removed again.
func organizationIdentityProvidersAfterFailedUnlink(identityProviders, oldIdentityProviders []keycloak.OrganizationIdentityProvider, remainingAliases []string) []keycloak.OrganizationIdentityProvider {
	remaining := make(map[string]bool, len(remainingAliases))
	for _, alias := range remainingAliases {
		remaining[alias] = true
	}

	kept := append([]keycloak.OrganizationIdentityProvider{}, identityProviders...)
	for _, identityProvider := range oldIdentityProviders {
		if remaining[identityProvider.Alias] {
			kept = append(kept, identityProvider)
		}
	}

	return kept
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

	identityProviders := getOrganizationIdentityProvidersFromData(data)
	if len(identityProviders) > 0 {
		if err = checkOrganizationIdentityProvidersSupported(ctx, keycloakClient, Organization); err != nil {
			return diag.FromErr(err)
		}
	}

	// The domains are created without routing: an identity provider can only be linked to the
	// organization, and a domain routed to it, once the organization exists.
	domains := Organization.Domains
	Organization.Domains = withoutDomainRouting(domains)
	if err = keycloakClient.NewOrganization(ctx, Organization); err != nil {
		return diag.FromErr(err)
	}
	Organization.Domains = domains
	if err = setOrganizationData(data, Organization); err != nil {
		return diag.FromErr(err)
	}

	if len(identityProviders) > 0 {
		if err = keycloakClient.LinkOrganizationIdentityProviders(ctx, Organization, identityProviders); err != nil {
			return diag.FromErr(err)
		}
		if err = keycloakClient.UpdateOrganizationWithDomainRouting(ctx, Organization); err != nil {
			return diag.FromErr(err)
		}
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

	stateIdentityProviders := getOrganizationIdentityProvidersFromData(data)
	identityProviders := make([]keycloak.OrganizationIdentityProvider, 0, len(stateIdentityProviders))
	for _, stateIdentityProvider := range stateIdentityProviders {
		identityProvider, err := keycloakClient.GetOrganizationIdentityProvider(ctx, realm, id, stateIdentityProvider.Alias)
		if err != nil {
			return diag.FromErr(err)
		}
		if identityProvider != nil {
			identityProviders = append(identityProviders, *identityProvider)
		}
	}
	if len(stateIdentityProviders) == 0 {
		// Without identity_provider blocks the routing of the domains belongs to the
		// identity provider resources that set organization_id.
		Organization.Domains = withoutDomainRouting(Organization.Domains)
	}
	if err = setOrganizationIdentityProvidersData(data, identityProviders); err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(setOrganizationData(data, Organization))
}

func resourceKeycloakOrganizationUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	Organization, err := getOrganizationFromData(data)
	if err != nil {
		return diag.FromErr(err)
	}

	identityProviders := getOrganizationIdentityProvidersFromData(data)
	oldIdentityProviderSet, _ := data.GetChange("identity_provider")
	oldIdentityProviders := getOrganizationIdentityProvidersFromSet(oldIdentityProviderSet.(*schema.Set))

	if len(identityProviders) > 0 {
		if err = checkOrganizationIdentityProvidersSupported(ctx, keycloakClient, Organization); err != nil {
			return diag.FromErr(err)
		}
		if err = keycloakClient.LinkOrganizationIdentityProviders(ctx, Organization, identityProviders); err != nil {
			return diag.FromErr(err)
		}
	}
	// The routing stays with this resource for the update that removes its last identity_provider
	// block, so that the update clears the routing before it unlinks.
	if len(identityProviders) > 0 || len(oldIdentityProviders) > 0 {
		err = keycloakClient.UpdateOrganizationWithDomainRouting(ctx, Organization)
	} else {
		err = keycloakClient.UpdateOrganization(ctx, Organization)
	}
	if err != nil {
		return diag.FromErr(err)
	}

	remainingAliases, err := keycloakClient.UnlinkOrganizationIdentityProviders(ctx, Organization, removedOrganizationIdentityProviderAliases(identityProviders, oldIdentityProviders))
	if err != nil {
		diags := diag.FromErr(err)
		if setErr := setOrganizationIdentityProvidersData(data, organizationIdentityProvidersAfterFailedUnlink(identityProviders, oldIdentityProviders, remainingAliases)); setErr != nil {
			diags = append(diags, diag.FromErr(setErr)...)
		}
		return diags
	}

	return diag.FromErr(setOrganizationData(data, Organization))
}

func resourceKeycloakOrganizationDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realm := data.Get("realm").(string)
	id := data.Id()

	return diag.FromErr(keycloakClient.DeleteOrganization(ctx, realm, id))
}
