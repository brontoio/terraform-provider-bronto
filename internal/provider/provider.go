package provider

import (
	"context"
	"os"
	bronto "terraform-provider-bronto/internal"
	"terraform-provider-bronto/internal/provider/datasource_dataset"
	"terraform-provider-bronto/internal/provider/datasource_slack_workspaces"
	"terraform-provider-bronto/internal/provider/datasource_user"
	"terraform-provider-bronto/internal/provider/resource_dashboard"
	"terraform-provider-bronto/internal/provider/resource_metric_definition"
	"terraform-provider-bronto/internal/provider/resource_slack_integration"
	"terraform-provider-bronto/internal/provider/resource_user"
	"terraform-provider-bronto/internal/provider/resource_webhook_integration"
	"terraform-provider-bronto/internal/provider/resource_widget"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	Endpoint types.String `tfsdk:"endpoint"`
	ApiKey   types.String `tfsdk:"api_key"`
}

type brontoProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

func (p *brontoProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage Bronto (https://bronto.io) resources as code: users, dashboards, widgets, metric definitions, and Slack/webhook integrations. " +
			"Multiple orgs (e.g. a root org and a child org, each with their own API key) can be managed from one configuration by declaring " +
			"multiple aliased instances of this provider and pointing each resource at the right one with provider = bronto.<alias> - see the " +
			"example below.",
		MarkdownDescription: "Manage [Bronto](https://bronto.io) resources as code: users, dashboards, widgets, metric definitions, and Slack/webhook integrations.\n\n" +
			"Multiple orgs (e.g. a root org and a child org, each with their own API key) can be managed from one configuration by declaring " +
			"multiple aliased instances of this provider and pointing each resource at the right one with `provider = bronto.<alias>` - see the " +
			"example below.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Required:            true,
				Description:         "The base URL of the Bronto API to use, e.g. https://api.eu.bronto.io. Can also be set via the BRONTO_ENDPOINT environment variable.",
				MarkdownDescription: "The base URL of the Bronto API to use, e.g. `https://api.eu.bronto.io`. Can also be set via the `BRONTO_ENDPOINT` environment variable.",
			},
			"api_key": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				Description:         "The API key used to authenticate with the Bronto API. Can also be set via the BRONTO_API_KEY environment variable.",
				MarkdownDescription: "The API key used to authenticate with the Bronto API. Can also be set via the `BRONTO_API_KEY` environment variable.",
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

	if config.Endpoint.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Unknown Bronto Endpoint",
			"The provider cannot create the Bronto API client as there is an unknown configuration value for the Bronto endpoint. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the BRONTO_ENDPOINT environment variable.",
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

	endpoint := os.Getenv("BRONTO_ENDPOINT")
	apkKey := os.Getenv("BRONTO_API_KEY")

	if !config.Endpoint.IsNull() {
		endpoint = config.Endpoint.ValueString()
	}

	if !config.ApiKey.IsNull() {
		apkKey = config.ApiKey.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if endpoint == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Missing Bronto Endpoint",
			"The provider cannot create the Bronto API client as there is a missing or empty value for the Bronto endpoint. "+
				"Set the endpoint value in the configuration or use the BRONTO_ENDPOINT environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if apkKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
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
	client, err := bronto.NewClient(&endpoint, &apkKey)
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
	return []func() datasource.DataSource{
		datasource_user.NewUserDataSource,
		datasource_slack_workspaces.NewSlackWorkspacesDataSource,
		datasource_dataset.NewDatasetDataSource,
	}
}

func (p *brontoProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resource_user.NewUserResource,
		resource_widget.NewWidgetResource,
		resource_dashboard.NewDashboardResource,
		resource_metric_definition.NewMetricDefinitionResource,
		resource_slack_integration.NewSlackIntegrationResource,
		resource_webhook_integration.NewWebhookIntegrationResource,
	}
}
