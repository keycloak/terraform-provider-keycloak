package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

var (
	keycloakLdapUserFederationEditModes             = []string{"READ_ONLY", "WRITABLE", "UNSYNCED"}
	keycloakLdapUserFederationVendors               = []string{"OTHER", "EDIRECTORY", "AD", "RHDS", "TIVOLI"}
	keycloakLdapUserFederationSearchScopes          = []string{"ONE_LEVEL", "SUBTREE"}
	keycloakLdapUserFederationTruststoreSpiSettings = []string{"ALWAYS", "ONLY_FOR_LDAPS", "NEVER"}
	keycloakUserFederationCachePolicies             = []string{"DEFAULT", "EVICT_DAILY", "EVICT_WEEKLY", "MAX_LIFESPAN", "NO_CACHE"}
)

func resourceKeycloakLdapUserFederation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakLdapUserFederationCreate,
		ReadContext:   resourceKeycloakLdapUserFederationRead,
		UpdateContext: resourceKeycloakLdapUserFederationUpdate,
		DeleteContext: resourceKeycloakLdapUserFederationDelete,
		// If this resource uses authentication, then this resource must be imported using the syntax {{realm_id}}/{{provider_id}}/{{bind_credential}}
		// Otherwise, this resource can be imported using {{realm}}/{{provider_id}}.
		// The Provider ID is displayed in the GUI when editing this provider
		Importer: &schema.ResourceImporter{
			StateContext: resourceKeycloakLdapUserFederationImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Display name of the provider when displayed in the console.",
			},
			"realm_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The realm this provider will provide user federation for.",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "When false, this provider will not be used when performing queries for users.",
			},
			"priority": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "Priority of this provider when looking up users. Lower values are first.",
			},
			"import_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "When true, LDAP users will be imported into the Keycloak database.",
			},
			"edit_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "READ_ONLY",
				ValidateFunc: validation.StringInSlice(keycloakLdapUserFederationEditModes, false),
				Description:  "READ_ONLY and WRITABLE are self-explanatory. UNSYNCED allows user data to be imported but not synced back to LDAP.",
			},
			"sync_registrations": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "When true, newly created users will be synced back to LDAP.",
			},
			"vendor": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "OTHER",
				ValidateFunc: validation.StringInSlice(keycloakLdapUserFederationVendors, false),
				Description:  "LDAP vendor. I am almost certain this field does nothing, but the UI indicates that it is required.",
			},
			"username_ldap_attribute": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the LDAP attribute to use as the Keycloak username.",
			},
			"rdn_ldap_attribute": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the LDAP attribute to use as the relative distinguished name.",
			},
			"uuid_ldap_attribute": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the LDAP attribute to use as a unique object identifier for objects in LDAP.",
			},
			"user_object_classes": {
				Type:        schema.TypeList,
				Required:    true,
				MinItems:    1,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "All values of LDAP objectClass attribute for users in LDAP.",
			},
			"connection_url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Connection URL to the LDAP server.",
			},
			"users_dn": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Full DN of LDAP tree where your users are.",
			},
			"bind_dn": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "DN of LDAP admin, which will be used by Keycloak to access LDAP server.",
			},
			"bind_credential": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
				DiffSuppressFunc: func(_, remoteBindCredential, _ string, _ *schema.ResourceData) bool {
					return remoteBindCredential == "**********"
				},
				Description: "Password of LDAP admin.",
			},
			"custom_user_search_filter": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Additional LDAP filter for filtering searched users. Must begin with '(' and end with ')'.",
			},
			"search_scope": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "ONE_LEVEL",
				ValidateFunc: validation.StringInSlice(keycloakLdapUserFederationSearchScopes, false),
				Description:  "ONE_LEVEL: only search for users in the DN specified by user_dn. SUBTREE: search entire LDAP subtree.",
			},

			"start_tls": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "When true, Keycloak will encrypt the connection to LDAP using STARTTLS, which will disable connection pooling.",
			},
			"use_password_modify_extended_op": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "When `true`, use the LDAPv3 Password Modify Extended Operation (RFC-3062).",
			},
			"validate_password_policy": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "When true, Keycloak will validate passwords using the realm policy before updating it.",
			},
			"trust_email": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "If enabled, email provided by this provider is not verified even if verification is enabled for the realm.",
			},
			"use_truststore_spi": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "ONLY_FOR_LDAPS",
				ValidateFunc: validation.StringInSlice(keycloakLdapUserFederationTruststoreSpiSettings, false),
			},
			"connection_timeout": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "LDAP connection timeout (duration string)",
				DiffSuppressFunc: suppressDurationStringDiff,
			},
			"read_timeout": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "LDAP read timeout (duration string)",
				DiffSuppressFunc: suppressDurationStringDiff,
			},
			"pagination": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "When true, Keycloak assumes the LDAP server supports pagination.",
			},

			"batch_size_for_sync": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1000,
				Description: "The number of users to sync within a single transaction.",
			},
			"full_sync_period": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      -1,
				ValidateFunc: validateSyncPeriod,
				Description:  "How frequently Keycloak should sync all LDAP users, in seconds. Omit this property to disable periodic full sync.",
			},
			"changed_sync_period": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      -1,
				ValidateFunc: validateSyncPeriod,
				Description:  "How frequently Keycloak should sync changed LDAP users, in seconds. Omit this property to disable periodic changed users sync.",
			},

			"kerberos": {
				Type:        schema.TypeSet,
				Optional:    true,
				MaxItems:    1,
				Description: "Settings regarding kerberos authentication for this realm.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kerberos_realm": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of the kerberos realm, e.g. FOO.LOCAL",
						},
						"server_principal": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The kerberos server principal, e.g. 'HTTP/host.foo.com@FOO.LOCAL'.",
						},
						"key_tab": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Path to the kerberos keytab file on the server with credentials of the service principal.",
						},
						"use_kerberos_for_password_authentication": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Use kerberos login module instead of ldap service api. Defaults to `false`.",
						},
					},
				},
			},
			"cache": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Settings regarding cache policy for this realm.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"policy": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "DEFAULT",
							ValidateFunc: validation.StringInSlice(keycloakUserFederationCachePolicies, false),
						},
						"max_lifespan": {
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: suppressDurationStringDiff,
							Description:      "Max lifespan of cache entry (duration string).",
						},
						"eviction_day": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      "-1",
							ValidateFunc: validation.All(validation.IntAtLeast(0), validation.IntAtMost(6)),
							Description:  "Day of the week the entry will become invalid on.",
						},
						"eviction_hour": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      "-1",
							ValidateFunc: validation.All(validation.IntAtLeast(0), validation.IntAtMost(23)),
							Description:  "Hour of day the entry will become invalid on.",
						},
						"eviction_minute": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      "-1",
							ValidateFunc: validation.All(validation.IntAtLeast(0), validation.IntAtMost(59)),
							Description:  "Minute of day the entry will become invalid on.",
						},
					},
				},
			},
			"delete_default_mappers": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
				Description: "When true, the provider will delete the default mappers which are normally created by Keycloak when creating an LDAP user federation provider.",
			},
		},
	}
}

