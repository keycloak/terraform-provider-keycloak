package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakAttributeImporterIdentityProviderMapper_basic(t *testing.T) {
	t.Parallel()
	mapperName := acctest.RandomWithPrefix("tf-acc")
	alias := acctest.RandomWithPrefix("tf-acc")
	userAttribute := acctest.RandomWithPrefix("tf-acc")
	claimName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakAttributeImporterIdentityProviderMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakAttributeImporterIdentityProviderMapper_basic(alias, mapperName, userAttribute, claimName),
				Check:  testAccCheckKeycloakAttributeImporterIdentityProviderMapperExists("keycloak_attribute_importer_identity_provider_mapper.oidc"),
			},
		},
	})
}

func TestAccKeycloakAttributeImporterIdentityProviderMapper_withExtraConfig(t *testing.T) {
	t.Parallel()
	mapperName := acctest.RandomWithPrefix("tf-acc")
	alias := acctest.RandomWithPrefix("tf-acc")
	userAttribute := acctest.RandomWithPrefix("tf-acc")
	claimName := acctest.RandomWithPrefix("tf-acc")
	syncMode := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakAttributeImporterIdentityProviderMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakAttributeImporterIdentityProviderMapper_withExtraConfig(alias, mapperName, userAttribute, claimName, syncMode),
				Check:  testAccCheckKeycloakAttributeImporterIdentityProviderMapperExists("keycloak_attribute_importer_identity_provider_mapper.oidc"),
			},
		},
	})
}

func TestAccKeycloakAttributeImporterIdentityProviderMapper_createAfterManualDestroy(t *testing.T) {
	t.Parallel()
	var mapper = &keycloak.IdentityProviderMapper{}

	mapperName := acctest.RandomWithPrefix("tf-acc")
	alias := acctest.RandomWithPrefix("tf-acc")
	userAttribute := acctest.RandomWithPrefix("tf-acc")
	claimName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakAttributeImporterIdentityProviderMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakAttributeImporterIdentityProviderMapper_basic(alias, mapperName, userAttribute, claimName),
				Check:  testAccCheckKeycloakAttributeImporterIdentityProviderMapperFetch("keycloak_attribute_importer_identity_provider_mapper.oidc", mapper),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteIdentityProviderMapper(testCtx, mapper.Realm, mapper.IdentityProviderAlias, mapper.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakAttributeImporterIdentityProviderMapper_basic(alias, mapperName, userAttribute, claimName),
				Check:  testAccCheckKeycloakAttributeImporterIdentityProviderMapperExists("keycloak_attribute_importer_identity_provider_mapper.oidc"),
			},
		},
	})
}

func TestAccKeycloakAttributeImporterIdentityProviderMapper_withExtraConfig_createAfterManualDestroy(t *testing.T) {
	t.Parallel()
	var mapper = &keycloak.IdentityProviderMapper{}

	mapperName := acctest.RandomWithPrefix("tf-acc")
	alias := acctest.RandomWithPrefix("tf-acc")
	userAttribute := acctest.RandomWithPrefix("tf-acc")
	claimName := acctest.RandomWithPrefix("tf-acc")
	syncMode := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakAttributeImporterIdentityProviderMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakAttributeImporterIdentityProviderMapper_withExtraConfig(alias, mapperName, userAttribute, claimName, syncMode),
				Check:  testAccCheckKeycloakAttributeImporterIdentityProviderMapperFetch("keycloak_attribute_importer_identity_provider_mapper.oidc", mapper),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteIdentityProviderMapper(testCtx, mapper.Realm, mapper.IdentityProviderAlias, mapper.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakAttributeImporterIdentityProviderMapper_basic(alias, mapperName, userAttribute, claimName),
				Check:  testAccCheckKeycloakAttributeImporterIdentityProviderMapperExists("keycloak_attribute_importer_identity_provider_mapper.oidc"),
			},
		},
	})
}

