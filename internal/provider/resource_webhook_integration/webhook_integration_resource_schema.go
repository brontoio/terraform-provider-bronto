package resource_webhook_integration

// Hand-written schema and model for the bronto_webhook_integration
// resource. Not run through the tfplugingen pipeline.
//
// payload is kept a plain string (not jsontypes.Normalized, unlike aux on
// the widget resource) even though it's typically JSON-shaped: the API
// types it as a bare string with minLength 0, and an empty string isn't
// valid JSON, which would trip up Normalized's semantic-equality handling.

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func WebhookIntegrationResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description:         "Manages a Bronto webhook integration.",
		MarkdownDescription: "Manages a Bronto webhook integration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The unique identifier for the webhook integration",
				MarkdownDescription: "The unique identifier for the webhook integration",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "The webhook name",
				MarkdownDescription: "The webhook name",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"url": schema.StringAttribute{
				Required:            true,
				Description:         "The webhook URL",
				MarkdownDescription: "The webhook URL",
				Validators: []validator.String{
					stringvalidator.LengthBetween(10, 2048),
				},
			},
			"custom_headers": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
				Description:         "Custom headers to include in the webhook request, e.g. an API key or auth token. Marked sensitive since header values commonly carry credentials, even though the API doesn't flag this field as such.",
				MarkdownDescription: "Custom headers to include in the webhook request, e.g. an API key or auth token. Marked sensitive since header values commonly carry credentials, even though the API doesn't flag this field as such.",
				Validators: []validator.Map{
					mapvalidator.SizeAtMost(10),
				},
			},
			"payload": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The desired payload structure, typically JSON, as a plain string template. May reference alert fields with $variable placeholders, e.g. $title, $description, $alert_id.",
				MarkdownDescription: "The desired payload structure, typically JSON, as a plain string template. May reference alert fields with `$variable` placeholders, e.g. `$title`, `$description`, `$alert_id`.",
				Validators: []validator.String{
					stringvalidator.LengthAtMost(2048),
				},
			},
		},
	}
}

type WebhookIntegrationModel struct {
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Url           types.String `tfsdk:"url"`
	CustomHeaders types.Map    `tfsdk:"custom_headers"`
	Payload       types.String `tfsdk:"payload"`
}
