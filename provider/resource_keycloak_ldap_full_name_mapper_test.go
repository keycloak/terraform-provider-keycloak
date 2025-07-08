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

func TestAccKeycloakLdapFullNameMapper_basic(t *testing.T) {
	t.Parallel()

	fullNameMapperName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapFullNameMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakLdapFullNameMapper_basic(fullNameMapperName),
				Check:  testAccCheckKeycloakLdapFullNameMapperExists("keycloak_ldap_full_name_mapper.full_name_mapper"),
			},
			{
				ResourceName:      "keycloak_ldap_full_name_mapper.full_name_mapper",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getLdapGenericMapperImportId("keycloak_ldap_full_name_mapper.full_name_mapper"),
			},
		},
	})
}

func TestAccKeycloakLdapFullNameMapper_createAfterManualDestroy(t *testing.T) {
	t.Parallel()

	var mapper = &keycloak.LdapFullNameMapper{}

	fullNameMapperName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapFullNameMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakLdapFullNameMapper_basic(fullNameMapperName),
				Check:  testAccCheckKeycloakLdapFullNameMapperFetch("keycloak_ldap_full_name_mapper.full_name_mapper", mapper),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteLdapFullNameMapper(testCtx, mapper.RealmId, mapper.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakLdapFullNameMapper_basic(fullNameMapperName),
				Check:  testAccCheckKeycloakLdapFullNameMapperFetch("keycloak_ldap_full_name_mapper.full_name_mapper", mapper),
			},
		},
	})
}

func TestAccKeycloakLdapFullNameMapper_readWriteValidation(t *testing.T) {
	t.Parallel()

	mapper := &keycloak.LdapFullNameMapper{
		LdapFullNameAttribute: "terraform-" + acctest.RandString(10),
		ReadOnly:              true,
		WriteOnly:             true,
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapFullNameMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakLdapFullNameMapper_basicFromInterface(mapper),
				ExpectError: regexp.MustCompile("validation error: ldap full name mapper cannot be both read only and write only"),
			},
		},
	})
}

// write_only can't be set to true if the user federation provider is not writable
func TestAccKeycloakLdapFullNameMapper_writableValidation(t *testing.T) {
	t.Parallel()

	mapperName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapFullNameMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakLdapFullNameMapper_writableInvalid(mapperName),
				ExpectError: regexp.MustCompile("validation error: ldap full name mapper cannot be write only when ldap provider is not writable"),
			},
			{
				Config: testKeycloakLdapFullNameMapper_writableValid(mapperName),
				Check:  testAccCheckKeycloakLdapFullNameMapperExists("keycloak_ldap_full_name_mapper.full_name_mapper"),
			},
		},
	})
}

func TestAccKeycloakLdapFullNameMapper_updateLdapUserFederation(t *testing.T) {
	t.Parallel()

	mapperName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapFullNameMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakLdapFullNameMapper_updateLdapUserFederationBefore(mapperName),
				Check:  testAccCheckKeycloakLdapFullNameMapperExists("keycloak_ldap_full_name_mapper.full_name_mapper"),
			},
			{
				Config: testKeycloakLdapFullNameMapper_updateLdapUserFederationAfter(mapperName),
				Check:  testAccCheckKeycloakLdapFullNameMapperExists("keycloak_ldap_full_name_mapper.full_name_mapper"),
			},
		},
	})
}

func testAccCheckKeycloakLdapFullNameMapperExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getLdapFullNameMapperFromState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckKeycloakLdapFullNameMapperFetch(resourceName string, mapper *keycloak.LdapFullNameMapper) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedMapper, err := getLdapFullNameMapperFromState(s, resourceName)
		if err != nil {
			return err
		}

		mapper.Id = fetchedMapper.Id
		mapper.RealmId = fetchedMapper.RealmId

		return nil
	}
}

func testAccCheckKeycloakLdapFullNameMapperDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_ldap_full_name_mapper" {
				continue
			}

			id := rs.Primary.ID
			realm := rs.Primary.Attributes["realm_id"]

			ldapFullNameMapper, _ := keycloakClient.GetLdapFullNameMapper(testCtx, realm, id)
			if ldapFullNameMapper != nil {
				return fmt.Errorf("ldap full name mapper with id %s still exists", id)
			}
		}

		return nil
	}
}

func getLdapFullNameMapperFromState(s *terraform.State, resourceName string) (*keycloak.LdapFullNameMapper, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm_id"]

	ldapFullNameMapper, err := keycloakClient.GetLdapFullNameMapper(testCtx, realm, id)
	if err != nil {
		return nil, fmt.Errorf("error getting ldap full name mapper with id %s: %s", id, err)
	}

	return ldapFullNameMapper, nil
}

