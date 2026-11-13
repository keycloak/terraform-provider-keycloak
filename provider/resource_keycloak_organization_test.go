package provider

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestOrganizationIdentityProvidersAfterFailedUnlink(t *testing.T) {
	t.Parallel()

	old := []keycloak.OrganizationIdentityProvider{
		{Alias: "kept", AutoMembership: true, MembershipType: "UNMANAGED"},
		{Alias: "unlinked", AutoMembership: true, MembershipType: "UNMANAGED"},
		{Alias: "refused", AutoMembership: true, MembershipType: "MANAGED"},
	}
	planned := []keycloak.OrganizationIdentityProvider{
		{Alias: "kept", AutoMembership: false, MembershipType: "UNMANAGED"},
		{Alias: "added", AutoMembership: true, MembershipType: "UNMANAGED"},
	}

	// The link that was not removed stays in state with its old settings, so that the next plan
	// removes it again. The planned links stay as planned: a refresh corrects them.
	got := organizationIdentityProvidersAfterFailedUnlink(planned, old, []string{"refused"})
	want := []keycloak.OrganizationIdentityProvider{planned[0], planned[1], old[2]}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestAccKeycloakOrganization_basic(t *testing.T) {
	organizationName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakOrganizationDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOrganization_basic(organizationName),
				Check:  testAccCheckKeycloakOrganizationExists("keycloak_organization.organization"),
			},
		},
	})
}

func TestAccKeycloakOrganization_basicUpdate(t *testing.T) {
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
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakOrganizationDestroy(),
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
	var organization = &keycloak.Organization{}

	organizationName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakOrganizationDestroy(),
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
	organizationName := acctest.RandomWithPrefix("tf-acc")
	extraDomain := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakOrganizationDestroy(),
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
	organizationName := acctest.RandomWithPrefix("tf-acc")
	attributeName := acctest.RandomWithPrefix("tf-acc")
	attributeValue := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakOrganizationDestroy(),
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

func TestAccKeycloakOrganization_multiValuedAttributeNoDrift(t *testing.T) {
	organizationName := acctest.RandomWithPrefix("tf-acc")
	attributeName := acctest.RandomWithPrefix("tf-acc-tenant-roles")

	orderedValues := []string{"role-1", "role-2", "role-3", "role-4", "role-5"}
	attributeValue := strings.Join(orderedValues, MULTIVALUE_ATTRIBUTE_SEPARATOR)

	resourceName := "keycloak_organization.organization"

	var organization *keycloak.Organization

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakOrganizationDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOrganization_attributes(organizationName, attributeName, attributeValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOrganizationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "attributes."+attributeName, attributeValue),
					func(s *terraform.State) error {
						var err error
						organization, err = getOrganizationFromState(s, resourceName)
						return err
					},
				),
			},
			{
				PreConfig: func() {
					// Keycloak won't persist a different order for a collection that already
					// holds the same items, so the values have to be replaced with a dummy
					// before they can be written back in a different order
					fetchedOrganization, err := keycloakClient.GetOrganization(testCtx, organization.Realm, organization.Id)
					if err != nil {
						t.Fatal(err)
					}
					fetchedOrganization.Attributes = map[string][]string{
						attributeName: {"dummy"},
					}
					if err = keycloakClient.UpdateOrganization(testCtx, fetchedOrganization); err != nil {
						t.Fatal(err)
					}
					fetchedOrganization.Attributes = map[string][]string{
						attributeName: {"role-3", "role-5", "role-1", "role-4", "role-2"},
					}
					if err = keycloakClient.UpdateOrganization(testCtx, fetchedOrganization); err != nil {
						t.Fatal(err)
					}
				},
				Config:             testKeycloakOrganization_attributes(organizationName, attributeName, attributeValue),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccKeycloakOrganization_identityProviders(t *testing.T) {
	skipIfVersionIsLessThan(testCtx, t, keycloakClient, keycloak.Version_26_8)

	organizationName := acctest.RandomWithPrefix("tf-acc")
	domainName := acctest.RandomWithPrefix("tf-acc") + ".example.com"
	firstAlias := acctest.RandomWithPrefix("tf-acc")
	secondAlias := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_organization.organization"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakOrganizationDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOrganization_identityProviders(organizationName, domainName, firstAlias, secondAlias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOrganizationHasIdentityProvider(resourceName, firstAlias, true, "UNMANAGED"),
					testAccCheckKeycloakOrganizationHasIdentityProvider(resourceName, secondAlias, false, "UNMANAGED"),
					testAccCheckKeycloakOrganizationDomainRoutesTo(resourceName, domainName, firstAlias, true),
				),
			},
			{
				// An import reads no identity_provider blocks, and so no domain routing either.
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdPrefix:     testAccRealm.Realm + "/",
				ImportStateVerifyIgnore: []string{"identity_provider", "domain"},
			},
			{
				Config: testKeycloakOrganization_identityProvidersUpdate(organizationName, domainName, firstAlias, secondAlias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOrganizationHasIdentityProvider(resourceName, firstAlias, true, "MANAGED"),
					testAccCheckKeycloakOrganizationHasNoIdentityProvider(resourceName, secondAlias),
					testAccCheckKeycloakOrganizationDomainRoutesTo(resourceName, domainName, firstAlias, false),
				),
			},
			{
				Config: testKeycloakOrganization_identityProvidersRemoved(organizationName, domainName, firstAlias, secondAlias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOrganizationHasNoIdentityProvider(resourceName, firstAlias),
					testAccCheckKeycloakOrganizationHasNoIdentityProvider(resourceName, secondAlias),
					testAccCheckKeycloakOrganizationDomainRoutesTo(resourceName, domainName, "", false),
				),
			},
		},
	})
}

