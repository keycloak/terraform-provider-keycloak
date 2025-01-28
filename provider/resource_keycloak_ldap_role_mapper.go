package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"regexp"
)

var (
	keycloakLdapRoleMapperModes                       = []string{"READ_ONLY", "LDAP_ONLY", "IMPORT"}
	keycloakLdapRoleMapperMembershipAttributeTypes    = []string{"DN", "UID"}
	keycloakLdapRoleMapperUserRolesRetrieveStrategies = []string{"LOAD_ROLES_BY_MEMBER_ATTRIBUTE", "GET_ROLES_FROM_USER_MEMBEROF_ATTRIBUTE", "LOAD_ROLES_BY_MEMBER_ATTRIBUTE_RECURSIVELY"}
)

func resourceKeycloakLdapRoleMapper() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakLdapRoleMapperCreate,
		ReadContext:   resourceKeycloakLdapRoleMapperRead,
		UpdateContext: resourceKeycloakLdapRoleMapperUpdate,
		DeleteContext: resourceKeycloakLdapRoleMapperDelete,
		// This resource can be imported using {{realm}}/{{provider_id}}/{{mapper_id}}. The Provider and Mapper IDs are displayed in the GUI
		Importer: &schema.ResourceImporter{
			StateContext: resourceKeycloakLdapGenericMapperImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Display name of the mapper when displayed in the console.",
			},
			"realm_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The realm in which the ldap user federation provider exists.",
			},
			"ldap_user_federation_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ldap user federation provider to attach this mapper to.",
			},
			"ldap_roles_dn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"role_name_ldap_attribute": {
				Type:     schema.TypeString,
				Required: true,
			},
			"role_object_classes": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"membership_ldap_attribute": {
				Type:     schema.TypeString,
				Required: true,
			},
			"membership_attribute_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "DN",
				ValidateFunc: validation.StringInSlice(keycloakLdapRoleMapperMembershipAttributeTypes, false),
			},
			"membership_user_ldap_attribute": {
				Type:     schema.TypeString,
				Required: true,
			},
			"roles_ldap_filter": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`\(.+\)`), "validation error: roles ldap filter must start with '(' and end with ')'"),
			},
			"mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "READ_ONLY",
				ValidateFunc: validation.StringInSlice(keycloakLdapRoleMapperModes, false),
			},
			"user_roles_retrieve_strategy": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "LOAD_ROLES_BY_MEMBER_ATTRIBUTE",
				ValidateFunc: validation.StringInSlice(keycloakLdapRoleMapperUserRolesRetrieveStrategies, false),
			},
			"memberof_ldap_attribute": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "memberOf",
			},
			"use_realm_roles_mapping": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"client_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func getLdapRoleMapperFromData(data *schema.ResourceData) *keycloak.LdapRoleMapper {
	var roleObjectClasses []string

	for _, roleObjectClass := range data.Get("role_object_classes").([]interface{}) {
		roleObjectClasses = append(roleObjectClasses, roleObjectClass.(string))
	}

	return &keycloak.LdapRoleMapper{
		Id:                   data.Id(),
		Name:                 data.Get("name").(string),
		RealmId:              data.Get("realm_id").(string),
		LdapUserFederationId: data.Get("ldap_user_federation_id").(string),

		LdapRolesDn:                 data.Get("ldap_roles_dn").(string),
		RoleNameLdapAttribute:       data.Get("role_name_ldap_attribute").(string),
		RoleObjectClasses:           roleObjectClasses,
		MembershipLdapAttribute:     data.Get("membership_ldap_attribute").(string),
		MembershipAttributeType:     data.Get("membership_attribute_type").(string),
		MembershipUserLdapAttribute: data.Get("membership_user_ldap_attribute").(string),
		RolesLdapFilter:             data.Get("roles_ldap_filter").(string),
		Mode:                        data.Get("mode").(string),
		UserRolesRetrieveStrategy:   data.Get("user_roles_retrieve_strategy").(string),
		MemberofLdapAttribute:       data.Get("memberof_ldap_attribute").(string),
		UseRealmRolesMapping:        data.Get("use_realm_roles_mapping").(bool),
		ClientId:                    data.Get("client_id").(string),
	}
}

func setLdapRoleMapperData(data *schema.ResourceData, ldapRoleMapper *keycloak.LdapRoleMapper) {
	data.SetId(ldapRoleMapper.Id)

	data.Set("name", ldapRoleMapper.Name)
	data.Set("realm_id", ldapRoleMapper.RealmId)
	data.Set("ldap_user_federation_id", ldapRoleMapper.LdapUserFederationId)

	data.Set("ldap_roles_dn", ldapRoleMapper.LdapRolesDn)
	data.Set("role_name_ldap_attribute", ldapRoleMapper.RoleNameLdapAttribute)
	data.Set("role_object_classes", ldapRoleMapper.RoleObjectClasses)
	data.Set("membership_ldap_attribute", ldapRoleMapper.MembershipLdapAttribute)
	data.Set("membership_attribute_type", ldapRoleMapper.MembershipAttributeType)
	data.Set("membership_user_ldap_attribute", ldapRoleMapper.MembershipUserLdapAttribute)
	data.Set("roles_ldap_filter", ldapRoleMapper.RolesLdapFilter)
	data.Set("mode", ldapRoleMapper.Mode)
	data.Set("user_roles_retrieve_strategy", ldapRoleMapper.UserRolesRetrieveStrategy)
	data.Set("memberof_ldap_attribute", ldapRoleMapper.MemberofLdapAttribute)
	data.Set("use_realm_roles_mapping", ldapRoleMapper.UseRealmRolesMapping)
	data.Set("client_id", ldapRoleMapper.ClientId)
}

func resourceKeycloakLdapRoleMapperCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	ldapRoleMapper := getLdapRoleMapperFromData(data)

	err := keycloakClient.NewLdapRoleMapper(ctx, ldapRoleMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	setLdapRoleMapperData(data, ldapRoleMapper)

	return resourceKeycloakLdapRoleMapperRead(ctx, data, meta)
}

func resourceKeycloakLdapRoleMapperRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	ldapRoleMapper, err := keycloakClient.GetLdapRoleMapper(ctx, realmId, id)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	setLdapRoleMapperData(data, ldapRoleMapper)

	return nil
}

func resourceKeycloakLdapRoleMapperUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	ldapRoleMapper := getLdapRoleMapperFromData(data)

	err := keycloakClient.UpdateLdapRoleMapper(ctx, ldapRoleMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	setLdapRoleMapperData(data, ldapRoleMapper)

	return nil
}

func resourceKeycloakLdapRoleMapperDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	return diag.FromErr(keycloakClient.DeleteLdapRoleMapper(ctx, realmId, id))
}
