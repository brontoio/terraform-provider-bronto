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

	if data.Tags.IsNull() {
		t.Errorf("Tags should not be null")
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
