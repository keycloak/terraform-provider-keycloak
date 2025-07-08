package provider

import (
	"encoding/xml"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccKeycloakDataSourceSamlClientInstallationProvider_basic(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_saml_client.saml_client"
	dataSourceName := "data.keycloak_saml_client_installation_provider.saml_sp_descriptor"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakSamlClientDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testDataSourceKeycloakSamlClientInstallationProvider_basic(clientId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "realm_id", resourceName, "realm_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "client_id", resourceName, "id"),
					resource.TestCheckResourceAttr(dataSourceName, "provider_id", "saml-sp-descriptor"),
					testAccCheckDataKeycloakSamlClientInstallationProvider(dataSourceName),
				),
			},
		},
	})
}

func testAccCheckDataKeycloakSamlClientInstallationProvider(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		value := rs.Primary.Attributes["value"]

		err := xml.Unmarshal([]byte(value), new(interface{}))
		if err != nil {
			return fmt.Errorf("invalid XML: %s\n%s", err, value)
		}

		return nil
	}
}

func testDataSourceKeycloakSamlClientInstallationProvider_basic(clientId string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_saml_client" "saml_client" {
	client_id = "%s"
	realm_id  = data.keycloak_realm.realm.id
}

data "keycloak_saml_client_installation_provider" "saml_sp_descriptor" {
  realm_id    = data.keycloak_realm.realm.id
  client_id   = keycloak_saml_client.saml_client.id
  provider_id = "saml-sp-descriptor"
}
	`, testAccRealm.Realm, clientId)
}
