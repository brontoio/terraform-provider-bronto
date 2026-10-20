package internal

import (
	"encoding/json"
	"strings"
)

// DashboardLayout mirrors the metric_ids grouping documented on
// Dashboard.layout / CreateDashboardRequest.layout. Note: the API's own
// spec is inconsistent here - the PATCH /dashboards/{id} example shows a
// different, widget-position shape (widget_layouts) under the same "layout"
// key that Dashboard.layout does not declare. Only the documented
// metric_ids form is supported.
type DashboardLayout struct {
	MetricIds []string `json:"metric_ids"`
}

// Dashboard is the wire representation of the Bronto Dashboards API's
// Dashboard schema.
type Dashboard struct {
	Id          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Widgets     []Widget          `json:"widgets"`
	Layout      *DashboardLayout  `json:"layout"`
	Tags        map[string]string `json:"tags"`
}

// CreateDashboardRequest is the body of POST /dashboards.
type CreateDashboardRequest struct {
	Name   string           `json:"name"`
	Layout *DashboardLayout `json:"layout,omitempty"`
}

// PatchDashboardNameRequest is one of the two mutually exclusive bodies
// accepted by PATCH /dashboards/{id}.
type PatchDashboardNameRequest struct {
	Name string `json:"name"`
}

// PatchDashboardLayoutRequest is the other body accepted by
// PATCH /dashboards/{id}; it cannot be combined with a name change in the
// same request.
type PatchDashboardLayoutRequest struct {
	Layout DashboardLayout `json:"layout"`
}

// AppendDashboardWidgetsRequest is the body of
// POST /dashboards/{id}/widgets.
type AppendDashboardWidgetsRequest struct {
	WidgetIds []string `json:"widget_ids"`
}

// DashboardsList is the response body of GET /dashboards.
type DashboardsList struct {
	Dashboards []Dashboard `json:"dashboards"`
}

// DecodeDashboard decodes a single Dashboard from a JSON response body.
func DecodeDashboard(body string) (Dashboard, error) {
	var d Dashboard
	err := json.NewDecoder(strings.NewReader(body)).Decode(&d)
	return d, err
}

// DecodeDashboardsList decodes a {"dashboards": [...]} response body.
func DecodeDashboardsList(body string) (DashboardsList, error) {
	var list DashboardsList
	err := json.NewDecoder(strings.NewReader(body)).Decode(&list)
	return list, err
}
