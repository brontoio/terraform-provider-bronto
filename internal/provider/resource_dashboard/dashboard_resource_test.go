package resource_dashboard

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	bronto "terraform-provider-bronto/internal"
)

func TestPopulateModelMapsAllFields(t *testing.T) {
	ctx := context.Background()

	d := bronto.Dashboard{
		Id:          "dash-1",
		Name:        "Application Performance Dashboard",
		Description: "Overview of application metrics and performance indicators",
		Widgets: []bronto.Widget{
			{Id: "widget-1"},
			{Id: "widget-2"},
		},
		Layout: &bronto.DashboardLayout{MetricIds: []string{"metric-1"}},
		Tags:   map[string]string{"team": "observability"},
	}

	var data DashboardModel
	diags := populateModel(ctx, &data, d)
	if diags.HasError() {
		t.Fatalf("populateModel: %v", diags)
	}

	if data.Id.ValueString() != "dash-1" {
		t.Errorf("Id = %s", data.Id.ValueString())
	}
	if data.Name.ValueString() != "Application Performance Dashboard" {
		t.Errorf("Name = %s", data.Name.ValueString())
	}
	if data.Description.ValueString() != "Overview of application metrics and performance indicators" {
		t.Errorf("Description = %s", data.Description.ValueString())
	}

	var widgetIds []string
	data.WidgetIds.ElementsAs(ctx, &widgetIds, false)
	if len(widgetIds) != 2 {
		t.Errorf("WidgetIds = %v, want 2 elements", widgetIds)
	}

	var layoutMetricIds []string
	data.LayoutMetricIds.ElementsAs(ctx, &layoutMetricIds, false)
	if len(layoutMetricIds) != 1 || layoutMetricIds[0] != "metric-1" {
		t.Errorf("LayoutMetricIds = %v", layoutMetricIds)
	}

}

func TestPopulateTagsEchoesExplicitPlanValue(t *testing.T) {
	ctx := context.Background()
	r := &dashboardResource{client: &bronto.Client{DefaultTags: map[string]string{"managed_by": "terraform"}}}

	planTags, _ := types.MapValueFrom(ctx, types.StringType, map[string]string{"team": "observability"})

	var data DashboardModel
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
	r := &dashboardResource{client: &bronto.Client{DefaultTags: map[string]string{"managed_by": "terraform"}}}

	var data DashboardModel
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

func TestPopulateModelNoWidgetsOrLayoutIsEmptyNotNull(t *testing.T) {
	var data DashboardModel
	diags := populateModel(context.Background(), &data, bronto.Dashboard{Id: "dash-1", Name: "Empty"})
	if diags.HasError() {
		t.Fatalf("populateModel: %v", diags)
	}

	if data.WidgetIds.IsNull() {
		t.Errorf("WidgetIds should be an empty set, not null, when there are no widgets")
	}
	if data.LayoutMetricIds.IsNull() {
		t.Errorf("LayoutMetricIds should be an empty set, not null, when there is no layout")
	}
	if !data.Description.IsNull() {
		t.Errorf("Description should be null when the API returns empty, got %q", data.Description.ValueString())
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
