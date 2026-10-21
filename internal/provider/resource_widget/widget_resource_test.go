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
		Sources: []bronto.WidgetSource{
			{Type: "metric", Id: "metric-1"},
			{Type: "metric", Id: "metric-2"},
		},
		WidgetIds: nil,
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
	if len(sources) != 2 {
		t.Fatalf("Sources = %d entries, want 2", len(sources))
	}
	if sources[0].Type.ValueString() != "metric" || sources[0].Id.ValueString() != "metric-1" {
		t.Errorf("Sources[0] = %+v", sources[0])
	}
	if data.WidgetIds.IsNull() {
		t.Errorf("WidgetIds should not be null (expected an empty set)")
	}
}

func TestUpdateTagSyncCurrentTagsComesFromTagsAll(t *testing.T) {
	ctx := context.Background()
	defaultTags := map[string]string{"managed_by": "terraform"}
	r := &widgetResource{client: &bronto.Client{DefaultTags: defaultTags}}

	stateTagsAll, _ := types.MapValueFrom(ctx, types.StringType, map[string]string{
		"managed_by": "terraform",
		"team":       "observability",
	})
	planTags := types.MapNull(types.StringType) // config doesn't declare its own tags

	currentTags, diags := extractStringMap(ctx, stateTagsAll)
	if diags.HasError() {
		t.Fatalf("extractStringMap(TagsAll): %v", diags)
	}
	specificTags, diags := extractStringMap(ctx, planTags)
	if diags.HasError() {
		t.Fatalf("extractStringMap(Tags): %v", diags)
	}
	desiredTags := bronto.MergeTags(r.client.DefaultTags, specificTags)

	if currentTags["managed_by"] != desiredTags["managed_by"] {
		t.Fatalf("currentTags[managed_by] = %q, desiredTags[managed_by] = %q - SyncTags would wrongly try to re-create an existing tag",
			currentTags["managed_by"], desiredTags["managed_by"])
	}
}

func TestPopulateTagsEchoesExplicitPlanValue(t *testing.T) {
	ctx := context.Background()
	r := &widgetResource{client: &bronto.Client{DefaultTags: map[string]string{"managed_by": "terraform"}}}

	planTags, _ := types.MapValueFrom(ctx, types.StringType, map[string]string{"team": "observability"})

	var data WidgetModel
	diags := r.populateTags(ctx, &data, planTags, map[string]string{"managed_by": "terraform", "team": "observability"})
	if diags.HasError() {
		t.Fatalf("populateTags: %v", diags)
	}

	if !data.Tags.Equal(planTags) {
		t.Errorf("Tags = %v, want it echoed back as the exact plan value %v", data.Tags, planTags)
	}

	var tagsAll map[string]string
	if diags := data.TagsAll.ElementsAs(ctx, &tagsAll, false); diags.HasError() {
		t.Fatalf("TagsAll.ElementsAs: %v", diags)
	}
	if tagsAll["managed_by"] != "terraform" || tagsAll["team"] != "observability" {
		t.Errorf("TagsAll = %v", tagsAll)
	}
}

func TestPopulateTagsDerivesOwnTagsWhenPlanUnset(t *testing.T) {
	ctx := context.Background()
	r := &widgetResource{client: &bronto.Client{DefaultTags: map[string]string{"managed_by": "terraform"}}}

	var data WidgetModel
	diags := r.populateTags(ctx, &data, types.MapNull(types.StringType), map[string]string{"managed_by": "terraform", "team": "observability"})
	if diags.HasError() {
		t.Fatalf("populateTags: %v", diags)
	}

	var tags map[string]string
	if diags := data.Tags.ElementsAs(ctx, &tags, false); diags.HasError() {
		t.Fatalf("Tags.ElementsAs: %v", diags)
	}
	if _, hasDefault := tags["managed_by"]; hasDefault {
		t.Errorf("Tags should exclude the default-derived key, got %v", tags)
	}
	if tags["team"] != "observability" {
		t.Errorf("Tags = %v, want team=observability", tags)
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
