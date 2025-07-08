package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccKeycloakDataSourceClientDescriptionConverter_basic(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.keycloak_client_description_converter.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccKeycloakClientDescriptionConverterConfig(clientId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "protocol", "saml"),
					resource.TestCheckResourceAttr(dataSourceName, "client_id", "FakeEntityId"),
					resource.TestCheckResourceAttr(dataSourceName, "realm_id", testAccRealm.Realm),
				),
			},
		},
	})
}

func testAccKeycloakClientDescriptionConverterConfig(clientId string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

data "keycloak_client_description_converter" "test" {
	realm_id = data.keycloak_realm.realm.id
	body     = <<EOF
	<md:EntityDescriptor xmlns:md="urn:oasis:names:tc:SAML:2.0:metadata" validUntil="2021-04-17T12:41:46Z" cacheDuration="PT604800S" entityID="FakeEntityId">
    <md:SPSSODescriptor AuthnRequestsSigned="false" WantAssertionsSigned="false" protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
        <md:KeyDescriptor use="signing">
			<ds:KeyInfo xmlns:ds="http://www.w3.org/2000/09/xmldsig#">
				<ds:X509Data>
					<ds:X509Certificate>MIICyDCCAjGgAwIBAgIBADANBgkqhkiG9w0BAQ0FADCBgDELMAkGA1UEBhMCdXMx
					CzAJBgNVBAgMAklBMSQwIgYDVQQKDBt0ZXJyYWZvcm0tcHJvdmlkZXIta2V5Y2xv
					YWsxHDAaBgNVBAMME21ycGFya2Vycy5naXRodWIuaW8xIDAeBgkqhkiG9w0BCQEW
					EW1pY2hhZWxAcGFya2VyLmdnMB4XDTE5MDEwODE0NDYzNloXDTI5MDEwNTE0NDYz
					NlowgYAxCzAJBgNVBAYTAnVzMQswCQYDVQQIDAJJQTEkMCIGA1UECgwbdGVycmFm
					b3JtLXByb3ZpZGVyLWtleWNsb2FrMRwwGgYDVQQDDBNtcnBhcmtlcnMuZ2l0aHVi
					LmlvMSAwHgYJKoZIhvcNAQkBFhFtaWNoYWVsQHBhcmtlci5nZzCBnzANBgkqhkiG
					9w0BAQEFAAOBjQAwgYkCgYEAxuZny7uyYxGVPtpie14gNQC4tT9sAvO2sVNDhuoe
					qIKLRpNwkHnwQmwe5OxSh9K0BPHp/DNuuVWUqvo4tniEYn3jBr7FwLYLTKojQIxj
					53S1UTT9EXq3eP5HsHMD0QnTuca2nlNYUDBm6ud2fQj0Jt5qLx86EbEC28N56IRv
					GX8CAwEAAaNQME4wHQYDVR0OBBYEFMLnbQh77j7vhGTpAhKpDhCrBsPZMB8GA1Ud
					IwQYMBaAFMLnbQh77j7vhGTpAhKpDhCrBsPZMAwGA1UdEwQFMAMBAf8wDQYJKoZI
					hvcNAQENBQADgYEAB8wGrAQY0pAfwbnYSyBt4STbebeRTu1/q1ucfrtc3qsegcd5
					n01xTR+T2uZJwqHFPpFjr4IPORiHx3+4BWCweslPD53qBjKUPXcbMO1Revjef6Tj
					K3K0AuJ94fxgXVoT61Nzu/a6Lj6RhzU/Dao9mlSbJY+YSbm+ZBpsuRUQ84s=</ds:X509Certificate>
				</ds:X509Data>
			</ds:KeyInfo>
		</md:KeyDescriptor>
		<md:NameIDFormat>urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified</md:NameIDFormat>
        <md:AssertionConsumerService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" Location="https://localhost/acs/saml/" index="1"/>
    </md:SPSSODescriptor>
</md:EntityDescriptor>
	EOF
}
`, testAccRealm.Realm)
}
