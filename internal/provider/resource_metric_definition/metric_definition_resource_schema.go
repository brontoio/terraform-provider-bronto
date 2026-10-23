package resource_metric_definition

// Hand-written schema and model for the bronto_metric_definition resource.
// Not run through the tfplugingen pipeline: queries/formulas are nested
// lists with their own validation and a nested (0-or-1) aggregation object,
// which the generator doesn't produce, and the API has no update endpoint
// at all (only create/get/list/delete) - every attribute here requires
// replacement on change instead of being reconciled in place.

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// timeFunctions are the aggregation time-bucket functions accepted by
// MetricQuery.aggregation.time, per the Metrics API's TimeFunction schema.
var timeFunctions = []string{
	"sum", "avg", "average", "min", "max", "mean", "count",
	"median", "p25", "p50", "p75", "p90", "p95", "p99", "p999",
	"histogram", "rate",
}

// reduceToValues are the reduction functions accepted by both
// MetricQuery.aggregation.reduce_to and MetricFormula.aggregation.reduce_to,
// per the Metrics API's ReduceTo schema.
var reduceToValues = []string{"avg", "sum", "min", "max"}

func MetricDefinitionResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description:         "Manages a Bronto metric definition.",
		MarkdownDescription: "Manages a Bronto metric definition.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The unique identifier for the metric definition",
				MarkdownDescription: "The unique identifier for the metric definition",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "An optional name for the metric definition",
				MarkdownDescription: "An optional name for the metric definition",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "An optional description for this metric definition",
				MarkdownDescription: "An optional description for this metric definition",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 1000),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"queries": schema.ListNestedAttribute{
				Required:            true,
				Description:         "The list of simple queries to be computed. The result can be referenced in formulas. A metric definition cannot be updated in place: changing this list replaces the resource.",
				MarkdownDescription: "The list of simple queries to be computed. The result can be referenced in formulas. A metric definition cannot be updated in place: changing this list replaces the resource.",
				Validators: []validator.List{
					listvalidator.SizeBetween(1, 6),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:            true,
							Description:         "The unique name of the query within the metric definition. Cannot contain spaces if the query is referenced in a formula.",
							MarkdownDescription: "The unique name of the query within the metric definition. Cannot contain spaces if the query is referenced in a formula.",
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 255),
							},
						},
						"select": schema.ListAttribute{
							ElementType:         types.StringType,
							Required:            true,
							Description:         "The single metric field to retrieve from the dataset.",
							MarkdownDescription: "The single metric field to retrieve from the dataset.",
							Validators: []validator.List{
								listvalidator.SizeBetween(1, 1),
							},
						},
						"from": schema.ListAttribute{
							ElementType:         types.StringType,
							Optional:            true,
							Computed:            true,
							Description:         "Log IDs to query. Exactly one of from or from_expr must be specified.",
							MarkdownDescription: "Log IDs to query. Exactly one of `from` or `from_expr` must be specified.",
							PlanModifiers: []planmodifier.List{
								listplanmodifier.UseStateForUnknown(),
							},
						},
						"from_expr": schema.StringAttribute{
							Optional:            true,
							Computed:            true,
							Description:         "An expression identifying the logs or metric to query. Exactly one of from or from_expr must be specified.",
							MarkdownDescription: "An expression identifying the logs or metric to query. Exactly one of `from` or `from_expr` must be specified.",
							Validators: []validator.String{
								stringvalidator.LengthAtMost(255),
							},
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"where": schema.StringAttribute{
							Required:            true,
							Description:         "Filters the results of the query. Can combine multiple terms using AND, OR, NOT.",
							MarkdownDescription: "Filters the results of the query. Can combine multiple terms using AND, OR, NOT.",
						},
						"groups": schema.ListAttribute{
							ElementType:         types.StringType,
							Optional:            true,
							Computed:            true,
							Description:         "Keys used to arrange the aggregated results into groups.",
							MarkdownDescription: "Keys used to arrange the aggregated results into groups.",
							Validators: []validator.List{
								listvalidator.SizeBetween(0, 5),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.UseStateForUnknown(),
							},
						},
						"aggregation": schema.SingleNestedAttribute{
							Optional:            true,
							Description:         "The aggregation to apply to the selected metric field.",
							MarkdownDescription: "The aggregation to apply to the selected metric field.",
							Attributes: map[string]schema.Attribute{
								"time": schema.StringAttribute{
									Required:            true,
									Description:         "The time-bucket aggregation function.",
									MarkdownDescription: "The time-bucket aggregation function.",
									Validators: []validator.String{
										stringvalidator.OneOf(timeFunctions...),
									},
								},
								"reduce_to": schema.StringAttribute{
									Optional:            true,
									Computed:            true,
									Description:         "The final reduction applied to the aggregated result. Defaults to avg.",
									MarkdownDescription: "The final reduction applied to the aggregated result. Defaults to `avg`.",
									Validators: []validator.String{
										stringvalidator.OneOf(reduceToValues...),
									},
									PlanModifiers: []planmodifier.String{
										stringplanmodifier.UseStateForUnknown(),
									},
								},
								"time_function_params": schema.ListAttribute{
									ElementType:         types.StringType,
									Optional:            true,
									Computed:            true,
									Description:         "Parameters for the time function.",
									MarkdownDescription: "Parameters for the time function.",
									Validators: []validator.List{
										listvalidator.SizeBetween(0, 1),
									},
									PlanModifiers: []planmodifier.List{
										listplanmodifier.UseStateForUnknown(),
									},
								},
							},
						},
						"unit_config": schema.StringAttribute{
							CustomType:          jsontypes.NormalizedType{},
							Optional:            true,
							Computed:            true,
							Description:         "Optional configuration describing the unit of the query's raw value, as a JSON-encoded string (e.g. jsonencode({...})). Used for UI display only.",
							MarkdownDescription: "Optional configuration describing the unit of the query's raw value, as a JSON-encoded string (e.g. `jsonencode({...})`). Used for UI display only.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
			"formulas": schema.ListNestedAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "A list of formulas to combine the queries results. A metric definition cannot be updated in place: changing this list replaces the resource.",
				MarkdownDescription: "A list of formulas to combine the queries results. A metric definition cannot be updated in place: changing this list replaces the resource.",
				Validators: []validator.List{
					listvalidator.SizeAtMost(5),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:            true,
							Description:         "The unique name, within the metric definition, of the formula.",
							MarkdownDescription: "The unique name, within the metric definition, of the formula.",
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 100),
							},
						},
						"expression": schema.StringAttribute{
							Required:            true,
							Description:         "The expression that defines the formula. Formula elements refer, by name, to queries defined in this metric definition.",
							MarkdownDescription: "The expression that defines the formula. Formula elements refer, by name, to queries defined in this metric definition.",
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 100),
							},
						},
						"reduce_to": schema.StringAttribute{
							Optional:            true,
							Computed:            true,
							Description:         "The reduction to apply to the timeseries to reduce it to a single value. Defaults to avg.",
							MarkdownDescription: "The reduction to apply to the timeseries to reduce it to a single value. Defaults to `avg`.",
							Validators: []validator.String{
								stringvalidator.OneOf(reduceToValues...),
							},
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"unit_config": schema.StringAttribute{
							CustomType:          jsontypes.NormalizedType{},
							Optional:            true,
							Computed:            true,
							Description:         "Optional configuration describing the unit of the formula's value, as a JSON-encoded string (e.g. jsonencode({...})). Used for UI display only.",
							MarkdownDescription: "Optional configuration describing the unit of the formula's value, as a JSON-encoded string (e.g. `jsonencode({...})`). Used for UI display only.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
		},
	}
}

