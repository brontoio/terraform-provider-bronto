package resource_slack_integration

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func SlackIntegrationResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description:         "Manages a Bronto Slack integration.",
		MarkdownDescription: "Manages a Bronto Slack integration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The unique identifier for the Slack integration",
				MarkdownDescription: "The unique identifier for the Slack integration",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "The name of the Slack integration",
				MarkdownDescription: "The name of the Slack integration",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"workspace_id": schema.StringAttribute{
				Required:            true,
				Description:         "The ID of the Slack workspace to use. The workspace must already be connected through the Bronto UI's OAuth flow first - it can't be created by this provider. Use the bronto_slack_workspaces data source to look up the ID of an already-connected workspace.",
				MarkdownDescription: "The ID of the Slack workspace to use. The workspace must already be connected through the Bronto UI's OAuth flow first - it can't be created by this provider. Use the `bronto_slack_workspaces` data source to look up the ID of an already-connected workspace.",
			},
			"channels": schema.SetNestedAttribute{
				Required:            true,
				Description:         "The Slack channels to send alerts to, and each channel's own settings.",
				MarkdownDescription: "The Slack channels to send alerts to, and each channel's own settings.",
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:            true,
							Description:         "The Slack channel name to send alerts to.",
							MarkdownDescription: "The Slack channel name to send alerts to.",
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
						"include_graph_image": schema.BoolAttribute{
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(true),
							Description:         "Whether alert notifications sent to this channel include the rendered graph image. Defaults to true.",
							MarkdownDescription: "Whether alert notifications sent to this channel include the rendered graph image. Defaults to `true`.",
						},
					},
				},
			},
		},
	}
}

type ChannelModel struct {
	Name              types.String `tfsdk:"name"`
	IncludeGraphImage types.Bool   `tfsdk:"include_graph_image"`
}

type SlackIntegrationModel struct {
	Id          types.String   `tfsdk:"id"`
	Name        types.String   `tfsdk:"name"`
	WorkspaceId types.String   `tfsdk:"workspace_id"`
	Channels    []ChannelModel `tfsdk:"channels"`
}