func TestAccKeycloakOrganization_identityProvidersKeepOtherLinks(t *testing.T) {
	skipIfVersionIsLessThan(testCtx, t, keycloakClient, keycloak.Version_26_8)

	organizationName := acctest.RandomWithPrefix("tf-acc")
	domainName := acctest.RandomWithPrefix("tf-acc") + ".example.com"
	firstAlias := acctest.RandomWithPrefix("tf-acc")
	secondAlias := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_organization.organization"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakOrganizationDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOrganization_identityProvidersWithOrganizationId(organizationName, domainName, firstAlias, secondAlias, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOrganizationHasIdentityProvider(resourceName, firstAlias, true, "UNMANAGED"),
					testAccCheckKeycloakOrganizationHasIdentityProvider(resourceName, secondAlias, true, "MANAGED"),
				),
			},
			{
				Config: testKeycloakOrganization_identityProvidersWithOrganizationId(organizationName, domainName, firstAlias, secondAlias, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOrganizationHasNoIdentityProvider(resourceName, firstAlias),
					testAccCheckKeycloakOrganizationHasIdentityProvider(resourceName, secondAlias, true, "MANAGED"),
				),
			},
		},
	})
}

func TestAccKeycloakOrganization_sharedIdentityProvider(t *testing.T) {
	skipIfVersionIsLessThan(testCtx, t, keycloakClient, keycloak.Version_26_8)

	firstOrganizationName := acctest.RandomWithPrefix("tf-acc")
	secondOrganizationName := acctest.RandomWithPrefix("tf-acc")
	firstDomainName := acctest.RandomWithPrefix("tf-acc") + ".example.com"
	secondDomainName := acctest.RandomWithPrefix("tf-acc") + ".example.com"
	alias := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakOrganizationDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOrganization_sharedIdentityProvider(firstOrganizationName, secondOrganizationName, firstDomainName, secondDomainName, alias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOrganizationHasIdentityProvider("keycloak_organization.first", alias, true, "UNMANAGED"),
					testAccCheckKeycloakOrganizationDomainRoutesTo("keycloak_organization.first", firstDomainName, alias, true),
					testAccCheckKeycloakOrganizationHasIdentityProvider("keycloak_organization.second", alias, false, "UNMANAGED"),
					testAccCheckKeycloakOrganizationDomainRoutesTo("keycloak_organization.second", secondDomainName, alias, false),
				),
			},
		},
	})
}

