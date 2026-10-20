package resource_dashboard

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	bronto "terraform-provider-bronto/internal"
)

var (
	_ resource.Resource                = &dashboardResource{}
	_ resource.ResourceWithConfigure   = &dashboardResource{}
	_ resource.ResourceWithImportState = &dashboardResource{}
)

func NewDashboardResource() resource.Resource {
	return &dashboardResource{}
}

type dashboardResource struct {
	client *bronto.Client
}

func (r *dashboardResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dashboard"
}

func (r *dashboardResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = DashboardResourceSchema(ctx)
}

func (r *dashboardResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *dashboardResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DashboardModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	layoutMetricIds, diags := extractStringSet(ctx, plan.LayoutMetricIds)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := bronto.CreateDashboardRequest{Name: plan.Name.ValueString()}
	if len(layoutMetricIds) > 0 {
		createReq.Layout = &bronto.DashboardLayout{MetricIds: layoutMetricIds}
	}

	body, err := json.Marshal(createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while encoding the create request. JSON Error: "+err.Error(),
		)
		return
	}

	httpResp, err := r.client.CreateRequest(ctx, body, "/dashboards")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while calling the create API. Error: "+err.Error(),
		)
		return
	}

	created, err := bronto.DecodeDashboard(httpResp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while parsing the resource creation response. JSON Error: "+err.Error(),
		)
		return
	}

	// The create endpoint doesn't accept widget_ids, so attach any
	// configured widgets with a follow-up append call.
	widgetIds, diags := extractStringSet(ctx, plan.WidgetIds)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(widgetIds) > 0 {
		if err := r.appendWidgets(ctx, created.Id, widgetIds); err != nil {
			resp.Diagnostics.AddError(
				"Unable to Create Resource",
				"An unexpected error occurred while attaching widgets to the new dashboard. Error: "+err.Error(),
			)
			return
		}
	}

	// The create endpoint doesn't accept tags either; the only way to set
	// them is one /tags API call per key.
	desiredTags, diags := extractStringMap(ctx, plan.Tags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(desiredTags) > 0 {
		if err := r.client.SyncTags(ctx, "dashboard", created.Id, nil, desiredTags); err != nil {
			resp.Diagnostics.AddError(
				"Unable to Create Resource",
				"An unexpected error occurred while applying tags to the new dashboard. Error: "+err.Error(),
			)
			return
		}
	}

	if len(widgetIds) > 0 || len(desiredTags) > 0 {
		created, err = r.getDashboard(ctx, created.Id)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Create Resource",
				"An unexpected error occurred while reading back the dashboard after attaching widgets and tags. Error: "+err.Error(),
			)
			return
		}
	}

	resp.Diagnostics.Append(populateModel(ctx, &plan, created)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dashboardResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DashboardModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	d, err := r.getDashboard(ctx, data.Id.ValueString())
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

	resp.Diagnostics.Append(populateModel(ctx, &data, d)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *dashboardResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DashboardModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state DashboardModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dashboardId := state.Id.ValueString()

	// PATCH only accepts one of {name} or {layout} per call, so issue up
	// to two separate requests rather than trying to merge them.
	if !plan.Name.Equal(state.Name) {
		body, err := json.Marshal(bronto.PatchDashboardNameRequest{Name: plan.Name.ValueString()})
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Update Resource",
				"An unexpected error occurred while encoding the name update request. JSON Error: "+err.Error(),
			)
			return
		}
		if _, err := r.client.UpdateRequest(ctx, body, "/dashboards/"+dashboardId); err != nil {
			resp.Diagnostics.AddError(
				"Unable to Update Resource",
				"An unexpected error occurred while updating the dashboard name. Error: "+err.Error(),
			)
			return
		}
	}

	if !plan.LayoutMetricIds.Equal(state.LayoutMetricIds) {
		layoutMetricIds, diags := extractStringSet(ctx, plan.LayoutMetricIds)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		body, err := json.Marshal(bronto.PatchDashboardLayoutRequest{
			Layout: bronto.DashboardLayout{MetricIds: layoutMetricIds},
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Update Resource",
				"An unexpected error occurred while encoding the layout update request. JSON Error: "+err.Error(),
			)
			return
		}
		if _, err := r.client.UpdateRequest(ctx, body, "/dashboards/"+dashboardId); err != nil {
			resp.Diagnostics.AddError(
				"Unable to Update Resource",
				"An unexpected error occurred while updating the dashboard layout. Error: "+err.Error(),
			)
			return
		}
	}

	if !plan.WidgetIds.Equal(state.WidgetIds) {
		currentWidgetIds, diags := extractStringSet(ctx, state.WidgetIds)
		resp.Diagnostics.Append(diags...)
		desiredWidgetIds, diags2 := extractStringSet(ctx, plan.WidgetIds)
		resp.Diagnostics.Append(diags2...)
		if resp.Diagnostics.HasError() {
			return
		}

		toAdd, toRemove := diffStringSets(currentWidgetIds, desiredWidgetIds)

		if len(toAdd) > 0 {
			if err := r.appendWidgets(ctx, dashboardId, toAdd); err != nil {
				resp.Diagnostics.AddError(
					"Unable to Update Resource",
					"An unexpected error occurred while attaching widgets to the dashboard. Error: "+err.Error(),
				)
				return
			}
		}

		for _, widgetId := range toRemove {
			if err := r.client.DeleteRequest(ctx, "/dashboards/"+dashboardId+"/widgets/"+widgetId); err != nil && !bronto.IsNotFound(err) {
				resp.Diagnostics.AddError(
					"Unable to Update Resource",
					fmt.Sprintf("An unexpected error occurred while removing widget %q from the dashboard. Error: %s", widgetId, err.Error()),
				)
				return
			}
		}
	}

	if !plan.Tags.Equal(state.Tags) {
		currentTags, diags := extractStringMap(ctx, state.Tags)
		resp.Diagnostics.Append(diags...)
		desiredTags, diags2 := extractStringMap(ctx, plan.Tags)
		resp.Diagnostics.Append(diags2...)
		if resp.Diagnostics.HasError() {
			return
		}

		if err := r.client.SyncTags(ctx, "dashboard", dashboardId, currentTags, desiredTags); err != nil {
			resp.Diagnostics.AddError(
				"Unable to Update Resource",
				"An unexpected error occurred while syncing dashboard tags. Error: "+err.Error(),
			)
			return
		}
	}

	updated, err := r.getDashboard(ctx, dashboardId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while reading back the updated dashboard. Error: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(populateModel(ctx, &plan, updated)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dashboardResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DashboardModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteRequest(ctx, "/dashboards/"+data.Id.ValueString())
	if err != nil && !bronto.IsNotFound(err) {
		resp.Diagnostics.AddError(
			"Unable to Delete Resource",
			"An unexpected error occurred while calling the delete API. Error: "+err.Error(),
		)
		return
	}
}

func (r *dashboardResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *dashboardResource) getDashboard(ctx context.Context, id string) (bronto.Dashboard, error) {
	httpResp, err := r.client.ReadRequest(ctx, "/dashboards/"+id)
	if err != nil {
		return bronto.Dashboard{}, err
	}
	return bronto.DecodeDashboard(httpResp)
}

func (r *dashboardResource) appendWidgets(ctx context.Context, dashboardId string, widgetIds []string) error {
	body, err := json.Marshal(bronto.AppendDashboardWidgetsRequest{WidgetIds: widgetIds})
	if err != nil {
		return err
	}
	return r.client.PostNoContentRequest(ctx, body, "/dashboards/"+dashboardId+"/widgets")
}

// populateModel maps an API dashboard response onto the Terraform model,
// including the id, so it can be used from Create, Read, and Update.
func populateModel(ctx context.Context, data *DashboardModel, d bronto.Dashboard) diag.Diagnostics {
	var diags diag.Diagnostics

	data.Id = types.StringValue(d.Id)
	data.Name = types.StringValue(d.Name)

	if d.Description == "" {
		data.Description = types.StringNull()
	} else {
		data.Description = types.StringValue(d.Description)
	}

	widgetIds := make([]string, 0, len(d.Widgets))
	for _, w := range d.Widgets {
		widgetIds = append(widgetIds, w.Id)
	}
	widgetIdsVal, widgetDiags := stringSetValue(ctx, widgetIds)
	diags.Append(widgetDiags...)
	data.WidgetIds = widgetIdsVal

	var layoutMetricIds []string
	if d.Layout != nil {
		layoutMetricIds = d.Layout.MetricIds
	}
	layoutVal, layoutDiags := stringSetValue(ctx, layoutMetricIds)
	diags.Append(layoutDiags...)
	data.LayoutMetricIds = layoutVal

	tagsVal, tagsDiags := types.MapValueFrom(ctx, types.StringType, d.Tags)
	diags.Append(tagsDiags...)
	data.Tags = tagsVal

	return diags
}

func extractStringMap(ctx context.Context, m types.Map) (map[string]string, diag.Diagnostics) {
	var diags diag.Diagnostics

	if m.IsNull() || m.IsUnknown() {
		return nil, diags
	}

	var values map[string]string
	diags.Append(m.ElementsAs(ctx, &values, false)...)
	return values, diags
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
// maps a nil slice to a null set, which would make these Optional+Computed
// attributes flip between null and an empty set depending on whether the
// API responds with an absent field or an empty array for the same "no
// values" state - exactly the kind of inconsistency that broke
// login_methods on the user resource.
func stringSetValue(ctx context.Context, values []string) (types.Set, diag.Diagnostics) {
	if values == nil {
		values = []string{}
	}
	return types.SetValueFrom(ctx, types.StringType, values)
}
