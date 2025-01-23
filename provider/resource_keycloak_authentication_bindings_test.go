package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccKeycloakAuthenticationBindings_browser(t *testing.T) {
	flow := "browser_flow"
	flowAlias := "browserCopyFlow"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakAuthenticationBindings(testAccRealm.Realm, flow, flowAlias),
				Check:  testAccCheckKeycloakAuthenticationBindingBrowserSet(testAccRealm.Realm, "BrowserFlow", flowAlias),
			},
		},
	})
}

// ensure that the separate flow bindings resource does not try to fight with the realm's flow binding settings
func TestAccKeycloakAuthenticationBindings_browserWithRealm(t *testing.T) {
	flow := "browser_flow"
	flowAlias := "browserCopyFlow"
	realmName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakAuthenticationBindingsWithRealm(realmName, flow, flowAlias),
				Check:  testAccCheckKeycloakAuthenticationBindingBrowserSet(realmName, "BrowserFlow", flowAlias),
			},
		},
	})
}

func TestAccKeycloakAuthenticationBindings_registration(t *testing.T) {
	flow := "registration_flow"
	flowAlias := "registrationCopyFlow"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakAuthenticationBindings(testAccRealm.Realm, flow, flowAlias),
				Check:  testAccCheckKeycloakAuthenticationBindingBrowserSet(testAccRealm.Realm, "RegistrationFlow", flowAlias),
			},
		},
	})
}

func TestAccKeycloakAuthenticationBindings_directGrant(t *testing.T) {
	flow := "direct_grant_flow"
	flowAlias := "directGrantCopyFlow"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakAuthenticationBindings(testAccRealm.Realm, flow, flowAlias),
				Check:  testAccCheckKeycloakAuthenticationBindingBrowserSet(testAccRealm.Realm, "DirectGrantFlow", flowAlias),
			},
		},
	})
}

func TestAccKeycloakAuthenticationBindings_resetCredentialsGrant(t *testing.T) {
	flow := "reset_credentials_flow"
	flowAlias := "resetCredentialsCopyFlow"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakAuthenticationBindings(testAccRealm.Realm, flow, flowAlias),
				Check:  testAccCheckKeycloakAuthenticationBindingBrowserSet(testAccRealm.Realm, "ResetCredentialsFlow", flowAlias),
			},
		},
	})
}

func TestAccKeycloakAuthenticationBindings_clientAuthenticationGrant(t *testing.T) {
	flow := "client_authentication_flow"
	flowAlias := "clientAuthenticationCopyFlow"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakAuthenticationBindings(testAccRealm.Realm, flow, flowAlias),
				Check:  testAccCheckKeycloakAuthenticationBindingBrowserSet(testAccRealm.Realm, "ClientAuthenticationFlow", flowAlias),
			},
		},
	})
}

func TestAccKeycloakAuthenticationBindings_dockerAuthenticationGrant(t *testing.T) {
	flow := "docker_authentication_flow"
	flowAlias := "dockerAuthenticationCopyFlow"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakAuthenticationBindings(testAccRealm.Realm, flow, flowAlias),
				Check:  testAccCheckKeycloakAuthenticationBindingBrowserSet(testAccRealm.Realm, "DockerAuthenticationFlow", flowAlias),
			},
		},
	})
}

func TestAccKeycloakAuthenticationBindings_firstBrokerLoginFlow(t *testing.T) {
	skipIfVersionIsLessThan(testCtx, t, keycloakClient, keycloak.Version_24)

	flow := "first_broker_login_flow"
	flowAlias := "firstBrokerLoginCopyFlow"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakAuthenticationBindings(testAccRealm.Realm, flow, flowAlias),
				Check:  testAccCheckKeycloakAuthenticationBindingBrowserSet(testAccRealm.Realm, "FirstBrokerLoginFlow", flowAlias),
			},
		},
	})
}

func TestAccKeycloakAuthenticationBindings_existingFlow(t *testing.T) {
	flow := "browser_flow"
	flowAlias := "browser"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakAuthenticationBindings_existingFlow(testAccRealm.Realm, flow, flowAlias),
				Check:  testAccCheckKeycloakAuthenticationBindingBrowserSet(testAccRealm.Realm, "BrowserFlow", flowAlias),
			},
		},
	})
}

func testKeycloakAuthenticationBindings(realm, flow, flowAlias string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_authentication_flow" "flow" {
	realm_id = data.keycloak_realm.realm.id
	alias    = "%s"
}

resource "keycloak_authentication_bindings" "authentication_binding" {
	realm_id	= data.keycloak_realm.realm.id
	%s	 		= keycloak_authentication_flow.flow.alias
}
	`, realm, flowAlias, flow)
}

func testKeycloakAuthenticationBindings_existingFlow(realm, flow, flowAlias string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_authentication_bindings" "authentication_binding" {
	realm_id	= data.keycloak_realm.realm.id
	%s	 		= "%s"
}
	`, realm, flow, flowAlias)
}

func testKeycloakAuthenticationBindingsWithRealm(realm, flow, flowAlias string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm        		= "%s"
	enabled     	 	= true
}

resource "keycloak_authentication_flow" "flow" {
	realm_id = keycloak_realm.realm.id
	alias    = "%s"
}

resource "keycloak_authentication_bindings" "authentication_binding" {
	realm_id	= keycloak_realm.realm.id
	%s	 		= keycloak_authentication_flow.flow.alias
}
	`, realm, flowAlias, flow)
}

func testAccCheckKeycloakAuthenticationBindingBrowserSet(realmName, binding, flowAlias string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		realm, err := keycloakClient.GetRealm(testCtx, realmName)
		if err != nil {
			return fmt.Errorf("error getting realm %s: %s", realmName, err)
		}

		v := reflect.ValueOf(*realm)
		bindingField := v.FieldByName(binding)
		b := bindingField.Interface().(*string)
		if *b != flowAlias {
			return fmt.Errorf("expected realm %s to have %s set to %s, but was %s", realm.Realm, binding, flowAlias, *b)
		}

		return nil
	}
}