func TestAccKeycloakOrganization_domainRoutingValidation(t *testing.T) {
	organizationName := acctest.RandomWithPrefix("tf-acc")
	domainName := acctest.RandomWithPrefix("tf-acc") + ".example.com"
	alias := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakOrganizationDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakOrganization_domainRoutedToUnlinkedIdentityProvider(organizationName, domainName, alias),
				ExpectError: regexp.MustCompile("which is not in an identity_provider block"),
			},
			{
				Config:      testKeycloakOrganization_domainAutoRedirectWithoutIdentityProvider(organizationName, domainName),
				ExpectError: regexp.MustCompile("requires identity_provider_alias"),
			},
			{
				Config:      testKeycloakOrganization_managedIdentityProviderWithoutAutoMembership(organizationName, domainName, alias),
				ExpectError: regexp.MustCompile("membership_type MANAGED requires auto_membership"),
			},
			{
				Config:      testKeycloakOrganization_duplicateIdentityProvider(organizationName, domainName, alias),
				ExpectError: regexp.MustCompile("is in more than one identity_provider block"),
			},
		},
	})
}

func TestAccKeycloakOrganization_identityProviderUnlinkedOutOfBand(t *testing.T) {
	skipIfVersionIsLessThan(testCtx, t, keycloakClient, keycloak.Version_26_8)

	organization := &keycloak.Organization{}
	organizationName := acctest.RandomWithPrefix("tf-acc")
	domainName := acctest.RandomWithPrefix("tf-acc") + ".example.com"
	firstAlias := acctest.RandomWithPrefix("tf-acc")
	secondAlias := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_organization.organization"

	unlinkSecond := func() {
		err := keycloakClient.UnlinkIdentityProviderFromOrganization(testCtx, organization.Realm, secondAlias, organization.Id)
		if err != nil {
			t.Fatal(err)
		}
	}

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakOrganizationDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOrganization_identityProviders(organizationName, domainName, firstAlias, secondAlias),
				Check:  testAccCheckKeycloakOrganizationFetch(resourceName, organization),
			},
			{
				PreConfig:          unlinkSecond,
				Config:             testKeycloakOrganization_identityProviders(organizationName, domainName, firstAlias, secondAlias),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testKeycloakOrganization_identityProviders(organizationName, domainName, firstAlias, secondAlias),
				Check:  testAccCheckKeycloakOrganizationHasIdentityProvider(resourceName, secondAlias, false, "UNMANAGED"),
			},
			{
				PreConfig: unlinkSecond,
				Config:    testKeycloakOrganization_identityProvidersUpdate(organizationName, domainName, firstAlias, secondAlias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOrganizationHasIdentityProvider(resourceName, firstAlias, true, "MANAGED"),
					testAccCheckKeycloakOrganizationHasNoIdentityProvider(resourceName, secondAlias),
				),
			},
		},
	})
}

