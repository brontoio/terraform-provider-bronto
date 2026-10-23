package resource_monitor

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	bronto "terraform-provider-bronto/internal"
)

func TestPopulateModelMapsAllFields(t *testing.T) {
	ctx := context.Background()
	warningThreshold := 900.0
	groupRetention := int64(600000)
	mutedUntil := int64(-1)
	lastTriggerTs := int64(1710948395538)

	m := bronto.Monitor{
		Id:                   "monitor-1",
		Name:                 "High Error Rate",
		Description:          "Average response time is high",
		ComparisonOperator:   "ABOVE",
		Threshold:            1000,
		WarningThreshold:     &warningThreshold,
		Window:               "Last 20 minutes",
		Actions:              []bronto.MonitorAction{{Type: "EMAIL", Email: "a@example.com"}},
		Status:               "OK",
		NoDataStatus:         "ALERT",
		MutedUntil:           &mutedUntil,
		NotifyOnce:           true,
		GroupRetention:       &groupRetention,
		LastTriggerTs:        &lastTriggerTs,
		Aux:                  json.RawMessage(`{"query_mode":"visual"}`),
		MonitorType:          "PATTERN",
		AiReportEnabled:      true,
		AiReportInstructions: "check deployments",
		MetricId:             "metric-1",
		MonitoredGroups: []bronto.MonitoredGroup{
			{GroupId: "region:EU", Groups: map[string]string{"region": "EU"}, Status: "OK"},
		},
	}

	var data MonitorModel
	diags := populateModel(ctx, &data, m)
	if diags.HasError() {
		t.Fatalf("populateModel: %v", diags)
	}

	if data.Id.ValueString() != "monitor-1" {
		t.Errorf("Id = %s", data.Id.ValueString())
	}
	if data.MetricDefId.ValueString() != "metric-1" {
		t.Errorf("MetricDefId = %s", data.MetricDefId.ValueString())
	}
	if data.Threshold.ValueFloat64() != 1000 {
		t.Errorf("Threshold = %v", data.Threshold.ValueFloat64())
	}
	if data.WarningThreshold.ValueFloat64() != 900 {
		t.Errorf("WarningThreshold = %v", data.WarningThreshold.ValueFloat64())
	}
	if data.MutedUntil.ValueInt64() != -1 {
		t.Errorf("MutedUntil = %v", data.MutedUntil.ValueInt64())
	}
	if data.LastTriggerTs.ValueInt64() != lastTriggerTs {
		t.Errorf("LastTriggerTs = %v", data.LastTriggerTs.ValueInt64())
	}
	if data.Status.ValueString() != "OK" {
		t.Errorf("Status = %s", data.Status.ValueString())
	}
	if data.Aux.ValueString() != `{"query_mode":"visual"}` {
		t.Errorf("Aux = %s", data.Aux.ValueString())
	}
	if len(data.Actions) != 1 || data.Actions[0].Type.ValueString() != "EMAIL" || data.Actions[0].Email.ValueString() != "a@example.com" {
		t.Errorf("Actions = %+v", data.Actions)
	}
	var groups []MonitoredGroupModel
	if diags := data.MonitoredGroups.ElementsAs(ctx, &groups, false); diags.HasError() {
		t.Fatalf("MonitoredGroups.ElementsAs: %v", diags)
	}
	if len(groups) != 1 || groups[0].GroupId.ValueString() != "region:EU" {
		t.Errorf("MonitoredGroups = %+v", groups)
	}
}

