package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"regexp"
	"testing"
)

func TestAccKeycloakLdapGroupMapper_basic(t *testing.T) {
	t.Parallel()

	groupMapperName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapGroupMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakLdapGroupMapper_basic(groupMapperName),
				Check:  testAccCheckKeycloakLdapGroupMapperExists("keycloak_ldap_group_mapper.group_mapper"),
			},
			{
				ResourceName:      "keycloak_ldap_group_mapper.group_mapper",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getLdapGenericMapperImportId("keycloak_ldap_group_mapper.group_mapper"),
			},
		},
	})
}

func TestAccKeycloakLdapGroupMapper_createAfterManualDestroy(t *testing.T) {
	t.Parallel()

	var mapper = &keycloak.LdapGroupMapper{}

	groupMapperName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapGroupMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakLdapGroupMapper_basic(groupMapperName),
				Check:  testAccCheckKeycloakLdapGroupMapperFetch("keycloak_ldap_group_mapper.group_mapper", mapper),
			},
			{
				PreConfig: func() {

					err := keycloakClient.DeleteLdapGroupMapper(testCtx, mapper.RealmId, mapper.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakLdapGroupMapper_basic(groupMapperName),
				Check:  testAccCheckKeycloakLdapGroupMapperExists("keycloak_ldap_group_mapper.group_mapper"),
			},
		},
	})
}

func TestAccKeycloakLdapGroupMapper_modeValidation(t *testing.T) {
	t.Parallel()

	groupMapperName := acctest.RandomWithPrefix("tf-acc")
	mode := randomStringInSlice(keycloakLdapGroupMapperModes)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapGroupMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakLdapGroupMapper_basicWithAttrValidation(groupMapperName, "mode", acctest.RandString(10)),
				ExpectError: regexp.MustCompile("expected mode to be one of .+ got .+"),
			},
			{
				Config: testKeycloakLdapGroupMapper_basicWithAttrValidation(groupMapperName, "mode", mode),
				Check:  testAccCheckKeycloakLdapGroupMapperExists("keycloak_ldap_group_mapper.group_mapper"),
			},
		},
	})
}

func TestAccKeycloakLdapGroupMapper_membershipAttributeTypeValidation(t *testing.T) {
	t.Parallel()

	groupMapperName := acctest.RandomWithPrefix("tf-acc")
	membershipAttributeType := randomStringInSlice(keycloakLdapGroupMapperMembershipAttributeTypes)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapGroupMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakLdapGroupMapper_basicWithAttrValidation(groupMapperName, "membership_attribute_type", acctest.RandString(10)),
				ExpectError: regexp.MustCompile("expected membership_attribute_type to be one of .+ got .+"),
			},
			{
				Config: testKeycloakLdapGroupMapper_basicWithAttrValidation(groupMapperName, "membership_attribute_type", membershipAttributeType),
				Check:  testAccCheckKeycloakLdapGroupMapperExists("keycloak_ldap_group_mapper.group_mapper"),
			},
		},
	})
}

func TestAccKeycloakLdapGroupMapper_userRolesRetrieveStrategyValidation(t *testing.T) {
	t.Parallel()

	groupMapperName := acctest.RandomWithPrefix("tf-acc")
	userRolesRetrieveStrategy := randomStringInSlice(keycloakLdapGroupMapperUserRolesRetrieveStrategies)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapGroupMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakLdapGroupMapper_basicWithAttrValidation(groupMapperName, "user_roles_retrieve_strategy", acctest.RandString(10)),
				ExpectError: regexp.MustCompile("expected user_roles_retrieve_strategy to be one of .+ got .+"),
			},
			{
				Config: testKeycloakLdapGroupMapper_basicWithAttrValidation(groupMapperName, "user_roles_retrieve_strategy", userRolesRetrieveStrategy),
				Check:  testAccCheckKeycloakLdapGroupMapperExists("keycloak_ldap_group_mapper.group_mapper"),
			},
		},
	})
}

