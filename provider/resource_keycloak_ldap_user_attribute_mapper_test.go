package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakLdapUserAttributeMapper_basic(t *testing.T) {
	t.Parallel()

	userAttributeMapperName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapUserAttributeMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakLdapUserAttributeMapper_basic(userAttributeMapperName),
				Check:  testAccCheckKeycloakLdapUserAttributeMapperExists("keycloak_ldap_user_attribute_mapper.username"),
			},
			{
				ResourceName:      "keycloak_ldap_user_attribute_mapper.username",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getLdapGenericMapperImportId("keycloak_ldap_user_attribute_mapper.username"),
			},
		},
	})
}

func TestAccKeycloakLdapUserAttributeMapper_createAfterManualDestroy(t *testing.T) {
	t.Parallel()

	var mapper = &keycloak.LdapUserAttributeMapper{}

	userAttributeMapperName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapUserAttributeMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakLdapUserAttributeMapper_basic(userAttributeMapperName),
				Check:  testAccCheckKeycloakLdapUserAttributeMapperFetch("keycloak_ldap_user_attribute_mapper.username", mapper),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteLdapUserAttributeMapper(testCtx, mapper.RealmId, mapper.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakLdapUserAttributeMapper_basic(userAttributeMapperName),
				Check:  testAccCheckKeycloakLdapUserAttributeMapperExists("keycloak_ldap_user_attribute_mapper.username"),
			},
		},
	})
}

func TestAccKeycloakLdapUserAttributeMapper_updateLdapUserFederation(t *testing.T) {
	t.Parallel()

	userAttributeMapperName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapUserAttributeMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakLdapUserAttributeMapper_updateLdapUserFederationBefore(userAttributeMapperName),
				Check:  testAccCheckKeycloakLdapUserAttributeMapperExists("keycloak_ldap_user_attribute_mapper.username"),
			},
			{
				Config: testKeycloakLdapUserAttributeMapper_updateLdapUserFederationAfter(userAttributeMapperName),
				Check:  testAccCheckKeycloakLdapUserAttributeMapperExists("keycloak_ldap_user_attribute_mapper.username"),
			},
		},
	})
}

func TestAccKeycloakLdapUserAttributeMapper_updateInPlace(t *testing.T) {
	t.Parallel()

	userAttributeMapperBefore := &keycloak.LdapUserAttributeMapper{
		Name:                    acctest.RandString(10),
		UserModelAttribute:      acctest.RandString(10),
		LdapAttribute:           acctest.RandString(10),
		IsMandatoryInLdap:       randomBool(),
		ReadOnly:                randomBool(),
		AlwaysReadValueFromLdap: true,
		AttributeDefaultValue:   acctest.RandString(10),
		IsBinaryAttribute:       randomBool(),
	}
	userAttributeMapperAfter := &keycloak.LdapUserAttributeMapper{
		Name:                    acctest.RandString(10),
		UserModelAttribute:      acctest.RandString(10),
		LdapAttribute:           acctest.RandString(10),
		IsMandatoryInLdap:       randomBool(),
		ReadOnly:                randomBool(),
		AlwaysReadValueFromLdap: randomBool(),
		AttributeDefaultValue:   acctest.RandString(10),
		IsBinaryAttribute:       false,
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapUserAttributeMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakLdapUserAttributeMapper_basicFromInterface(userAttributeMapperBefore),
				Check:  testAccCheckKeycloakLdapUserAttributeMapperExists("keycloak_ldap_user_attribute_mapper.username"),
			},
			{
				Config: testKeycloakLdapUserAttributeMapper_basicFromInterface(userAttributeMapperAfter),
				Check:  testAccCheckKeycloakLdapUserAttributeMapperExists("keycloak_ldap_user_attribute_mapper.username"),
			},
		},
	})
}

