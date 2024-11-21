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
	keycloakLdapGroupMapperModes                       = []string{"READ_ONLY", "LDAP_ONLY", "IMPORT"}
	keycloakLdapGroupMapperMembershipAttributeTypes    = []string{"DN", "UID"}
	keycloakLdapGroupMapperUserRolesRetrieveStrategies = []string{"LOAD_GROUPS_BY_MEMBER_ATTRIBUTE", "GET_GROUPS_FROM_USER_MEMBEROF_ATTRIBUTE", "LOAD_GROUPS_BY_MEMBER_ATTRIBUTE_RECURSIVELY"}
)

func resourceKeycloakLdapGroupMapper() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakLdapGroupMapperCreate,
		ReadContext:   resourceKeycloakLdapGroupMapperRead,
		UpdateContext: resourceKeycloakLdapGroupMapperUpdate,
		DeleteContext: resourceKeycloakLdapGroupMapperDelete,
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
			"ldap_groups_dn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"group_name_ldap_attribute": {
				Type:     schema.TypeString,
				Required: true,
			},
			"group_object_classes": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"preserve_group_inheritance": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"ignore_missing_groups": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"membership_ldap_attribute": {
				Type:     schema.TypeString,
				Required: true,
			},
			"membership_attribute_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "DN",
				ValidateFunc: validation.StringInSlice(keycloakLdapGroupMapperMembershipAttributeTypes, false),
			},
			"membership_user_ldap_attribute": {
				Type:     schema.TypeString,
				Required: true,
			},
			"groups_ldap_filter": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`\(.+\)`), "validation error: groups ldap filter must start with '(' and end with ')'"),
			},
			"mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "READ_ONLY",
				ValidateFunc: validation.StringInSlice(keycloakLdapGroupMapperModes, false),
			},
			"user_roles_retrieve_strategy": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "LOAD_GROUPS_BY_MEMBER_ATTRIBUTE",
				ValidateFunc: validation.StringInSlice(keycloakLdapGroupMapperUserRolesRetrieveStrategies, false),
			},
			"memberof_ldap_attribute": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "memberOf",
			},
			"mapped_group_attributes": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"drop_non_existing_groups_during_sync": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"groups_path": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func getLdapGroupMapperFromData(ctx context.Context, keycloakClient *keycloak.KeycloakClient, data *schema.ResourceData) (*keycloak.LdapGroupMapper, error) {
	var groupObjectClasses []string

	for _, groupObjectClass := range data.Get("group_object_classes").([]interface{}) {
		groupObjectClasses = append(groupObjectClasses, groupObjectClass.(string))
	}

	var mappedGroupAttributes []string

	for _, mappedGroupAttribute := range data.Get("mapped_group_attributes").([]interface{}) {
		mappedGroupAttributes = append(mappedGroupAttributes, mappedGroupAttribute.(string))
	}

	mapper := &keycloak.LdapGroupMapper{
		Id:                   data.Id(),
		Name:                 data.Get("name").(string),
		RealmId:              data.Get("realm_id").(string),
		LdapUserFederationId: data.Get("ldap_user_federation_id").(string),

		LdapGroupsDn:                    data.Get("ldap_groups_dn").(string),
		GroupNameLdapAttribute:          data.Get("group_name_ldap_attribute").(string),
		GroupObjectClasses:              groupObjectClasses,
		PreserveGroupInheritance:        data.Get("preserve_group_inheritance").(bool),
		IgnoreMissingGroups:             data.Get("ignore_missing_groups").(bool),
		MembershipLdapAttribute:         data.Get("membership_ldap_attribute").(string),
		MembershipAttributeType:         data.Get("membership_attribute_type").(string),
		MembershipUserLdapAttribute:     data.Get("membership_user_ldap_attribute").(string),
		GroupsLdapFilter:                data.Get("groups_ldap_filter").(string),
		Mode:                            data.Get("mode").(string),
		UserRolesRetrieveStrategy:       data.Get("user_roles_retrieve_strategy").(string),
		MemberofLdapAttribute:           data.Get("memberof_ldap_attribute").(string),
		MappedGroupAttributes:           mappedGroupAttributes,
		DropNonExistingGroupsDuringSync: data.Get("drop_non_existing_groups_during_sync").(bool),
	}
	versionOk, err := keycloakClient.VersionIsGreaterThanOrEqualTo(ctx, keycloak.Version_11)
	if err != nil {
		return nil, err
	}

	if groupsPath, ok := data.GetOk("groups_path"); ok && versionOk {
		mapper.GroupsPath = groupsPath.(string)
	}

	return mapper, nil
}

