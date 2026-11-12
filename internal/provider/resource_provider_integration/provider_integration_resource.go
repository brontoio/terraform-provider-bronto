package resource_provider_integration

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	bronto "terraform-provider-bronto/internal"
	tfutil "terraform-provider-bronto/internal/provider/tfutil"
)

var (
	_ resource.Resource                = &providerIntegrationResource{}
	_ resource.ResourceWithConfigure   = &providerIntegrationResource{}
	_ resource.ResourceWithImportState = &providerIntegrationResource{}
)

func NewProviderIntegrationResource() resource.Resource {
	return &providerIntegrationResource{}
}

type providerIntegrationResource struct {
	client *bronto.Client
}

func (r *providerIntegrationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_provider_integration"
}

func (r *providerIntegrationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ProviderIntegrationResourceSchema(ctx)
}

func (r *providerIntegrationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *providerIntegrationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ProviderIntegrationModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.Secret.IsNull() || plan.Secret.ValueString() == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("secret"),
			"Missing Secret",
			"secret is required when creating a provider integration - the API has no way to authenticate with the downstream provider without one.",
		)
		return
	}

	reqBody := bronto.CreateProviderIntegrationRequest{
		Provider: plan.ProviderType.ValueString(),
		Url:      plan.Url.ValueString(),
		Secret:   plan.Secret.ValueString(),
		AuthType: plan.AuthType.ValueString(),
	}
	if !plan.Name.IsNull() && !plan.Name.IsUnknown() {
		reqBody.Name = plan.Name.ValueString()
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Encode Request",
			"An unexpected error occurred while encoding the provider integration create request. JSON Error: "+err.Error(),
		)
		return
	}

	httpResp, err := r.client.CreateRequest(ctx, body, "/integrations/providers")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while calling the create API. Error: "+err.Error(),
		)
		return
	}

	created, err := bronto.DecodeProviderIntegration(httpResp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while parsing the resource creation response. JSON Error: "+err.Error(),
		)
		return
	}

	// The create endpoint doesn't accept tags; the only way to set them is
	// one /tags API call per key.
	desiredTags, diags := tfutil.DesiredTags(ctx, plan.Tags, r.client.DefaultTags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(desiredTags) > 0 {
		if err := r.client.SyncTags(ctx, "integration", created.Id, nil, desiredTags); err != nil {
			resp.Diagnostics.AddError(
				"Unable to Create Resource",
				"An unexpected error occurred while applying tags to the new provider integration. Error: "+err.Error(),
			)
			return
		}

		created, err = r.getProviderIntegration(ctx, created.Id)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Create Resource",
				"An unexpected error occurred while reading back the provider integration after applying tags. Error: "+err.Error(),
			)
			return
		}
	}

	populateModel(&plan, created)
	resp.Diagnostics.Append(tfutil.PopulateTags(ctx, &plan.Tags, &plan.TagsAll, plan.Tags, created.Tags, r.client.DefaultTags)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *providerIntegrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProviderIntegrationModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found, err := r.getProviderIntegration(ctx, data.Id.ValueString())
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

	// populateModel never touches data.Secret: the API never returns it, so
	// whatever's already in state is left exactly as-is.
	populateModel(&data, found)
	resp.Diagnostics.Append(tfutil.PopulateTags(ctx, &data.Tags, &data.TagsAll, types.MapNull(types.StringType), found.Tags, r.client.DefaultTags)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *providerIntegrationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ProviderIntegrationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ProviderIntegrationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	reqBody := bronto.UpdateProviderIntegrationRequest{
		Url:      plan.Url.ValueString(),
		AuthType: plan.AuthType.ValueString(),
	}
	if !plan.Name.IsNull() && !plan.Name.IsUnknown() {
		reqBody.Name = plan.Name.ValueString()
	}
	// Only send a new secret when the config actually changed it - omitting
	// it (via the json "omitempty" tag) tells the API to keep the one
	// already stored, per the update endpoint's documented behavior.
	if !plan.Secret.Equal(state.Secret) && !plan.Secret.IsNull() && !plan.Secret.IsUnknown() {
		reqBody.Secret = plan.Secret.ValueString()
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Encode Request",
			"An unexpected error occurred while encoding the provider integration update request. JSON Error: "+err.Error(),
		)
		return
	}

	httpResp, err := r.client.PutRequestWithBody(ctx, body, "/integrations/providers/"+state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while calling the update API. Error: "+err.Error(),
		)
		return
	}

	updated, err := bronto.DecodeProviderIntegration(httpResp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while parsing the resource update response. JSON Error: "+err.Error(),
		)
		return
	}

	// Always reconcile tags, even if plan.Tags == state.Tags - see
	// tfutil.PopulateTags for why currentTags must come from state.TagsAll.
	currentTags, diags := tfutil.ExtractStringMap(ctx, state.TagsAll)
	resp.Diagnostics.Append(diags...)
	desiredTags, diags2 := tfutil.DesiredTags(ctx, plan.Tags, r.client.DefaultTags)
	resp.Diagnostics.Append(diags2...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.SyncTags(ctx, "integration", updated.Id, currentTags, desiredTags); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while syncing provider integration tags. Error: "+err.Error(),
		)
		return
	}

	updated, err = r.getProviderIntegration(ctx, updated.Id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while reading back the provider integration after syncing tags. Error: "+err.Error(),
		)
		return
	}

	populateModel(&plan, updated)
	resp.Diagnostics.Append(tfutil.PopulateTags(ctx, &plan.Tags, &plan.TagsAll, plan.Tags, updated.Tags, r.client.DefaultTags)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *providerIntegrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ProviderIntegrationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteRequest(ctx, "/integrations/providers/"+data.Id.ValueString())
	if err != nil && !bronto.IsNotFound(err) {
		resp.Diagnostics.AddError(
			"Unable to Delete Resource",
			"An unexpected error occurred while calling the delete API. Error: "+err.Error(),
		)
		return
	}
}

func (r *providerIntegrationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *providerIntegrationResource) getProviderIntegration(ctx context.Context, id string) (bronto.ProviderIntegration, error) {
	httpResp, err := r.client.ReadRequest(ctx, "/integrations/providers/"+id)
	if err != nil {
		return bronto.ProviderIntegration{}, err
	}
	return bronto.DecodeProviderIntegration(httpResp)
}

// populateModel maps an API provider integration response onto the
// Terraform model, including the id, so it can be used from Create, Read,
// and Update. It deliberately never touches Secret: the API never returns
// it, so the caller's existing value (from state or plan) is preserved.
func populateModel(data *ProviderIntegrationModel, p bronto.ProviderIntegration) {
	data.Id = types.StringValue(p.Id)
	data.Name = types.StringValue(p.Name)
	data.ProviderType = types.StringValue(p.Provider)
	data.Url = types.StringValue(p.Url)
	data.HasSecret = types.BoolValue(p.HasSecret)
	data.AuthType = types.StringValue(p.AuthType)
}