func TestAccKeycloakAttributeImporterIdentityProviderMapper_basicUpdateAll(t *testing.T) {
	t.Parallel()
	identityProviderAliasName := acctest.RandomWithPrefix("tf-acc")

	firstMapper := &keycloak.IdentityProviderMapper{
		Realm:                 testAccRealm.Realm,
		IdentityProviderAlias: identityProviderAliasName,
		Name:                  acctest.RandString(10),
		Config: &keycloak.IdentityProviderMapperConfig{
			UserAttribute: acctest.RandString(10),
			Attribute:     acctest.RandString(10),
		},
	}

	secondMapper := &keycloak.IdentityProviderMapper{
		Realm:                 testAccRealm.Realm,
		IdentityProviderAlias: identityProviderAliasName,
		Name:                  acctest.RandString(10),
		Config: &keycloak.IdentityProviderMapperConfig{
			UserAttribute: acctest.RandString(10),
			Attribute:     acctest.RandString(10),
		},
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakAttributeImporterIdentityProviderMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakAttributeImporterIdentityProviderMapper_basicFromInterface(firstMapper),
				Check:  testAccCheckKeycloakAttributeImporterIdentityProviderMapperExists("keycloak_attribute_importer_identity_provider_mapper.saml"),
			},
			{
				Config: testKeycloakAttributeImporterIdentityProviderMapper_basicFromInterface(secondMapper),
				Check:  testAccCheckKeycloakAttributeImporterIdentityProviderMapperExists("keycloak_attribute_importer_identity_provider_mapper.saml"),
			},
		},
	})
}

func testAccCheckKeycloakAttributeImporterIdentityProviderMapperExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getKeycloakAttributeImporterIdentityProviderMapperFromState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckKeycloakAttributeImporterIdentityProviderMapperFetch(resourceName string, mapper *keycloak.IdentityProviderMapper) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedMapper, err := getKeycloakAttributeImporterIdentityProviderMapperFromState(s, resourceName)
		if err != nil {
			return err
		}

		mapper.IdentityProviderAlias = fetchedMapper.IdentityProviderAlias
		mapper.Realm = fetchedMapper.Realm
		mapper.Id = fetchedMapper.Id

		return nil
	}
}

func testAccCheckKeycloakAttributeImporterIdentityProviderMapperDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_attribute_importer_identity_provider_mapper" {
				continue
			}

			realm := rs.Primary.Attributes["realm"]
			alias := rs.Primary.Attributes["identity_provider_alias"]
			id := rs.Primary.ID

			mapper, _ := keycloakClient.GetIdentityProviderMapper(testCtx, realm, alias, id)
			if mapper != nil {
				return fmt.Errorf("oidc config with id %s still exists", id)
			}
		}

		return nil
	}
}

func getKeycloakAttributeImporterIdentityProviderMapperFromState(s *terraform.State, resourceName string) (*keycloak.IdentityProviderMapper, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	realm := rs.Primary.Attributes["realm"]
	alias := rs.Primary.Attributes["identity_provider_alias"]
	id := rs.Primary.ID

	mapper, err := keycloakClient.GetIdentityProviderMapper(testCtx, realm, alias, id)
	if err != nil {
		return nil, fmt.Errorf("error getting identity provider mapper config with id %s: %s", id, err)
	}

	return mapper, nil
}

func testKeycloakAttributeImporterIdentityProviderMapper_basic(alias, name, userAttribute, claimName string) string {
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

resource keycloak_attribute_importer_identity_provider_mapper oidc {
	realm                   = data.keycloak_realm.realm.id
	name                    = "%s"
	identity_provider_alias = keycloak_oidc_identity_provider.oidc.alias
	user_attribute          = "%s"
	claim_name              = "%s"
}
	`, testAccRealm.Realm, alias, name, userAttribute, claimName)
}

func testKeycloakAttributeImporterIdentityProviderMapper_withExtraConfig(alias, name, userAttribute, claimName, syncMode string) string {
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

resource keycloak_attribute_importer_identity_provider_mapper oidc {
	realm                   = data.keycloak_realm.realm.id
	name                    = "%s"
	identity_provider_alias = keycloak_oidc_identity_provider.oidc.alias
	user_attribute          = "%s"
	claim_name              = "%s"
	extra_config 			= {
		syncMode = "%s"
	}
}
	`, testAccRealm.Realm, alias, name, userAttribute, claimName, syncMode)
}

func testKeycloakAttributeImporterIdentityProviderMapper_basicFromInterface(mapper *keycloak.IdentityProviderMapper) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_saml_identity_provider" "saml" {
	realm                      = data.keycloak_realm.realm.id
	alias                      = "%s"
	entity_id                  = "https://example.com/entity_id"
	single_sign_on_service_url = "https://example.com/auth"
}

resource keycloak_attribute_importer_identity_provider_mapper saml {
	realm                   = data.keycloak_realm.realm.id
	name                    = "%s"
	identity_provider_alias = keycloak_saml_identity_provider.saml.alias
	attribute_name          = "%s"
	user_attribute          = "%s"
}
	`, testAccRealm.Realm, mapper.IdentityProviderAlias, mapper.Name, mapper.Config.Attribute, mapper.Config.UserAttribute)
}