func setLdapGroupMapperData(ctx context.Context, keycloakClient *keycloak.KeycloakClient, data *schema.ResourceData, ldapGroupMapper *keycloak.LdapGroupMapper) error {
	data.SetId(ldapGroupMapper.Id)

	data.Set("name", ldapGroupMapper.Name)
	data.Set("realm_id", ldapGroupMapper.RealmId)
	data.Set("ldap_user_federation_id", ldapGroupMapper.LdapUserFederationId)

	data.Set("ldap_groups_dn", ldapGroupMapper.LdapGroupsDn)
	data.Set("group_name_ldap_attribute", ldapGroupMapper.GroupNameLdapAttribute)
	data.Set("group_object_classes", ldapGroupMapper.GroupObjectClasses)
	data.Set("preserve_group_inheritance", ldapGroupMapper.PreserveGroupInheritance)
	data.Set("ignore_missing_groups", ldapGroupMapper.IgnoreMissingGroups)
	data.Set("membership_ldap_attribute", ldapGroupMapper.MembershipLdapAttribute)
	data.Set("membership_attribute_type", ldapGroupMapper.MembershipAttributeType)
	data.Set("membership_user_ldap_attribute", ldapGroupMapper.MembershipUserLdapAttribute)
	data.Set("groups_ldap_filter", ldapGroupMapper.GroupsLdapFilter)
	data.Set("mode", ldapGroupMapper.Mode)
	data.Set("user_roles_retrieve_strategy", ldapGroupMapper.UserRolesRetrieveStrategy)
	data.Set("memberof_ldap_attribute", ldapGroupMapper.MemberofLdapAttribute)
	data.Set("mapped_group_attributes", ldapGroupMapper.MappedGroupAttributes)
	data.Set("drop_non_existing_groups_during_sync", ldapGroupMapper.DropNonExistingGroupsDuringSync)

	versionOk, err := keycloakClient.VersionIsGreaterThanOrEqualTo(ctx, keycloak.Version_11)
	if err != nil {
		return err
	}

	if ldapGroupMapper.GroupsPath != "" && versionOk {
		data.Set("groups_path", ldapGroupMapper.GroupsPath)
	}

	return nil
}

func resourceKeycloakLdapGroupMapperCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	ldapGroupMapper, err := getLdapGroupMapperFromData(ctx, keycloakClient, data)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.ValidateLdapGroupMapper(ctx, ldapGroupMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.NewLdapGroupMapper(ctx, ldapGroupMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	err = setLdapGroupMapperData(ctx, keycloakClient, data, ldapGroupMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceKeycloakLdapGroupMapperRead(ctx, data, meta)
}

func resourceKeycloakLdapGroupMapperRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	ldapGroupMapper, err := keycloakClient.GetLdapGroupMapper(ctx, realmId, id)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	err = setLdapGroupMapperData(ctx, keycloakClient, data, ldapGroupMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakLdapGroupMapperUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	ldapGroupMapper, err := getLdapGroupMapperFromData(ctx, keycloakClient, data)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.ValidateLdapGroupMapper(ctx, ldapGroupMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.UpdateLdapGroupMapper(ctx, ldapGroupMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	err = setLdapGroupMapperData(ctx, keycloakClient, data, ldapGroupMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakLdapGroupMapperDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	return diag.FromErr(keycloakClient.DeleteLdapGroupMapper(ctx, realmId, id))
}
