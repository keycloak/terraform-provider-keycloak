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

func TestAccKeycloakLdapRoleMapper_basic(t *testing.T) {
	t.Parallel()

	roleMapperName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapRoleMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakLdapRoleMapper_basic(roleMapperName),
				Check:  testAccCheckKeycloakLdapRoleMapperExists("keycloak_ldap_role_mapper.role_mapper"),
			},
			{
				ResourceName:      "keycloak_ldap_role_mapper.role_mapper",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getLdapGenericMapperImportId("keycloak_ldap_role_mapper.role_mapper"),
			},
		},
	})
}

func TestAccKeycloakLdapRoleMapper_createAfterManualDestroy(t *testing.T) {
	t.Parallel()

	var mapper = &keycloak.LdapRoleMapper{}

	roleMapperName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapRoleMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakLdapRoleMapper_basic(roleMapperName),
				Check:  testAccCheckKeycloakLdapRoleMapperFetch("keycloak_ldap_role_mapper.role_mapper", mapper),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteLdapRoleMapper(testCtx, mapper.RealmId, mapper.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakLdapRoleMapper_basic(roleMapperName),
				Check:  testAccCheckKeycloakLdapRoleMapperExists("keycloak_ldap_role_mapper.role_mapper"),
			},
		},
	})
}

func TestAccKeycloakLdapRoleMapper_modeValidation(t *testing.T) {
	t.Parallel()

	roleMapperName := acctest.RandomWithPrefix("tf-acc")
	mode := randomStringInSlice(keycloakLdapRoleMapperModes)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapRoleMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakLdapRoleMapper_basicWithAttrValidation(roleMapperName, "mode", acctest.RandString(10)),
				ExpectError: regexp.MustCompile("expected mode to be one of .+ got .+"),
			},
			{
				Config: testKeycloakLdapRoleMapper_basicWithAttrValidation(roleMapperName, "mode", mode),
				Check:  testAccCheckKeycloakLdapRoleMapperExists("keycloak_ldap_role_mapper.role_mapper"),
			},
		},
	})
}

func TestAccKeycloakLdapRoleMapper_membershipAttributeTypeValidation(t *testing.T) {
	t.Parallel()

	roleMapperName := acctest.RandomWithPrefix("tf-acc")
	membershipAttributeType := randomStringInSlice(keycloakLdapRoleMapperMembershipAttributeTypes)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapRoleMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakLdapRoleMapper_basicWithAttrValidation(roleMapperName, "membership_attribute_type", acctest.RandString(10)),
				ExpectError: regexp.MustCompile("expected membership_attribute_type to be one of .+ got .+"),
			},
			{
				Config: testKeycloakLdapRoleMapper_basicWithAttrValidation(roleMapperName, "membership_attribute_type", membershipAttributeType),
				Check:  testAccCheckKeycloakLdapRoleMapperExists("keycloak_ldap_role_mapper.role_mapper"),
			},
		},
	})
}

func TestAccKeycloakLdapRoleMapper_userRolesRetrieveStrategyValidation(t *testing.T) {
	t.Parallel()

	roleMapperName := acctest.RandomWithPrefix("tf-acc")
	userRolesRetrieveStrategy := randomStringInSlice(keycloakLdapRoleMapperUserRolesRetrieveStrategies)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapRoleMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakLdapRoleMapper_basicWithAttrValidation(roleMapperName, "user_roles_retrieve_strategy", acctest.RandString(10)),
				ExpectError: regexp.MustCompile("expected user_roles_retrieve_strategy to be one of .+ got .+"),
			},
			{
				Config: testKeycloakLdapRoleMapper_basicWithAttrValidation(roleMapperName, "user_roles_retrieve_strategy", userRolesRetrieveStrategy),
				Check:  testAccCheckKeycloakLdapRoleMapperExists("keycloak_ldap_role_mapper.role_mapper"),
			},
		},
	})
}

func TestAccKeycloakLdapRoleMapper_rolesLdapFilterValidation(t *testing.T) {
	t.Parallel()

	roleMapperName := acctest.RandomWithPrefix("tf-acc")
	rolesLdapFilter := "(" + acctest.RandString(10) + ")"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapRoleMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakLdapRoleMapper_basicWithAttrValidation(roleMapperName, "roles_ldap_filter", acctest.RandString(10)),
				ExpectError: regexp.MustCompile(`validation error: roles ldap filter must start with '\(' and end with '\)'`),
			},
			{
				Config: testKeycloakLdapRoleMapper_basicWithAttrValidation(roleMapperName, "roles_ldap_filter", rolesLdapFilter),
				Check:  testAccCheckKeycloakLdapRoleMapperExists("keycloak_ldap_role_mapper.role_mapper"),
			},
		},
	})
}