func TestAccKeycloakOrganization_identityProvidersFromOrganizationId(t *testing.T) {
	skipIfVersionIsLessThan(testCtx, t, keycloakClient, keycloak.Version_26_8)

	organization := &keycloak.Organization{}
	organizationName := acctest.RandomWithPrefix("tf-acc")
	domainName := acctest.RandomWithPrefix("tf-acc") + ".example.com"
	alias := acctest.RandomWithPrefix("tf-acc")
	username := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_organization.organization"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakOrganizationDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOrganization_organizationIdOnIdentityProvider(organizationName, domainName, alias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOrganizationHasIdentityProvider(resourceName, alias, true, "MANAGED"),
					testAccCheckKeycloakOrganizationDomainRoutesTo(resourceName, domainName, alias, true),
					testAccCheckKeycloakOrganizationFetch(resourceName, organization),
				),
			},
			{
				// The move removes the link and adds it again. A managed member must stay: removing
				// a managed member deletes the user.
				PreConfig: func() {
					user := &keycloak.User{RealmId: organization.Realm, Username: username, Enabled: true}
					if err := keycloakClient.NewUser(testCtx, user); err != nil {
						t.Fatal(err)
					}
					if err := keycloakClient.AddUserToOrganization(testCtx, organization.Realm, organization.Id, user.Id); err != nil {
						t.Fatal(err)
					}
					if err := keycloakClient.UpdateOrganizationMembershipType(testCtx, organization.Realm, organization.Id, user.Id, "MANAGED"); err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakOrganization_organizationIdReplacedByBlock(organizationName, domainName, alias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOrganizationHasIdentityProvider(resourceName, alias, true, "MANAGED"),
					testAccCheckKeycloakOrganizationDomainRoutesTo(resourceName, domainName, alias, true),
					testAccCheckKeycloakOrganizationHasManagedMember(resourceName, username),
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

func testAccCheckKeycloakOrganizationHasIdentityProvider(resourceName, alias string, autoMembership bool, membershipType string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		organization, err := getOrganizationFromState(state, resourceName)
		if err != nil {
			return err
		}

		identityProvider, err := keycloakClient.GetOrganizationIdentityProvider(testCtx, organization.Realm, organization.Id, alias)
		if err != nil {
			return err
		}

		if identityProvider == nil {
			return fmt.Errorf("expected organization %s to have identity provider %s", organization.Name, alias)
		}

		if identityProvider.AutoMembership != autoMembership || identityProvider.MembershipType != membershipType {
			return fmt.Errorf("expected identity provider %s of organization %s to have auto membership %t and membership type %s, got %t and %s", alias, organization.Name, autoMembership, membershipType, identityProvider.AutoMembership, identityProvider.MembershipType)
		}

		return nil
	}
}

func testAccCheckKeycloakOrganizationHasNoIdentityProvider(resourceName, alias string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		organization, err := getOrganizationFromState(state, resourceName)
		if err != nil {
			return err
		}

		identityProvider, err := keycloakClient.GetOrganizationIdentityProvider(testCtx, organization.Realm, organization.Id, alias)
		if err != nil {
			return err
		}

		if identityProvider != nil {
			return fmt.Errorf("expected organization %s to not have identity provider %s", organization.Name, alias)
		}

		return nil
	}
}

func testAccCheckKeycloakOrganizationHasManagedMember(resourceName, username string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		organization, err := getOrganizationFromState(state, resourceName)
		if err != nil {
			return err
		}

		members, err := keycloakClient.GetOrganizationMembers(testCtx, organization.Realm, organization.Id)
		if err != nil {
			return err
		}

		for _, member := range members {
			if member.Username != username {
				continue
			}
			if member.MembershipType != "MANAGED" {
				return fmt.Errorf("expected member %s of organization %s to be MANAGED, got %s", username, organization.Name, member.MembershipType)
			}
			return nil
		}

		return fmt.Errorf("expected organization %s to have managed member %s", organization.Name, username)
	}
}

func testAccCheckKeycloakOrganizationDomainRoutesTo(resourceName, domainName, alias string, autoRedirect bool) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		organization, err := getOrganizationFromState(state, resourceName)
		if err != nil {
			return err
		}

		for _, domain := range organization.Domains {
			if domain.Name != domainName {
				continue
			}
			if domain.IdentityProviderAlias != alias || domain.AutoRedirect != autoRedirect {
				return fmt.Errorf("expected domain %s of organization %s to route to identity provider %q with auto redirect %t, got %q and %t", domainName, organization.Name, alias, autoRedirect, domain.IdentityProviderAlias, domain.AutoRedirect)
			}
			return nil
		}

		return fmt.Errorf("expected organization %s to have domain %s", organization.Name, domainName)
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

func testKeycloakOrganization_identityProviderResources(firstAlias, secondAlias string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_oidc_identity_provider" "first" {
	realm             = data.keycloak_realm.realm.id
	alias             = "%s"
	authorization_url = "https://example.com/auth"
	token_url         = "https://example.com/token"
	client_id         = "example_id"
	client_secret     = "example_token"
}

