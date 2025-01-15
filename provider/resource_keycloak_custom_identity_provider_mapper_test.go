package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakCustomIdentityProviderMapper_basic(t *testing.T) {
	t.Parallel()
	mapperName := acctest.RandomWithPrefix("tf-acc")
	alias := acctest.RandomWithPrefix("tf-acc")
	mapperType := "oidc-user-attribute-idp-mapper"
	userAttribute := acctest.RandomWithPrefix("tf-acc")
	claimName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakCustomIdentityProviderMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakCustomIdentityProviderMapper_basic(alias, mapperType, mapperName, userAttribute, claimName),
				Check:  testAccCheckKeycloakCustomIdentityProviderMapperExists("keycloak_custom_identity_provider_mapper.oidc"),
			},
		},
	})
}

func TestAccKeycloakCustomIdentityProviderMapper_withExtraConfig(t *testing.T) {
	t.Parallel()
	mapperName := acctest.RandomWithPrefix("tf-acc")
	alias := acctest.RandomWithPrefix("tf-acc")
	mapperType := "oidc-user-attribute-idp-mapper"
	userAttribute := acctest.RandomWithPrefix("tf-acc")
	claimName := acctest.RandomWithPrefix("tf-acc")
	syncMode := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakCustomIdentityProviderMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakCustomIdentityProviderMapper_withExtraConfig(alias, mapperType, mapperName, userAttribute, claimName, syncMode),
				Check:  testAccCheckKeycloakCustomIdentityProviderMapperExists("keycloak_custom_identity_provider_mapper.oidc"),
			},
		},
	})
}

func TestAccKeycloakCustomIdentityProviderMapper_createAfterManualDestroy(t *testing.T) {
	t.Parallel()
	var mapper = &keycloak.IdentityProviderMapper{}

	mapperName := acctest.RandomWithPrefix("tf-acc")
	alias := acctest.RandomWithPrefix("tf-acc")
	mapperType := "oidc-user-attribute-idp-mapper"
	userAttribute := acctest.RandomWithPrefix("tf-acc")
	claimName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakCustomIdentityProviderMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakCustomIdentityProviderMapper_basic(alias, mapperType, mapperName, userAttribute, claimName),
				Check:  testAccCheckKeycloakCustomIdentityProviderMapperFetch("keycloak_custom_identity_provider_mapper.oidc", mapper),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteIdentityProviderMapper(testCtx, mapper.Realm, mapper.IdentityProviderAlias, mapper.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakCustomIdentityProviderMapper_basic(alias, mapperType, mapperName, userAttribute, claimName),
				Check:  testAccCheckKeycloakCustomIdentityProviderMapperExists("keycloak_custom_identity_provider_mapper.oidc"),
			},
		},
	})
}

func TestAccKeycloakCustomIdentityProviderMapper_withExtraConfig_createAfterManualDestroy(t *testing.T) {
	t.Parallel()
	var mapper = &keycloak.IdentityProviderMapper{}

	mapperName := acctest.RandomWithPrefix("tf-acc")
	alias := acctest.RandomWithPrefix("tf-acc")
	mapperType := "oidc-user-attribute-idp-mapper"
	userAttribute := acctest.RandomWithPrefix("tf-acc")
	claimName := acctest.RandomWithPrefix("tf-acc")
	syncMode := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakCustomIdentityProviderMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakCustomIdentityProviderMapper_withExtraConfig(alias, mapperType, mapperName, userAttribute, claimName, syncMode),
				Check:  testAccCheckKeycloakCustomIdentityProviderMapperFetch("keycloak_custom_identity_provider_mapper.oidc", mapper),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteIdentityProviderMapper(testCtx, mapper.Realm, mapper.IdentityProviderAlias, mapper.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakCustomIdentityProviderMapper_basic(alias, mapperType, mapperName, userAttribute, claimName),
				Check:  testAccCheckKeycloakCustomIdentityProviderMapperExists("keycloak_custom_identity_provider_mapper.oidc"),
			},
		},
	})
}

