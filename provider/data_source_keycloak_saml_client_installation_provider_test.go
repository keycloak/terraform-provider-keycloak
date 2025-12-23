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
	dataSourceName := "data.keycloak_saml_client_installation_provider.descriptor"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakSamlClientDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testDataSourceKeycloakSamlClientInstallationProvider(clientId, "saml-sp-descriptor"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "realm_id", resourceName, "realm_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "client_id", resourceName, "id"),
					resource.TestCheckResourceAttr(dataSourceName, "provider_id", "saml-sp-descriptor"),
					resource.TestCheckResourceAttr(dataSourceName, "zip_files.%", "0"),
					testAccCheckDataKeycloakSamlClientInstallationProvider_isXML(dataSourceName, "value"),
				),
			},
		},
	})
}

func TestAccKeycloakDataSourceSamlClientInstallationProvider_zip(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_saml_client.saml_client"
	dataSourceName := "data.keycloak_saml_client_installation_provider.descriptor"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakSamlClientDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testDataSourceKeycloakSamlClientInstallationProvider(clientId, "mod-auth-mellon"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "realm_id", resourceName, "realm_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "client_id", resourceName, "id"),
					resource.TestCheckResourceAttr(dataSourceName, "provider_id", "mod-auth-mellon"),
					resource.TestCheckResourceAttr(dataSourceName, "zip_files.%", "4"),
					testAccCheckDataKeycloakSamlClientInstallationProvider_isXML(dataSourceName, "zip_files.idp-metadata.xml"),
					testAccCheckDataKeycloakSamlClientInstallationProvider_isXML(dataSourceName, "zip_files.sp-metadata.xml"),
					resource.TestCheckResourceAttrSet(dataSourceName, "zip_files.client-cert.pem"),
					resource.TestCheckResourceAttrSet(dataSourceName, "zip_files.client-private-key.pem"),
				),
			},
		},
	})
}

func testAccCheckDataKeycloakSamlClientInstallationProvider_isXML(resourceName string, attributeName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		value := rs.Primary.Attributes[attributeName]

		err := xml.Unmarshal([]byte(value), new(interface{}))
		if err != nil {
			return fmt.Errorf("invalid XML: %s\n%s", err, value)
		}

		return nil
	}
}

func testDataSourceKeycloakSamlClientInstallationProvider(clientId string, providerId string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_saml_client" "saml_client" {
	client_id = "%s"
	realm_id  = data.keycloak_realm.realm.id
}

data "keycloak_saml_client_installation_provider" "descriptor" {
  realm_id    = data.keycloak_realm.realm.id
  client_id   = keycloak_saml_client.saml_client.id
  provider_id = "%s"
}
	`, testAccRealm.Realm, clientId, providerId)
}