resource "keycloak_oidc_identity_provider" "second" {
	realm             = data.keycloak_realm.realm.id
	alias             = "%s"
	authorization_url = "https://example.com/auth"
	token_url         = "https://example.com/token"
	client_id         = "example_id"
	client_secret     = "example_token"
}
	`, testAccRealm.Realm, firstAlias, secondAlias)
}

func testKeycloakOrganization_identityProviders(organizationName, domainName, firstAlias, secondAlias string) string {
	return testKeycloakOrganization_identityProviderResources(firstAlias, secondAlias) + fmt.Sprintf(`
resource "keycloak_organization" "organization" {
	name  = "%s"
	realm = data.keycloak_realm.realm.id

	identity_provider {
		alias = keycloak_oidc_identity_provider.first.alias
	}

	identity_provider {
		alias           = keycloak_oidc_identity_provider.second.alias
		auto_membership = false
	}

	domain {
		name                    = "%s"
		identity_provider_alias = keycloak_oidc_identity_provider.first.alias
		auto_redirect           = true
	}
}
	`, organizationName, domainName)
}

func testKeycloakOrganization_identityProvidersUpdate(organizationName, domainName, firstAlias, secondAlias string) string {
	return testKeycloakOrganization_identityProviderResources(firstAlias, secondAlias) + fmt.Sprintf(`
resource "keycloak_organization" "organization" {
	name  = "%s"
	realm = data.keycloak_realm.realm.id

	identity_provider {
		alias           = keycloak_oidc_identity_provider.first.alias
		membership_type = "MANAGED"
	}

	domain {
		name                    = "%s"
		identity_provider_alias = keycloak_oidc_identity_provider.first.alias
	}
}
	`, organizationName, domainName)
}

func testKeycloakOrganization_identityProvidersRemoved(organizationName, domainName, firstAlias, secondAlias string) string {
	return testKeycloakOrganization_identityProviderResources(firstAlias, secondAlias) + fmt.Sprintf(`
resource "keycloak_organization" "organization" {
	name  = "%s"
	realm = data.keycloak_realm.realm.id

	domain {
		name = "%s"
	}
}
	`, organizationName, domainName)
}

func testKeycloakOrganization_identityProvidersWithOrganizationId(organizationName, domainName, firstAlias, secondAlias string, linkFirst bool) string {
	identityProvider := ""
	if linkFirst {
		identityProvider = `
	identity_provider {
		alias = keycloak_oidc_identity_provider.first.alias
	}`
	}

	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_oidc_identity_provider" "first" {
	realm             = data.keycloak_realm.realm.id
	alias             = "%s"
	authorization_url = "https://example.com/auth"
	token_url         = "https://example.com/token"
	client_id         = "example_id"
	client_secret     = "example_token"
}

resource "keycloak_organization" "organization" {
	name  = "%s"
	realm = data.keycloak_realm.realm.id
%s

	domain {
		name = "%s"
	}
}

resource "keycloak_oidc_identity_provider" "second" {
	realm             = data.keycloak_realm.realm.id
	alias             = "%s"
	authorization_url = "https://example.com/auth"
	token_url         = "https://example.com/token"
	client_id         = "example_id"
	client_secret     = "example_token"

	organization_id = keycloak_organization.organization.id
}
	`, testAccRealm.Realm, firstAlias, organizationName, identityProvider, domainName, secondAlias)
}

func testKeycloakOrganization_organizationIdOnIdentityProvider(organizationName, domainName, alias string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_organization" "organization" {
	name  = "%s"
	realm = data.keycloak_realm.realm.id

	domain {
		name = "%s"
	}
}

