package resource_widget

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
	_ resource.Resource                = &widgetResource{}
	_ resource.ResourceWithConfigure   = &widgetResource{}
	_ resource.ResourceWithImportState = &widgetResource{}
)

func NewWidgetResource() resource.Resource {
	return &widgetResource{}
}

type widgetResource struct {
	client *bronto.Client
}

func (r *widgetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_widget"
}

func (r *widgetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = WidgetResourceSchema(ctx)
}

func (r *widgetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *widgetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WidgetModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, diags := encodeWidgetRequest(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.CreateRequest(ctx, body, "/widgets")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while calling the create API. Error: "+err.Error(),
		)
		return
	}

	created, err := bronto.DecodeWidget(httpResp)
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

func (r *widgetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data WidgetModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.ReadRequest(ctx, "/widgets/"+data.Id.ValueString())
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

	w, err := bronto.DecodeWidget(httpResp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Resource",
			"An unexpected error occurred while parsing the resource read response. JSON Error: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(populateModel(ctx, &data, w)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *widgetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan WidgetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state WidgetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	widgetId := state.Id.ValueString()

	body, diags := encodeWidgetRequest(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// PUT fully replaces the widget but responds 204 No Content, so the
	// updated resource has to be fetched separately afterward.
	err := r.client.PutRequest(ctx, body, "/widgets/"+widgetId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while calling the update API. Error: "+err.Error(),
		)
		return
	}

	httpResp, err := r.client.ReadRequest(ctx, "/widgets/"+widgetId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while reading back the updated resource. Error: "+err.Error(),
		)
		return
	}

	updated, err := bronto.DecodeWidget(httpResp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while parsing the resource update response. JSON Error: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(populateModel(ctx, &plan, updated)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *widgetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data WidgetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteRequest(ctx, "/widgets/"+data.Id.ValueString())
	if err != nil && !bronto.IsNotFound(err) {
		resp.Diagnostics.AddError(
			"Unable to Delete Resource",
			"An unexpected error occurred while calling the delete API. Error: "+err.Error(),
		)
		return
	}
}

func (r *widgetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func encodeWidgetRequest(ctx context.Context, data WidgetModel) ([]byte, diag.Diagnostics) {
	var diags diag.Diagnostics

	metricIds, metricDiags := extractStringSet(ctx, data.MetricIds)
	diags.Append(metricDiags...)

	widgetIds, widgetDiags := extractStringSet(ctx, data.WidgetIds)
	diags.Append(widgetDiags...)

	if diags.HasError() {
		return nil, diags
	}

	reqBody := bronto.WidgetRequest{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		Type:        data.Type.ValueString(),
		MetricIds:   metricIds,
		WidgetIds:   widgetIds,
	}
	if !data.Aux.IsNull() && !data.Aux.IsUnknown() {
		reqBody.Aux = json.RawMessage(data.Aux.ValueString())
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		diags.AddError(
			"Unable to Encode Request",
			"An unexpected error occurred while encoding the widget request. JSON Error: "+err.Error(),
		)
		return nil, diags
	}

	return body, diags
}

// populateModel maps an API widget response onto the Terraform model,
// including the id, so it can be used from Create, Read, and Update.
func populateModel(ctx context.Context, data *WidgetModel, w bronto.Widget) diag.Diagnostics {
	var diags diag.Diagnostics

	data.Id = types.StringValue(w.Id)
	data.Name = types.StringValue(w.Name)
	data.Type = types.StringValue(w.Type)

	// The API can't distinguish "not set" from "" for these fields; treat
	// an empty response value as unset to avoid permanent drift against a
	// null config.
	if w.Description == "" {
		data.Description = types.StringNull()
	} else {
		data.Description = types.StringValue(w.Description)
	}

	if len(w.Aux) == 0 {
		data.Aux = jsontypes.NewNormalizedNull()
	} else {
		data.Aux = jsontypes.NewNormalizedValue(string(w.Aux))
	}

	metricIdsVal, metricDiags := stringSetValue(ctx, w.MetricIds)
	diags.Append(metricDiags...)
	data.MetricIds = metricIdsVal

	widgetIdsVal, widgetDiags := stringSetValue(ctx, w.WidgetIds)
	diags.Append(widgetDiags...)
	data.WidgetIds = widgetIdsVal

	tagsVal, tagsDiags := types.MapValueFrom(ctx, types.StringType, w.Tags)
	diags.Append(tagsDiags...)
	data.Tags = tagsVal

	return diags
}

func extractStringSet(ctx context.Context, set types.Set) ([]string, diag.Diagnostics) {
	var diags diag.Diagnostics

	if set.IsNull() || set.IsUnknown() {
		return nil, diags
	}

	var values []string
	diags.Append(set.ElementsAs(ctx, &values, false)...)
	return values, diags
}

// stringSetValue converts a []string into a types.Set, normalizing a nil
// slice to an empty (but non-null) set first. types.SetValueFrom on its own
// maps a nil slice to a null set, which would make this Optional+Computed
// attribute flip between null and an empty set depending on whether the API
// responds with an absent field or an empty array for the same "no values"
// state - exactly the kind of inconsistency that broke login_methods on the
// user resource.
func stringSetValue(ctx context.Context, values []string) (types.Set, diag.Diagnostics) {
	if values == nil {
		values = []string{}
	}
	return types.SetValueFrom(ctx, types.StringType, values)
}