func TestAccKeycloakLdapRoleMapper_updateLdapUserFederationForceNew(t *testing.T) {
	t.Parallel()

	roleMapperName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapRoleMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakLdapRoleMapper_updateLdapUserFederationBefore(roleMapperName),
				Check:  testAccCheckKeycloakLdapRoleMapperExists("keycloak_ldap_role_mapper.role_mapper"),
			},
			{
				Config: testKeycloakLdapRoleMapper_updateLdapUserFederationAfter(roleMapperName),
				Check:  testAccCheckKeycloakLdapRoleMapperExists("keycloak_ldap_role_mapper.role_mapper"),
			},
		},
	})
}

func TestAccKeycloakLdapRoleMapper_updateLdapUserFederationInPlace(t *testing.T) {
	t.Parallel()

	clientId := acctest.RandomWithPrefix("tf-acc")
	useRealmRolesMapping := randomBool()

	roleMapperOne := &keycloak.LdapRoleMapper{
		Name:                        acctest.RandString(10),
		RealmId:                     testAccRealmUserFederation.Realm,
		LdapRolesDn:                 acctest.RandString(10),
		RoleNameLdapAttribute:       acctest.RandString(10),
		RoleObjectClasses:           []string{acctest.RandString(10), acctest.RandString(10)},
		MembershipLdapAttribute:     acctest.RandString(10),
		MembershipAttributeType:     "DN",
		MembershipUserLdapAttribute: acctest.RandString(10),
		RolesLdapFilter:             "(" + acctest.RandString(10) + ")",
		Mode:                        randomStringInSlice(keycloakLdapRoleMapperModes),
		UserRolesRetrieveStrategy:   randomStringInSlice(keycloakLdapRoleMapperUserRolesRetrieveStrategies),
		MemberofLdapAttribute:       acctest.RandString(10),
		UseRealmRolesMapping:        useRealmRolesMapping,
		ClientId:                    clientId,
	}

	roleMapperTwo := &keycloak.LdapRoleMapper{
		Name:                        acctest.RandString(10),
		RealmId:                     testAccRealmUserFederation.Realm,
		LdapRolesDn:                 acctest.RandString(10),
		RoleNameLdapAttribute:       acctest.RandString(10),
		RoleObjectClasses:           []string{acctest.RandString(10), acctest.RandString(10), acctest.RandString(10)},
		MembershipLdapAttribute:     acctest.RandString(10),
		MembershipAttributeType:     randomStringInSlice(keycloakLdapRoleMapperMembershipAttributeTypes),
		MembershipUserLdapAttribute: acctest.RandString(10),
		RolesLdapFilter:             "(" + acctest.RandString(10) + ")",
		Mode:                        randomStringInSlice(keycloakLdapRoleMapperModes),
		UserRolesRetrieveStrategy:   randomStringInSlice(keycloakLdapRoleMapperUserRolesRetrieveStrategies),
		MemberofLdapAttribute:       acctest.RandString(10),
		UseRealmRolesMapping:        !useRealmRolesMapping,
		ClientId:                    clientId,
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapRoleMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakLdapRoleMapper_basicFromInterface(roleMapperOne),
				Check:  testAccCheckKeycloakLdapRoleMapperExists("keycloak_ldap_role_mapper.role_mapper"),
			},
			{
				Config: testKeycloakLdapRoleMapper_basicFromInterface(roleMapperTwo),
				Check:  testAccCheckKeycloakLdapRoleMapperExists("keycloak_ldap_role_mapper.role_mapper"),
			},
		},
	})
}

func testAccCheckKeycloakLdapRoleMapperExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getLdapRoleMapperFromState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckKeycloakLdapRoleMapperFetch(resourceName string, mapper *keycloak.LdapRoleMapper) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedMapper, err := getLdapRoleMapperFromState(s, resourceName)
		if err != nil {
			return err
		}

		mapper.Id = fetchedMapper.Id
		mapper.RealmId = fetchedMapper.RealmId

		return nil
	}
}

func testAccCheckKeycloakLdapRoleMapperDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_ldap_role_mapper" {
				continue
			}

			id := rs.Primary.ID
			realm := rs.Primary.Attributes["realm_id"]

			ldapRoleMapper, _ := keycloakClient.GetLdapRoleMapper(testCtx, realm, id)
			if ldapRoleMapper != nil {
				return fmt.Errorf("ldap role mapper with id %s still exists", id)
			}
		}

		return nil
	}
}

func getLdapRoleMapperFromState(s *terraform.State, resourceName string) (*keycloak.LdapRoleMapper, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm_id"]

	ldapRoleMapper, err := keycloakClient.GetLdapRoleMapper(testCtx, realm, id)
	if err != nil {
		return nil, fmt.Errorf("error getting ldap role mapper with id %s: %s", id, err)
	}

	return ldapRoleMapper, nil
}

