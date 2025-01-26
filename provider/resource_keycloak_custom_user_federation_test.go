package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakCustomUserFederation_basic(t *testing.T) {
	t.Parallel()

	skipIfEnvSet(t, "CI") // temporary while I figure out how to load this custom provider in CI

	name := acctest.RandomWithPrefix("tf-acc")
	providerId := "custom"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakCustomUserFederationDestroy(),
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

	skipIfEnvSet(t, "CI") // temporary while I figure out how to load this custom provider in CI

	name := acctest.RandomWithPrefix("tf-acc")
	configValue := acctest.RandomWithPrefix("tf-acc")
	providerId := "custom"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakCustomUserFederationDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakCustomUserFederation_customConfig(name, providerId, configValue),
				Check:  testAccCheckKeycloakCustomUserFederationExistsWithCustomConfig("keycloak_custom_user_federation.custom", configValue),
			},
		},
	})

	configValue = configValue + "," + acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakCustomUserFederationDestroy(),
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

	skipIfEnvSet(t, "CI") // temporary while I figure out how to load this custom provider in CI

	var customFederation = &keycloak.CustomUserFederation{}

	name := acctest.RandomWithPrefix("tf-acc")
	providerId := "custom"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakCustomUserFederationDestroy(),
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
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakCustomUserFederationDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakCustomUserFederation_basic(name, providerId),
				ExpectError: regexp.MustCompile("custom user federation provider with id .+ is not installed on the server"),
			},
		},
	})
}

func TestAccKeycloakCustomUserFederation_ParentIdDifferentFromRealmName(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")
	internalId := acctest.RandomWithPrefix("tf-acc")
	name := acctest.RandomWithPrefix("tf-acc")
	providerId := "custom"

	realm := &keycloak.Realm{
		Realm: realmName,
		Id:    internalId,
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakCustomUserFederationDestroy(),
		Steps: []resource.TestStep{
			{
				ResourceName:  "keycloak_realm.realm",
				ImportStateId: realmName,
				ImportState:   true,
				PreConfig: func() {
					err := keycloakClient.NewRealm(testCtx, realm)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakCustomUserFederation_parentId(realmName, name, providerId, internalId),
				Check:  testAccCheckKeycloakCustomUserFederationExists("keycloak_custom_user_federation.custom"),
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

func testKeycloakCustomUserFederation_parentId(realm, name, providerId, parentId string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_custom_user_federation" "custom" {
	name        = "%s"
	realm_id    = keycloak_realm.realm.id
	provider_id = "%s"
    parent_id   = "%s"

	full_sync_period    = 30
	changed_sync_period = 60

	enabled     = true
}
	`, realm, name, providerId, parentId)
}
