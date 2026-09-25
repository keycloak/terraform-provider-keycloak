package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tf5server"
	"github.com/keycloak/terraform-provider-keycloak/provider"
)

func main() {

	var debugMode bool
	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	// the SDKv2 based provider is muxed with a framework based provider, which hosts framework-only features like actions
	muxServer, err := provider.MuxProviderServer(context.Background(), provider.KeycloakProvider(nil))
	if err != nil {
		log.Fatal(err)
	}

	var serveOpts []tf5server.ServeOpt
	if debugMode {
		serveOpts = append(serveOpts, tf5server.WithManagedDebug())
	}

	// The provider address is only used for logging and for the reattach configuration in debug mode, the same as
	// the ProviderAddr of the plugin.ServeOpts of the SDKv2. It does not affect how Terraform resolves the provider.
	err = tf5server.Serve("terraform.local/keycloak/keycloak", func() tfprotov5.ProviderServer {
		return muxServer
	}, serveOpts...)
	if err != nil {
		log.Fatal(err)
	}
}