func TestAccKeycloakLdapGroupMapper_groupsLdapFilterValidation(t *testing.T) {
	t.Parallel()

	groupMapperName := acctest.RandomWithPrefix("tf-acc")
	groupsLdapFilter := "(" + acctest.RandString(10) + ")"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapGroupMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakLdapGroupMapper_basicWithAttrValidation(groupMapperName, "groups_ldap_filter", acctest.RandString(10)),
				ExpectError: regexp.MustCompile(`validation error: groups ldap filter must start with '\(' and end with '\)'`),
			},
			{
				Config: testKeycloakLdapGroupMapper_basicWithAttrValidation(groupMapperName, "groups_ldap_filter", groupsLdapFilter),
				Check:  testAccCheckKeycloakLdapGroupMapperExists("keycloak_ldap_group_mapper.group_mapper"),
			},
		},
	})
}

func TestAccKeycloakLdapGroupMapper_groupInheritanceValidation(t *testing.T) {
	t.Parallel()

	groupMapperName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapGroupMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakLdapGroupMapper_groupInheritanceValidation(groupMapperName),
				ExpectError: regexp.MustCompile("validation error: group inheritance cannot be preserved while membership attribute type is UID"),
			},
		},
	})
}

func TestAccKeycloakLdapGroupMapper_updateLdapUserFederationForceNew(t *testing.T) {
	t.Parallel()

	groupMapperName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapGroupMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakLdapGroupMapper_updateLdapUserFederationBefore(groupMapperName),
				Check:  testAccCheckKeycloakLdapGroupMapperExists("keycloak_ldap_group_mapper.group_mapper"),
			},
			{
				Config: testKeycloakLdapGroupMapper_updateLdapUserFederationAfter(groupMapperName),
				Check:  testAccCheckKeycloakLdapGroupMapperExists("keycloak_ldap_group_mapper.group_mapper"),
			},
		},
	})
}

func TestAccKeycloakLdapGroupMapper_updateLdapUserFederationInPlace(t *testing.T) {
	t.Parallel()

	preserveGroupInheritance := true
	ignoreMissingGroups := randomBool()
	dropNonExistingGroupsDuringSync := randomBool()

	groupMapperOne := &keycloak.LdapGroupMapper{
		Name:                            acctest.RandString(10),
		RealmId:                         testAccRealmUserFederation.Realm,
		LdapGroupsDn:                    acctest.RandString(10),
		GroupNameLdapAttribute:          acctest.RandString(10),
		GroupObjectClasses:              []string{acctest.RandString(10), acctest.RandString(10)},
		PreserveGroupInheritance:        preserveGroupInheritance,
		IgnoreMissingGroups:             ignoreMissingGroups,
		MembershipLdapAttribute:         acctest.RandString(10),
		MembershipAttributeType:         "DN",
		MembershipUserLdapAttribute:     acctest.RandString(10),
		GroupsLdapFilter:                "(" + acctest.RandString(10) + ")",
		Mode:                            randomStringInSlice(keycloakLdapGroupMapperModes),
		UserRolesRetrieveStrategy:       randomStringInSlice(keycloakLdapGroupMapperUserRolesRetrieveStrategies),
		MemberofLdapAttribute:           acctest.RandString(10),
		MappedGroupAttributes:           []string{acctest.RandString(10), acctest.RandString(10), acctest.RandString(10)},
		DropNonExistingGroupsDuringSync: dropNonExistingGroupsDuringSync,
	}

	groupMapperTwo := &keycloak.LdapGroupMapper{
		Name:                            acctest.RandString(10),
		RealmId:                         testAccRealmUserFederation.Realm,
		LdapGroupsDn:                    acctest.RandString(10),
		GroupNameLdapAttribute:          acctest.RandString(10),
		GroupObjectClasses:              []string{acctest.RandString(10), acctest.RandString(10), acctest.RandString(10)},
		PreserveGroupInheritance:        !preserveGroupInheritance,
		IgnoreMissingGroups:             !ignoreMissingGroups,
		MembershipLdapAttribute:         acctest.RandString(10),
		MembershipAttributeType:         randomStringInSlice(keycloakLdapGroupMapperMembershipAttributeTypes),
		MembershipUserLdapAttribute:     acctest.RandString(10),
		GroupsLdapFilter:                "(" + acctest.RandString(10) + ")",
		Mode:                            randomStringInSlice(keycloakLdapGroupMapperModes),
		UserRolesRetrieveStrategy:       randomStringInSlice(keycloakLdapGroupMapperUserRolesRetrieveStrategies),
		MemberofLdapAttribute:           acctest.RandString(10),
		MappedGroupAttributes:           []string{acctest.RandString(10)},
		DropNonExistingGroupsDuringSync: !dropNonExistingGroupsDuringSync,
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapGroupMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakLdapGroupMapper_basicFromInterface(groupMapperOne),
				Check:  testAccCheckKeycloakLdapGroupMapperExists("keycloak_ldap_group_mapper.group_mapper"),
			},
			{
				Config: testKeycloakLdapGroupMapper_basicFromInterface(groupMapperTwo),
				Check:  testAccCheckKeycloakLdapGroupMapperExists("keycloak_ldap_group_mapper.group_mapper"),
			},
		},
	})
}

