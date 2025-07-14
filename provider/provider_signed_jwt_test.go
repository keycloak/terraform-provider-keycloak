package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccKeycloakProvider_signedJWT(t *testing.T) {
	t.Parallel()
	jwtSigningKey := "-----BEGIN PRIVATE KEY-----\r\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDEgtQWZtM/nSIX\r\npuJ9aJ9elurICPD1FEKOaIQBg6MihHYDDQxkIT93FkGLTRPMNFNgSXUKKj7IW9Ih\r\nxpv7v4ltsRaNmT0n2CFmVDGQI9404M1vz7V6Gj70feHWtcwzF42kvsCMEETwsW0j\r\nkOSX2RXun7RPaLSfXavki0w1ql3/nKVxQMuFgmMZrQzpIGh/EPpRjEWgL9HRYlp4\r\nX5wbW/GGzDuUldJoJBWhBWb8uOVSJcmXcgZ45k5LxkGTjTXPlgYJorVSdS8bkoBx\r\nBRa3COwilTwUFiBNna4HwnMmLFiKEYMhCGO+HvyFm4AKzhhShHQSm8VrISWsD05F\r\ng0Uo/LEHAgMBAAECggEANe9gjat4NKAIoOw9gsUp5LjQRMnrdKC0aci23oGGT22C\r\nxHCa44qalDFoGPc1RVlhPu66cGlK5QwKnxmXa1/VNOWjdobGGb8A38ig99pYXTQM\r\nPrGIMjSs7cb1Ksyn+KfwyPRP/cFjYpqYBWh5zVGYau+rehYXaRw5FxfCeYJCnWqi\r\ndPKZPb8num2NOr4ts3zFGbf7Ni47k4Ma07alEuAi9Whi+ONagFI9P0sgyYl63Zen\r\nJTshWkKNuljkvKfNS7a8PGkSar2iYDJE3GkfiGP2sOJzpvPDQbWDhZMFhlHzS9hK\r\n8BELVcR/qIaSlXY4AekeQTONaXOwwX9Fj/7fKMVUZQKBgQDmiGuaAfDGfAY3uO73\r\n6keESta35yoFz2IarzTW2GhJagNfJX7wUOLvjPEmgwI+QU4PA9hisEePSyjHaz4V\r\nRLn47pmV9aQDwKfKjZdLL8G63gMbhWup5keTKpXU8LpWWiTK0C9ouV/M2HO2IymC\r\n0dv1RHaVpBhOSMbxX26TQ8O+3QKBgQDaOEEXaEe7uAqIlj0W7HbZevaGLz4PVE17\r\nYqWBD50T+hiXTgCRnOee5sQPVjUrmBjk7QQ8sBentoVlL418XFH1MY+IcNy9B7/d\r\nKpxpwPCEjJchFdKAZvDs9QFdf2VOAvxM1HAZM01qdGMrq/wXLTMHPnB3Os3cTWaB\r\ncWgvcvInMwKBgQCp0pkhjIhoTvjtl4hCjQ0+ATuHofys5waoDaVpF2ZLnpL5Rk/q\r\njEuAmF0VN7ExVz4/hV+j46PzhTR3IyNK26P8Ixh1Bc1bDlMMvZ1UP8wA8odrgK+9\r\nKuxTFy3k/ajm7+TmmtIx3U0bQ+CJrgFoY1wbo+GPfqCBGs+jA+AbD/Jk6QKBgG3F\r\nVID41PTJ/Ip+wNYyNwrpfu86/oXpi1xg4A5PE14ENbCO7VxSSHU3cjKg0/hM92DZ\r\nFYONtSiJeQrQY+TF7/heaOxikbeJGWugzrOn+ZVDv5ZGCvDKV7FrAbfNqOEYQWBI\r\nkOcsVmoRh/1k81eZRg0DzME9VGbYjJLawGT19nffAoGADDoFR5YT3Hj1gi4xrCv7\r\nUsdorG0VqXJNpkFEv53MHUO5zZocd9Iv7zcQ+weTDlJE38nTEixBcfXHzyPcY8IW\r\nrRMDTOcfL/vPSADig7pRyLankKEtaS2QE2BKwDP7hkoQMkllVDe4OPflghmCclYd\r\nDamFJ27n4ONrNiy142vb+MY=\r\n-----END PRIVATE KEY-----"
	testAccProvider = KeycloakProvider(keycloakClient)

	os.Setenv("KEYCLOAK_CLIENT_ID", "terraform-jwt")
	os.Setenv("KEYCLOAK_JWT_SIGNING_KEY", jwtSigningKey)

	defer func() {
		os.Setenv("KEYCLOAK_CLIENT_ID", "terraform")
		os.Unsetenv("KEYCLOAK_JWT_SIGNING_KEY")
	}()

	clientId := acctest.RandomWithPrefix("tf-acc")
	resource.Test(t, resource.TestCase{
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"keycloak": func() (*schema.Provider, error) {
				return testAccProvider, nil
			},
		},
		PreCheck: func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClient_basic(clientId),
			},
		},
	})
}
