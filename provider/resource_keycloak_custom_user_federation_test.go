package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakCustomUserFederation_basic(t *testing.T) {
	t.Parallel()

	name := acctest.RandomWithPrefix("tf-acc")
	providerId := "custom"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakCustomUserFederationDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakCustomUserFederation_basic(name, providerId),
				Check:  testAccCheckKeycloakCustomUserFederationExists("keycloak_custom_user_federation.custom"),
			},
			{
				ResourceName:        "keycloak_custom_user_federation.custom",
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: testAccRealm.Realm + "/",
			},
		},
	})
}

func TestAccKeycloakCustomUserFederation_customConfig(t *testing.T) {
	t.Parallel()

	name := acctest.RandomWithPrefix("tf-acc")
	configValue := acctest.RandomWithPrefix("tf-acc")
	providerId := "custom"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakCustomUserFederationDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakCustomUserFederation_customConfig(name, providerId, configValue),
				Check:  testAccCheckKeycloakCustomUserFederationExistsWithCustomConfig("keycloak_custom_user_federation.custom", configValue),
			},
		},
	})

	configValue = configValue + "," + acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakCustomUserFederationDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakCustomUserFederation_customConfig(name, providerId, configValue),
				Check:  testAccCheckKeycloakCustomUserFederationExistsWithCustomConfig("keycloak_custom_user_federation.custom", configValue),
			},
		},
	})
}

func TestAccKeycloakCustomUserFederation_createAfterManualDestroy(t *testing.T) {
	t.Parallel()

	var customFederation = &keycloak.CustomUserFederation{}

	name := acctest.RandomWithPrefix("tf-acc")
	providerId := "custom"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakCustomUserFederationDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakCustomUserFederation_basic(name, providerId),
				Check:  testAccCheckKeycloakCustomUserFederationFetch("keycloak_custom_user_federation.custom", customFederation),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteCustomUserFederation(testCtx, customFederation.RealmId, customFederation.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakCustomUserFederation_basic(name, providerId),
				Check:  testAccCheckKeycloakCustomUserFederationExists("keycloak_custom_user_federation.custom"),
			},
		},
	})
}

func TestAccKeycloakCustomUserFederation_validation(t *testing.T) {
	t.Parallel()

	name := acctest.RandomWithPrefix("tf-acc")
	providerId := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakCustomUserFederationDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakCustomUserFederation_basic(name, providerId),
				ExpectError: regexp.MustCompile("custom user federation provider with id .+ is not installed on the server"),
			},
		},
	})
}

func TestAccKeycloakCustomUserFederation_parentIdDifferentFromRealmName(t *testing.T) {
	t.Parallel()

	realmName := acctest.RandomWithPrefix("tf-acc")
	internalId := acctest.RandomWithPrefix("tf-acc")
	name := acctest.RandomWithPrefix("tf-acc")
	providerId := "custom"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakCustomUserFederationDestroy(),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					err := keycloakClient.NewRealm(testCtx, &keycloak.Realm{
						Realm: realmName,
						Id:    internalId,
					})
					if err != nil {
						t.Fatal(err)
					}

					t.Cleanup(func() {
						if err := keycloakClient.DeleteRealm(testCtx, realmName); err != nil {
							t.Logf("failed to clean up realm %s: %s", realmName, err)
						}
					})
				},
				// create: keycloak must default the omitted parentId to the realm's internal id
				Config: testKeycloakCustomUserFederation_parentId(realmName, name, providerId, "0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakCustomUserFederationExists("keycloak_custom_user_federation.custom"),
					testAccCheckKeycloakCustomUserFederationParentId("keycloak_custom_user_federation.custom", internalId),
					resource.TestCheckResourceAttr("keycloak_custom_user_federation.custom", "parent_id", internalId),
				),
			},
			{
				// update: the omitted parentId must leave the stored parent untouched
				Config: testKeycloakCustomUserFederation_parentId(realmName, name, providerId, "10"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakCustomUserFederationExists("keycloak_custom_user_federation.custom"),
					testAccCheckKeycloakCustomUserFederationParentId("keycloak_custom_user_federation.custom", internalId),
					resource.TestCheckResourceAttr("keycloak_custom_user_federation.custom", "parent_id", internalId),
				),
			},
		},
	})
}

