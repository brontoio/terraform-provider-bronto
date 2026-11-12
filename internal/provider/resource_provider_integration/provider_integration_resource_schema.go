package resource_provider_integration

// The secret is write-only: the API never returns it, only whether one is
// currently stored (has_secret). It's Optional (not Computed) and Read
// deliberately never touches it, so whatever value is already in state or
// plan is left alone - the standard pattern for a credential the API
// won't echo back.
//
// The API's own field for which downstream provider to send alerts to is
// named "provider", which can't be used as a schema attribute name here:
// it collides with Terraform's reserved provider = ... resource
// meta-argument. It's exposed as provider_type instead, matching the name
// already used for bronto_sso_idp's equivalent field.

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// integrationProviders are the downstream alerting providers accepted by
// ProviderIntegrationResource.provider.
var integrationProviders = []string{"aws-devops-agent", "pagerduty", "jsm", "rootly"}

// authTypes are the request-authentication schemes accepted by
// ProviderIntegrationResource.auth_type.
var authTypes = []string{"hmac", "api_key"}

func ProviderIntegrationResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description:         "Manages a Bronto provider integration (e.g. PagerDuty, JSM, Rootly, or the AWS DevOps Agent) that alerts are sent to.",
		MarkdownDescription: "Manages a Bronto provider integration (e.g. PagerDuty, JSM, Rootly, or the AWS DevOps Agent) that alerts are sent to.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The unique identifier for the provider integration",
				MarkdownDescription: "The unique identifier for the provider integration",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The integration name. Defaults to the provider's name.",
				MarkdownDescription: "The integration name. Defaults to the provider's name.",
				Validators: []validator.String{
					stringvalidator.LengthAtMost(255),
				},
			},
			"provider_type": schema.StringAttribute{
				Required:            true,
				Description:         "The downstream provider alerts are sent to: aws-devops-agent, pagerduty, jsm, or rootly. Cannot be changed after creation.",
				MarkdownDescription: "The downstream provider alerts are sent to: `aws-devops-agent`, `pagerduty`, `jsm`, or `rootly`. Cannot be changed after creation.",
				Validators: []validator.String{
					stringvalidator.OneOf(integrationProviders...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"url": schema.StringAttribute{
				Required:            true,
				Description:         "The provider URL to send alerts to. Must use https.",
				MarkdownDescription: "The provider URL to send alerts to. Must use https.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(10, 2048),
				},
			},
			"secret": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				Description:         "The secret used to authenticate with the provider (an HMAC signing secret or API key). Write-only: the API never returns it, so this is never refreshed from the API, only ever sent on create or when explicitly changed.",
				MarkdownDescription: "The secret used to authenticate with the provider (an HMAC signing secret or API key). Write-only: the API never returns it, so this is never refreshed from the API, only ever sent on create or when explicitly changed.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 4096),
				},
			},
			"has_secret": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether a secret is currently stored for this integration.",
				MarkdownDescription: "Whether a secret is currently stored for this integration.",
			},
			"auth_type": schema.StringAttribute{
				Required: true,
				Description: "How requests to the provider are authenticated. hmac signs each request with the secret, which is never sent; api_key sends the " +
					"secret with each request in the form the provider expects. Not every provider supports both.",
				MarkdownDescription: "How requests to the provider are authenticated. `hmac` signs each request with the secret, which is never sent; `api_key` sends the " +
					"secret with each request in the form the provider expects. Not every provider supports both.",
				Validators: []validator.String{
					stringvalidator.OneOf(authTypes...),
				},
			},
			"tags": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "A map of key value pairs associated with the provider integration, not including any inherited from the provider's default_tags - see tags_all for the full set. " +
					"Applied via individual tag API calls, not the provider integration's own create/update endpoints, so each changed key costs a separate request.",
				MarkdownDescription: "A map of key value pairs associated with the provider integration, not including any inherited from the provider's `default_tags` - see `tags_all` for the full set. " +
					"Applied via individual tag API calls, not the provider integration's own create/update endpoints, so each changed key costs a separate request.",
			},
			"tags_all": schema.MapAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				Description:         "The full set of tags on the provider integration, including both tags and any merged in from the provider's default_tags.",
				MarkdownDescription: "The full set of tags on the provider integration, including both `tags` and any merged in from the provider's `default_tags`.",
			},
		},
	}
}

type ProviderIntegrationModel struct {
	Id           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	ProviderType types.String `tfsdk:"provider_type"`
	Url          types.String `tfsdk:"url"`
	Secret       types.String `tfsdk:"secret"`
	HasSecret    types.Bool   `tfsdk:"has_secret"`
	AuthType     types.String `tfsdk:"auth_type"`
	Tags         types.Map    `tfsdk:"tags"`
	TagsAll      types.Map    `tfsdk:"tags_all"`
}
