package resource_widget

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	bronto "terraform-provider-bronto/internal"
)

func TestPopulateModelMapsAllFields(t *testing.T) {
	ctx := context.Background()

	w := bronto.Widget{
		Id:          "widget-1",
		Name:        "5xx errors by service",
		Description: "This widget shows the 5xx errors by service",
		Type:        "treemap",
		Aux:         json.RawMessage(`{"color_name":"Purple"}`),
		MetricIds:   []string{"metric-1", "metric-2"},
		WidgetIds:   nil,
		Tags:        map[string]string{"team": "observability"},
	}

	var data WidgetModel
	diags := populateModel(ctx, &data, w)
	if diags.HasError() {
		t.Fatalf("populateModel: %v", diags)
	}

	if data.Id.ValueString() != "widget-1" {
		t.Errorf("Id = %s", data.Id.ValueString())
	}
	if data.Name.ValueString() != "5xx errors by service" {
		t.Errorf("Name = %s", data.Name.ValueString())
	}
	if data.Description.ValueString() != "This widget shows the 5xx errors by service" {
		t.Errorf("Description = %s", data.Description.ValueString())
	}
	if data.Type.ValueString() != "treemap" {
		t.Errorf("Type = %s", data.Type.ValueString())
	}
	if data.Aux.ValueString() != `{"color_name":"Purple"}` {
		t.Errorf("Aux = %s", data.Aux.ValueString())
	}
	if data.MetricIds.IsNull() {
		t.Errorf("MetricIds should not be null")
	}
	if data.WidgetIds.IsNull() {
		t.Errorf("WidgetIds should not be null (expected an empty set)")
	}
	if data.Tags.IsNull() {
		t.Errorf("Tags should not be null")
	}
}

func TestPopulateModelEmptyDescriptionAndAuxAreNull(t *testing.T) {
	var data WidgetModel
	diags := populateModel(context.Background(), &data, bronto.Widget{Id: "w", Name: "n", Type: "line"})
	if diags.HasError() {
		t.Fatalf("populateModel: %v", diags)
	}
	if !data.Description.IsNull() {
		t.Errorf("Description should be null when the API returns empty, got %q", data.Description.ValueString())
	}
	if !data.Aux.IsNull() {
		t.Errorf("Aux should be null when the API returns no aux, got %q", data.Aux.ValueString())
	}
}

func TestEncodeWidgetRequestIncludesAux(t *testing.T) {
	ctx := context.Background()

	metricIds, _ := types.SetValueFrom(ctx, types.StringType, []string{"m1"})

	data := WidgetModel{
		Name:      types.StringValue("Widget"),
		Type:      types.StringValue("line"),
		Aux:       jsontypes.NewNormalizedValue(`{"mode":"sql"}`),
		MetricIds: metricIds,
		WidgetIds: types.SetNull(types.StringType),
	}

	body, diags := encodeWidgetRequest(ctx, data)
	if diags.HasError() {
		t.Fatalf("encodeWidgetRequest: %v", diags)
	}

	var decoded bronto.WidgetRequest
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("unmarshal request body: %v", err)
	}

	if decoded.Name != "Widget" {
		t.Errorf("Name = %s", decoded.Name)
	}
	if string(decoded.Aux) != `{"mode":"sql"}` {
		t.Errorf("Aux = %s", string(decoded.Aux))
	}
	if len(decoded.MetricIds) != 1 || decoded.MetricIds[0] != "m1" {
		t.Errorf("MetricIds = %v", decoded.MetricIds)
	}
	if len(decoded.WidgetIds) != 0 {
		t.Errorf("WidgetIds = %v, want empty", decoded.WidgetIds)
	}
}

func TestEncodeWidgetRequestOmitsAuxWhenNull(t *testing.T) {
	ctx := context.Background()

	data := WidgetModel{
		Name:      types.StringValue("Widget"),
		Type:      types.StringValue("line"),
		Aux:       jsontypes.NewNormalizedNull(),
		MetricIds: types.SetNull(types.StringType),
		WidgetIds: types.SetNull(types.StringType),
	}

	body, diags := encodeWidgetRequest(ctx, data)
	if diags.HasError() {
		t.Fatalf("encodeWidgetRequest: %v", diags)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		t.Fatalf("unmarshal request body: %v", err)
	}
	if _, ok := raw["aux"]; ok {
		t.Errorf("expected aux to be omitted, got: %s", raw["aux"])
	}
}

func TestExtractStringSetNull(t *testing.T) {
	values, diags := extractStringSet(context.Background(), types.SetNull(types.StringType))
	if diags.HasError() {
		t.Fatalf("extractStringSet: %v", diags)
	}
	if values != nil {
		t.Errorf("expected nil for a null set, got %v", values)
	}
}
