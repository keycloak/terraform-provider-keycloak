package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"testing"
)

func TestAccKeycloakLdapMsadLdsUserAccountControlMapper_basic(t *testing.T) {
	t.Parallel()

	msadLdsUacMapperName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapMsadLdsUserAccountControlMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakLdapMsadLdsUserAccountControlMapper_basic(msadLdsUacMapperName),
				Check:  testAccCheckKeycloakLdapMsadLdsUserAccountControlMapperExists("keycloak_ldap_msad_lds_user_account_control_mapper.uac_mapper"),
			},
			{
				ResourceName:      "keycloak_ldap_msad_lds_user_account_control_mapper.uac_mapper",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getLdapGenericMapperImportId("keycloak_ldap_msad_lds_user_account_control_mapper.uac_mapper"),
			},
		},
	})
}

func TestAccKeycloakLdapMsadLdsUserAccountControlMapper_createAfterManualDestroy(t *testing.T) {
	t.Parallel()

	var mapper = &keycloak.LdapMsadLdsUserAccountControlMapper{}

	msadLdsUacMapperName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapMsadLdsUserAccountControlMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakLdapMsadLdsUserAccountControlMapper_basic(msadLdsUacMapperName),
				Check:  testAccCheckKeycloakLdapMsadLdsUserAccountControlMapperFetch("keycloak_ldap_msad_lds_user_account_control_mapper.uac_mapper", mapper),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteLdapMsadLdsUserAccountControlMapper(testCtx, mapper.RealmId, mapper.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakLdapMsadLdsUserAccountControlMapper_basic(msadLdsUacMapperName),
				Check:  testAccCheckKeycloakLdapMsadLdsUserAccountControlMapperExists("keycloak_ldap_msad_lds_user_account_control_mapper.uac_mapper"),
			},
		},
	})
}

func TestAccKeycloakLdapMsadLdsUserAccountControlMapper_updateLdapUserFederation(t *testing.T) {
	t.Parallel()

	msadLdsUacMapperName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapMsadLdsUserAccountControlMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakLdapMsadLdsUserAccountControlMapper_updateLdapUserFederationBefore(msadLdsUacMapperName),
				Check:  testAccCheckKeycloakLdapMsadLdsUserAccountControlMapperExists("keycloak_ldap_msad_lds_user_account_control_mapper.uac_mapper"),
			},
			{
				Config: testKeycloakLdapMsadLdsUserAccountControlMapper_updateLdapUserFederationAfter(msadLdsUacMapperName),
				Check:  testAccCheckKeycloakLdapMsadLdsUserAccountControlMapperExists("keycloak_ldap_msad_lds_user_account_control_mapper.uac_mapper"),
			},
		},
	})
}

func testAccCheckKeycloakLdapMsadLdsUserAccountControlMapperExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getLdapMsadLdsUserAccountControlMapperFromState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckKeycloakLdapMsadLdsUserAccountControlMapperFetch(resourceName string, mapper *keycloak.LdapMsadLdsUserAccountControlMapper) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedMapper, err := getLdapMsadLdsUserAccountControlMapperFromState(s, resourceName)
		if err != nil {
			return err
		}

		mapper.Id = fetchedMapper.Id
		mapper.RealmId = fetchedMapper.RealmId

		return nil
	}
}

func testAccCheckKeycloakLdapMsadLdsUserAccountControlMapperDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_ldap_msad_lds_user_account_control_mapper" {
				continue
			}

			id := rs.Primary.ID
			realm := rs.Primary.Attributes["realm_id"]

			ldapMsadLdsUserAccountControlMapper, _ := keycloakClient.GetLdapMsadLdsUserAccountControlMapper(testCtx, realm, id)
			if ldapMsadLdsUserAccountControlMapper != nil {
				return fmt.Errorf("ldap msad-lds uac mapper with id %s still exists", id)
			}
		}

		return nil
	}
}

func getLdapMsadLdsUserAccountControlMapperFromState(s *terraform.State, resourceName string) (*keycloak.LdapMsadLdsUserAccountControlMapper, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm_id"]

	ldapMsadLdsUserAccountControlMapper, err := keycloakClient.GetLdapMsadLdsUserAccountControlMapper(testCtx, realm, id)
	if err != nil {
		return nil, fmt.Errorf("error getting ldap msad-lds uac mapper with id %s: %s", id, err)
	}

	return ldapMsadLdsUserAccountControlMapper, nil
}

func testKeycloakLdapMsadLdsUserAccountControlMapper_basic(msadLdsUacMapperName string) string {
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

resource "keycloak_ldap_msad_lds_user_account_control_mapper" "uac_mapper" {
	name                               = "%s"
	realm_id                           = data.keycloak_realm.realm.id
	ldap_user_federation_id            = keycloak_ldap_user_federation.openldap.id
}
	`, testAccRealmUserFederation.Realm, msadLdsUacMapperName)
}

func testKeycloakLdapMsadLdsUserAccountControlMapper_updateLdapUserFederationBefore(msadLdsUacMapperName string) string {
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

resource "keycloak_ldap_msad_lds_user_account_control_mapper" "uac_mapper" {
	name                               = "%s"
	realm_id                           = data.keycloak_realm.realm_one.id
	ldap_user_federation_id            = keycloak_ldap_user_federation.openldap_one.id
}
	`, testAccRealmUserFederation.Realm, testAccRealmTwo.Realm, msadLdsUacMapperName)
}

func testKeycloakLdapMsadLdsUserAccountControlMapper_updateLdapUserFederationAfter(msadLdsUacMapperName string) string {
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

resource "keycloak_ldap_msad_lds_user_account_control_mapper" "uac_mapper" {
	name                               = "%s"
	realm_id                           = data.keycloak_realm.realm_two.id
	ldap_user_federation_id            = keycloak_ldap_user_federation.openldap_two.id
}
	`, testAccRealmUserFederation.Realm, testAccRealmTwo.Realm, msadLdsUacMapperName)
}
