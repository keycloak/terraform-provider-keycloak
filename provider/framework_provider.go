package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	fwprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	fwschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

// keycloakFrameworkProvider hosts everything that can only be implemented with the terraform-plugin-framework,
// e.g. actions. It is served next to the SDKv2 based provider via MuxProviderServer.
//
// It does not configure a Keycloak client on its own. Instead it reuses the client of the SDKv2 provider, which
// is always configured first by the mux server.
type keycloakFrameworkProvider struct {
	sdkProvider *schema.Provider
}

var _ fwprovider.ProviderWithActions = &keycloakFrameworkProvider{}

func KeycloakFrameworkProvider(sdkProvider *schema.Provider) fwprovider.Provider {
	return &keycloakFrameworkProvider{sdkProvider: sdkProvider}
}

// MuxProviderServer combines the SDKv2 based provider with the framework based provider into a single provider server.
func MuxProviderServer(ctx context.Context, sdkProvider *schema.Provider) (tfprotov5.ProviderServer, error) {
	muxServer, err := tf5muxserver.NewMuxServer(ctx,
		// the order is important, the framework provider relies on the SDKv2 provider being configured first
		sdkProvider.GRPCProvider,
		providerserver.NewProtocol5(KeycloakFrameworkProvider(sdkProvider)),
	)
	if err != nil {
		return nil, err
	}

	return muxServer.ProviderServer(), nil
}

func (p *keycloakFrameworkProvider) Metadata(_ context.Context, _ fwprovider.MetadataRequest, resp *fwprovider.MetadataResponse) {
	resp.TypeName = "keycloak"
}

// Schema derives the provider schema from the SDKv2 provider, since all muxed providers must expose an identical
// provider schema. This way the provider configuration only needs to be maintained in one place.
func (p *keycloakFrameworkProvider) Schema(ctx context.Context, _ fwprovider.SchemaRequest, resp *fwprovider.SchemaResponse) {
	sdkSchema, err := p.sdkProvider.GRPCProvider().GetProviderSchema(ctx, &tfprotov5.GetProviderSchemaRequest{})
	if err != nil {
		resp.Diagnostics.AddError("unable to read SDKv2 provider schema", err.Error())
		return
	}

	attributes := make(map[string]fwschema.Attribute)
	for _, attribute := range sdkSchema.Provider.Block.Attributes {
		var deprecationMessage string
		if attribute.Deprecated {
			deprecationMessage = "Deprecated"
		}

		switch {
		case attribute.Type.Is(tftypes.String):
			attributes[attribute.Name] = fwschema.StringAttribute{
				Required:           attribute.Required,
				Optional:           attribute.Optional,
				Sensitive:          attribute.Sensitive,
				Description:        attribute.Description,
				DeprecationMessage: deprecationMessage,
			}
		case attribute.Type.Is(tftypes.Bool):
			attributes[attribute.Name] = fwschema.BoolAttribute{
				Required:           attribute.Required,
				Optional:           attribute.Optional,
				Sensitive:          attribute.Sensitive,
				Description:        attribute.Description,
				DeprecationMessage: deprecationMessage,
			}
		case attribute.Type.Is(tftypes.Number):
			// covers TypeInt as well as TypeFloat of the SDKv2
			attributes[attribute.Name] = fwschema.NumberAttribute{
				Required:           attribute.Required,
				Optional:           attribute.Optional,
				Sensitive:          attribute.Sensitive,
				Description:        attribute.Description,
				DeprecationMessage: deprecationMessage,
			}
		case attribute.Type.Is(tftypes.Map{ElementType: tftypes.String}):
			attributes[attribute.Name] = fwschema.MapAttribute{
				ElementType:        types.StringType,
				Required:           attribute.Required,
				Optional:           attribute.Optional,
				Sensitive:          attribute.Sensitive,
				Description:        attribute.Description,
				DeprecationMessage: deprecationMessage,
			}
		default:
			// fail early with a clear message if the SDKv2 provider schema starts to use other types
			resp.Diagnostics.AddError("unsupported provider attribute type", fmt.Sprintf("attribute %q has type %s, which cannot be mapped to the framework provider schema", attribute.Name, attribute.Type))
		}
	}

	resp.Schema = fwschema.Schema{Attributes: attributes}
}

func (p *keycloakFrameworkProvider) Configure(_ context.Context, _ fwprovider.ConfigureRequest, resp *fwprovider.ConfigureResponse) {
	keycloakClient, ok := p.sdkProvider.Meta().(*keycloak.KeycloakClient)
	if !ok || keycloakClient == nil {
		resp.Diagnostics.AddError("keycloak client not available", "The SDKv2 based provider must be configured before the framework based provider.")
		return
	}

	resp.ActionData = keycloakClient
	resp.ResourceData = keycloakClient
	resp.DataSourceData = keycloakClient
}

func (p *keycloakFrameworkProvider) Resources(_ context.Context) []func() resource.Resource {
	return nil
}

func (p *keycloakFrameworkProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}

func (p *keycloakFrameworkProvider) Actions(_ context.Context) []func() action.Action {
	return []func() action.Action{
		NewUserFederationSyncAction,
	}
}
