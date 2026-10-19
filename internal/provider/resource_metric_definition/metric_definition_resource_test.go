package resource_metric_definition

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

	m := bronto.MetricDefinition{
		Id:          "metric-def-1",
		Name:        "api_response_time",
		Description: "Tracks response time for API endpoints",
		Queries: []bronto.MetricQuery{
			{
				Name:     "production_api_response_time",
				Select:   []string{"response_time_ms"},
				FromExpr: "environment = 'production'",
				Where:    "response_time_ms > 0",
				Groups:   []string{"endpoint", "method"},
				Aggregation: &bronto.Aggregation{
					Time:     "avg",
					ReduceTo: "avg",
				},
				UnitConfig: json.RawMessage(`{"type":"data_size","input":"bytes"}`),
			},
		},
		Formulas: []bronto.MetricFormula{
			{
				Name:        "response_time_seconds",
				Expression:  "production_api_response_time / 1000",
				Aggregation: &bronto.FormulaAggregation{ReduceTo: "avg"},
			},
		},
	}

	var data MetricDefinitionModel
	diags := populateModel(ctx, &data, m)
	if diags.HasError() {
		t.Fatalf("populateModel: %v", diags)
	}

	if data.Id.ValueString() != "metric-def-1" {
		t.Errorf("Id = %s", data.Id.ValueString())
	}
	if data.Name.ValueString() != "api_response_time" {
		t.Errorf("Name = %s", data.Name.ValueString())
	}
	if data.Description.ValueString() != "Tracks response time for API endpoints" {
		t.Errorf("Description = %s", data.Description.ValueString())
	}

	if len(data.Queries) != 1 {
		t.Fatalf("Queries = %d entries, want 1", len(data.Queries))
	}
	q := data.Queries[0]
	if q.Name.ValueString() != "production_api_response_time" {
		t.Errorf("Query.Name = %s", q.Name.ValueString())
	}
	if q.From.IsNull() {
		t.Errorf("Query.From should be an empty list, not null, when the query used from_expr")
	}
	if q.FromExpr.ValueString() != "environment = 'production'" {
		t.Errorf("Query.FromExpr = %s", q.FromExpr.ValueString())
	}
	if q.UnitConfig.ValueString() != `{"type":"data_size","input":"bytes"}` {
		t.Errorf("Query.UnitConfig = %s", q.UnitConfig.ValueString())
	}
	if q.Aggregation == nil {
		t.Fatalf("Query.Aggregation should not be nil")
	}
	if q.Aggregation.Time.ValueString() != "avg" {
		t.Errorf("Query.Aggregation.Time = %s", q.Aggregation.Time.ValueString())
	}
	if q.Aggregation.TimeFunctionParams.IsNull() {
		t.Errorf("Query.Aggregation.TimeFunctionParams should be an empty list, not null")
	}

	var formulas []FormulaModel
	if diags := data.Formulas.ElementsAs(ctx, &formulas, false); diags.HasError() {
		t.Fatalf("Formulas.ElementsAs: %v", diags)
	}
	if len(formulas) != 1 {
		t.Fatalf("Formulas = %d entries, want 1", len(formulas))
	}
	if formulas[0].ReduceTo.ValueString() != "avg" {
		t.Errorf("Formula.ReduceTo = %s", formulas[0].ReduceTo.ValueString())
	}
	if !formulas[0].UnitConfig.IsNull() {
		t.Errorf("Formula.UnitConfig should be null when the API returns none")
	}
}

func TestPopulateModelEmptyNameAndDescriptionAreNull(t *testing.T) {
	m := bronto.MetricDefinition{
		Id: "metric-def-1",
		Queries: []bronto.MetricQuery{
			{Name: "q1", Select: []string{"field"}, Where: "x > 0", From: []string{"log-1"}},
		},
	}

	var data MetricDefinitionModel
	diags := populateModel(context.Background(), &data, m)
	if diags.HasError() {
		t.Fatalf("populateModel: %v", diags)
	}

	if !data.Name.IsNull() {
		t.Errorf("Name should be null when the API returns empty, got %q", data.Name.ValueString())
	}
	if !data.Description.IsNull() {
		t.Errorf("Description should be null when the API returns empty, got %q", data.Description.ValueString())
	}
	if data.Queries[0].Aggregation != nil {
		t.Errorf("Query.Aggregation should be nil when the API returns none")
	}
	if data.Queries[0].Groups.IsNull() {
		t.Errorf("Query.Groups should be an empty list, not null, when there are no groups")
	}
	if data.Formulas.IsNull() {
		t.Errorf("Formulas should be an empty list, not null, when there are no formulas")
	}
}

