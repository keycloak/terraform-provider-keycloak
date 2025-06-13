package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakDataSourceOrganization_basic(t *testing.T) {
	skipIfVersionIsLessThan(testCtx, t, keycloakClient, keycloak.Version_26)
	orgName := acctest.RandomWithPrefix("tf-acc-test")
	domainName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.keycloak_organization.test"
	resourceName := "keycloak_organization.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccKeycloakOrganizationConfig(orgName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "realm", resourceName, "realm"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "enabled", resourceName, "enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "redirect_url", resourceName, "redirect_url"),
					resource.TestCheckResourceAttrPair(dataSourceName, "domain", resourceName, "domain"),
					resource.TestCheckResourceAttrPair(dataSourceName, "attributes", resourceName, "attributes"),
				),
			},
		},
	})
}

func testAccKeycloakOrganizationConfig(orgName, domainName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_organization" "test" {
	name		 = "%s"
	alias		 = "%s"
	realm        = data.keycloak_realm.realm.id
	enabled      = true
	description  = "a test organization"
	redirect_url = "http://localhost:5555"
	domain {
		name 	 = "%s"
		verified = true
	}
	attributes = {
		"key1" = "value1"
	}
}

data "keycloak_organization" "test" {
	realm = data.keycloak_realm.realm.id
	name  = keycloak_organization.test.name

	depends_on = [
		keycloak_organization.test,
	]
}
`, testAccRealm.Realm, orgName, orgName, domainName)
}
