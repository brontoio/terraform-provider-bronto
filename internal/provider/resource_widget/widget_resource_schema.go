package resource_widget

// Hand-written schema and model for the bronto_widget resource. Not run
// through the tfplugingen pipeline: the aux field needs jsontypes.Normalized
// and metric_ids/widget_ids need Set semantics (the API doesn't preserve
// element order, as already found to be true of login_methods on the user
// resource), neither of which the generator produces automatically.

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func WidgetResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description:         "Manages a Bronto widget.",
		MarkdownDescription: "Manages a Bronto widget.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The unique identifier for the widget",
				MarkdownDescription: "The unique identifier for the widget",
			},
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "The widget name",
				MarkdownDescription: "The widget name",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Description:         "The widget description",
				MarkdownDescription: "The widget description",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 1024),
				},
			},
			"type": schema.StringAttribute{
				Required:            true,
				Description:         "The widget type, e.g. line, bar, area, score, pie, treemap, top-list",
				MarkdownDescription: "The widget type, e.g. `line`, `bar`, `area`, `score`, `pie`, `treemap`, `top-list`",
			},
			"aux": schema.StringAttribute{
				CustomType:          jsontypes.NormalizedType{},
				Optional:            true,
				Computed:            true,
				Description:         "Additional widget configuration, as a JSON-encoded string (e.g. jsonencode({...}))",
				MarkdownDescription: "Additional widget configuration, as a JSON-encoded string (e.g. `jsonencode({...})`)",
			},
			"metric_ids": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The metric IDs associated with the widget. A widget can include either metrics or widgets, but not both.",
				MarkdownDescription: "The metric IDs associated with the widget. A widget can include either metrics or widgets, but not both.",
			},
			"widget_ids": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The IDs of other widgets included in this (composite) widget. A widget can include either metrics or widgets, but not both.",
				MarkdownDescription: "The IDs of other widgets included in this (composite) widget. A widget can include either metrics or widgets, but not both.",
			},
			"tags": schema.MapAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				Description:         "A map of key value pairs associated with the widget",
				MarkdownDescription: "A map of key value pairs associated with the widget",
			},
		},
	}
}

type WidgetModel struct {
	Id          types.String         `tfsdk:"id"`
	Name        types.String         `tfsdk:"name"`
	Description types.String         `tfsdk:"description"`
	Type        types.String         `tfsdk:"type"`
	Aux         jsontypes.Normalized `tfsdk:"aux"`
	MetricIds   types.Set            `tfsdk:"metric_ids"`
	WidgetIds   types.Set            `tfsdk:"widget_ids"`
	Tags        types.Map            `tfsdk:"tags"`
}