func testKeycloakLdapRoleMapper_basic(roleMapperName string) string {
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

resource "keycloak_ldap_role_mapper" "role_mapper" {
	name                    = "%s"
	realm_id                = data.keycloak_realm.realm.id
	ldap_user_federation_id = "${keycloak_ldap_user_federation.openldap.id}"

	ldap_roles_dn                 = "dc=example,dc=org"
        role_name_ldap_attribute      = "cn"
        role_object_classes           = [
                "group"
        ]
        membership_ldap_attribute      = "member"
        membership_user_ldap_attribute = "sAMAccountName"
        memberof_ldap_attribute        = "memberOf"
}
	`, testAccRealmUserFederation.Realm, roleMapperName)
}

func testKeycloakLdapRoleMapper_basicWithAttrValidation(roleMapperName, attr, val string) string {
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

resource "keycloak_ldap_role_mapper" "role_mapper" {
	name                        = "%s"
	realm_id                    = data.keycloak_realm.realm.id
	ldap_user_federation_id     = "${keycloak_ldap_user_federation.openldap.id}"

	%s                          = "%s"

	ldap_roles_dn                 = "dc=example,dc=org"
	role_name_ldap_attribute      = "cn"
	role_object_classes           = [
		"group"
	]
	membership_ldap_attribute      = "member"
	membership_user_ldap_attribute = "sAMAccountName"
	memberof_ldap_attribute        = "memberOf"
}
	`, testAccRealmUserFederation.Realm, roleMapperName, attr, val)
}

func testKeycloakLdapRoleMapper_basicFromInterface(mapper *keycloak.LdapRoleMapper) string {
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

resource "keycloak_ldap_role_mapper" "role_mapper" {
	name                        = "%s"
	realm_id                    = data.keycloak_realm.realm.id
	ldap_user_federation_id     = "${keycloak_ldap_user_federation.openldap.id}"

	ldap_roles_dn                  = "%s"
	role_name_ldap_attribute       = "%s"
	role_object_classes            = %s
	membership_ldap_attribute      = "%s"
	membership_attribute_type      = "%s"
	membership_user_ldap_attribute = "%s"
	roles_ldap_filter              = "%s"
	mode                           = "%s"
	user_roles_retrieve_strategy   = "%s"
	memberof_ldap_attribute        = "%s"
	use_realm_roles_mapping        = %t
	client_id                      = "%s"
}
	`, testAccRealmUserFederation.Realm, mapper.Name, mapper.LdapRolesDn, mapper.RoleNameLdapAttribute, arrayOfStringsForTerraformResource(mapper.RoleObjectClasses), mapper.MembershipLdapAttribute, mapper.MembershipAttributeType, mapper.MembershipUserLdapAttribute, mapper.RolesLdapFilter, mapper.Mode, mapper.UserRolesRetrieveStrategy, mapper.MemberofLdapAttribute, mapper.UseRealmRolesMapping, mapper.ClientId)
}

func testKeycloakLdapRoleMapper_updateLdapUserFederationBefore(roleMapperName string) string {
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

resource "keycloak_ldap_role_mapper" "role_mapper" {
	name                        = "%s"
	realm_id                    = data.keycloak_realm.realm_one.id
	ldap_user_federation_id     = "${keycloak_ldap_user_federation.openldap_one.id}"

	ldap_roles_dn                 = "dc=example,dc=org"
	role_name_ldap_attribute      = "cn"
	role_object_classes           = [
		"group"
	]
	membership_attribute_type      = "DN"
	membership_ldap_attribute      = "member"
	membership_user_ldap_attribute = "sAMAccountName"
	memberof_ldap_attribute        = "memberOf"
}
	`, testAccRealmUserFederation.Realm, testAccRealmTwo.Realm, roleMapperName)
}

func testKeycloakLdapRoleMapper_updateLdapUserFederationAfter(roleMapperName string) string {
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

resource "keycloak_ldap_role_mapper" "role_mapper" {
	name                        = "%s"
	realm_id                    = data.keycloak_realm.realm_two.id
	ldap_user_federation_id     = "${keycloak_ldap_user_federation.openldap_two.id}"

	ldap_roles_dn                 = "dc=example,dc=org"
	role_name_ldap_attribute      = "cn"
	role_object_classes           = [
		"group"
	]
	membership_attribute_type      = "DN"
	membership_ldap_attribute      = "member"
	membership_user_ldap_attribute = "sAMAccountName"
	memberof_ldap_attribute        = "memberOf"
}
	`, testAccRealmUserFederation.Realm, testAccRealmTwo.Realm, roleMapperName)
}