func getLdapGenericMapperImportId(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}

		id := rs.Primary.ID
		realmId := rs.Primary.Attributes["realm_id"]
		ldapUserFederationId := rs.Primary.Attributes["ldap_user_federation_id"]

		return fmt.Sprintf("%s/%s/%s", realmId, ldapUserFederationId, id), nil
	}
}

func testKeycloakLdapFullNameMapper_basic(mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_ldap_user_federation" "openldap" {
	name     = "openldap"
	realm_id = data.keycloak_realm.realm.id

	enabled = true

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

resource "keycloak_ldap_full_name_mapper" "full_name_mapper" {
	name                     = "%s"
	realm_id                 = data.keycloak_realm.realm.id
	ldap_user_federation_id  = keycloak_ldap_user_federation.openldap.id

	ldap_full_name_attribute = "cn"
}
	`, testAccRealmUserFederation.Realm, mapperName)
}

func testKeycloakLdapFullNameMapper_basicFromInterface(mapper *keycloak.LdapFullNameMapper) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_ldap_user_federation" "openldap" {
	name     = "openldap"
	realm_id = data.keycloak_realm.realm.id

	enabled = true

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

resource "keycloak_ldap_full_name_mapper" "full_name_mapper" {
	name                    = "%s"
	realm_id                = data.keycloak_realm.realm.id
	ldap_user_federation_id = keycloak_ldap_user_federation.openldap.id

	ldap_full_name_attribute = "%s"
	read_only                = %t
	write_only               = %t
}
	`, testAccRealmUserFederation.Realm, mapper.Name, mapper.LdapFullNameAttribute, mapper.ReadOnly, mapper.WriteOnly)
}

func testKeycloakLdapFullNameMapper_updateLdapUserFederationBefore(mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm_one" {
	realm = "%s"
}

data "keycloak_realm" "realm_two" {
	realm = "%s"
}

resource "keycloak_ldap_user_federation" "openldap_one" {
	name     = "openldap"
	realm_id = data.keycloak_realm.realm_one.id

	enabled = true

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
	name     = "openldap"
	realm_id = data.keycloak_realm.realm_two.id

	enabled = true

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

resource "keycloak_ldap_full_name_mapper" "full_name_mapper" {
	name                    = "%s"
	realm_id                = data.keycloak_realm.realm_one.id
	ldap_user_federation_id = keycloak_ldap_user_federation.openldap_one.id

	ldap_full_name_attribute = "cn"
}
	`, testAccRealmUserFederation.Realm, testAccRealmTwo.Realm, mapperName)
}

func testKeycloakLdapFullNameMapper_updateLdapUserFederationAfter(mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm_one" {
	realm = "%s"
}

data "keycloak_realm" "realm_two" {
	realm = "%s"
}

resource "keycloak_ldap_user_federation" "openldap_one" {
	name     = "openldap"
	realm_id = data.keycloak_realm.realm_one.id

	enabled = true

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
	name     = "openldap"
	realm_id = data.keycloak_realm.realm_two.id

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

resource "keycloak_ldap_full_name_mapper" "full_name_mapper" {
	name                    = "%s"
	realm_id                = data.keycloak_realm.realm_two.id
	ldap_user_federation_id = keycloak_ldap_user_federation.openldap_two.id

	ldap_full_name_attribute = "cn"
}
	`, testAccRealmUserFederation.Realm, testAccRealmTwo.Realm, mapperName)
}

func testKeycloakLdapFullNameMapper_writableInvalid(mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_ldap_user_federation" "openldap" {
	name     = "openldap"
	realm_id = data.keycloak_realm.realm.id

	enabled   = true
	edit_mode = "READ_ONLY"

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

resource "keycloak_ldap_full_name_mapper" "full_name_mapper" {
	name                    = "%s"
	realm_id                = data.keycloak_realm.realm.id
	ldap_user_federation_id = keycloak_ldap_user_federation.openldap.id

	ldap_full_name_attribute = "cn"
	write_only               = true
}
	`, testAccRealmUserFederation.Realm, mapperName)
}

func testKeycloakLdapFullNameMapper_writableValid(mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_ldap_user_federation" "openldap" {
	name     = "openldap"
	realm_id = data.keycloak_realm.realm.id

	enabled   = true
	edit_mode = "WRITABLE"

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

resource "keycloak_ldap_full_name_mapper" "full_name_mapper" {
	name                    = "%s"
	realm_id                = data.keycloak_realm.realm.id
	ldap_user_federation_id = keycloak_ldap_user_federation.openldap.id

	ldap_full_name_attribute = "cn"
	write_only               = true
}
	`, testAccRealmUserFederation.Realm, mapperName)
}
