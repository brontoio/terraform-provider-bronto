package resource_widget

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	bronto "terraform-provider-bronto/internal"
)

func TestPopulateModelMapsAllFields(t *testing.T) {
	ctx := context.Background()

	normalizeData := true
	w := bronto.Widget{
		Id:          "widget-1",
		Name:        "5xx errors by service",
		Description: "This widget shows the 5xx errors by service",
		Type:        "treemap",
		Aux:         json.RawMessage(`{"color_name":"Purple"}`),
		Sources: []bronto.WidgetSource{
			{Type: "metric", Id: "metric-1"},
			{Type: "metric", Id: "metric-2"},
			{Type: "saved_search", Id: "search-1"},
		},
		WidgetIds: nil,
		Display:   bronto.WidgetDisplay{NormalizeData: &normalizeData, ColourName: "blue"},
		Sort:      bronto.WidgetSort{By: "name", Direction: "asc"},
		Tags:      map[string]string{"team": "observability"},
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
	var sources []SourceModel
	if diags := data.Sources.ElementsAs(ctx, &sources, false); diags.HasError() {
		t.Fatalf("Sources.ElementsAs: %v", diags)
	}
	if len(sources) != 3 {
		t.Fatalf("Sources = %d entries, want 3", len(sources))
	}
	if sources[0].Type.ValueString() != "metric" || sources[0].Id.ValueString() != "metric-1" {
		t.Errorf("Sources[0] = %+v", sources[0])
	}
	if sources[2].Type.ValueString() != "saved_search" || sources[2].Id.ValueString() != "search-1" {
		t.Errorf("Sources[2] = %+v", sources[2])
	}
	if data.WidgetIds.IsNull() {
		t.Errorf("WidgetIds should not be null (expected an empty set)")
	}

	var display DisplayModel
	if diags := data.Display.As(ctx, &display, basetypes.ObjectAsOptions{}); diags.HasError() {
		t.Fatalf("Display.As: %v", diags)
	}
	if !display.NormalizeData.ValueBool() {
		t.Errorf("Display.NormalizeData = false, want true")
	}
	if display.ColourName.ValueString() != "blue" {
		t.Errorf("Display.ColourName = %s", display.ColourName.ValueString())
	}
	if !display.StackedBars.IsNull() {
		t.Errorf("Display.StackedBars should be null when the API omits it")
	}

	var sort SortModel
	if diags := data.Sort.As(ctx, &sort, basetypes.ObjectAsOptions{}); diags.HasError() {
		t.Fatalf("Sort.As: %v", diags)
	}
	if sort.By.ValueString() != "name" || sort.Direction.ValueString() != "asc" {
		t.Errorf("Sort = %+v", sort)
	}
}

func TestPopulateModelEmptyDisplayAndSortFieldsAreNull(t *testing.T) {
	ctx := context.Background()
	var data WidgetModel
	diags := populateModel(ctx, &data, bronto.Widget{Id: "w", Name: "n", Type: "line"})
	if diags.HasError() {
		t.Fatalf("populateModel: %v", diags)
	}

	var display DisplayModel
	if diags := data.Display.As(ctx, &display, basetypes.ObjectAsOptions{}); diags.HasError() {
		t.Fatalf("Display.As: %v", diags)
	}
	if !display.NormalizeData.IsNull() || !display.StackedBars.IsNull() || !display.ColourName.IsNull() {
		t.Errorf("Display fields should all be null when the API omits them, got %+v", display)
	}

	var sort SortModel
	if diags := data.Sort.As(ctx, &sort, basetypes.ObjectAsOptions{}); diags.HasError() {
		t.Fatalf("Sort.As: %v", diags)
	}
	if !sort.By.IsNull() || !sort.Direction.IsNull() {
		t.Errorf("Sort fields should all be null when the API omits them, got %+v", sort)
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

	sources, diags := types.ListValueFrom(ctx, sourceObjectType, []SourceModel{
		{Type: types.StringValue("metric"), Id: types.StringValue("m1")},
	})
	if diags.HasError() {
		t.Fatalf("types.ListValueFrom: %v", diags)
	}

	data := WidgetModel{
		Name:      types.StringValue("Widget"),
		Type:      types.StringValue("line"),
		Aux:       jsontypes.NewNormalizedValue(`{"mode":"sql"}`),
		Sources:   sources,
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
	if len(decoded.Sources) != 1 || decoded.Sources[0].Type != "metric" || decoded.Sources[0].Id != "m1" {
		t.Errorf("Sources = %v", decoded.Sources)
	}
	if len(decoded.WidgetIds) != 0 {
		t.Errorf("WidgetIds = %v, want empty", decoded.WidgetIds)
	}
}

func TestEncodeWidgetRequestIncludesDisplayAndSort(t *testing.T) {
	ctx := context.Background()

	display, diags := types.ObjectValueFrom(ctx, displayObjectType.AttrTypes, DisplayModel{
		NormalizeData: types.BoolValue(true),
		StackedBars:   types.BoolNull(),
		ColourName:    types.StringValue("blue"),
	})
	if diags.HasError() {
		t.Fatalf("types.ObjectValueFrom(display): %v", diags)
	}
	sort, diags := types.ObjectValueFrom(ctx, sortObjectType.AttrTypes, SortModel{
		By:        types.StringValue("name"),
		Direction: types.StringNull(),
	})
	if diags.HasError() {
		t.Fatalf("types.ObjectValueFrom(sort): %v", diags)
	}

	data := WidgetModel{
		Name:      types.StringValue("Widget"),
		Type:      types.StringValue("line"),
		Aux:       jsontypes.NewNormalizedNull(),
		Sources:   types.ListNull(sourceObjectType),
		WidgetIds: types.SetNull(types.StringType),
		Display:   display,
		Sort:      sort,
	}

	body, diags := encodeWidgetRequest(ctx, data)
	if diags.HasError() {
		t.Fatalf("encodeWidgetRequest: %v", diags)
	}

	var decoded bronto.WidgetRequest
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("unmarshal request body: %v", err)
	}

	if decoded.Display == nil {
		t.Fatalf("Display should not be nil")
	}
	if decoded.Display.NormalizeData == nil || !*decoded.Display.NormalizeData {
		t.Errorf("Display.NormalizeData = %v, want true", decoded.Display.NormalizeData)
	}
	if decoded.Display.StackedBars != nil {
		t.Errorf("Display.StackedBars should be omitted when null, got %v", *decoded.Display.StackedBars)
	}
	if decoded.Display.ColourName != "blue" {
		t.Errorf("Display.ColourName = %s", decoded.Display.ColourName)
	}
	if decoded.Sort == nil {
		t.Fatalf("Sort should not be nil")
	}
	if decoded.Sort.By != "name" {
		t.Errorf("Sort.By = %s", decoded.Sort.By)
	}
	if decoded.Sort.Direction != "" {
		t.Errorf("Sort.Direction should be omitted when null, got %s", decoded.Sort.Direction)
	}
}

func TestEncodeWidgetRequestOmitsDisplayAndSortWhenUnset(t *testing.T) {
	ctx := context.Background()

	data := WidgetModel{
		Name:      types.StringValue("Widget"),
		Type:      types.StringValue("line"),
		Aux:       jsontypes.NewNormalizedNull(),
		Sources:   types.ListNull(sourceObjectType),
		WidgetIds: types.SetNull(types.StringType),
		Display:   types.ObjectNull(displayObjectType.AttrTypes),
		Sort:      types.ObjectNull(sortObjectType.AttrTypes),
	}

	body, diags := encodeWidgetRequest(ctx, data)
	if diags.HasError() {
		t.Fatalf("encodeWidgetRequest: %v", diags)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		t.Fatalf("unmarshal request body: %v", err)
	}
	if _, ok := raw["display"]; ok {
		t.Errorf("expected display to be omitted, got: %s", raw["display"])
	}
	if _, ok := raw["sort"]; ok {
		t.Errorf("expected sort to be omitted, got: %s", raw["sort"])
	}
}

func TestEncodeWidgetRequestOmitsAuxWhenNull(t *testing.T) {
	ctx := context.Background()

	data := WidgetModel{
		Name:      types.StringValue("Widget"),
		Type:      types.StringValue("line"),
		Aux:       jsontypes.NewNormalizedNull(),
		Sources:   types.ListNull(sourceObjectType),
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
