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

func TestAccKeycloakHardcodedAttributeMapper_basic(t *testing.T) {
	t.Parallel()
	attributeName := acctest.RandomWithPrefix("tf-acc")
	attributeValue := acctest.RandomWithPrefix("tf-acc")
	attributeMapperName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakHardcodedAttributeMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakHardcodedAttributeMapper(attributeMapperName, attributeName, attributeValue),
				Check:  testAccCheckKeycloakHardcodedAttributeMapperExists("keycloak_hardcoded_attribute_mapper.hardcoded_attribute_mapper"),
			},
			{
				ResourceName:      "keycloak_hardcoded_attribute_mapper.hardcoded_attribute_mapper",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getLdapGenericMapperImportId("keycloak_hardcoded_attribute_mapper.hardcoded_attribute_mapper"),
			},
		},
	})
}

func TestAccKeycloakHardcodedAttributeMapper_createAfterManualDestroy(t *testing.T) {
	t.Parallel()
	var mapper = &keycloak.HardcodedAttributeMapper{}

	attributeName := acctest.RandomWithPrefix("tf-acc")
	attributeValue := acctest.RandomWithPrefix("tf-acc")
	attributeMapperName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakHardcodedAttributeMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakHardcodedAttributeMapper(attributeMapperName, attributeName, attributeValue),
				Check:  testAccCheckKeycloakHardcodedAttributeMapperFetch("keycloak_hardcoded_attribute_mapper.hardcoded_attribute_mapper", mapper),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteHardcodedAttributeMapper(context.Background(), mapper.RealmId, mapper.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakHardcodedAttributeMapper(attributeMapperName, attributeName, attributeValue),
				Check:  testAccCheckKeycloakHardcodedAttributeMapperExists("keycloak_hardcoded_attribute_mapper.hardcoded_attribute_mapper"),
			},
		},
	})
}

func testAccCheckKeycloakHardcodedAttributeMapperExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getHardcodedAttributeMapperFromState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckKeycloakHardcodedAttributeMapperFetch(resourceName string, mapper *keycloak.HardcodedAttributeMapper) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedMapper, err := getHardcodedAttributeMapperFromState(s, resourceName)
		if err != nil {
			return err
		}

		mapper.Id = fetchedMapper.Id
		mapper.RealmId = fetchedMapper.RealmId

		return nil
	}
}

func testAccCheckKeycloakHardcodedAttributeMapperDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_hardcoded_attribute_mapper" {
				continue
			}

			id := rs.Primary.ID
			realm := rs.Primary.Attributes["realm_id"]

			ldapMapper, _ := keycloakClient.GetHardcodedAttributeMapper(context.Background(), realm, id)
			if ldapMapper != nil {
				return fmt.Errorf("Hardcoded attribute mapper with id %s still exists", id)
			}
		}

		return nil
	}
}

func getHardcodedAttributeMapperFromState(s *terraform.State, resourceName string) (*keycloak.HardcodedAttributeMapper, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm_id"]

	ldapMapper, err := keycloakClient.GetHardcodedAttributeMapper(context.Background(), realm, id)
	if err != nil {
		return nil, fmt.Errorf("error getting attribute mapper with id %s: %s", id, err)
	}

	return ldapMapper, nil
}

func testKeycloakHardcodedAttributeMapper(attributeMapperName, attributeName, attributeValue string) string {
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

resource "keycloak_hardcoded_attribute_mapper" "hardcoded_attribute_mapper" {
	name                        = "%s"
	realm_id                    = data.keycloak_realm.realm.id
	ldap_user_federation_id     = keycloak_ldap_user_federation.openldap.id
	attribute_name              = "%s"
	attribute_value             = "%s"
}
	`, testAccRealmUserFederation.Realm, attributeMapperName, attributeName, attributeValue)
}
