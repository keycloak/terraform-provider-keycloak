package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakLdapHardcodedAttributeMapper_basic(t *testing.T) {
	t.Parallel()
	attributeName := acctest.RandomWithPrefix("tf-acc")
	attributeValue := acctest.RandomWithPrefix("tf-acc")
	attributeMapperName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapHardcodedAttributeMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakLdapHardcodedAttributeMapper(attributeMapperName, attributeName, attributeValue),
				Check:  testAccCheckKeycloakLdapHardcodedAttributeMapperExists("keycloak_ldap_hardcoded_attribute_mapper.hardcoded_attribute_mapper"),
			},
			{
				ResourceName:      "keycloak_ldap_hardcoded_attribute_mapper.hardcoded_attribute_mapper",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getLdapGenericMapperImportId("keycloak_ldap_hardcoded_attribute_mapper.hardcoded_attribute_mapper"),
			},
		},
	})
}

func TestAccKeycloakLdapHardcodedAttributeMapper_createAfterManualDestroy(t *testing.T) {
	t.Parallel()
	var mapper = &keycloak.LdapHardcodedAttributeMapper{}

	attributeName := acctest.RandomWithPrefix("tf-acc")
	attributeValue := acctest.RandomWithPrefix("tf-acc")
	attributeMapperName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapHardcodedAttributeMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakLdapHardcodedAttributeMapper(attributeMapperName, attributeName, attributeValue),
				Check:  testAccCheckKeycloakLdapHardcodedAttributeMapperFetch("keycloak_ldap_hardcoded_attribute_mapper.hardcoded_attribute_mapper", mapper),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteLdapHardcodedAttributeMapper(context.Background(), mapper.RealmId, mapper.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakLdapHardcodedAttributeMapper(attributeMapperName, attributeName, attributeValue),
				Check:  testAccCheckKeycloakLdapHardcodedAttributeMapperExists("keycloak_ldap_hardcoded_attribute_mapper.hardcoded_attribute_mapper"),
			},
		},
	})
}

func testAccCheckKeycloakLdapHardcodedAttributeMapperExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getLdapHardcodedAttributeMapperFromState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckKeycloakLdapHardcodedAttributeMapperFetch(resourceName string, mapper *keycloak.LdapHardcodedAttributeMapper) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedMapper, err := getLdapHardcodedAttributeMapperFromState(s, resourceName)
		if err != nil {
			return err
		}

		mapper.Id = fetchedMapper.Id
		mapper.RealmId = fetchedMapper.RealmId

		return nil
	}
}

func testAccCheckKeycloakLdapHardcodedAttributeMapperDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_ldap_hardcoded_attribute_mapper" {
				continue
			}

			id := rs.Primary.ID
			realm := rs.Primary.Attributes["realm_id"]

			ldapMapper, _ := keycloakClient.GetLdapHardcodedAttributeMapper(context.Background(), realm, id)
			if ldapMapper != nil {
				return fmt.Errorf("ldap hardcoded attribute mapper with id %s still exists", id)
			}
		}

		return nil
	}
}

func getLdapHardcodedAttributeMapperFromState(s *terraform.State, resourceName string) (*keycloak.LdapHardcodedAttributeMapper, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm_id"]

	ldapMapper, err := keycloakClient.GetLdapHardcodedAttributeMapper(context.Background(), realm, id)
	if err != nil {
		return nil, fmt.Errorf("error getting ldap attribute mapper with id %s: %s", id, err)
	}

	return ldapMapper, nil
}

func testKeycloakLdapHardcodedAttributeMapper(attributeMapperName, attributeName, attributeValue string) string {
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

resource "keycloak_ldap_hardcoded_attribute_mapper" "hardcoded_attribute_mapper" {
	name                        = "%s"
	realm_id                    = data.keycloak_realm.realm.id
	ldap_user_federation_id     = keycloak_ldap_user_federation.openldap.id
	attribute_name              = "%s"
	attribute_value             = "%s"
}
	`, testAccRealmUserFederation.Realm, attributeMapperName, attributeName, attributeValue)
}
