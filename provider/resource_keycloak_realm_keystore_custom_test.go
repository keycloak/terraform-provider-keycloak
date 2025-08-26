package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mrparkers/terraform-provider-keycloak/keycloak"
	"strconv"
	"testing"
)

const providerId = "rsa-generated"
const providerType = "org.keycloak.keys.KeyProvider"

func TestAccKeycloakRealmKeystoreCustom_basic(t *testing.T) {
	t.Parallel()

	name := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckRealmKeystoreCustomDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmKeystoreCustom_basic(name),
				Check:  testAccCheckRealmKeystoreCustomExists("keycloak_realm_keystore_custom.realm_custom"),
			},
		},
	})
}

func TestAccKeycloakRealmKeystoreCustom_createAfterManualDestroy(t *testing.T) {
	t.Parallel()

	var custom = &keycloak.RealmKeystoreCustom{}

	fullNameKeystoreName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckRealmKeystoreCustomDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmKeystoreCustom_basic(fullNameKeystoreName),
				Check:  testAccCheckRealmKeystoreCustomFetch("keycloak_realm_keystore_custom.realm_custom", custom),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteRealmKeystoreCustom(testCtx, custom.RealmId, custom.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakRealmKeystoreCustom_basic(fullNameKeystoreName),
				Check:  testAccCheckRealmKeystoreCustomFetch("keycloak_realm_keystore_custom.realm_custom", custom),
			},
		},
	})
}

func TestAccKeycloakRealmKeystoreCustom_updateCustomGenerated(t *testing.T) {
	t.Parallel()

	enabled := randomBool()
	active := randomBool()
	priority := acctest.RandIntRange(0, 100)

	groupKeystoreOne := &keycloak.RealmKeystoreCustom{
		Name:         acctest.RandString(10),
		RealmId:      testAccRealmUserFederation.Realm,
		Enabled:      enabled,
		Active:       active,
		Priority:     priority,
		ProviderId:   providerId,
		ProviderType: providerType,
		ExtraConfig: map[string]interface{}{
			"algorithm": "RS384",
		},
	}

	groupKeystoreTwo := &keycloak.RealmKeystoreCustom{
		Name:         acctest.RandString(10),
		RealmId:      testAccRealmUserFederation.Realm,
		Enabled:      enabled,
		Active:       active,
		Priority:     priority,
		ProviderId:   providerId,
		ProviderType: providerType,
		ExtraConfig: map[string]interface{}{
			"algorithm": "RS384",
		},
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckRealmKeystoreCustomDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmKeystoreCustom_basicFromInterface(groupKeystoreOne),
				Check:  testAccCheckRealmKeystoreCustomExists("keycloak_realm_keystore_custom.realm_custom"),
			},
			{
				Config: testKeycloakRealmKeystoreCustom_basicFromInterface(groupKeystoreTwo),
				Check:  testAccCheckRealmKeystoreCustomExists("keycloak_realm_keystore_custom.realm_custom"),
			},
		},
	})
}

func testAccCheckRealmKeystoreCustomExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getKeycloakRealmKeystoreCustomFromState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckRealmKeystoreCustomFetch(resourceName string, keystore *keycloak.RealmKeystoreCustom) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedKeystore, err := getKeycloakRealmKeystoreCustomFromState(s, resourceName)
		if err != nil {
			return err
		}

		keystore.Id = fetchedKeystore.Id
		keystore.RealmId = fetchedKeystore.RealmId

		return nil
	}
}

func testAccCheckRealmKeystoreCustomDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_realm_keystore_custom" {
				continue
			}

			id := rs.Primary.ID
			realm := rs.Primary.Attributes["realm_id"]

			ldapGroupKeystore, _ := keycloakClient.GetRealmKeystoreCustom(testCtx, realm, id)
			if ldapGroupKeystore != nil {
				return fmt.Errorf("custom keystore with id %s still exists", id)
			}
		}

		return nil
	}
}

func getKeycloakRealmKeystoreCustomFromState(s *terraform.State,
	resourceName string) (*keycloak.RealmKeystoreCustom,
	error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm_id"]

	realmKeystore, err := keycloakClient.GetRealmKeystoreCustom(testCtx, realm, id)
	if err != nil {
		return nil, fmt.Errorf("error getting custom keystore with id %s: %s", id, err)
	}

	return realmKeystore, nil
}

func testKeycloakRealmKeystoreCustom_basic(name string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_realm_keystore_custom" "realm_custom" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id

    provider_id   = "%s"
	provider_type = "%s"

	extra_config = {
		algorithm = "RS384"
	}
}
	`, testAccRealmUserFederation.Realm, name, providerId, providerType)
}

func testKeycloakRealmKeystoreCustom_basicFromInterface(keystore *keycloak.RealmKeystoreCustom) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_realm_keystore_custom" "realm_custom" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id

    priority      = %s
    provider_id   = "%s"
	provider_type = "%s"

	extra_config = {
		algorithm = "RS384"
	}
}
	`, testAccRealmUserFederation.Realm, keystore.Name, strconv.Itoa(keystore.Priority), providerId, providerType)
}
