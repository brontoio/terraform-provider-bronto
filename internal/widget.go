package internal

import (
	"encoding/json"
	"strings"
)

// WidgetSource is a reference to a metric, monitor, or SLO that supplies data to a widget.
type WidgetSource struct {
	Type string `json:"type"`
	Id   string `json:"id"`
}

// Widget is the wire representation of the Bronto Widgets API's Widget
// schema.
type Widget struct {
	Id          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Type        string            `json:"type"`
	Aux         json.RawMessage   `json:"aux,omitempty"`
	Sources     []WidgetSource    `json:"sources"`
	WidgetIds   []string          `json:"widget_ids"`
	Tags        map[string]string `json:"tags"`
}

// WidgetRequest is the body of both POST /widgets (create) and
// PUT /widgets/{id} (full replace) - both accept the same shape.
type WidgetRequest struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Type        string          `json:"type"`
	Aux         json.RawMessage `json:"aux,omitempty"`
	Sources     []WidgetSource  `json:"sources,omitempty"`
	WidgetIds   []string        `json:"widget_ids,omitempty"`
}

// WidgetsList is the response body of GET /widgets.
type WidgetsList struct {
	Widgets []Widget `json:"widgets"`
}

// DecodeWidget decodes a single Widget from a JSON response body.
func DecodeWidget(body string) (Widget, error) {
	var w Widget
	err := json.NewDecoder(strings.NewReader(body)).Decode(&w)
	return w, err
}

// DecodeWidgetsList decodes a {"widgets": [...]} response body.
func DecodeWidgetsList(body string) (WidgetsList, error) {
	var list WidgetsList
	err := json.NewDecoder(strings.NewReader(body)).Decode(&list)
	return list, err
}
