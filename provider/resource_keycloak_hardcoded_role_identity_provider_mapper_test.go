package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakHardcodedRoleIdentityProviderMapper_basic(t *testing.T) {
	t.Parallel()

	mapperName := acctest.RandomWithPrefix("tf-acc")
	alias := acctest.RandomWithPrefix("tf-acc")
	role := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakHardcodedRoleIdentityProviderMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakHardcodedRoleIdentityProviderMapper_basic(alias, mapperName, role),
				Check:  testAccCheckKeycloakHardcodedRoleIdentityProviderMapperExists("keycloak_hardcoded_role_identity_provider_mapper.oidc"),
			},
		},
	})
}

func TestAccKeycloakHardcodedRoleIdentityProviderMapper_withExtraConfig(t *testing.T) {
	t.Parallel()

	mapperName := acctest.RandomWithPrefix("tf-acc")
	alias := acctest.RandomWithPrefix("tf-acc")
	role := acctest.RandomWithPrefix("tf-acc")
	syncMode := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakHardcodedRoleIdentityProviderMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakHardcodedRoleIdentityProviderMapper_withExtraConfig(alias, mapperName, role, syncMode),
				Check:  testAccCheckKeycloakHardcodedRoleIdentityProviderMapperExists("keycloak_hardcoded_role_identity_provider_mapper.oidc"),
			},
		},
	})
}

func TestAccKeycloakHardcodedRoleIdentityProviderMapper_createAfterManualDestroy(t *testing.T) {
	t.Parallel()

	var mapper = &keycloak.IdentityProviderMapper{}

	mapperName := acctest.RandomWithPrefix("tf-acc")
	alias := acctest.RandomWithPrefix("tf-acc")
	role := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakHardcodedRoleIdentityProviderMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakHardcodedRoleIdentityProviderMapper_basic(alias, mapperName, role),
				Check:  testAccCheckKeycloakHardcodedRoleIdentityProviderMapperFetch("keycloak_hardcoded_role_identity_provider_mapper.oidc", mapper),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteIdentityProviderMapper(testCtx, mapper.Realm, mapper.IdentityProviderAlias, mapper.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakHardcodedRoleIdentityProviderMapper_basic(alias, mapperName, role),
				Check:  testAccCheckKeycloakHardcodedRoleIdentityProviderMapperExists("keycloak_hardcoded_role_identity_provider_mapper.oidc"),
			},
		},
	})
}

func TestAccKeycloakHardcodedRoleIdentityProviderMapper_withExtraConfig_createAfterManualDestroy(t *testing.T) {
	t.Parallel()

	var mapper = &keycloak.IdentityProviderMapper{}

	mapperName := acctest.RandomWithPrefix("tf-acc")
	alias := acctest.RandomWithPrefix("tf-acc")
	role := acctest.RandomWithPrefix("tf-acc")
	syncMode := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakHardcodedRoleIdentityProviderMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakHardcodedRoleIdentityProviderMapper_withExtraConfig(alias, mapperName, role, syncMode),
				Check:  testAccCheckKeycloakHardcodedRoleIdentityProviderMapperFetch("keycloak_hardcoded_role_identity_provider_mapper.oidc", mapper),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteIdentityProviderMapper(testCtx, mapper.Realm, mapper.IdentityProviderAlias, mapper.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakHardcodedRoleIdentityProviderMapper_basic(alias, mapperName, role),
				Check:  testAccCheckKeycloakHardcodedRoleIdentityProviderMapperExists("keycloak_hardcoded_role_identity_provider_mapper.oidc"),
			},
		},
	})
}

func TestAccKeycloakHardcodedRoleIdentityProviderMapper_basicUpdateAll(t *testing.T) {
	t.Parallel()

	identityProviderAliasName := acctest.RandomWithPrefix("tf-acc")

	firstMapper := &keycloak.IdentityProviderMapper{
		Realm:                 testAccRealm.Realm,
		IdentityProviderAlias: identityProviderAliasName,
		Name:                  acctest.RandString(10),
		Config: &keycloak.IdentityProviderMapperConfig{
			Role: acctest.RandString(10),
		},
	}

	secondMapper := &keycloak.IdentityProviderMapper{
		Realm:                 testAccRealm.Realm,
		IdentityProviderAlias: identityProviderAliasName,
		Name:                  acctest.RandString(10),
		Config: &keycloak.IdentityProviderMapperConfig{
			Role: acctest.RandString(10),
		},
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakHardcodedRoleIdentityProviderMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakHardcodedRoleIdentityProviderMapper_basicFromInterface(firstMapper),
				Check:  testAccCheckKeycloakHardcodedRoleIdentityProviderMapperExists("keycloak_hardcoded_role_identity_provider_mapper.saml"),
			},
			{
				Config: testKeycloakHardcodedRoleIdentityProviderMapper_basicFromInterface(secondMapper),
				Check:  testAccCheckKeycloakHardcodedRoleIdentityProviderMapperExists("keycloak_hardcoded_role_identity_provider_mapper.saml"),
			},
		},
	})
}

func testAccCheckKeycloakHardcodedRoleIdentityProviderMapperExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getKeycloakHardcodedRoleIdentityProviderMapperFromState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckKeycloakHardcodedRoleIdentityProviderMapperFetch(resourceName string, mapper *keycloak.IdentityProviderMapper) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedMapper, err := getKeycloakHardcodedRoleIdentityProviderMapperFromState(s, resourceName)
		if err != nil {
			return err
		}

		mapper.IdentityProviderAlias = fetchedMapper.IdentityProviderAlias
		mapper.Realm = fetchedMapper.Realm
		mapper.Id = fetchedMapper.Id

		return nil
	}
}

func testAccCheckKeycloakHardcodedRoleIdentityProviderMapperDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_hardcoded_role_identity_provider_mapper" {
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

func getKeycloakHardcodedRoleIdentityProviderMapperFromState(s *terraform.State, resourceName string) (*keycloak.IdentityProviderMapper, error) {
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

func testKeycloakHardcodedRoleIdentityProviderMapper_basic(alias, name, role string) string {
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

resource keycloak_hardcoded_role_identity_provider_mapper oidc {
	realm                   = data.keycloak_realm.realm.id
	name                    = "%s"
	identity_provider_alias = keycloak_oidc_identity_provider.oidc.alias
	role                    = "%s"
}
	`, testAccRealm.Realm, alias, name, role)
}

func testKeycloakHardcodedRoleIdentityProviderMapper_withExtraConfig(alias, name, role, syncMode string) string {
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

resource keycloak_hardcoded_role_identity_provider_mapper oidc {
	realm                   = data.keycloak_realm.realm.id
	name                    = "%s"
	identity_provider_alias = keycloak_oidc_identity_provider.oidc.alias
	role                    = "%s"
	extra_config 			= {
		syncMode = "%s"
	}
}
	`, testAccRealm.Realm, alias, name, role, syncMode)
}

func testKeycloakHardcodedRoleIdentityProviderMapper_basicFromInterface(mapper *keycloak.IdentityProviderMapper) string {
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

resource keycloak_hardcoded_role_identity_provider_mapper saml {
	realm                   = data.keycloak_realm.realm.id
	name                    = "%s"
	identity_provider_alias = keycloak_saml_identity_provider.saml.alias
	role                    = "%s"
}
	`, testAccRealm.Realm, mapper.IdentityProviderAlias, mapper.Name, mapper.Config.Role)
}
