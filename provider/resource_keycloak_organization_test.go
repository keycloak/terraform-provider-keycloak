package provider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakOrganization_basic(t *testing.T) {
	skipIfVersionIsLessThan(testCtx, t, keycloakClient, keycloak.Version_26)
	organizationName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOrganizationDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOrganization_basic(organizationName),
				Check:  testAccCheckKeycloakOrganizationExists("keycloak_organization.organization"),
			},
		},
	})
}

func TestAccKeycloakOrganization_basicUpdate(t *testing.T) {
	skipIfVersionIsLessThan(testCtx, t, keycloakClient, keycloak.Version_26)
	organizationName := acctest.RandomWithPrefix("tf-acc")
	firstEnabled := randomBool()
	domainOne := acctest.RandomWithPrefix("tf-acc")
	domainTwo := acctest.RandomWithPrefix("tf-acc")

	firstOrg := &keycloak.Organization{
		Realm:       testAccRealm.Realm,
		Name:        organizationName,
		Alias:       organizationName,
		Enabled:     firstEnabled,
		Description: acctest.RandomWithPrefix("tf-acc"),
		RedirectUrl: "https://example.com",
		Domains: []keycloak.OrganizationDomain{
			{
				Name:     domainOne,
				Verified: firstEnabled,
			},
		},
	}

	secondOrg := &keycloak.Organization{
		Realm:       testAccRealm.Realm,
		Name:        organizationName,
		Alias:       organizationName,
		Enabled:     !firstEnabled,
		Description: acctest.RandomWithPrefix("tf-acc"),
		RedirectUrl: "https://example.org",
		Domains: []keycloak.OrganizationDomain{
			{
				Name:     domainTwo,
				Verified: !firstEnabled,
			},
		},
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOrganizationDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOrganization_all(firstOrg),
				Check:  testAccCheckKeycloakOrganizationExists("keycloak_organization.organization"),
			},
			{
				Config: testKeycloakOrganization_all(secondOrg),
				Check:  testAccCheckKeycloakOrganizationExists("keycloak_organization.organization"),
			},
			{
				Config: testKeycloakOrganization_basic(organizationName),
				Check:  testAccCheckKeycloakOrganizationExists("keycloak_organization.organization"),
			},
		},
	})
}

func TestAccKeycloakOrganization_createAfterManualDestroy(t *testing.T) {
	skipIfVersionIsLessThan(testCtx, t, keycloakClient, keycloak.Version_26)
	var organization = &keycloak.Organization{}

	organizationName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOrganizationDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOrganization_basic(organizationName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOrganizationExists("keycloak_organization.organization"),
					testAccCheckKeycloakOrganizationFetch("keycloak_organization.organization", organization),
				),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteOrganization(testCtx, organization.Realm, organization.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakOrganization_basic(organizationName),
				Check:  testAccCheckKeycloakOrganizationExists("keycloak_organization.organization"),
			},
		},
	})
}

func TestAccKeycloakOrganization_basicWithMultipleDomains(t *testing.T) {
	skipIfVersionIsLessThan(testCtx, t, keycloakClient, keycloak.Version_26)
	organizationName := acctest.RandomWithPrefix("tf-acc")
	extraDomain := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOrganizationDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOrganization_multipleDomains(organizationName, extraDomain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOrganizationExists("keycloak_organization.organization"),
					testAccCheckKeycloakOrganizationHasDomains("keycloak_organization.organization", extraDomain),
				),
			},
		},
	})
}

func TestAccKeycloakOrganization_basicWithAttributes(t *testing.T) {
	skipIfVersionIsLessThan(testCtx, t, keycloakClient, keycloak.Version_26)
	organizationName := acctest.RandomWithPrefix("tf-acc")
	attributeName := acctest.RandomWithPrefix("tf-acc")
	attributeValue := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOrganizationDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOrganization_attributes(organizationName, attributeName, attributeValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOrganizationExists("keycloak_organization.organization"),
					testAccCheckKeycloakOrganizationHasAttribute("keycloak_organization.organization", attributeName, attributeValue),
				),
			},
		},
	})
}

func testAccCheckKeycloakOrganizationExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getOrganizationFromState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckKeycloakOrganizationDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for name, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_organization" || strings.HasPrefix(name, "data") {
				continue
			}

			id := rs.Primary.ID
			realm := rs.Primary.Attributes["realm"]

			organization, _ := keycloakClient.GetOrganization(testCtx, realm, id)
			if organization != nil {
				return fmt.Errorf("%s with id %s still exists", name, id)
			}
		}

		return nil
	}
}

func testAccCheckKeycloakOrganizationFetch(resourceName string, organization *keycloak.Organization) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		fetchedOrganization, err := getOrganizationFromState(state, resourceName)
		if err != nil {
			return err
		}

		organization.Id = fetchedOrganization.Id
		organization.Name = fetchedOrganization.Name
		organization.Realm = fetchedOrganization.Realm

		return nil
	}
}

func testAccCheckKeycloakOrganizationHasDomains(resourceName, domainName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		organization, err := getOrganizationFromState(state, resourceName)
		if err != nil {
			return err
		}

		if len(organization.Domains) != 2 || (organization.Domains[0].Name != domainName && organization.Domains[1].Name != domainName) {
			return fmt.Errorf("expected organization %s to have domain with domainName %s", organization.Name, domainName)
		}

		return nil
	}
}

func testAccCheckKeycloakOrganizationHasAttribute(resourceName, attributeName, attributeValue string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		organization, err := getOrganizationFromState(state, resourceName)
		if err != nil {
			return err
		}

		if len(organization.Attributes) != 1 || organization.Attributes[attributeName][0] != attributeValue {
			return fmt.Errorf("expected organization %s to have attribute %s with value %s", organization.Name, attributeName, attributeValue)
		}

		return nil
	}
}

func getOrganizationFromState(s *terraform.State, resourceName string) (*keycloak.Organization, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm"]

	organization, err := keycloakClient.GetOrganization(testCtx, realm, id)
	if err != nil {
		return nil, fmt.Errorf("error getting organization with id %s: %s", id, err)
	}

	return organization, nil
}

func testKeycloakOrganization_basic(organization string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_organization" "organization" {
	name  = "%s"
	realm = data.keycloak_realm.realm.id

	domain {
		name     = "example.com"
		verified = true
	}
}
	`, testAccRealm.Realm, organization)
}

func testKeycloakOrganization_all(org *keycloak.Organization) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_organization" "organization" {
	name         = "%s"
	alias        = "%s"
	realm        = data.keycloak_realm.realm.id
	enabled      = %t
	description  = "%s"
	redirect_url = "%s"

	domain {
		name 	 = "%s"
		verified = %t
	}
}
	`, testAccRealm.Realm, org.Name, org.Alias, org.Enabled, org.Description, org.RedirectUrl, org.Domains[0].Name, org.Domains[0].Verified)
}

func testKeycloakOrganization_multipleDomains(organizationName, domainName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_organization" "organization" {
	name  = "%s"
	realm = data.keycloak_realm.realm.id

	domain {
		name     = "example.com"
	}

	domain {
		name 	 = "%s"
	}
}
	`, testAccRealm.Realm, organizationName, domainName)
}

func testKeycloakOrganization_attributes(organizationName, attributeName, attributeValue string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_organization" "organization" {
	name     = "%s"
	realm    = data.keycloak_realm.realm.id

	domain {
		name     = "example.com"
		verified = true
	}

	attributes = {
		"%s" = "%s"
	}
}
	`, testAccRealm.Realm, organizationName, attributeName, attributeValue)
}
