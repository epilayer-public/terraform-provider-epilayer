package provider

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/epilayer-public/epilayer-go"
	"github.com/epilayer-public/terraform-provider-epilayer/internal/providerenhancer"
	"github.com/epilayer-public/terraform-provider-epilayer/internal/timedurationvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure EpiLayerProvider satisfies various provider interfaces.
var (
	_ provider.Provider = &EpiLayerProvider{}
)

// EpiLayerProvider defines the provider implementation.
type EpiLayerProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// EpiLayerProviderModel describes the provider data model.
type EpiLayerProviderModel struct {
	Endpoint        types.String `tfsdk:"endpoint"`
	Token           types.String `tfsdk:"token"`
	PollingInterval types.String `tfsdk:"polling_interval"`
}

func (p *EpiLayerProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "epilayer"
	resp.Version = p.version
}

func (p *EpiLayerProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The EpiLayer provider is used to interact with resources supported by [EpiLayer](https://www.epilayer.eu/). The provider needs to be configured with the proper credentials before it can be used.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf(
					"EpiLayer API endpoint. May also be provided via `EPILAYER_ENDPOINT` environment variable. If neither is provided, defaults to `%s`.",
					epilayer.DefaultEndpoint),
				Optional: true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "EpiLayer API token. May also be provided via `EPILAYER_TOKEN` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"polling_interval": providerenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The polling interval.",
				Optional:            true,
				Validators: []validator.String{
					timedurationvalidator.Positive(),
				},
			}),
		},
	}
}

func (p *EpiLayerProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data EpiLayerProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.Endpoint.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Unknown EpiLayer API endpoint",
			"The provider cannot create the EpiLayer API client as there is an unknown configuration value for the EpiLayer API endpoint. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the EPILAYER_ENDPOINT environment variable.",
		)
	}

	if data.Token.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Unknown EpiLayer API token",
			"The provider cannot create the EpiLayer API client as there is an unknown configuration value for the EpiLayer API token. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the EPILAYER_TOKEN environment variable.",
		)
	}

	if data.PollingInterval.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("polling_interval"),
			"Unknown Polling Interval",
			"The provider cannot create the EpiLayer API client as there is an unknown configuration value for the Polling Interval. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or remove it to use the default.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := os.Getenv("EPILAYER_ENDPOINT")
	token := os.Getenv("EPILAYER_TOKEN")
	pollingInterval := 2 * time.Second

	if !data.Endpoint.IsNull() {
		endpoint = data.Endpoint.ValueString()
	}
	if !data.Token.IsNull() {
		token = data.Token.ValueString()
	}
	if !data.PollingInterval.IsNull() {
		duration, err := time.ParseDuration(data.PollingInterval.ValueString())
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("polling_interval"),
				"Timeout Cannot Be Parsed",
				err.Error(),
			)
			return
		}

		pollingInterval = duration
	}

	if endpoint == "" {
		endpoint = epilayer.DefaultEndpoint
	}

	if token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Missing EpiLayer API token",
			"The provider cannot create the EpiLayer API client as there is a missing or empty value for the EpiLayer API token. "+
				"Set the token value in the configuration or use the EPILAYER_TOKEN environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	providerClient, err := NewClient(ctx, ClientConfig{
		ClientConfig: epilayer.ClientConfig{
			Endpoint: endpoint,
			Token:    token,
		},
		PollingInterval: pollingInterval,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create EpiLayer API Client",
			"An unexpected error occurred when creating the EpiLayer API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	resp.DataSourceData = providerClient
	resp.ResourceData = providerClient
}

func (p *EpiLayerProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewInstanceResource,
		NewInstanceStatusResource,
		NewSSHKeyResource,
		NewFloatingIPResource,
		NewVolumeResource,
		NewFilesystemResource,
		NewSecurityGroupResource,
		NewSnapshotResource,
		NewPrivateNetworkResource,
		NewKubernetesClusterResource,
	}
}

func (p *EpiLayerProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewImagesDataSource,
		NewKubernetesClusterDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &EpiLayerProvider{
			version: version,
		}
	}
}
