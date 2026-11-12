package internal

import (
	"encoding/json"
	"strings"
)

// WidgetSource is a reference to a metric, monitor, SLO, or saved search
// that supplies data to a widget.
type WidgetSource struct {
	Type string `json:"type"`
	Id   string `json:"id"`
}

// WidgetDisplay holds a widget's display options. Each field applies only to
// some widget types and is ignored by the others. A field that's unset is
// omitted by the API (never defaulted server-side); NormalizeData and
// StackedBars are pointers so an explicit false can be told apart from unset.
type WidgetDisplay struct {
	NormalizeData *bool  `json:"normalize_data,omitempty"`
	StackedBars   *bool  `json:"stacked_bars,omitempty"`
	ColourName    string `json:"colour_name,omitempty"`
}

// WidgetSort holds how a widget orders its series or rows (e.g. top-list,
// table). Ignored by widget types that don't sort. By/Direction are left
// empty rather than defaulted when unset - the API documents "value"/"desc"
// as the effective default for display purposes only, it doesn't write them
// into the stored value.
type WidgetSort struct {
	By        string `json:"by,omitempty"`
	Direction string `json:"direction,omitempty"`
}

// Widget is the wire representation of the Bronto Widgets API's Widget
// schema. Display and Sort are always present in a response (even if their
// own fields are all empty), unlike WidgetRequest's, which are optional.
type Widget struct {
	Id          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Type        string            `json:"type"`
	Aux         json.RawMessage   `json:"aux,omitempty"`
	Sources     []WidgetSource    `json:"sources"`
	WidgetIds   []string          `json:"widget_ids"`
	Display     WidgetDisplay     `json:"display"`
	Sort        WidgetSort        `json:"sort"`
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
	Display     *WidgetDisplay  `json:"display,omitempty"`
	Sort        *WidgetSort     `json:"sort,omitempty"`
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