func validateSyncPeriod(i interface{}, k string) (s []string, errs []error) {
	num, ok := i.(int)
	if !ok {
		errs = append(errs, fmt.Errorf("expected type of %s to be int", k))
	}

	if num < 1 && num != -1 {
		errs = append(errs, fmt.Errorf("expected %s to be either -1 (disabled), or greater than zero, got %d", k, num))
	}

	return
}

func getLdapUserFederationFromData(data *schema.ResourceData, realmInternalId string) *keycloak.LdapUserFederation {
	var userObjectClasses []string

	for _, userObjectClass := range data.Get("user_object_classes").([]interface{}) {
		userObjectClasses = append(userObjectClasses, userObjectClass.(string))
	}

	ldapUserFederation := &keycloak.LdapUserFederation{
		Id:      data.Id(),
		Name:    data.Get("name").(string),
		RealmId: realmInternalId,

		Enabled:  data.Get("enabled").(bool),
		Priority: data.Get("priority").(int),

		ImportEnabled:     data.Get("import_enabled").(bool),
		EditMode:          data.Get("edit_mode").(string),
		SyncRegistrations: data.Get("sync_registrations").(bool),

		Vendor:                 data.Get("vendor").(string),
		UsernameLDAPAttribute:  data.Get("username_ldap_attribute").(string),
		RdnLDAPAttribute:       data.Get("rdn_ldap_attribute").(string),
		UuidLDAPAttribute:      data.Get("uuid_ldap_attribute").(string),
		UserObjectClasses:      userObjectClasses,
		ConnectionUrl:          data.Get("connection_url").(string),
		UsersDn:                data.Get("users_dn").(string),
		BindDn:                 data.Get("bind_dn").(string),
		BindCredential:         data.Get("bind_credential").(string),
		CustomUserSearchFilter: data.Get("custom_user_search_filter").(string),
		SearchScope:            data.Get("search_scope").(string),

		StartTls:                    data.Get("start_tls").(bool),
		UsePasswordModifyExtendedOp: data.Get("use_password_modify_extended_op").(bool),
		ValidatePasswordPolicy:      data.Get("validate_password_policy").(bool),
		TrustEmail:                  data.Get("trust_email").(bool),
		UseTruststoreSpi:            data.Get("use_truststore_spi").(string),
		ConnectionTimeout:           data.Get("connection_timeout").(string),
		ReadTimeout:                 data.Get("read_timeout").(string),
		Pagination:                  data.Get("pagination").(bool),

		BatchSizeForSync:  data.Get("batch_size_for_sync").(int),
		FullSyncPeriod:    data.Get("full_sync_period").(int),
		ChangedSyncPeriod: data.Get("changed_sync_period").(int),
	}

	if cache, ok := data.GetOk("cache"); ok {
		cache := cache.([]interface{})
		cacheData := cache[0].(map[string]interface{})

		evictionDay := cacheData["eviction_day"].(int)
		evictionHour := cacheData["eviction_hour"].(int)
		evictionMinute := cacheData["eviction_minute"].(int)

		ldapUserFederation.MaxLifespan = cacheData["max_lifespan"].(string)

		ldapUserFederation.EvictionDay = &evictionDay
		ldapUserFederation.EvictionHour = &evictionHour
		ldapUserFederation.EvictionMinute = &evictionMinute
		ldapUserFederation.CachePolicy = cacheData["policy"].(string)
	}

	if kerberos, ok := data.GetOk("kerberos"); ok {
		ldapUserFederation.AllowKerberosAuthentication = true
		kerberosSettingsData := kerberos.(*schema.Set).List()[0]
		kerberosSettings := kerberosSettingsData.(map[string]interface{})

		ldapUserFederation.KerberosRealm = kerberosSettings["kerberos_realm"].(string)
		ldapUserFederation.ServerPrincipal = kerberosSettings["server_principal"].(string)
		ldapUserFederation.UseKerberosForPasswordAuthentication = kerberosSettings["use_kerberos_for_password_authentication"].(bool)
		ldapUserFederation.KeyTab = kerberosSettings["key_tab"].(string)
	} else {
		ldapUserFederation.AllowKerberosAuthentication = false
	}

	return ldapUserFederation
}

