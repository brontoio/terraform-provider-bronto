package resource_widget

// Hand-written schema and model for the bronto_widget resource. Not run
// through the tfplugingen pipeline: the aux field needs jsontypes.Normalized,
// widget_ids needs Set semantics (the API doesn't preserve element order, as
// already found to be true of login_methods on the user resource), and
// sources is an ordered list of nested {type, id} objects - none of which
// the generator produces automatically.

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// widgetSourceTypes are the entity kinds a widget source can reference, per
// the Widgets API's WidgetSource schema.
var widgetSourceTypes = []string{"metric", "monitor", "slo"}

func WidgetResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description:         "Manages a Bronto widget.",
		MarkdownDescription: "Manages a Bronto widget.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The unique identifier for the widget",
				MarkdownDescription: "The unique identifier for the widget",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
			"sources": schema.ListNestedAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The metrics, monitors, and SLOs that supply data to the widget, in display order. A widget can include either sources or widget_ids, but not both.",
				MarkdownDescription: "The metrics, monitors, and SLOs that supply data to the widget, in display order. A widget can include either `sources` or `widget_ids`, but not both.",
				Validators: []validator.List{
					listvalidator.SizeBetween(0, 20),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Required:            true,
							Description:         "The kind of entity referenced: metric, monitor, or slo.",
							MarkdownDescription: "The kind of entity referenced: `metric`, `monitor`, or `slo`.",
							Validators: []validator.String{
								stringvalidator.OneOf(widgetSourceTypes...),
							},
						},
						"id": schema.StringAttribute{
							Required:            true,
							Description:         "The unique identifier of the referenced metric definition, monitor, or SLO.",
							MarkdownDescription: "The unique identifier of the referenced metric definition, monitor, or SLO.",
						},
					},
				},
			},
			"widget_ids": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The IDs of other widgets included in this (composite) widget. A widget can include either sources or widget_ids, but not both.",
				MarkdownDescription: "The IDs of other widgets included in this (composite) widget. A widget can include either `sources` or `widget_ids`, but not both.",
			},
			"tags": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "A map of key value pairs associated with the widget, not including any inherited from the provider's default_tags - see tags_all for the full set. " +
					"Applied via individual tag API calls, not the widget's own create/update endpoints, so each changed key costs a separate request.",
				MarkdownDescription: "A map of key value pairs associated with the widget, not including any inherited from the provider's `default_tags` - see `tags_all` for the full set. " +
					"Applied via individual tag API calls, not the widget's own create/update endpoints, so each changed key costs a separate request.",
			},
			"tags_all": schema.MapAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				Description:         "The full set of tags on the widget, including both tags and any merged in from the provider's default_tags.",
				MarkdownDescription: "The full set of tags on the widget, including both `tags` and any merged in from the provider's `default_tags`.",
			},
		},
	}
}

type SourceModel struct {
	Type types.String `tfsdk:"type"`
	Id   types.String `tfsdk:"id"`
}

// sourceObjectType mirrors SourceModel's attribute types. "sources" is
// Optional+Computed (normalized to an empty, not null, list when the API
// returns none), which means its value can be entirely unknown during
// planning - something only a wrapped attr.Value type can represent, unlike
// a plain []SourceModel slice (see the same reasoning on bronto_metric_
// definition's "formulas" attribute).
var sourceObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"type": types.StringType,
		"id":   types.StringType,
	},
}

type WidgetModel struct {
	Id          types.String         `tfsdk:"id"`
	Name        types.String         `tfsdk:"name"`
	Description types.String         `tfsdk:"description"`
	Type        types.String         `tfsdk:"type"`
	Aux         jsontypes.Normalized `tfsdk:"aux"`
	Sources     types.List           `tfsdk:"sources"`
	WidgetIds   types.Set            `tfsdk:"widget_ids"`
	Tags        types.Map            `tfsdk:"tags"`
	TagsAll     types.Map            `tfsdk:"tags_all"`
}