func TestAccKeycloakLdapGroupMapper_groupsPath(t *testing.T) {
	t.Parallel()

	if ok, _ := keycloakClient.VersionIsGreaterThanOrEqualTo(testCtx, keycloak.Version_11); !ok {
		t.Skip()
	}

	groupName := acctest.RandomWithPrefix("tf-acc")
	groupMapperName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapGroupMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakLdapGroupMapper_groupsPath(groupName, groupMapperName),
				Check:  testAccCheckKeycloakLdapGroupMapperExists("keycloak_ldap_group_mapper.group_mapper"),
			},
		},
	})
}

func testAccCheckKeycloakLdapGroupMapperExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getLdapGroupMapperFromState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckKeycloakLdapGroupMapperFetch(resourceName string, mapper *keycloak.LdapGroupMapper) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedMapper, err := getLdapGroupMapperFromState(s, resourceName)
		if err != nil {
			return err
		}

		mapper.Id = fetchedMapper.Id
		mapper.RealmId = fetchedMapper.RealmId

		return nil
	}
}

func testAccCheckKeycloakLdapGroupMapperDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_ldap_group_mapper" {
				continue
			}

			id := rs.Primary.ID
			realm := rs.Primary.Attributes["realm_id"]

			ldapGroupMapper, _ := keycloakClient.GetLdapGroupMapper(testCtx, realm, id)
			if ldapGroupMapper != nil {
				return fmt.Errorf("ldap group mapper with id %s still exists", id)
			}
		}

		return nil
	}
}

func getLdapGroupMapperFromState(s *terraform.State, resourceName string) (*keycloak.LdapGroupMapper, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm_id"]

	ldapGroupMapper, err := keycloakClient.GetLdapGroupMapper(testCtx, realm, id)
	if err != nil {
		return nil, fmt.Errorf("error getting ldap group mapper with id %s: %s", id, err)
	}

	return ldapGroupMapper, nil
}

func testKeycloakLdapGroupMapper_basic(groupMapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_ldap_user_federation" "openldap" {
	name                    = "openldap"
	realm_id                = data.keycloak_realm.realm.id

	enabled                 = true

	username_ldap_attribute = "cn"
	rdn_ldap_attribute      = "cn"
	uuid_ldap_attribute     = "entryDN"
	user_object_classes     = [
		"simpleSecurityObject",
		"organizationalRole"
	]
	connection_url          = "ldap://openldap"
	users_dn                = "dc=example,dc=org"
	bind_dn                 = "cn=admin,dc=example,dc=org"
	bind_credential         = "admin"
}