// the deprecated parent_id attribute must keep working for existing configurations that set it explicitly
func TestAccKeycloakCustomUserFederation_explicitParentId(t *testing.T) {
	t.Parallel()

	realmName := acctest.RandomWithPrefix("tf-acc")
	internalId := acctest.RandomWithPrefix("tf-acc")
	name := acctest.RandomWithPrefix("tf-acc")
	providerId := "custom"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakCustomUserFederationDestroy(),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					err := keycloakClient.NewRealm(testCtx, &keycloak.Realm{
						Realm: realmName,
						Id:    internalId,
					})
					if err != nil {
						t.Fatal(err)
					}

					t.Cleanup(func() {
						if err := keycloakClient.DeleteRealm(testCtx, realmName); err != nil {
							t.Logf("failed to clean up realm %s: %s", realmName, err)
						}
					})
				},
				Config: testKeycloakCustomUserFederation_explicitParentId(realmName, name, providerId, internalId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakCustomUserFederationExists("keycloak_custom_user_federation.custom"),
					testAccCheckKeycloakCustomUserFederationParentId("keycloak_custom_user_federation.custom", internalId),
				),
			},
		},
	})
}

func testAccCheckKeycloakCustomUserFederationExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getCustomUserFederationFromState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckKeycloakCustomUserFederationExistsWithCustomConfig(resourceName, customConfigValue string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedFederation, err := getCustomUserFederationFromState(s, resourceName)
		if err != nil {
			return err
		}

		if len(fetchedFederation.Config["dummyConfig"]) <= 0 || fetchedFederation.Config["dummyConfig"][0] != customConfigValue {
			return fmt.Errorf("expected user federation provider to have config with a custom key 'dummyConfig' with a value %s", customConfigValue)
		}

		return nil
	}
}

func testAccCheckKeycloakCustomUserFederationFetch(resourceName string, federation *keycloak.CustomUserFederation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedFederation, err := getCustomUserFederationFromState(s, resourceName)
		if err != nil {
			return err
		}

		federation.Id = fetchedFederation.Id
		federation.RealmId = fetchedFederation.RealmId

		return nil
	}
}

func testAccCheckKeycloakCustomUserFederationParentId(resourceName, expectedParentId string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		custom, err := getCustomUserFederationFromState(s, resourceName)
		if err != nil {
			return err
		}

		if custom.ParentId != expectedParentId {
			return fmt.Errorf("expected custom user federation %s to have parent id %s, but got %s", custom.Id, expectedParentId, custom.ParentId)
		}

		return nil
	}
}

func testAccCheckKeycloakCustomUserFederationDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_custom_user_federation" {
				continue
			}

			id := rs.Primary.ID
			realm := rs.Primary.Attributes["realm_id"]

			custom, _ := keycloakClient.GetCustomUserFederation(testCtx, realm, id)
			if custom != nil {
				return fmt.Errorf("custom user federation with id %s still exists", id)
			}
		}

		return nil
	}
}

func getCustomUserFederationFromState(s *terraform.State, resourceName string) (*keycloak.CustomUserFederation, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm_id"]

	custom, err := keycloakClient.GetCustomUserFederation(testCtx, realm, id)
	if err != nil {
		return nil, fmt.Errorf("error getting custom user federation with id %s: %s", id, err)
	} else if custom.FullSyncPeriod != 30 {
		return nil, fmt.Errorf("expected fullSyncPeriod to equal %d, actual value = %d", 30, custom.FullSyncPeriod)
	} else if custom.ChangedSyncPeriod != 60 {
		return nil, fmt.Errorf("expected changedSyncPeriod to equal %d, actual value = %d", 60, custom.ChangedSyncPeriod)
	}

	return custom, nil
}

func testKeycloakCustomUserFederation_basic(name, providerId string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_custom_user_federation" "custom" {
	name        = "%s"
	realm_id    = data.keycloak_realm.realm.id
	provider_id = "%s"

	full_sync_period    = 30
	changed_sync_period = 60

	enabled     = true
}
	`, testAccRealm.Realm, name, providerId)
}

func testKeycloakCustomUserFederation_customConfig(name, providerId, customConfigValue string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_custom_user_federation" "custom" {
	name        = "%s"
	realm_id    = data.keycloak_realm.realm.id
	provider_id = "%s"

	full_sync_period    = 30
	changed_sync_period = 60

	enabled     = true

	config 		= {
		dummyConfig = "%s"
	}
}
	`, testAccRealm.Realm, name, providerId, customConfigValue)
}

func testKeycloakCustomUserFederation_parentId(realmName, name, providerId, priority string) string {
	return fmt.Sprintf(`
resource "keycloak_custom_user_federation" "custom" {
	name        = "%s"
	realm_id    = "%s"
	provider_id = "%s"
	priority    = %s

	full_sync_period    = 30
	changed_sync_period = 60

	enabled     = true
}
	`, name, realmName, providerId, priority)
}

func testKeycloakCustomUserFederation_explicitParentId(realmName, name, providerId, parentId string) string {
	return fmt.Sprintf(`
resource "keycloak_custom_user_federation" "custom" {
	name        = "%s"
	realm_id    = "%s"
	provider_id = "%s"
	parent_id   = "%s"

	full_sync_period    = 30
	changed_sync_period = 60

	enabled     = true
}
	`, name, realmName, providerId, parentId)
}
