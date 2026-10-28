package resource_monitor

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// comparisonOperators are the Monitors API's ComparisonOperator values.
var comparisonOperators = []string{
	"BELOW", "BELOW_OR_EQUAL", "ABOVE", "ABOVE_OR_EQUAL", "EQUAL", "NOT_EQUAL", "OUTSIDE",
}

// monitorStatuses are the Monitors API's Status values, shared by the
// monitor's own status and no_data_status.
var monitorStatuses = []string{"OK", "NO_DATA", "ALERT", "WARN"}

// monitorTypes are the Monitors API's MonitorType values.
var monitorTypes = []string{"PATTERN", "USAGE", "CHANGE_DETECTION", "ANOMALY_DETECTION"}

// changeTypes are the Monitors API's ChangeType values, for
// CHANGE_DETECTION monitors only.
var changeTypes = []string{"DIFFERENCE", "PERCENTAGE"}

// actionTypes are the Monitors API's Action.type values.
var actionTypes = []string{"EMAIL", "INTEGRATION"}

func MonitorResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Manages a Bronto monitor. Only monitors that reference an existing metric definition " +
			"(via metric_id) are supported - not the API's alternative form with inline queries/formulas.",
		MarkdownDescription: "Manages a Bronto monitor. Only monitors that reference an existing metric definition " +
			"(via `metric_id`) are supported - not the API's alternative form with inline queries/formulas.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The unique identifier for the monitor",
				MarkdownDescription: "The unique identifier for the monitor",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "The name of the monitor",
				MarkdownDescription: "The name of the monitor",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Description:         "Extra contextual information to add to the monitor",
				MarkdownDescription: "Extra contextual information to add to the monitor",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 4000),
				},
			},
			"metric_id": schema.StringAttribute{
				Required: true,
				Description: "The id of the bronto_metric_definition this monitor tracks. The monitor references it " +
					"rather than owning its lifecycle - deleting the monitor does not delete the metric definition.",
				MarkdownDescription: "The id of the `bronto_metric_definition` this monitor tracks. The monitor references it " +
					"rather than owning its lifecycle - deleting the monitor does not delete the metric definition.",
			},
			"comparison_operator": schema.StringAttribute{
				Required: true,
				Description: "The comparison operator of the monitor. OUTSIDE is two-sided and only allowed for " +
					"ANOMALY_DETECTION monitors.",
				MarkdownDescription: "The comparison operator of the monitor. `OUTSIDE` is two-sided and only allowed for " +
					"`ANOMALY_DETECTION` monitors.",
				Validators: []validator.String{
					stringvalidator.OneOf(comparisonOperators...),
				},
			},
			"threshold": schema.Float64Attribute{
				Required: true,
				Description: "The threshold the monitor compares against. For PATTERN, USAGE and CHANGE_DETECTION " +
					"monitors this is a metric value. For ANOMALY_DETECTION monitors it is a count of baseline spread " +
					"units instead, and must be greater than 0.",
				MarkdownDescription: "The threshold the monitor compares against. For `PATTERN`, `USAGE` and `CHANGE_DETECTION` " +
					"monitors this is a metric value. For `ANOMALY_DETECTION` monitors it is a count of baseline spread " +
					"units instead, and must be greater than 0.",
			},
			"warning_threshold": schema.Float64Attribute{
				Optional: true,
				Description: "The warning threshold value, in the same units as threshold. For ANOMALY_DETECTION " +
					"monitors it must be greater than 0 and less than threshold.",
				MarkdownDescription: "The warning threshold value, in the same units as `threshold`. For `ANOMALY_DETECTION` " +
					"monitors it must be greater than 0 and less than `threshold`.",
			},
			"window": schema.StringAttribute{
				Required:            true,
				Description:         "The time window to evaluate the monitor, e.g. 'Last 20 minutes'. Minimum five minutes, maximum one day.",
				MarkdownDescription: "The time window to evaluate the monitor, e.g. `Last 20 minutes`. Minimum five minutes, maximum one day.",
			},
			"actions": schema.SetNestedAttribute{
				Required:            true,
				Description:         "The notifications to send when the monitor fires or resolves.",
				MarkdownDescription: "The notifications to send when the monitor fires or resolves.",
				Validators: []validator.Set{
					setvalidator.SizeBetween(1, 10),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Required:            true,
							Description:         "The kind of action: EMAIL or INTEGRATION.",
							MarkdownDescription: "The kind of action: `EMAIL` or `INTEGRATION`.",
							Validators: []validator.String{
								stringvalidator.OneOf(actionTypes...),
							},
						},
						"email": schema.StringAttribute{
							Optional:            true,
							Description:         "The email address to notify. Required when type is EMAIL.",
							MarkdownDescription: "The email address to notify. Required when `type` is `EMAIL`.",
						},
						"integration_id": schema.StringAttribute{
							Optional: true,
							Description: "The id of the bronto_slack_integration or bronto_webhook_integration to notify. " +
								"Required when type is INTEGRATION.",
							MarkdownDescription: "The id of the `bronto_slack_integration` or `bronto_webhook_integration` to notify. " +
								"Required when `type` is `INTEGRATION`.",
						},
					},
				},
			},
			"no_data_status": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The status to use for a group with no data (no events matched in the time window). Defaults to OK.",
				MarkdownDescription: "The status to use for a group with no data (no events matched in the time window). Defaults to `OK`.",
				Validators: []validator.String{
					stringvalidator.OneOf(monitorStatuses...),
				},
				// Without this, leaving no_data_status unconfigured makes it unknown on every plan.
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"notify_once": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				Description: "Notify only the groups changing the global monitor status, i.e. notifications are sent only " +
					"for the first group entering ALERT status or for the last group resolving. Defaults to false.",
				MarkdownDescription: "Notify only the groups changing the global monitor status, i.e. notifications are sent only " +
					"for the first group entering `ALERT` status or for the last group resolving. Defaults to `false`.",
			},
			"group_retention": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Description:         "Remove groups with no data after this many milliseconds. Defaults to 0 (never).",
				MarkdownDescription: "Remove groups with no data after this many milliseconds. Defaults to `0` (never).",
				// Same reasoning as no_data_status.
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"aux": schema.StringAttribute{
				CustomType:          jsontypes.NormalizedType{},
				Optional:            true,
				Computed:            true,
				Description:         "Additional monitor configuration, as a JSON-encoded string (e.g. jsonencode({...}))",
				MarkdownDescription: "Additional monitor configuration, as a JSON-encoded string (e.g. `jsonencode({...})`)",
				// encodeUpdateMonitorRequest already guards on
				// IsNull/IsUnknown before setting Aux, so an unconfigured
				// aux happens to be omitted (not reset) today even without
				// this modifier - but relying on that is fragile, and it
				// still causes needless "(known after apply)" plan noise
				// on every unrelated update. Carry forward state like the
				// other Optional+Computed, no-Default attributes here.
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"muted_until": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Description: "Epoch millis timestamp the monitor is muted until. -1 mutes indefinitely. Omit (or 0) for " +
					"unmuted. Changed after creation via a dedicated mute/unmute call, not the monitor's own update endpoint.",
				MarkdownDescription: "Epoch millis timestamp the monitor is muted until. `-1` mutes indefinitely. Omit (or `0`) for " +
					"unmuted. Changed after creation via a dedicated mute/unmute call, not the monitor's own update endpoint.",
				// Left unconfigured, this must carry forward whatever's in  state rather than going unknown on every update.
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"monitor_type": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("PATTERN"),
				Description: "What the monitor compares its metric against. PATTERN and USAGE compare it to a fixed " +
					"threshold, CHANGE_DETECTION to the same metric at an earlier time (see compare_to and change_type), " +
					"and ANOMALY_DETECTION to a trailing baseline (see evaluation_window and anomaly_algorithm). Defaults to PATTERN.",
				MarkdownDescription: "What the monitor compares its metric against. `PATTERN` and `USAGE` compare it to a fixed " +
					"threshold, `CHANGE_DETECTION` to the same metric at an earlier time (see `compare_to` and `change_type`), " +
					"and `ANOMALY_DETECTION` to a trailing baseline (see `evaluation_window` and `anomaly_algorithm`). Defaults to `PATTERN`.",
				Validators: []validator.String{
					stringvalidator.OneOf(monitorTypes...),
				},
			},
			"compare_to": schema.StringAttribute{
				Optional:            true,
				Description:         "For CHANGE_DETECTION monitors only. How far back the comparison query runs, e.g. '2 hours ago'.",
				MarkdownDescription: "For `CHANGE_DETECTION` monitors only. How far back the comparison query runs, e.g. `2 hours ago`.",
			},
			"change_type": schema.StringAttribute{
				Optional:            true,
				Description:         "For CHANGE_DETECTION monitors only. How the change value is calculated: DIFFERENCE or PERCENTAGE.",
				MarkdownDescription: "For `CHANGE_DETECTION` monitors only. How the change value is calculated: `DIFFERENCE` or `PERCENTAGE`.",
				Validators: []validator.String{
					stringvalidator.OneOf(changeTypes...),
				},
			},
			"evaluation_window": schema.StringAttribute{
				Optional: true,
				Description: "For ANOMALY_DETECTION monitors only, where it is required. The trailing period used to " +
					"build the baseline, e.g. '7 days'.",
				MarkdownDescription: "For `ANOMALY_DETECTION` monitors only, where it is required. The trailing period used to " +
					"build the baseline, e.g. `7 days`.",
			},
			"anomaly_algorithm": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Description: "For ANOMALY_DETECTION monitors only - the API rejects this field for any other monitor_type. " +
					"How the baseline is estimated; MEDIAN_MAD is the only supported value.",
				MarkdownDescription: "For `ANOMALY_DETECTION` monitors only - the API rejects this field for any other `monitor_type`. " +
					"How the baseline is estimated; `MEDIAN_MAD` is the only supported value.",
				Validators: []validator.String{
					stringvalidator.OneOf("MEDIAN_MAD"),
				},
				// Same reasoning as no_data_status/group_retention.
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ai_report_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				Description:         "Whether an automatic AI report is generated when the monitor triggers. Defaults to false.",
				MarkdownDescription: "Whether an automatic AI report is generated when the monitor triggers. Defaults to `false`.",
			},
			"ai_report_instructions": schema.StringAttribute{
				Optional: true,
				Description: "Instructions for the AI on what to include in the report, e.g. check for recent " +
					"deployments, verify the health of the following dependencies: ...",
				MarkdownDescription: "Instructions for the AI on what to include in the report, e.g. check for recent " +
					"deployments, verify the health of the following dependencies: ...",
			},
		},
	}
}

