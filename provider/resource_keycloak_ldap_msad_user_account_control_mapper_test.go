package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"testing"
)

func TestAccKeycloakLdapMsadUserAccountControlMapper_basic(t *testing.T) {
	t.Parallel()

	msadUacMapperName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapMsadUserAccountControlMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakLdapMsadUserAccountControlMapper_basic(msadUacMapperName, randomBool()),
				Check:  testAccCheckKeycloakLdapMsadUserAccountControlMapperExists("keycloak_ldap_msad_user_account_control_mapper.uac_mapper"),
			},
			{
				ResourceName:      "keycloak_ldap_msad_user_account_control_mapper.uac_mapper",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getLdapGenericMapperImportId("keycloak_ldap_msad_user_account_control_mapper.uac_mapper"),
			},
		},
	})
}

func TestAccKeycloakLdapMsadUserAccountControlMapper_createAfterManualDestroy(t *testing.T) {
	t.Parallel()

	var mapper = &keycloak.LdapMsadUserAccountControlMapper{}

	msadUacMapperName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapMsadUserAccountControlMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakLdapMsadUserAccountControlMapper_basic(msadUacMapperName, randomBool()),
				Check:  testAccCheckKeycloakLdapMsadUserAccountControlMapperFetch("keycloak_ldap_msad_user_account_control_mapper.uac_mapper", mapper),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteLdapMsadUserAccountControlMapper(testCtx, mapper.RealmId, mapper.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakLdapMsadUserAccountControlMapper_basic(msadUacMapperName, randomBool()),
				Check:  testAccCheckKeycloakLdapMsadUserAccountControlMapperExists("keycloak_ldap_msad_user_account_control_mapper.uac_mapper"),
			},
		},
	})
}

func TestAccKeycloakLdapMsadUserAccountControlMapper_updateLdapUserFederation(t *testing.T) {
	t.Parallel()

	msadUacMapperName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapMsadUserAccountControlMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakLdapMsadUserAccountControlMapper_updateLdapUserFederationBefore(msadUacMapperName),
				Check:  testAccCheckKeycloakLdapMsadUserAccountControlMapperExists("keycloak_ldap_msad_user_account_control_mapper.uac_mapper"),
			},
			{
				Config: testKeycloakLdapMsadUserAccountControlMapper_updateLdapUserFederationAfter(msadUacMapperName),
				Check:  testAccCheckKeycloakLdapMsadUserAccountControlMapperExists("keycloak_ldap_msad_user_account_control_mapper.uac_mapper"),
			},
		},
	})
}

func TestAccKeycloakLdapMsadUserAccountControlMapper_updateInPlace(t *testing.T) {
	t.Parallel()

	passwordHintsEnabled := randomBool()

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapMsadUserAccountControlMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakLdapMsadUserAccountControlMapper_basic(acctest.RandString(10), passwordHintsEnabled),
				Check:  testAccCheckKeycloakLdapMsadUserAccountControlMapperExists("keycloak_ldap_msad_user_account_control_mapper.uac_mapper"),
			},
			{
				Config: testKeycloakLdapMsadUserAccountControlMapper_basic(acctest.RandString(10), !passwordHintsEnabled),
				Check:  testAccCheckKeycloakLdapMsadUserAccountControlMapperExists("keycloak_ldap_msad_user_account_control_mapper.uac_mapper"),
			},
		},
	})
}

func testAccCheckKeycloakLdapMsadUserAccountControlMapperExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getLdapMsadUserAccountControlMapperFromState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckKeycloakLdapMsadUserAccountControlMapperFetch(resourceName string, mapper *keycloak.LdapMsadUserAccountControlMapper) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedMapper, err := getLdapMsadUserAccountControlMapperFromState(s, resourceName)
		if err != nil {
			return err
		}

		mapper.Id = fetchedMapper.Id
		mapper.RealmId = fetchedMapper.RealmId

		return nil
	}
}

func testAccCheckKeycloakLdapMsadUserAccountControlMapperDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_ldap_msad_user_account_control_mapper" {
				continue
			}

			id := rs.Primary.ID
			realm := rs.Primary.Attributes["realm_id"]

			ldapMsadUserAccountControlMapper, _ := keycloakClient.GetLdapMsadUserAccountControlMapper(testCtx, realm, id)
			if ldapMsadUserAccountControlMapper != nil {
				return fmt.Errorf("ldap msad uac mapper with id %s still exists", id)
			}
		}

		return nil
	}
}

func getLdapMsadUserAccountControlMapperFromState(s *terraform.State, resourceName string) (*keycloak.LdapMsadUserAccountControlMapper, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm_id"]

	ldapMsadUserAccountControlMapper, err := keycloakClient.GetLdapMsadUserAccountControlMapper(testCtx, realm, id)
	if err != nil {
		return nil, fmt.Errorf("error getting ldap msad uac mapper with id %s: %s", id, err)
	}

	return ldapMsadUserAccountControlMapper, nil
}

func testKeycloakLdapMsadUserAccountControlMapper_basic(msadUacMapperName string, passwordHintsEnabled bool) string {
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

resource "keycloak_ldap_msad_user_account_control_mapper" "uac_mapper" {
	name                               = "%s"
	realm_id                           = data.keycloak_realm.realm.id
	ldap_user_federation_id            = "${keycloak_ldap_user_federation.openldap.id}"

	ldap_password_policy_hints_enabled = %t
}
	`, testAccRealmUserFederation.Realm, msadUacMapperName, passwordHintsEnabled)
}

func testKeycloakLdapMsadUserAccountControlMapper_updateLdapUserFederationBefore(msadUacMapperName string) string {
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

resource "keycloak_ldap_msad_user_account_control_mapper" "uac_mapper" {
	name                               = "%s"
	realm_id                           = data.keycloak_realm.realm_one.id
	ldap_user_federation_id            = "${keycloak_ldap_user_federation.openldap_one.id}"
}
	`, testAccRealmUserFederation.Realm, testAccRealmTwo.Realm, msadUacMapperName)
}

func testKeycloakLdapMsadUserAccountControlMapper_updateLdapUserFederationAfter(msadUacMapperName string) string {
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

resource "keycloak_ldap_msad_user_account_control_mapper" "uac_mapper" {
	name                               = "%s"
	realm_id                           = data.keycloak_realm.realm_two.id
	ldap_user_federation_id            = "${keycloak_ldap_user_federation.openldap_two.id}"
}
	`, testAccRealmUserFederation.Realm, testAccRealmTwo.Realm, msadUacMapperName)
}
