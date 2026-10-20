package resource_dashboard

// Hand-written schema and model for the bronto_dashboard resource. Not run
// through the tfplugingen pipeline: widget_ids and layout_metric_ids need
// Set semantics (order isn't preserved by the API, same reasoning as
// login_methods on the user resource), and reconciling widget membership
// via separate append/remove endpoints doesn't fit the generator's model.

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func DashboardResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description:         "Manages a Bronto dashboard.",
		MarkdownDescription: "Manages a Bronto dashboard.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The unique identifier for the dashboard",
				MarkdownDescription: "The unique identifier for the dashboard",
			},
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "The name of the dashboard",
				MarkdownDescription: "The name of the dashboard",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"description": schema.StringAttribute{
				Computed:            true,
				Description:         "The description of the dashboard. The API does not currently support setting this on create or update.",
				MarkdownDescription: "The description of the dashboard. The API does not currently support setting this on create or update.",
			},
			"widget_ids": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The IDs of the widgets attached to this dashboard",
				MarkdownDescription: "The IDs of the widgets attached to this dashboard",
			},
			"layout_metric_ids": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The IDs of the metrics (LBMs) associated with the dashboard's layout",
				MarkdownDescription: "The IDs of the metrics (LBMs) associated with the dashboard's layout",
			},
			"tags": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "A map of key value pairs associated with the dashboard. Applied via individual tag API calls, not the dashboard's own create/update endpoints, so each changed key costs a separate request.",
				MarkdownDescription: "A map of key value pairs associated with the dashboard. Applied via individual tag API calls, not the dashboard's own create/update endpoints, so each changed key costs a separate request.",
			},
		},
	}
}

type DashboardModel struct {
	Id              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	WidgetIds       types.Set    `tfsdk:"widget_ids"`
	LayoutMetricIds types.Set    `tfsdk:"layout_metric_ids"`
	Tags            types.Map    `tfsdk:"tags"`
}