func TestPopulateModelAbsentOptionalFieldsAreNull(t *testing.T) {
	ctx := context.Background()
	m := bronto.Monitor{
		Id:                 "monitor-1",
		Name:               "m",
		ComparisonOperator: "ABOVE",
		Threshold:          1,
		Window:             "Last 5 minutes",
		MonitorType:        "PATTERN",
	}

	var data MonitorModel
	diags := populateModel(ctx, &data, m)
	if diags.HasError() {
		t.Fatalf("populateModel: %v", diags)
	}

	if !data.Description.IsNull() {
		t.Errorf("Description should be null when absent, got %q", data.Description.ValueString())
	}
	if !data.WarningThreshold.IsNull() {
		t.Errorf("WarningThreshold should be null when absent")
	}
	// Unlike the other optional fields, an absent muted_until normalizes to
	// 0 (not null): the API omits it entirely to mean "unmuted", and 0 is
	// what a config sets to mean the same thing - see populateModel.
	if data.MutedUntil.ValueInt64() != 0 {
		t.Errorf("MutedUntil should be 0 when absent, got %v", data.MutedUntil.ValueInt64())
	}
	if !data.MetricDefId.IsNull() {
		t.Errorf("MetricDefId should be null when absent")
	}
	if data.Aux.IsNull() == false {
		t.Errorf("Aux should be null when absent")
	}
	if len(data.Actions) != 0 {
		t.Errorf("Actions should be empty, got %+v", data.Actions)
	}
	var groups []MonitoredGroupModel
	if diags := data.MonitoredGroups.ElementsAs(ctx, &groups, false); diags.HasError() {
		t.Fatalf("MonitoredGroups.ElementsAs: %v", diags)
	}
	if len(groups) != 0 {
		t.Errorf("MonitoredGroups should be empty, got %+v", groups)
	}
}

func TestEncodeCreateMonitorRequest(t *testing.T) {
	ctx := context.Background()
	data := MonitorModel{
		Name:               types.StringValue("High Error Rate"),
		ComparisonOperator: types.StringValue("ABOVE"),
		Threshold:          types.Float64Value(1000),
		Window:             types.StringValue("Last 20 minutes"),
		MetricDefId:        types.StringValue("metric-1"),
		MonitorType:        types.StringValue("PATTERN"),
		NotifyOnce:         types.BoolValue(false),
		AiReportEnabled:    types.BoolValue(false),
		MutedUntil:         types.Int64Value(-1),
		Actions: []ActionModel{
			{Type: types.StringValue("EMAIL"), Email: types.StringValue("a@example.com")},
		},
	}

	body, diags := encodeCreateMonitorRequest(ctx, data)
	if diags.HasError() {
		t.Fatalf("encodeCreateMonitorRequest: %v", diags)
	}

	var decoded bronto.CreateMonitorRequest
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("unmarshal request body: %v", err)
	}

	if decoded.MetricDefId != "metric-1" {
		t.Errorf("MetricDefId = %s", decoded.MetricDefId)
	}
	if decoded.Threshold != 1000 {
		t.Errorf("Threshold = %v", decoded.Threshold)
	}
	if decoded.MuteUntil == nil || *decoded.MuteUntil != -1 {
		t.Errorf("MuteUntil = %v", decoded.MuteUntil)
	}
	if len(decoded.Actions) != 1 || decoded.Actions[0].Email != "a@example.com" {
		t.Errorf("Actions = %+v", decoded.Actions)
	}
}

func TestEncodeUpdateMonitorRequestOmitsMuteUntil(t *testing.T) {
	ctx := context.Background()
	data := MonitorModel{
		Name:               types.StringValue("m"),
		ComparisonOperator: types.StringValue("ABOVE"),
		Threshold:          types.Float64Value(1),
		Window:             types.StringValue("Last 5 minutes"),
		MetricDefId:        types.StringValue("metric-1"),
	}

	body, diags := encodeUpdateMonitorRequest(ctx, data)
	if diags.HasError() {
		t.Fatalf("encodeUpdateMonitorRequest: %v", diags)
	}

	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		t.Fatalf("unmarshal request body: %v", err)
	}
	if _, ok := raw["mute_until"]; ok {
		t.Errorf("mute_until must not appear in the update request body, got %v", raw)
	}
}