func testAccCheckKeycloakLdapUserAttributeMapperExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getLdapUserAttributeMapperFromState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckKeycloakLdapUserAttributeMapperFetch(resourceName string, mapper *keycloak.LdapUserAttributeMapper) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedMapper, err := getLdapUserAttributeMapperFromState(s, resourceName)
		if err != nil {
			return err
		}

		mapper.Id = fetchedMapper.Id
		mapper.RealmId = fetchedMapper.RealmId

		return nil
	}
}

func testAccCheckKeycloakLdapUserAttributeMapperDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_ldap_user_attribute_mapper" {
				continue
			}

			id := rs.Primary.ID
			realm := rs.Primary.Attributes["realm_id"]

			ldapUserAttributeMapper, _ := keycloakClient.GetLdapUserAttributeMapper(testCtx, realm, id)
			if ldapUserAttributeMapper != nil {
				return fmt.Errorf("ldap user attribute mapper with id %s still exists", id)
			}
		}

		return nil
	}
}

func getLdapUserAttributeMapperFromState(s *terraform.State, resourceName string) (*keycloak.LdapUserAttributeMapper, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm_id"]

	ldapUserAttributeMapper, err := keycloakClient.GetLdapUserAttributeMapper(testCtx, realm, id)
	if err != nil {
		return nil, fmt.Errorf("error getting ldap user attribute mapper with id %s: %s", id, err)
	}

	return ldapUserAttributeMapper, nil
}

func testKeycloakLdapUserAttributeMapper_basic(userAttributeMapperName string) string {
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

resource "keycloak_ldap_user_attribute_mapper" "username" {
	name                        = "%s"
	realm_id                    = data.keycloak_realm.realm.id
	ldap_user_federation_id     = "${keycloak_ldap_user_federation.openldap.id}"

	user_model_attribute        = "username"
	ldap_attribute              = "cn"
}
	`, testAccRealmUserFederation.Realm, userAttributeMapperName)
}

func testKeycloakLdapUserAttributeMapper_basicFromInterface(mapper *keycloak.LdapUserAttributeMapper) string {
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

resource "keycloak_ldap_user_attribute_mapper" "username" {
	name                        = "%s"
	realm_id                    = data.keycloak_realm.realm.id
	ldap_user_federation_id     = "${keycloak_ldap_user_federation.openldap.id}"

	user_model_attribute        = "%s"
	ldap_attribute              = "%s"

	read_only                   = %t
	always_read_value_from_ldap = %t
	is_mandatory_in_ldap        = %t
	attribute_default_value     = "%s"
	is_binary_attribute         = %t
}
	`, testAccRealmUserFederation.Realm, mapper.Name, mapper.UserModelAttribute, mapper.LdapAttribute, mapper.ReadOnly, mapper.AlwaysReadValueFromLdap, mapper.IsMandatoryInLdap,
		mapper.AttributeDefaultValue, mapper.IsBinaryAttribute)
}

func testKeycloakLdapUserAttributeMapper_updateLdapUserFederationBefore(userAttributeMapperName string) string {
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

resource "keycloak_ldap_user_attribute_mapper" "username" {
	name                        = "%s"
	realm_id                    = data.keycloak_realm.realm_one.id
	ldap_user_federation_id     = "${keycloak_ldap_user_federation.openldap_one.id}"

	user_model_attribute        = "username"
	ldap_attribute              = "cn"
}
	`, testAccRealmUserFederation.Realm, testAccRealmTwo.Realm, userAttributeMapperName)
}

func testKeycloakLdapUserAttributeMapper_updateLdapUserFederationAfter(userAttributeMapperName string) string {
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

resource "keycloak_ldap_user_attribute_mapper" "username" {
	name                        = "%s"
	realm_id                    = data.keycloak_realm.realm_two.id
	ldap_user_federation_id     = "${keycloak_ldap_user_federation.openldap_two.id}"

	user_model_attribute        = "username"
	ldap_attribute              = "cn"
}
	`, testAccRealmUserFederation.Realm, testAccRealmTwo.Realm, userAttributeMapperName)
}
