package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakLdapCustomMapper_basic(t *testing.T) {
	t.Parallel()

	customMapperName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapCustomMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakLdapCustomMapper_basic(customMapperName),
				Check:  testAccCheckKeycloakLdapCustomMapperExists("keycloak_ldap_custom_mapper.sample_mapper"),
			},
			{
				ResourceName:      "keycloak_ldap_custom_mapper.sample_mapper",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getLdapGenericMapperImportId("keycloak_ldap_custom_mapper.sample_mapper"),
			},
		},
	})
}

func TestAccKeycloakLdapCustomMapper_createAfterManualDestroy(t *testing.T) {
	t.Parallel()

	var mapper = &keycloak.LdapCustomMapper{}

	customMapperName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapCustomMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakLdapCustomMapper_basic(customMapperName),
				Check:  testAccCheckKeycloakLdapCustomMapperFetch("keycloak_ldap_custom_mapper.sample_mapper", mapper),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteLdapCustomMapper(testCtx, mapper.RealmId, mapper.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakLdapCustomMapper_basic(customMapperName),
				Check:  testAccCheckKeycloakLdapCustomMapperExists("keycloak_ldap_custom_mapper.sample_mapper"),
			},
		},
	})
}

func TestAccKeycloakLdapCustomMapper_updateLdapUserFederation(t *testing.T) {
	t.Parallel()

	customMapperName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapCustomMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakLdapCustomMapper_updateLdapUserFederationBefore(customMapperName),
				Check:  testAccCheckKeycloakLdapCustomMapperExists("keycloak_ldap_custom_mapper.sample_mapper"),
			},
			{
				Config: testKeycloakLdapCustomMapper_updateLdapUserFederationAfter(customMapperName),
				Check:  testAccCheckKeycloakLdapCustomMapperExists("keycloak_ldap_custom_mapper.sample_mapper"),
			},
		},
	})
}

func TestAccKeycloakLdapCustomMapper_updateInPlace(t *testing.T) {
	t.Parallel()

	customMapperBefore := &keycloak.LdapCustomMapper{
		Name:         acctest.RandString(10),
		ProviderId:   "msad-user-account-control-mapper",
		ProviderType: "org.keycloak.storage.ldap.mappers.LDAPStorageMapper",
	}
	customMapperAfter := &keycloak.LdapCustomMapper{
		Name:         acctest.RandString(10),
		ProviderId:   "msad-user-account-control-mapper",
		ProviderType: "org.keycloak.storage.ldap.mappers.LDAPStorageMapper",
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapCustomMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakLdapCustomMapper_basicFromInterface(customMapperBefore),
				Check:  testAccCheckKeycloakLdapCustomMapperExists("keycloak_ldap_custom_mapper.sample_mapper"),
			},
			{
				Config: testKeycloakLdapCustomMapper_basicFromInterface(customMapperAfter),
				Check:  testAccCheckKeycloakLdapCustomMapperExists("keycloak_ldap_custom_mapper.sample_mapper"),
			},
		},
	})
}

func testAccCheckKeycloakLdapCustomMapperExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getLdapCustomMapperFromState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckKeycloakLdapCustomMapperFetch(resourceName string, mapper *keycloak.LdapCustomMapper) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedMapper, err := getLdapCustomMapperFromState(s, resourceName)
		if err != nil {
			return err
		}

		mapper.Id = fetchedMapper.Id
		mapper.RealmId = fetchedMapper.RealmId

		return nil
	}
}

func testAccCheckKeycloakLdapCustomMapperDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_ldap_custom_mapper" {
				continue
			}

			id := rs.Primary.ID
			realm := rs.Primary.Attributes["realm_id"]

			ldapCustomMapper, _ := keycloakClient.GetLdapCustomMapper(testCtx, realm, id)
			if ldapCustomMapper != nil {
				return fmt.Errorf("ldap user attribute mapper with id %s still exists", id)
			}
		}

		return nil
	}
}

func getLdapCustomMapperFromState(s *terraform.State, resourceName string) (*keycloak.LdapCustomMapper, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm_id"]

	ldapCustomMapper, err := keycloakClient.GetLdapCustomMapper(testCtx, realm, id)
	if err != nil {
		return nil, fmt.Errorf("error getting ldap user attribute mapper with id %s: %s", id, err)
	}

	return ldapCustomMapper, nil
}

func testKeycloakLdapCustomMapper_basic(customMapperName string) string {
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

resource "keycloak_ldap_custom_mapper" "sample_mapper" {
	name                        = "%s"
	realm_id                    = data.keycloak_realm.realm.id
	ldap_user_federation_id     = "${keycloak_ldap_user_federation.openldap.id}"

	provider_id        			= "user-attribute-ldap-mapper"
	provider_type               = "org.keycloak.storage.ldap.mappers.LDAPStorageMapper"
    config = {
	  "user.model.attribute"    = "username"
	  "ldap.attribute"          = "cn"
    }
}
	`, testAccRealmUserFederation.Realm, customMapperName)
}

func testKeycloakLdapCustomMapper_basicFromInterface(mapper *keycloak.LdapCustomMapper) string {
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

resource "keycloak_ldap_custom_mapper" "sample_mapper" {
	name                        = "%s"
	realm_id                    = data.keycloak_realm.realm.id
	ldap_user_federation_id     = "${keycloak_ldap_user_federation.openldap.id}"

	provider_id			        = "%s"
	provider_type               = "%s"

}
	`, testAccRealmUserFederation.Realm, mapper.Name, mapper.ProviderId, mapper.ProviderType)
}

func testKeycloakLdapCustomMapper_updateLdapUserFederationBefore(customMapperName string) string {
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

resource "keycloak_ldap_custom_mapper" "sample_mapper" {
	name                        = "%s"
	realm_id                    = data.keycloak_realm.realm_one.id
	ldap_user_federation_id     = "${keycloak_ldap_user_federation.openldap_one.id}"

	provider_id        			= "msad-user-account-control-mapper"
	provider_type               = "org.keycloak.storage.ldap.mappers.LDAPStorageMapper"
}
	`, testAccRealmUserFederation.Realm, testAccRealmTwo.Realm, customMapperName)
}

func testKeycloakLdapCustomMapper_updateLdapUserFederationAfter(customMapperName string) string {
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

resource "keycloak_ldap_custom_mapper" "sample_mapper" {
	name                        = "%s"
	realm_id                    = data.keycloak_realm.realm_two.id
	ldap_user_federation_id     = "${keycloak_ldap_user_federation.openldap_two.id}"

	provider_id        			= "msad-user-account-control-mapper"
	provider_type               = "org.keycloak.storage.ldap.mappers.LDAPStorageMapper"
}
	`, testAccRealmUserFederation.Realm, testAccRealmTwo.Realm, customMapperName)
}