func TestAccKeycloakCustomIdentityProviderMapper_basicUpdateAll(t *testing.T) {
	t.Parallel()
	identityProviderAliasName := acctest.RandomWithPrefix("tf-acc")
	identityProviderMapper := "saml-user-attribute-idp-mapper"

	firstMapper := &keycloak.IdentityProviderMapper{
		Realm:                  testAccRealm.Realm,
		IdentityProviderAlias:  identityProviderAliasName,
		IdentityProviderMapper: identityProviderMapper,
		Name:                   acctest.RandString(10),
		Config: &keycloak.IdentityProviderMapperConfig{
			UserAttribute: acctest.RandString(10),
			Attribute:     acctest.RandString(10),
		},
	}

	secondMapper := &keycloak.IdentityProviderMapper{
		Realm:                  testAccRealm.Realm,
		IdentityProviderAlias:  identityProviderAliasName,
		IdentityProviderMapper: identityProviderMapper,
		Name:                   acctest.RandString(10),
		Config: &keycloak.IdentityProviderMapperConfig{
			UserAttribute: acctest.RandString(10),
			Attribute:     acctest.RandString(10),
		},
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakCustomIdentityProviderMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakCustomIdentityProviderMapper_basicFromInterface(firstMapper),
				Check:  testAccCheckKeycloakCustomIdentityProviderMapperExists("keycloak_custom_identity_provider_mapper.saml"),
			},
			{
				Config: testKeycloakCustomIdentityProviderMapper_basicFromInterface(secondMapper),
				Check:  testAccCheckKeycloakCustomIdentityProviderMapperExists("keycloak_custom_identity_provider_mapper.saml"),
			},
		},
	})
}

func testAccCheckKeycloakCustomIdentityProviderMapperExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getKeycloakCustomIdentityProviderMapperFromState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckKeycloakCustomIdentityProviderMapperFetch(resourceName string, mapper *keycloak.IdentityProviderMapper) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedMapper, err := getKeycloakCustomIdentityProviderMapperFromState(s, resourceName)
		if err != nil {
			return err
		}

		mapper.IdentityProviderAlias = fetchedMapper.IdentityProviderAlias
		mapper.Realm = fetchedMapper.Realm
		mapper.Id = fetchedMapper.Id

		return nil
	}
}

func testAccCheckKeycloakCustomIdentityProviderMapperDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_custom_identity_provider_mapper" {
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

func getKeycloakCustomIdentityProviderMapperFromState(s *terraform.State, resourceName string) (*keycloak.IdentityProviderMapper, error) {
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

func testKeycloakCustomIdentityProviderMapper_basic(alias, mapperType, name, userAttribute, claimName string) string {
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

resource keycloak_custom_identity_provider_mapper oidc {
	realm                    = data.keycloak_realm.realm.id
	name                     = "%s"
	identity_provider_alias  = keycloak_oidc_identity_provider.oidc.alias
	identity_provider_mapper = "%s"
	extra_config 			= {
		userAttribute = "%s"
		claim         = "%s"
	}
}
	`, testAccRealm.Realm, alias, name, mapperType, userAttribute, claimName)
}

func testKeycloakCustomIdentityProviderMapper_withExtraConfig(alias, mapperType, name, userAttribute, claimName, syncMode string) string {
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

resource keycloak_custom_identity_provider_mapper oidc {
	realm                    = data.keycloak_realm.realm.id
	name                     = "%s"
	identity_provider_alias  = keycloak_oidc_identity_provider.oidc.alias
	identity_provider_mapper = "%s"
	extra_config 			= {
		syncMode      = "%s"
		userAttribute = "%s"
		claim         = "%s"
	}
}
	`, testAccRealm.Realm, alias, name, mapperType, syncMode, userAttribute, claimName)
}

func testKeycloakCustomIdentityProviderMapper_basicFromInterface(mapper *keycloak.IdentityProviderMapper) string {
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

resource keycloak_custom_identity_provider_mapper saml {
	realm                    = data.keycloak_realm.realm.id
	name                     = "%s"
	identity_provider_alias  = keycloak_saml_identity_provider.saml.alias
	identity_provider_mapper = "%s"
	extra_config 			= {
		attribute     = "%s"
		userAttribute = "%s"
	}
}
	`, testAccRealm.Realm, mapper.IdentityProviderAlias, mapper.Name, mapper.IdentityProviderMapper, mapper.Config.Attribute, mapper.Config.UserAttribute)
}
