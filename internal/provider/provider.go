package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"os"
	bronto "terraform-provider-bronto/internal"
	"terraform-provider-bronto/internal/provider/resource_user"
)

var (
	_ provider.Provider = &brontoProvider{}
)

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &brontoProvider{
			version: version,
		}
	}
}

type brontoProviderModel struct {
	Region types.String `tfsdk:"region"`
	ApiKey types.String `tfsdk:"api_key"`
}

type brontoProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

func (p *brontoProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"region": schema.StringAttribute{
				Required: true,
			},
			"api_key": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
		},
	}

}

func (p *brontoProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config brontoProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.Region.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("region"),
			"Unknown Bronto Region",
			"The provider cannot create the Bronto API client as there is an unknown configuration value for the Bronto Region. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the BRONTO_REGION environment variable.",
		)
	}

	if config.ApiKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown Bronto API Key",
			"The provider cannot create the Bronto API client as there is an unknown configuration value for the Bronto API Key. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the BRONTO_API_KEY environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	region := os.Getenv("BRONTO_REGION")
	apkKey := os.Getenv("BRONTO_API_KEY")

	if !config.Region.IsNull() {
		region = config.Region.ValueString()
	}

	if !config.ApiKey.IsNull() {
		apkKey = config.ApiKey.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if region == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("region"),
			"Missing Bronto Region",
			"The provider cannot create the Bronto API client as there is a missing or empty value for the Bronto Region. "+
				"Set the host value in the configuration or use the BRONTO_REGION environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if apkKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("apk_key"),
			"Missing Bronto API Key",
			"The provider cannot create the Bronto API client as there is a missing or empty value for the Bronto API Key. "+
				"Set the host value in the configuration or use the BRONTO_API_KEY environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create a new Bronto client using the configuration values
	client, err := bronto.NewClient(&region, &apkKey)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Bronto API client",
			"An unexpected error occurred when creating the Bronto API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Bronto Client Error: "+err.Error(),
		)
		return
	}

	// Make the Bronto client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *brontoProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "bronto"
	resp.Version = p.version
}

func (p *brontoProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *brontoProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resource_user.NewUserResource,
	}
}
