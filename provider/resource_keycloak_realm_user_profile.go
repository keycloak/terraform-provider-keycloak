package provider

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakRealmUserProfile() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakRealmUserProfileCreate,
		ReadContext:   resourceKeycloakRealmUserProfileRead,
		DeleteContext: resourceKeycloakRealmUserProfileDelete,
		UpdateContext: resourceKeycloakRealmUserProfileUpdate,
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"attribute": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"display_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"group": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"enabled_when_scope": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"required_for_roles": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"required_for_scopes": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"permissions": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"view": {
										Type:     schema.TypeSet,
										Set:      schema.HashString,
										Required: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"edit": {
										Type:     schema.TypeSet,
										Set:      schema.HashString,
										Required: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"validator": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"config": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"annotations": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"group": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"display_header": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"display_description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"annotations": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

func getRealmUserProfileAttributeFromData(m map[string]interface{}) *keycloak.RealmUserProfileAttribute {
	attribute := &keycloak.RealmUserProfileAttribute{
		Name:        m["name"].(string),
		DisplayName: m["display_name"].(string),
		Group:       m["group"].(string),
	}

	if v, ok := m["permissions"]; ok && len(v.([]interface{})) > 0 {
		permissions := keycloak.RealmUserProfilePermissions{
			Edit: make([]string, 0),
			View: make([]string, 0),
		}

		permissionsConfig := v.([]interface{})[0].(map[string]interface{})

		if v, ok := permissionsConfig["view"]; ok {
			permView := make([]string, 0)
			for _, perm := range v.(*schema.Set).List() {
				permView = append(permView, perm.(string))
			}
			permissions.View = permView
		}

		if v, ok := permissionsConfig["edit"]; ok {
			permEdit := make([]string, 0)
			for _, perm := range v.(*schema.Set).List() {
				permEdit = append(permEdit, perm.(string))
			}
			permissions.Edit = permEdit
		}

		attribute.Permissions = &permissions
	}

	if v, ok := m["enabled_when_scope"]; ok && len(interfaceSliceToStringSlice(v.(*schema.Set).List())) != 0 {
		attribute.Selector = &keycloak.RealmUserProfileSelector{
			Scopes: interfaceSliceToStringSlice(v.(*schema.Set).List()),
		}
	}

	if v, ok := m["validator"]; ok {
		validations := make(map[string]keycloak.RealmUserProfileValidationConfig)

		for _, validator := range v.(*schema.Set).List() {
			validationConfig := validator.(map[string]interface{})

			name := validationConfig["name"].(string)

			if name == "" {
				continue
			}

			config := make(map[string]interface{})
			if v, ok := validationConfig["config"]; ok {
				for key, value := range v.(map[string]interface{}) {
					if strings.HasPrefix(value.(string), "[") {
						t := make([]interface{}, 0)
						json.Unmarshal([]byte(value.(string)), &t)

						config[key] = t
					} else {
						config[key] = value
					}
				}
			}

			validations[name] = config
		}

		attribute.Validations = validations
	}

	required := &keycloak.RealmUserProfileRequired{}

	if v, ok := m["required_for_roles"]; ok {
		required.Roles = interfaceSliceToStringSlice(v.(*schema.Set).List())
	}
	if v, ok := m["required_for_scopes"]; ok {
		required.Scopes = interfaceSliceToStringSlice(v.(*schema.Set).List())
	}

	if len(required.Roles) != 0 || len(required.Scopes) != 0 {
		attribute.Required = required
	}

	if v, ok := m["annotations"]; ok {
		annotations := make(map[string]interface{})

		for key, value := range v.(map[string]interface{}) {

			if strings.HasPrefix(value.(string), "{") {
				var t interface{}
				json.Unmarshal([]byte(value.(string)), &t)
				annotations[key] = t
			} else {
				annotations[key] = value
			}
		}
		attribute.Annotations = annotations
	}

	return attribute

}

func getRealmUserProfileAttributesFromData(lst []interface{}) []*keycloak.RealmUserProfileAttribute {
	attributes := make([]*keycloak.RealmUserProfileAttribute, 0)

	for _, m := range lst {
		userProfileAttribute := getRealmUserProfileAttributeFromData(m.(map[string]interface{}))
		if userProfileAttribute.Name != "" {
			attributes = append(attributes, userProfileAttribute)
		}
	}

	return attributes
}

func getRealmUserProfileGroupFromData(m map[string]interface{}) *keycloak.RealmUserProfileGroup {
	group := keycloak.RealmUserProfileGroup{
		DisplayDescription: m["display_description"].(string),
		DisplayHeader:      m["display_header"].(string),
		Name:               m["name"].(string),
	}

	if v, ok := m["annotations"]; ok {
		annotations := make(map[string]interface{})

		for key, value := range v.(map[string]interface{}) {
			if strings.HasPrefix(value.(string), "{") {
				var t interface{}
				json.Unmarshal([]byte(value.(string)), &t)

				annotations[key] = t
			} else {
				annotations[key] = value
			}
		}

		group.Annotations = annotations
	}

	return &group

}
func getRealmUserProfileGroupsFromData(lst []interface{}) []*keycloak.RealmUserProfileGroup {
	groups := make([]*keycloak.RealmUserProfileGroup, 0)

	for _, m := range lst {
		userProfileGroup := getRealmUserProfileGroupFromData(m.(map[string]interface{}))
		if userProfileGroup.Name != "" {
			groups = append(groups, userProfileGroup)
		}
	}

	return groups
}

func getRealmUserProfileFromData(data *schema.ResourceData) *keycloak.RealmUserProfile {
	realmUserProfile := &keycloak.RealmUserProfile{}

	realmUserProfile.Attributes = getRealmUserProfileAttributesFromData(data.Get("attribute").([]interface{}))
	realmUserProfile.Groups = getRealmUserProfileGroupsFromData(data.Get("group").(*schema.Set).List())

	return realmUserProfile
}

func getRealmUserProfileAttributeData(attr *keycloak.RealmUserProfileAttribute) map[string]interface{} {
	attributeData := make(map[string]interface{})

	attributeData["name"] = attr.Name

	attributeData["display_name"] = attr.DisplayName
	attributeData["group"] = attr.Group
	if attr.Selector != nil && len(attr.Selector.Scopes) != 0 {
		attributeData["enabled_when_scope"] = attr.Selector.Scopes
	}

	attributeData["required_for_roles"] = make([]string, 0)
	attributeData["required_for_scopes"] = make([]string, 0)
	if attr.Required != nil {
		attributeData["required_for_roles"] = attr.Required.Roles
		attributeData["required_for_scopes"] = attr.Required.Scopes
	}

	if attr.Permissions != nil {
		permission := make(map[string]interface{})

		permission["edit"] = attr.Permissions.Edit
		permission["view"] = attr.Permissions.View

		attributeData["permissions"] = []interface{}{permission}
	}

	if attr.Validations != nil {
		validations := make([]interface{}, 0)
		for name, config := range attr.Validations {
			validator := make(map[string]interface{})

			validator["name"] = name

			c := make(map[string]interface{})
			for k, v := range config {
				if _, ok := v.([]interface{}); ok {
					t, _ := json.Marshal(v)
					c[k] = string(t)
				} else {
					c[k] = v
				}
			}

			validator["config"] = c

			validations = append(validations, validator)
		}
		attributeData["validator"] = validations
	}

	if attr.Annotations != nil {
		annotations := make(map[string]interface{})

		for k, v := range attr.Annotations {
			if _, ok := v.(map[string]interface{}); ok {
				t, _ := json.Marshal(v)
				annotations[k] = string(t)
			} else {
				annotations[k] = v
			}
		}

		attributeData["annotations"] = annotations
	}

	return attributeData
}

func getRealmUserProfileGroupData(group *keycloak.RealmUserProfileGroup) map[string]interface{} {
	groupData := make(map[string]interface{})

	groupData["name"] = group.Name
	groupData["display_header"] = group.DisplayHeader
	groupData["display_description"] = group.DisplayDescription

	annotations := make(map[string]interface{})

	for k, v := range group.Annotations {
		if _, ok := v.(map[string]interface{}); ok {
			t, _ := json.Marshal(v)
			annotations[k] = string(t)
		} else {
			annotations[k] = v
		}
	}

	groupData["annotations"] = annotations

	return groupData
}

func setRealmUserProfileData(data *schema.ResourceData, realmUserProfile *keycloak.RealmUserProfile) {
	attributes := make([]interface{}, 0)
	for _, attr := range realmUserProfile.Attributes {
		attributes = append(attributes, getRealmUserProfileAttributeData(attr))
	}
	data.Set("attribute", attributes)

	groups := make([]interface{}, 0)
	for _, group := range realmUserProfile.Groups {
		groups = append(groups, getRealmUserProfileGroupData(group))
	}
	data.Set("group", groups)
}

func resourceKeycloakRealmUserProfileCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	realmId := data.Get("realm_id").(string)
	data.SetId(realmId)

	realmUserProfile := getRealmUserProfileFromData(data)

	err := keycloakClient.UpdateRealmUserProfile(ctx, realmId, realmUserProfile)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceKeycloakRealmUserProfileRead(ctx, data, meta)
}

func resourceKeycloakRealmUserProfileRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)

	realmUserProfile, err := keycloakClient.GetRealmUserProfile(ctx, realmId)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	setRealmUserProfileData(data, realmUserProfile)

	return nil
}

func resourceKeycloakRealmUserProfileDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	realmId := data.Get("realm_id").(string)

	// The realm user profile cannot be deleted, so instead we set it back to its "zero" values.
	realmUserProfile := &keycloak.RealmUserProfile{
		Attributes: []*keycloak.RealmUserProfileAttribute{},
		Groups:     []*keycloak.RealmUserProfileGroup{},
	}

	err := keycloakClient.UpdateRealmUserProfile(ctx, realmId, realmUserProfile)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakRealmUserProfileUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	realmUserProfile := getRealmUserProfileFromData(data)

	err := keycloakClient.UpdateRealmUserProfile(ctx, realmId, realmUserProfile)
	if err != nil {
		return diag.FromErr(err)
	}

	setRealmUserProfileData(data, realmUserProfile)

	return nil
}