resource "keycloak_ldap_group_mapper" "group_mapper" {
	name                        = "%s"
	realm_id                    = data.keycloak_realm.realm.id
	ldap_user_federation_id     = "${keycloak_ldap_user_federation.openldap.id}"

	ldap_groups_dn                 = "dc=example,dc=org"
	group_name_ldap_attribute      = "cn"
	group_object_classes           = [
		"groupOfNames"
	]
	membership_attribute_type      = "DN"
	membership_ldap_attribute      = "member"
	membership_user_ldap_attribute = "cn"
	memberof_ldap_attribute        = "memberOf"
}
	`, testAccRealmUserFederation.Realm, groupMapperName)
}

func testKeycloakLdapGroupMapper_basicWithAttrValidation(groupMapperName, attr, val string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_ldap_user_federation" "openldap" {
	name                    = "openldap"
	realm_id                = data.keycloak_realm.realm.id

	enabled                 = true

	username_ldap_attribute = "cn"
	rdn_ldap_attribute      = "cn"
	uuid_ldap_attribute     = "entryDN"
	user_object_classes     = [
		"simpleSecurityObject",
		"organizationalRole"
	]
	connection_url          = "ldap://openldap"
	users_dn                = "dc=example,dc=org"
	bind_dn                 = "cn=admin,dc=example,dc=org"
	bind_credential         = "admin"
}

resource "keycloak_ldap_group_mapper" "group_mapper" {
	name                        = "%s"
	realm_id                    = data.keycloak_realm.realm.id
	ldap_user_federation_id     = "${keycloak_ldap_user_federation.openldap.id}"

	%s                          = "%s"

	ldap_groups_dn                 = "dc=example,dc=org"
	group_name_ldap_attribute      = "cn"
	group_object_classes           = [
		"groupOfNames"
	]
	membership_ldap_attribute      = "member"
	membership_user_ldap_attribute = "cn"
	memberof_ldap_attribute        = "memberOf"
	groups_path                    = "/"
}
	`, testAccRealmUserFederation.Realm, groupMapperName, attr, val)
}

func testKeycloakLdapGroupMapper_groupInheritanceValidation(groupMapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_ldap_user_federation" "openldap" {
	name                    = "openldap"
	realm_id                = data.keycloak_realm.realm.id

	enabled                 = true

	username_ldap_attribute = "cn"
	rdn_ldap_attribute      = "cn"
	uuid_ldap_attribute     = "entryDN"
	user_object_classes     = [
		"simpleSecurityObject",
		"organizationalRole"
	]
	connection_url          = "ldap://openldap"
	users_dn                = "dc=example,dc=org"
	bind_dn                 = "cn=admin,dc=example,dc=org"
	bind_credential         = "admin"
}