type AggregationModel struct {
	Time               types.String `tfsdk:"time"`
	ReduceTo           types.String `tfsdk:"reduce_to"`
	TimeFunctionParams types.List   `tfsdk:"time_function_params"`
}

type QueryModel struct {
	Name        types.String         `tfsdk:"name"`
	Select      types.List           `tfsdk:"select"`
	From        types.List           `tfsdk:"from"`
	FromExpr    types.String         `tfsdk:"from_expr"`
	Where       types.String         `tfsdk:"where"`
	Groups      types.List           `tfsdk:"groups"`
	Aggregation *AggregationModel    `tfsdk:"aggregation"`
	UnitConfig  jsontypes.Normalized `tfsdk:"unit_config"`
}

type FormulaModel struct {
	Name       types.String         `tfsdk:"name"`
	Expression types.String         `tfsdk:"expression"`
	ReduceTo   types.String         `tfsdk:"reduce_to"`
	UnitConfig jsontypes.Normalized `tfsdk:"unit_config"`
}

// formulaObjectType mirrors FormulaModel's attribute types. The top-level
// "formulas" attribute is Optional+Computed (it's normalized to an empty,
// not null, list when the API returns none), which means its value can be
// entirely unknown during planning - something only a wrapped attr.Value
// type can represent, unlike a plain []FormulaModel slice.
var formulaObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"name":        types.StringType,
		"expression":  types.StringType,
		"reduce_to":   types.StringType,
		"unit_config": jsontypes.NormalizedType{},
	},
}

type MetricDefinitionModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Queries     []QueryModel `tfsdk:"queries"`
	Formulas    types.List   `tfsdk:"formulas"`
}