type ActionModel struct {
	Type          types.String `tfsdk:"type"`
	Email         types.String `tfsdk:"email"`
	IntegrationId types.String `tfsdk:"integration_id"`
}

type MonitorModel struct {
	Id                   types.String         `tfsdk:"id"`
	Name                 types.String         `tfsdk:"name"`
	Description          types.String         `tfsdk:"description"`
	MetricId             types.String         `tfsdk:"metric_id"`
	ComparisonOperator   types.String         `tfsdk:"comparison_operator"`
	Threshold            types.Float64        `tfsdk:"threshold"`
	WarningThreshold     types.Float64        `tfsdk:"warning_threshold"`
	Window               types.String         `tfsdk:"window"`
	Actions              []ActionModel        `tfsdk:"actions"`
	NoDataStatus         types.String         `tfsdk:"no_data_status"`
	NotifyOnce           types.Bool           `tfsdk:"notify_once"`
	GroupRetention       types.Int64          `tfsdk:"group_retention"`
	Aux                  jsontypes.Normalized `tfsdk:"aux"`
	MutedUntil           types.Int64          `tfsdk:"muted_until"`
	MonitorType          types.String         `tfsdk:"monitor_type"`
	CompareTo            types.String         `tfsdk:"compare_to"`
	ChangeType           types.String         `tfsdk:"change_type"`
	EvaluationWindow     types.String         `tfsdk:"evaluation_window"`
	AnomalyAlgorithm     types.String         `tfsdk:"anomaly_algorithm"`
	AiReportEnabled      types.Bool           `tfsdk:"ai_report_enabled"`
	AiReportInstructions types.String         `tfsdk:"ai_report_instructions"`
}
