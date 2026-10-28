package resource_monitor

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	bronto "terraform-provider-bronto/internal"
)

var (
	_ resource.Resource                = &monitorResource{}
	_ resource.ResourceWithConfigure   = &monitorResource{}
	_ resource.ResourceWithImportState = &monitorResource{}
)

func NewMonitorResource() resource.Resource {
	return &monitorResource{}
}

type monitorResource struct {
	client *bronto.Client
}

func (r *monitorResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitor"
}

func (r *monitorResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = MonitorResourceSchema(ctx)
}

func (r *monitorResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*bronto.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *bronto.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *monitorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan MonitorModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, diags := encodeCreateMonitorRequest(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.CreateRequest(ctx, body, "/monitors")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while calling the create API. Error: "+err.Error(),
		)
		return
	}

	created, err := bronto.DecodeMonitor(httpResp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while parsing the resource creation response. JSON Error: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(populateModel(ctx, &plan, created)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *monitorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MonitorModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	m, err := r.getMonitor(ctx, data.Id.ValueString())
	if err != nil {
		if bronto.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Read Resource",
			"An unexpected error occurred while calling the read API. Error: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(populateModel(ctx, &data, m)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *monitorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan MonitorModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state MonitorModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	monitorId := state.Id.ValueString()

	body, diags := encodeUpdateMonitorRequest(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// PUT fully replaces the monitor (except mute state) and responds 200
	// with the updated resource.
	httpResp, err := r.client.PutRequestWithBody(ctx, body, "/monitors/"+monitorId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while calling the update API. Error: "+err.Error(),
		)
		return
	}
	if _, err := bronto.DecodeMonitor(httpResp); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while parsing the resource update response. JSON Error: "+err.Error(),
		)
		return
	}

	// Muting can only be changed through a dedicated endpoint, not the PUT
	// above. A null/unset muted_until means unmuted (mute_until: 0).
	if !plan.MutedUntil.Equal(state.MutedUntil) {
		muteUntil := int64(0)
		if !plan.MutedUntil.IsNull() && !plan.MutedUntil.IsUnknown() {
			muteUntil = plan.MutedUntil.ValueInt64()
		}
		muteBody, err := json.Marshal(bronto.MuteMonitorRequest{MuteUntil: muteUntil})
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Update Resource",
				"An unexpected error occurred while encoding the mute request. JSON Error: "+err.Error(),
			)
			return
		}
		if _, err := r.client.PostRequest(ctx, muteBody, "/monitors/"+monitorId+"/status"); err != nil {
			resp.Diagnostics.AddError(
				"Unable to Update Resource",
				"An unexpected error occurred while updating the monitor's mute state. Error: "+err.Error(),
			)
			return
		}
	}

	updated, err := r.getMonitor(ctx, monitorId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while reading back the updated resource. Error: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(populateModel(ctx, &plan, updated)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *monitorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data MonitorModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteRequest(ctx, "/monitors/"+data.Id.ValueString())
	if err != nil && !bronto.IsNotFound(err) {
		resp.Diagnostics.AddError(
			"Unable to Delete Resource",
			"An unexpected error occurred while calling the delete API. Error: "+err.Error(),
		)
		return
	}
}

func (r *monitorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *monitorResource) getMonitor(ctx context.Context, id string) (bronto.Monitor, error) {
	httpResp, err := r.client.ReadRequest(ctx, "/monitors/"+id)
	if err != nil {
		return bronto.Monitor{}, err
	}
	return bronto.DecodeMonitor(httpResp)
}

func encodeCreateMonitorRequest(ctx context.Context, data MonitorModel) ([]byte, diag.Diagnostics) {
	var diags diag.Diagnostics

	actions, actionDiags := extractActions(data.Actions)
	diags.Append(actionDiags...)
	if diags.HasError() {
		return nil, diags
	}

	reqBody := bronto.CreateMonitorRequest{
		Name:                 data.Name.ValueString(),
		Description:          data.Description.ValueString(),
		ComparisonOperator:   data.ComparisonOperator.ValueString(),
		Threshold:            data.Threshold.ValueFloat64(),
		WarningThreshold:     float64PointerValue(data.WarningThreshold),
		Window:               data.Window.ValueString(),
		Actions:              actions,
		NoDataStatus:         data.NoDataStatus.ValueString(),
		NotifyOnce:           boolPointerValue(data.NotifyOnce),
		GroupRetention:       int64PointerValue(data.GroupRetention),
		MuteUntil:            int64PointerValue(data.MutedUntil),
		MonitorType:          data.MonitorType.ValueString(),
		CompareTo:            data.CompareTo.ValueString(),
		ChangeType:           data.ChangeType.ValueString(),
		EvaluationWindow:     data.EvaluationWindow.ValueString(),
		AnomalyAlgorithm:     data.AnomalyAlgorithm.ValueString(),
		AiReportEnabled:      boolPointerValue(data.AiReportEnabled),
		AiReportInstructions: data.AiReportInstructions.ValueString(),
		MetricId:             data.MetricId.ValueString(),
	}
	if !data.Aux.IsNull() && !data.Aux.IsUnknown() {
		reqBody.Aux = json.RawMessage(data.Aux.ValueString())
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		diags.AddError(
			"Unable to Encode Request",
			"An unexpected error occurred while encoding the monitor create request. JSON Error: "+err.Error(),
		)
		return nil, diags
	}
	return body, diags
}

func encodeUpdateMonitorRequest(ctx context.Context, data MonitorModel) ([]byte, diag.Diagnostics) {
	var diags diag.Diagnostics

	actions, actionDiags := extractActions(data.Actions)
	diags.Append(actionDiags...)
	if diags.HasError() {
		return nil, diags
	}

	reqBody := bronto.UpdateMonitorRequest{
		Name:                 data.Name.ValueString(),
		Description:          data.Description.ValueString(),
		ComparisonOperator:   data.ComparisonOperator.ValueString(),
		Threshold:            data.Threshold.ValueFloat64(),
		WarningThreshold:     float64PointerValue(data.WarningThreshold),
		Window:               data.Window.ValueString(),
		Actions:              actions,
		NoDataStatus:         data.NoDataStatus.ValueString(),
		NotifyOnce:           boolPointerValue(data.NotifyOnce),
		GroupRetention:       int64PointerValue(data.GroupRetention),
		MonitorType:          data.MonitorType.ValueString(),
		CompareTo:            data.CompareTo.ValueString(),
		ChangeType:           data.ChangeType.ValueString(),
		EvaluationWindow:     data.EvaluationWindow.ValueString(),
		AnomalyAlgorithm:     data.AnomalyAlgorithm.ValueString(),
		AiReportEnabled:      boolPointerValue(data.AiReportEnabled),
		AiReportInstructions: data.AiReportInstructions.ValueString(),
		MetricId:             data.MetricId.ValueString(),
	}
	if !data.Aux.IsNull() && !data.Aux.IsUnknown() {
		reqBody.Aux = json.RawMessage(data.Aux.ValueString())
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		diags.AddError(
			"Unable to Encode Request",
			"An unexpected error occurred while encoding the monitor update request. JSON Error: "+err.Error(),
		)
		return nil, diags
	}
	return body, diags
}

func extractActions(models []ActionModel) ([]bronto.MonitorAction, diag.Diagnostics) {
	var diags diag.Diagnostics
	actions := make([]bronto.MonitorAction, 0, len(models))
	for _, a := range models {
		actions = append(actions, bronto.MonitorAction{
			Type:          a.Type.ValueString(),
			Email:         a.Email.ValueString(),
			IntegrationId: a.IntegrationId.ValueString(),
		})
	}
	return actions, diags
}

// populateModel maps an API monitor response onto the Terraform model,
// including the id, so it can be used from Create, Read, and Update.
func populateModel(ctx context.Context, data *MonitorModel, m bronto.Monitor) diag.Diagnostics {
	var diags diag.Diagnostics

	data.Id = types.StringValue(m.Id)
	data.Name = types.StringValue(m.Name)
	data.ComparisonOperator = types.StringValue(m.ComparisonOperator)
	data.Threshold = types.Float64Value(m.Threshold)
	data.Window = types.StringValue(m.Window)
	data.NotifyOnce = types.BoolValue(m.NotifyOnce)
	data.MonitorType = types.StringValue(m.MonitorType)
	data.AiReportEnabled = types.BoolValue(m.AiReportEnabled)
	data.Status = types.StringValue(m.Status)

	if m.Description == "" {
		data.Description = types.StringNull()
	} else {
		data.Description = types.StringValue(m.Description)
	}

	if m.MetricId == "" {
		data.MetricId = types.StringNull()
	} else {
		data.MetricId = types.StringValue(m.MetricId)
	}

	if m.NoDataStatus == "" {
		data.NoDataStatus = types.StringNull()
	} else {
		data.NoDataStatus = types.StringValue(m.NoDataStatus)
	}

	if m.CompareTo == "" {
		data.CompareTo = types.StringNull()
	} else {
		data.CompareTo = types.StringValue(m.CompareTo)
	}

	if m.ChangeType == "" {
		data.ChangeType = types.StringNull()
	} else {
		data.ChangeType = types.StringValue(m.ChangeType)
	}

	if m.EvaluationWindow == "" {
		data.EvaluationWindow = types.StringNull()
	} else {
		data.EvaluationWindow = types.StringValue(m.EvaluationWindow)
	}

	if m.AnomalyAlgorithm == "" {
		data.AnomalyAlgorithm = types.StringNull()
	} else {
		data.AnomalyAlgorithm = types.StringValue(m.AnomalyAlgorithm)
	}

	if m.AiReportInstructions == "" {
		data.AiReportInstructions = types.StringNull()
	} else {
		data.AiReportInstructions = types.StringValue(m.AiReportInstructions)
	}

	data.WarningThreshold = float64PointerToValue(m.WarningThreshold)
	data.GroupRetention = int64PointerToValue(m.GroupRetention)
	data.LastTriggerTs = int64PointerToValue(m.LastTriggerTs)

	// Unlike group_retention/last_trigger_ts, the API omits muted_until
	// entirely when unmuted rather than returning 0 - normalize that to 0
	// (not null) to match what "unmuted" means on write (muted_until = 0
	// or omitted), so a configured 0 doesn't flip to null on refresh.
	if m.MutedUntil == nil {
		data.MutedUntil = types.Int64Value(0)
	} else {
		data.MutedUntil = types.Int64Value(*m.MutedUntil)
	}

	if len(m.Aux) == 0 {
		data.Aux = jsontypes.NewNormalizedNull()
	} else {
		data.Aux = jsontypes.NewNormalizedValue(string(m.Aux))
	}

	actions := make([]ActionModel, 0, len(m.Actions))
	for _, a := range m.Actions {
		action := ActionModel{Type: types.StringValue(a.Type)}
		if a.Email == "" {
			action.Email = types.StringNull()
		} else {
			action.Email = types.StringValue(a.Email)
		}
		if a.IntegrationId == "" {
			action.IntegrationId = types.StringNull()
		} else {
			action.IntegrationId = types.StringValue(a.IntegrationId)
		}
		actions = append(actions, action)
	}
	data.Actions = actions

	groups := make([]MonitoredGroupModel, 0, len(m.MonitoredGroups))
	for _, g := range m.MonitoredGroups {
		groupsVal, groupDiags := stringMapValue(ctx, g.Groups)
		diags.Append(groupDiags...)
		groups = append(groups, MonitoredGroupModel{
			GroupId: types.StringValue(g.GroupId),
			Groups:  groupsVal,
			Status:  types.StringValue(g.Status),
		})
	}
	monitoredGroupsVal, monitoredGroupsDiags := types.ListValueFrom(ctx, monitoredGroupObjectType, groups)
	diags.Append(monitoredGroupsDiags...)
	data.MonitoredGroups = monitoredGroupsVal

	return diags
}

// stringMapValue converts a map[string]string into a types.Map, normalizing
// a nil map to an empty (but non-null) map first, so the API responding
// with an absent field or an empty object for "no values" map to the same
// state value.
func stringMapValue(ctx context.Context, values map[string]string) (types.Map, diag.Diagnostics) {
	if values == nil {
		values = map[string]string{}
	}
	return types.MapValueFrom(ctx, types.StringType, values)
}

func float64PointerValue(v types.Float64) *float64 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	val := v.ValueFloat64()
	return &val
}

func float64PointerToValue(v *float64) types.Float64 {
	if v == nil {
		return types.Float64Null()
	}
	return types.Float64Value(*v)
}

func int64PointerValue(v types.Int64) *int64 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	val := v.ValueInt64()
	return &val
}

func int64PointerToValue(v *int64) types.Int64 {
	if v == nil {
		return types.Int64Null()
	}
	return types.Int64Value(*v)
}

func boolPointerValue(v types.Bool) *bool {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	val := v.ValueBool()
	return &val
}
