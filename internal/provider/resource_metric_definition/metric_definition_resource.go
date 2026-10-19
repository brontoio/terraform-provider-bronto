package resource_metric_definition

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	bronto "terraform-provider-bronto/internal"
)

var (
	_ resource.Resource                = &metricDefinitionResource{}
	_ resource.ResourceWithConfigure   = &metricDefinitionResource{}
	_ resource.ResourceWithImportState = &metricDefinitionResource{}
)

func NewMetricDefinitionResource() resource.Resource {
	return &metricDefinitionResource{}
}

type metricDefinitionResource struct {
	client *bronto.Client
}

func (r *metricDefinitionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_metric_definition"
}

func (r *metricDefinitionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = MetricDefinitionResourceSchema(ctx)
}

func (r *metricDefinitionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*bronto.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *bronto.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *metricDefinitionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan MetricDefinitionModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, diags := encodeMetricDefinitionRequest(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.CreateRequest(ctx, body, "/metrics/definitions")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while calling the create API. Error: "+err.Error(),
		)
		return
	}

	created, err := bronto.DecodeMetricDefinition(httpResp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while parsing the resource creation response. JSON Error: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(populateModel(ctx, &plan, created)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *metricDefinitionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MetricDefinitionModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	m, err := r.getMetricDefinition(ctx, data.Id.ValueString())
	if err != nil {
		if bronto.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Read Resource",
			"An unexpected error occurred while calling the read API. Error: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(populateModel(ctx, &data, m)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update is never expected to run a real change through: every configurable
// attribute carries a RequiresReplace plan modifier because the Metrics API
// has no update endpoint for a metric definition (only create/get/list/
// delete). It's implemented defensively anyway, re-reading the resource
// rather than assuming nothing changed.
func (r *metricDefinitionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan MetricDefinitionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state MetricDefinitionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.Id = state.Id

	m, err := r.getMetricDefinition(ctx, state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while reading back the metric definition. Error: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(populateModel(ctx, &plan, m)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *metricDefinitionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data MetricDefinitionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteRequest(ctx, "/metrics/definitions/"+data.Id.ValueString())
	if err != nil && !bronto.IsNotFound(err) {
		resp.Diagnostics.AddError(
			"Unable to Delete Resource",
			"An unexpected error occurred while calling the delete API. Error: "+err.Error(),
		)
		return
	}
}

func (r *metricDefinitionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *metricDefinitionResource) getMetricDefinition(ctx context.Context, id string) (bronto.MetricDefinition, error) {
	httpResp, err := r.client.ReadRequest(ctx, "/metrics/definitions/"+id)
	if err != nil {
		return bronto.MetricDefinition{}, err
	}
	return bronto.DecodeMetricDefinition(httpResp)
}

func encodeMetricDefinitionRequest(ctx context.Context, data MetricDefinitionModel) ([]byte, diag.Diagnostics) {
	var diags diag.Diagnostics

	queries, queryDiags := extractQueries(ctx, data.Queries)
	diags.Append(queryDiags...)

	formulas, formulaDiags := extractFormulas(ctx, data.Formulas)
	diags.Append(formulaDiags...)

	if diags.HasError() {
		return nil, diags
	}

	reqBody := bronto.CreateMetricDefinitionRequest{
		Queries:  queries,
		Formulas: formulas,
	}
	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		reqBody.Name = data.Name.ValueString()
	}
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		reqBody.Description = data.Description.ValueString()
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		diags.AddError(
			"Unable to Encode Request",
			"An unexpected error occurred while encoding the metric definition request. JSON Error: "+err.Error(),
		)
		return nil, diags
	}

	return body, diags
}

func extractQueries(ctx context.Context, queries []QueryModel) ([]bronto.MetricQuery, diag.Diagnostics) {
	var diags diag.Diagnostics

	result := make([]bronto.MetricQuery, 0, len(queries))
	for _, q := range queries {
		selectVals, d := extractStringList(ctx, q.Select)
		diags.Append(d...)

		fromVals, d := extractStringList(ctx, q.From)
		diags.Append(d...)

		groupVals, d := extractStringList(ctx, q.Groups)
		diags.Append(d...)

		mq := bronto.MetricQuery{
			Name:   q.Name.ValueString(),
			Select: selectVals,
			From:   fromVals,
			Where:  q.Where.ValueString(),
			Groups: groupVals,
		}

		if !q.FromExpr.IsNull() && !q.FromExpr.IsUnknown() {
			mq.FromExpr = q.FromExpr.ValueString()
		}
		if !q.UnitConfig.IsNull() && !q.UnitConfig.IsUnknown() {
			mq.UnitConfig = json.RawMessage(q.UnitConfig.ValueString())
		}

		if q.Aggregation != nil {
			tfp, d := extractStringList(ctx, q.Aggregation.TimeFunctionParams)
			diags.Append(d...)

			agg := &bronto.Aggregation{
				Time:               q.Aggregation.Time.ValueString(),
				TimeFunctionParams: tfp,
			}
			if !q.Aggregation.ReduceTo.IsNull() && !q.Aggregation.ReduceTo.IsUnknown() {
				agg.ReduceTo = q.Aggregation.ReduceTo.ValueString()
			}
			mq.Aggregation = agg
		}

		result = append(result, mq)
	}

	return result, diags
}

func extractFormulas(ctx context.Context, list types.List) ([]bronto.MetricFormula, diag.Diagnostics) {
	var diags diag.Diagnostics

	if list.IsNull() || list.IsUnknown() {
		return nil, diags
	}

	var formulas []FormulaModel
	diags.Append(list.ElementsAs(ctx, &formulas, false)...)
	if diags.HasError() {
		return nil, diags
	}

	result := make([]bronto.MetricFormula, 0, len(formulas))
	for _, f := range formulas {
		mf := bronto.MetricFormula{
			Name:       f.Name.ValueString(),
			Expression: f.Expression.ValueString(),
		}

		if !f.UnitConfig.IsNull() && !f.UnitConfig.IsUnknown() {
			mf.UnitConfig = json.RawMessage(f.UnitConfig.ValueString())
		}

		if !f.ReduceTo.IsNull() && !f.ReduceTo.IsUnknown() {
			mf.Aggregation = &bronto.FormulaAggregation{ReduceTo: f.ReduceTo.ValueString()}
		}

		result = append(result, mf)
	}

	return result, diags
}

// populateModel maps an API metric definition response onto the Terraform
// model, including the id, so it can be used from Create, Read, and Update.
func populateModel(ctx context.Context, data *MetricDefinitionModel, m bronto.MetricDefinition) diag.Diagnostics {
	var diags diag.Diagnostics

	data.Id = types.StringValue(m.Id)

	// The API can't distinguish "not set" from "" for these fields; treat
	// an empty response value as unset to avoid permanent drift against a
	// null config.
	if m.Name == "" {
		data.Name = types.StringNull()
	} else {
		data.Name = types.StringValue(m.Name)
	}

	if m.Description == "" {
		data.Description = types.StringNull()
	} else {
		data.Description = types.StringValue(m.Description)
	}

	queries, queryDiags := convertQueries(ctx, m.Queries)
	diags.Append(queryDiags...)
	data.Queries = queries

	formulas, formulaDiags := convertFormulas(ctx, m.Formulas)
	diags.Append(formulaDiags...)
	data.Formulas = formulas

	return diags
}

func convertQueries(ctx context.Context, queries []bronto.MetricQuery) ([]QueryModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	result := make([]QueryModel, 0, len(queries))
	for _, q := range queries {
		selectVal, d := stringListValue(ctx, q.Select)
		diags.Append(d...)

		fromVal, d := stringListValue(ctx, q.From)
		diags.Append(d...)

		groupsVal, d := stringListValue(ctx, q.Groups)
		diags.Append(d...)

		qm := QueryModel{
			Name:   types.StringValue(q.Name),
			Select: selectVal,
			From:   fromVal,
			Where:  types.StringValue(q.Where),
			Groups: groupsVal,
		}

		if q.FromExpr == "" {
			qm.FromExpr = types.StringNull()
		} else {
			qm.FromExpr = types.StringValue(q.FromExpr)
		}

		if len(q.UnitConfig) == 0 {
			qm.UnitConfig = jsontypes.NewNormalizedNull()
		} else {
			qm.UnitConfig = jsontypes.NewNormalizedValue(string(q.UnitConfig))
		}

		if q.Aggregation != nil {
			tfpVal, d := stringListValue(ctx, q.Aggregation.TimeFunctionParams)
			diags.Append(d...)

			reduceTo := types.StringNull()
			if q.Aggregation.ReduceTo != "" {
				reduceTo = types.StringValue(q.Aggregation.ReduceTo)
			}

			qm.Aggregation = &AggregationModel{
				Time:               types.StringValue(q.Aggregation.Time),
				ReduceTo:           reduceTo,
				TimeFunctionParams: tfpVal,
			}
		}

		result = append(result, qm)
	}

	return result, diags
}

func convertFormulas(ctx context.Context, formulas []bronto.MetricFormula) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	result := make([]FormulaModel, 0, len(formulas))
	for _, f := range formulas {
		fm := FormulaModel{
			Name:       types.StringValue(f.Name),
			Expression: types.StringValue(f.Expression),
			ReduceTo:   types.StringNull(),
		}

		if f.Aggregation != nil && f.Aggregation.ReduceTo != "" {
			fm.ReduceTo = types.StringValue(f.Aggregation.ReduceTo)
		}

		if len(f.UnitConfig) == 0 {
			fm.UnitConfig = jsontypes.NewNormalizedNull()
		} else {
			fm.UnitConfig = jsontypes.NewNormalizedValue(string(f.UnitConfig))
		}

		result = append(result, fm)
	}

	listVal, listDiags := types.ListValueFrom(ctx, formulaObjectType, result)
	diags.Append(listDiags...)
	return listVal, diags
}

func extractStringList(ctx context.Context, list types.List) ([]string, diag.Diagnostics) {
	var diags diag.Diagnostics

	if list.IsNull() || list.IsUnknown() {
		return nil, diags
	}

	var values []string
	diags.Append(list.ElementsAs(ctx, &values, false)...)
	return values, diags
}

// stringListValue converts a []string into a types.List, normalizing a nil
// slice to an empty (but non-null) list first. types.ListValueFrom on its
// own maps a nil slice to a null list, which would make an Optional+Computed
// attribute flip between null and an empty list depending on whether the API
// responds with an absent field or an empty array for the same "no values"
// state - exactly the kind of inconsistency that broke login_methods on the
// user resource.
func stringListValue(ctx context.Context, values []string) (types.List, diag.Diagnostics) {
	if values == nil {
		values = []string{}
	}
	return types.ListValueFrom(ctx, types.StringType, values)
}