func TestExtractQueriesRoundTripsAggregation(t *testing.T) {
	ctx := context.Background()

	selectList, _ := types.ListValueFrom(ctx, types.StringType, []string{"field"})
	fromExprList := types.ListNull(types.StringType)
	groupsList := types.ListNull(types.StringType)
	tfpList, _ := types.ListValueFrom(ctx, types.StringType, []string{"count"})

	queries := []QueryModel{
		{
			Name:     types.StringValue("q1"),
			Select:   selectList,
			From:     fromExprList,
			FromExpr: types.StringValue("service = 'api'"),
			Where:    types.StringValue("status >= 400"),
			Groups:   groupsList,
			Aggregation: &AggregationModel{
				Time:               types.StringValue("rate"),
				ReduceTo:           types.StringValue("sum"),
				TimeFunctionParams: tfpList,
			},
			UnitConfig: jsontypes.NewNormalizedNull(),
		},
	}

	encoded, diags := extractQueries(ctx, queries)
	if diags.HasError() {
		t.Fatalf("extractQueries: %v", diags)
	}
	if len(encoded) != 1 {
		t.Fatalf("encoded = %d entries, want 1", len(encoded))
	}

	q := encoded[0]
	if q.FromExpr != "service = 'api'" {
		t.Errorf("FromExpr = %s", q.FromExpr)
	}
	if q.Aggregation == nil {
		t.Fatalf("Aggregation should not be nil")
	}
	if q.Aggregation.Time != "rate" || q.Aggregation.ReduceTo != "sum" {
		t.Errorf("Aggregation = %+v", q.Aggregation)
	}
	if len(q.Aggregation.TimeFunctionParams) != 1 || q.Aggregation.TimeFunctionParams[0] != "count" {
		t.Errorf("Aggregation.TimeFunctionParams = %v", q.Aggregation.TimeFunctionParams)
	}

	// Round-trip through the wire (de)serialization to confirm the
	// single-element aggregation array shape is handled both ways.
	body, err := json.Marshal(q)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded bronto.MetricQuery
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.Aggregation == nil || decoded.Aggregation.Time != "rate" {
		t.Errorf("decoded.Aggregation = %+v", decoded.Aggregation)
	}
}

func TestExtractFormulasOmitsAggregationWhenReduceToUnset(t *testing.T) {
	ctx := context.Background()

	formulas, diags := types.ListValueFrom(ctx, formulaObjectType, []FormulaModel{
		{
			Name:       types.StringValue("f1"),
			Expression: types.StringValue("q1 + q2"),
			ReduceTo:   types.StringNull(),
			UnitConfig: jsontypes.NewNormalizedNull(),
		},
	})
	if diags.HasError() {
		t.Fatalf("types.ListValueFrom: %v", diags)
	}

	encoded, diags := extractFormulas(ctx, formulas)
	if diags.HasError() {
		t.Fatalf("extractFormulas: %v", diags)
	}
	if len(encoded) != 1 {
		t.Fatalf("encoded = %d entries, want 1", len(encoded))
	}
	if encoded[0].Aggregation != nil {
		t.Errorf("Aggregation should be nil when reduce_to is unset, got %+v", encoded[0].Aggregation)
	}
}

func TestExtractStringListNull(t *testing.T) {
	values, diags := extractStringList(context.Background(), types.ListNull(types.StringType))
	if diags.HasError() {
		t.Fatalf("extractStringList: %v", diags)
	}
	if values != nil {
		t.Errorf("expected nil for a null list, got %v", values)
	}
}
