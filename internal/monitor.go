package internal

import (
	"encoding/json"
	"strings"
)

// MonitorAction is one notification action a monitor takes when it fires
// or resolves, per the Monitors API's Action schema.
type MonitorAction struct {
	Type          string `json:"type"`
	Email         string `json:"email,omitempty"`
	IntegrationId string `json:"integration_id,omitempty"`
}

// Monitor is the wire representation of the Monitors API's Monitor schema.
// Omits status, monitored_groups, and last_trigger_ts: the provider doesn't
// surface them (see monitor_resource_schema.go), so there's nothing that
// reads them here.
type Monitor struct {
	Id                   string          `json:"id"`
	Name                 string          `json:"name"`
	Description          string          `json:"description"`
	ComparisonOperator   string          `json:"comparison_operator"`
	Threshold            float64         `json:"threshold"`
	WarningThreshold     *float64        `json:"warning_threshold,omitempty"`
	Window               string          `json:"window"`
	Actions              []MonitorAction `json:"actions"`
	NoDataStatus         string          `json:"no_data_status,omitempty"`
	MutedUntil           *int64          `json:"muted_until,omitempty"`
	NotifyOnce           bool            `json:"notify_once"`
	GroupRetention       *int64          `json:"group_retention,omitempty"`
	Aux                  json.RawMessage `json:"aux,omitempty"`
	MonitorType          string          `json:"monitor_type"`
	CompareTo            string          `json:"compare_to,omitempty"`
	ChangeType           string          `json:"change_type,omitempty"`
	EvaluationWindow     string          `json:"evaluation_window,omitempty"`
	AnomalyAlgorithm     string          `json:"anomaly_algorithm,omitempty"`
	AiReportEnabled      bool            `json:"ai_report_enabled"`
	AiReportInstructions string          `json:"ai_report_instructions,omitempty"`
	MetricId             string          `json:"metric_id,omitempty"`
}

// CreateMonitorRequest is the body of POST /monitors. Only the metric_id
// (reference an existing metric definition) form is supported - not the
// alternative inline queries/formulas form - mirroring bronto_widget's
// sources, which also only ever references metric definitions (and
// monitors, SLOs) by id rather than embedding query definitions.
type CreateMonitorRequest struct {
	Name                 string          `json:"name"`
	Description          string          `json:"description,omitempty"`
	ComparisonOperator   string          `json:"comparison_operator"`
	Threshold            float64         `json:"threshold"`
	WarningThreshold     *float64        `json:"warning_threshold,omitempty"`
	Window               string          `json:"window"`
	Actions              []MonitorAction `json:"actions"`
	NoDataStatus         string          `json:"no_data_status,omitempty"`
	NotifyOnce           *bool           `json:"notify_once,omitempty"`
	GroupRetention       *int64          `json:"group_retention,omitempty"`
	Aux                  json.RawMessage `json:"aux,omitempty"`
	MuteUntil            *int64          `json:"mute_until,omitempty"`
	MonitorType          string          `json:"monitor_type,omitempty"`
	CompareTo            string          `json:"compare_to,omitempty"`
	ChangeType           string          `json:"change_type,omitempty"`
	EvaluationWindow     string          `json:"evaluation_window,omitempty"`
	AnomalyAlgorithm     string          `json:"anomaly_algorithm,omitempty"`
	AiReportEnabled      *bool           `json:"ai_report_enabled,omitempty"`
	AiReportInstructions string          `json:"ai_report_instructions,omitempty"`
	MetricId             string          `json:"metric_id"`
}

// UpdateMonitorRequest is the body of PUT /monitors/{id}. Unlike create, it
// has no mute_until field - muting is only changeable through the
// dedicated POST /monitors/{id}/status endpoint (see MuteMonitorRequest).
type UpdateMonitorRequest struct {
	Name                 string          `json:"name"`
	Description          string          `json:"description,omitempty"`
	ComparisonOperator   string          `json:"comparison_operator"`
	Threshold            float64         `json:"threshold"`
	WarningThreshold     *float64        `json:"warning_threshold,omitempty"`
	Window               string          `json:"window"`
	Actions              []MonitorAction `json:"actions"`
	NoDataStatus         string          `json:"no_data_status,omitempty"`
	NotifyOnce           *bool           `json:"notify_once,omitempty"`
	GroupRetention       *int64          `json:"group_retention,omitempty"`
	Aux                  json.RawMessage `json:"aux,omitempty"`
	MonitorType          string          `json:"monitor_type,omitempty"`
	CompareTo            string          `json:"compare_to,omitempty"`
	ChangeType           string          `json:"change_type,omitempty"`
	EvaluationWindow     string          `json:"evaluation_window,omitempty"`
	AnomalyAlgorithm     string          `json:"anomaly_algorithm,omitempty"`
	AiReportEnabled      *bool           `json:"ai_report_enabled,omitempty"`
	AiReportInstructions string          `json:"ai_report_instructions,omitempty"`
	MetricId             string          `json:"metric_id,omitempty"`
}

// MuteMonitorRequest is the body of POST /monitors/{id}/status, the only
// way to change a monitor's mute state after creation.
type MuteMonitorRequest struct {
	MuteUntil int64 `json:"mute_until"`
}

// MonitorsList is the response body of GET /monitors.
type MonitorsList struct {
	Monitors []Monitor `json:"monitors"`
}

// DecodeMonitor decodes a single Monitor from a JSON response body.
func DecodeMonitor(body string) (Monitor, error) {
	var m Monitor
	err := json.NewDecoder(strings.NewReader(body)).Decode(&m)
	return m, err
}
