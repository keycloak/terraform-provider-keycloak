package main

import (
	"flag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/keycloak/terraform-provider-keycloak/provider"
)

func main() {

	var debugMode bool
	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := &plugin.ServeOpts{
		ProviderFunc: func() *schema.Provider {
			return provider.KeycloakProvider(nil)
		},
		Debug: debugMode,
		// using local provider address for debugging:
		ProviderAddr: "terraform.local/keycloak/keycloak",
	}
	plugin.Serve(opts)
}