resource "keycloak_ldap_group_mapper" "group_mapper" {
	name                        = "%s"
	realm_id                    = data.keycloak_realm.realm.id
	ldap_user_federation_id     = "${keycloak_ldap_user_federation.openldap.id}"

	membership_attribute_type      = "UID"
	preserve_group_inheritance     = true

	ldap_groups_dn                 = "dc=example,dc=org"
	group_name_ldap_attribute      = "cn"
	group_object_classes           = [
		"groupOfNames"
	]
	membership_ldap_attribute      = "member"
	membership_user_ldap_attribute = "cn"
	memberof_ldap_attribute        = "memberOf"
	groups_path                    = "/"
}
	`, testAccRealmUserFederation.Realm, groupMapperName)
}

func testKeycloakLdapGroupMapper_basicFromInterface(mapper *keycloak.LdapGroupMapper) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_ldap_user_federation" "openldap" {
	name                    = "openldap"
	realm_id                = data.keycloak_realm.realm.id

	enabled                 = true

	username_ldap_attribute = "cn"
	rdn_ldap_attribute      = "cn"
	uuid_ldap_attribute     = "entryDN"
	user_object_classes     = [
		"simpleSecurityObject",
		"organizationalRole"
	]
	connection_url          = "ldap://openldap"
	users_dn                = "dc=example,dc=org"
	bind_dn                 = "cn=admin,dc=example,dc=org"
	bind_credential         = "admin"
}

resource "keycloak_ldap_group_mapper" "group_mapper" {
	name                        = "%s"
	realm_id                    = data.keycloak_realm.realm.id
	ldap_user_federation_id     = "${keycloak_ldap_user_federation.openldap.id}"

	ldap_groups_dn                       = "%s"
	group_name_ldap_attribute            = "%s"
	group_object_classes                 = %s
	preserve_group_inheritance           = %t
	ignore_missing_groups                = %t
	membership_ldap_attribute            = "%s"
	membership_attribute_type            = "%s"
	membership_user_ldap_attribute       = "%s"
	groups_ldap_filter                   = "%s"
	mode                                 = "%s"
	user_roles_retrieve_strategy         = "%s"
	memberof_ldap_attribute              = "%s"
	mapped_group_attributes              = %s
	drop_non_existing_groups_during_sync = %t
}
	`, testAccRealmUserFederation.Realm, mapper.Name, mapper.LdapGroupsDn, mapper.GroupNameLdapAttribute, arrayOfStringsForTerraformResource(mapper.GroupObjectClasses), mapper.PreserveGroupInheritance, mapper.IgnoreMissingGroups, mapper.MembershipLdapAttribute, mapper.MembershipAttributeType, mapper.MembershipUserLdapAttribute, mapper.GroupsLdapFilter, mapper.Mode, mapper.UserRolesRetrieveStrategy, mapper.MemberofLdapAttribute, arrayOfStringsForTerraformResource(mapper.MappedGroupAttributes), mapper.DropNonExistingGroupsDuringSync)
}

func testKeycloakLdapGroupMapper_updateLdapUserFederationBefore(groupMapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm_one" {
	realm = "%s"
}

data "keycloak_realm" "realm_two" {
	realm = "%s"
}

resource "keycloak_ldap_user_federation" "openldap_one" {
	name                    = "openldap"
	realm_id                = data.keycloak_realm.realm_one.id

	enabled                 = true

	username_ldap_attribute = "cn"
	rdn_ldap_attribute      = "cn"
	uuid_ldap_attribute     = "entryDN"
	user_object_classes     = [
		"simpleSecurityObject",
		"organizationalRole"
	]
	connection_url          = "ldap://openldap"
	users_dn                = "dc=example,dc=org"
	bind_dn                 = "cn=admin,dc=example,dc=org"
	bind_credential         = "admin"
}

resource "keycloak_ldap_user_federation" "openldap_two" {
	name                    = "openldap"
	realm_id                = data.keycloak_realm.realm_two.id

	enabled                 = true

	username_ldap_attribute = "cn"
	rdn_ldap_attribute      = "cn"
	uuid_ldap_attribute     = "entryDN"
	user_object_classes     = [
		"simpleSecurityObject",
		"organizationalRole"
	]
	connection_url          = "ldap://openldap"
	users_dn                = "dc=example,dc=org"
	bind_dn                 = "cn=admin,dc=example,dc=org"
	bind_credential         = "admin"
}