resource "keycloak_oidc_identity_provider" "oidc" {
	realm             = data.keycloak_realm.realm.id
	alias             = "%s"
	authorization_url = "https://example.com/auth"
	token_url         = "https://example.com/token"
	client_id         = "example_id"
	client_secret     = "example_token"

	organization_id                 = keycloak_organization.organization.id
	org_domain                      = "%s"
	org_redirect_mode_email_matches = true
}
	`, testAccRealm.Realm, organizationName, domainName, alias, domainName)
}

func testKeycloakOrganization_organizationIdReplacedByBlock(organizationName, domainName, alias string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_oidc_identity_provider" "oidc" {
	realm             = data.keycloak_realm.realm.id
	alias             = "%s"
	authorization_url = "https://example.com/auth"
	token_url         = "https://example.com/token"
	client_id         = "example_id"
	client_secret     = "example_token"
}

resource "keycloak_organization" "organization" {
	name  = "%s"
	realm = data.keycloak_realm.realm.id

	identity_provider {
		alias           = keycloak_oidc_identity_provider.oidc.alias
		membership_type = "MANAGED"
	}

	domain {
		name                    = "%s"
		identity_provider_alias = keycloak_oidc_identity_provider.oidc.alias
		auto_redirect           = true
	}
}
	`, testAccRealm.Realm, alias, organizationName, domainName)
}

func testKeycloakOrganization_managedIdentityProviderWithoutAutoMembership(organizationName, domainName, alias string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_organization" "organization" {
	name  = "%s"
	realm = data.keycloak_realm.realm.id

	identity_provider {
		alias           = "%s"
		auto_membership = false
		membership_type = "MANAGED"
	}

	domain {
		name = "%s"
	}
}
	`, testAccRealm.Realm, organizationName, alias, domainName)
}

func testKeycloakOrganization_duplicateIdentityProvider(organizationName, domainName, alias string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_organization" "organization" {
	name  = "%s"
	realm = data.keycloak_realm.realm.id

	identity_provider {
		alias           = "%s"
		auto_membership = false
	}

	identity_provider {
		alias           = "%s"
		auto_membership = true
	}

	domain {
		name = "%s"
	}
}
	`, testAccRealm.Realm, organizationName, alias, alias, domainName)
}

func testKeycloakOrganization_sharedIdentityProvider(firstOrganizationName, secondOrganizationName, firstDomainName, secondDomainName, alias string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_oidc_identity_provider" "shared" {
	realm             = data.keycloak_realm.realm.id
	alias             = "%s"
	authorization_url = "https://example.com/auth"
	token_url         = "https://example.com/token"
	client_id         = "example_id"
	client_secret     = "example_token"
}

resource "keycloak_organization" "first" {
	name  = "%s"
	realm = data.keycloak_realm.realm.id

	identity_provider {
		alias = keycloak_oidc_identity_provider.shared.alias
	}

	domain {
		name                    = "%s"
		identity_provider_alias = keycloak_oidc_identity_provider.shared.alias
		auto_redirect           = true
	}
}

resource "keycloak_organization" "second" {
	name  = "%s"
	realm = data.keycloak_realm.realm.id

	identity_provider {
		alias           = keycloak_oidc_identity_provider.shared.alias
		auto_membership = false
	}

	domain {
		name                    = "%s"
		identity_provider_alias = keycloak_oidc_identity_provider.shared.alias
	}
}
	`, testAccRealm.Realm, alias, firstOrganizationName, firstDomainName, secondOrganizationName, secondDomainName)
}

func testKeycloakOrganization_domainRoutedToUnlinkedIdentityProvider(organizationName, domainName, alias string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_organization" "organization" {
	name  = "%s"
	realm = data.keycloak_realm.realm.id

	domain {
		name                    = "%s"
		identity_provider_alias = "%s"
	}
}
	`, testAccRealm.Realm, organizationName, domainName, alias)
}

func testKeycloakOrganization_domainAutoRedirectWithoutIdentityProvider(organizationName, domainName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_organization" "organization" {
	name  = "%s"
	realm = data.keycloak_realm.realm.id

	domain {
		name          = "%s"
		auto_redirect = true
	}
}
	`, testAccRealm.Realm, organizationName, domainName)
}