func setLdapUserFederationData(data *schema.ResourceData, ldap *keycloak.LdapUserFederation, realmId string) {
	data.SetId(ldap.Id)

	data.Set("name", ldap.Name)
	data.Set("realm_id", realmId)

	data.Set("enabled", ldap.Enabled)
	data.Set("priority", ldap.Priority)

	data.Set("import_enabled", ldap.ImportEnabled)
	data.Set("edit_mode", ldap.EditMode)
	data.Set("sync_registrations", ldap.SyncRegistrations)

	data.Set("vendor", ldap.Vendor)
	data.Set("username_ldap_attribute", ldap.UsernameLDAPAttribute)
	data.Set("rdn_ldap_attribute", ldap.RdnLDAPAttribute)
	data.Set("uuid_ldap_attribute", ldap.UuidLDAPAttribute)
	data.Set("user_object_classes", ldap.UserObjectClasses)
	data.Set("connection_url", ldap.ConnectionUrl)
	data.Set("users_dn", ldap.UsersDn)
	data.Set("bind_dn", ldap.BindDn)
	data.Set("bind_credential", ldap.BindCredential)
	data.Set("custom_user_search_filter", ldap.CustomUserSearchFilter)
	data.Set("search_scope", ldap.SearchScope)

	data.Set("start_tls", ldap.StartTls)
	data.Set("use_password_modify_extended_op", ldap.UsePasswordModifyExtendedOp)
	data.Set("validate_password_policy", ldap.ValidatePasswordPolicy)
	data.Set("trust_email", ldap.TrustEmail)
	data.Set("use_truststore_spi", ldap.UseTruststoreSpi)
	data.Set("connection_timeout", ldap.ConnectionTimeout)
	data.Set("read_timeout", ldap.ReadTimeout)
	data.Set("pagination", ldap.Pagination)

	if ldap.AllowKerberosAuthentication {
		kerberosSettings := make(map[string]interface{})

		kerberosSettings["server_principal"] = ldap.ServerPrincipal
		kerberosSettings["use_kerberos_for_password_authentication"] = ldap.UseKerberosForPasswordAuthentication
		kerberosSettings["key_tab"] = ldap.KeyTab
		kerberosSettings["kerberos_realm"] = ldap.KerberosRealm

		data.Set("kerberos", []interface{}{kerberosSettings})
	} else {
		data.Set("kerberos", nil)
	}

	data.Set("batch_size_for_sync", ldap.BatchSizeForSync)
	data.Set("full_sync_period", ldap.FullSyncPeriod)
	data.Set("changed_sync_period", ldap.ChangedSyncPeriod)

	if _, ok := data.GetOk("cache"); ok {
		cachePolicySettings := make(map[string]interface{})

		if ldap.MaxLifespan != "" {
			cachePolicySettings["max_lifespan"] = ldap.MaxLifespan
		}

		if ldap.EvictionDay != nil {
			cachePolicySettings["eviction_day"] = *ldap.EvictionDay
		}
		if ldap.EvictionHour != nil {
			cachePolicySettings["eviction_hour"] = *ldap.EvictionHour
		}
		if ldap.EvictionMinute != nil {
			cachePolicySettings["eviction_minute"] = *ldap.EvictionMinute
		}

		cachePolicySettings["policy"] = ldap.CachePolicy

		data.Set("cache", []interface{}{cachePolicySettings})
	}
}

func resourceKeycloakLdapUserFederationCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)

	realm, err := keycloakClient.GetRealm(ctx, realmId)
	if err != nil {
		return diag.FromErr(err)
	}

	ldap := getLdapUserFederationFromData(data, realm.Id)

	err = keycloakClient.ValidateLdapUserFederation(ctx, ldap)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.NewLdapUserFederation(ctx, realmId, ldap)
	if err != nil {
		return diag.FromErr(err)
	}

	if data.Get("delete_default_mappers").(bool) {
		err = keycloakClient.DeleteLdapUserFederationMappers(ctx, realmId, ldap.Id)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	setLdapUserFederationData(data, ldap, realmId)

	return resourceKeycloakLdapUserFederationRead(ctx, data, meta)
}

func resourceKeycloakLdapUserFederationRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	ldap, err := keycloakClient.GetLdapUserFederation(ctx, realmId, id)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	ldap.BindCredential = data.Get("bind_credential").(string) // we can't trust the API to set this field correctly since it just responds with "**********"
	setLdapUserFederationData(data, ldap, realmId)

	return nil
}

func resourceKeycloakLdapUserFederationUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)

	realm, err := keycloakClient.GetRealm(ctx, realmId)
	if err != nil {
		return diag.FromErr(err)
	}

	ldap := getLdapUserFederationFromData(data, realm.Id)

	err = keycloakClient.ValidateLdapUserFederation(ctx, ldap)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.UpdateLdapUserFederation(ctx, realmId, ldap)
	if err != nil {
		return diag.FromErr(err)
	}

	setLdapUserFederationData(data, ldap, realmId)

	return nil
}

func resourceKeycloakLdapUserFederationDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	return diag.FromErr(keycloakClient.DeleteLdapUserFederation(ctx, realmId, id))
}

func resourceKeycloakLdapUserFederationImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	parts := strings.Split(d.Id(), "/")

	var realmId, id string
	switch {
	case len(parts) == 2:
		realmId = parts[0]
		id = parts[1]
	case len(parts) == 3:
		realmId = parts[0]
		id = parts[1]
		d.Set("bind_credential", parts[2])
	default:
		return nil, fmt.Errorf("Invalid import. Supported import formats: {{realmId}}/{{userFederationId}}, {{realmId}}/{{userFederationId}}/{{bindCredentials}}")
	}

	_, err := keycloakClient.GetLdapUserFederation(ctx, realmId, id)
	if err != nil {
		return nil, err
	}

	d.Set("realm_id", realmId)
	d.Set("delete_default_mappers", false) // this is only valid on create, so we assume this is false
	d.SetId(id)

	diagnostics := resourceKeycloakLdapUserFederationRead(ctx, d, meta)
	if diagnostics.HasError() {
		return nil, errors.New(diagnostics[0].Summary)
	}

	return []*schema.ResourceData{d}, nil
}