resource "keycloak_ldap_group_mapper" "group_mapper" {
	name                        = "%s"
	realm_id                    = data.keycloak_realm.realm_one.id
	ldap_user_federation_id     = "${keycloak_ldap_user_federation.openldap_one.id}"

	ldap_groups_dn                 = "dc=example,dc=org"
	group_name_ldap_attribute      = "cn"
	group_object_classes           = [
		"groupOfNames"
	]
	membership_attribute_type      = "DN"
	membership_ldap_attribute      = "member"
	membership_user_ldap_attribute = "cn"
	memberof_ldap_attribute        = "memberOf"
	groups_path                    = "/"
}
	`, testAccRealmUserFederation.Realm, testAccRealmTwo.Realm, groupMapperName)
}

func testKeycloakLdapGroupMapper_updateLdapUserFederationAfter(groupMapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm_one" {
	realm = "%s"
}

data "keycloak_realm" "realm_two" {
	realm = "%s"
}

resource "keycloak_ldap_user_federation" "openldap_one" {
	name                    = "openldap"
	realm_id                = data.keycloak_realm.realm_one.id

	enabled                 = true

	username_ldap_attribute = "cn"
	rdn_ldap_attribute      = "cn"
	uuid_ldap_attribute     = "entryDN"
	user_object_classes     = [
		"simpleSecurityObject",
		"organizationalRole"
	]
	connection_url          = "ldap://openldap"
	users_dn                = "dc=example,dc=org"
	bind_dn                 = "cn=admin,dc=example,dc=org"
	bind_credential         = "admin"
}

resource "keycloak_ldap_user_federation" "openldap_two" {
	name                    = "openldap"
	realm_id                = data.keycloak_realm.realm_two.id

	enabled                 = true

	username_ldap_attribute = "cn"
	rdn_ldap_attribute      = "cn"
	uuid_ldap_attribute     = "entryDN"
	user_object_classes     = [
		"simpleSecurityObject",
		"organizationalRole"
	]
	connection_url          = "ldap://openldap"
	users_dn                = "dc=example,dc=org"
	bind_dn                 = "cn=admin,dc=example,dc=org"
	bind_credential         = "admin"
}

resource "keycloak_ldap_group_mapper" "group_mapper" {
	name                        = "%s"
	realm_id                    = data.keycloak_realm.realm_two.id
	ldap_user_federation_id     = "${keycloak_ldap_user_federation.openldap_two.id}"

	ldap_groups_dn                 = "dc=example,dc=org"
	group_name_ldap_attribute      = "cn"
	group_object_classes           = [
		"groupOfNames"
	]
	membership_attribute_type      = "DN"
	membership_ldap_attribute      = "member"
	membership_user_ldap_attribute = "cn"
	memberof_ldap_attribute        = "memberOf"
}
	`, testAccRealmUserFederation.Realm, testAccRealmTwo.Realm, groupMapperName)
}

func testKeycloakLdapGroupMapper_groupsPath(groupName, groupMapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_group" "group" {
	realm_id = data.keycloak_realm.realm.id
	name = "%s"
}

resource "keycloak_ldap_user_federation" "openldap" {
	name                    = "openldap"
	realm_id                = data.keycloak_realm.realm.id

	enabled                 = true

	username_ldap_attribute = "cn"
	rdn_ldap_attribute      = "cn"
	uuid_ldap_attribute     = "entryDN"
	user_object_classes     = [
		"simpleSecurityObject",
		"organizationalRole"
	]
	connection_url          = "ldap://openldap"
	users_dn                = "dc=example,dc=org"
	bind_dn                 = "cn=admin,dc=example,dc=org"
	bind_credential         = "admin"
}

resource "keycloak_ldap_group_mapper" "group_mapper" {
	name                        = "%s"
	realm_id                    = data.keycloak_realm.realm.id
	ldap_user_federation_id     = keycloak_ldap_user_federation.openldap.id

	ldap_groups_dn                 = "dc=example,dc=org"
	group_name_ldap_attribute      = "cn"
	group_object_classes           = [
		"groupOfNames"
	]
	membership_attribute_type      = "DN"
	membership_ldap_attribute      = "member"
	membership_user_ldap_attribute = "cn"
	memberof_ldap_attribute        = "memberOf"

	groups_path = keycloak_group.group.path
}
	`, testAccRealmUserFederation.Realm, groupName, groupMapperName)
}
