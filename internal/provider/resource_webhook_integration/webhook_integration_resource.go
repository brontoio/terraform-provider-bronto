package resource_webhook_integration

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	bronto "terraform-provider-bronto/internal"
	tfutil "terraform-provider-bronto/internal/provider/tfutil"
)

var (
	_ resource.Resource                = &webhookIntegrationResource{}
	_ resource.ResourceWithConfigure   = &webhookIntegrationResource{}
	_ resource.ResourceWithImportState = &webhookIntegrationResource{}
)

func NewWebhookIntegrationResource() resource.Resource {
	return &webhookIntegrationResource{}
}

type webhookIntegrationResource struct {
	client *bronto.Client
}

func (r *webhookIntegrationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook_integration"
}

func (r *webhookIntegrationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = WebhookIntegrationResourceSchema(ctx)
}

func (r *webhookIntegrationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *webhookIntegrationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WebhookIntegrationModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, diags := encodeWebhookIntegrationRequest(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.CreateRequest(ctx, body, "/integrations/webhooks")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while calling the create API. Error: "+err.Error(),
		)
		return
	}

	created, err := bronto.DecodeWebhookIntegration(httpResp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while parsing the resource creation response. JSON Error: "+err.Error(),
		)
		return
	}

	// The create endpoint doesn't accept tags; the only way to set them is
	// one /tags API call per key.
	desiredTags, diags2 := tfutil.DesiredTags(ctx, plan.Tags, r.client.DefaultTags)
	resp.Diagnostics.Append(diags2...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(desiredTags) > 0 {
		if err := r.client.SyncTags(ctx, "integration", created.Id, nil, desiredTags); err != nil {
			resp.Diagnostics.AddError(
				"Unable to Create Resource",
				"An unexpected error occurred while applying tags to the new webhook integration. Error: "+err.Error(),
			)
			return
		}

		created, err = r.getWebhookIntegration(ctx, created.Id)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Create Resource",
				"An unexpected error occurred while reading back the webhook integration after applying tags. Error: "+err.Error(),
			)
			return
		}
	}

	resp.Diagnostics.Append(populateModel(ctx, &plan, created)...)
	resp.Diagnostics.Append(tfutil.PopulateTags(ctx, &plan.Tags, &plan.TagsAll, plan.Tags, created.Tags, r.client.DefaultTags)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *webhookIntegrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data WebhookIntegrationModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.ReadRequest(ctx, "/integrations/webhooks/"+data.Id.ValueString())
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

	found, err := bronto.DecodeWebhookIntegration(httpResp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Resource",
			"An unexpected error occurred while parsing the resource read response. JSON Error: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(populateModel(ctx, &data, found)...)
	resp.Diagnostics.Append(tfutil.PopulateTags(ctx, &data.Tags, &data.TagsAll, types.MapNull(types.StringType), found.Tags, r.client.DefaultTags)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *webhookIntegrationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan WebhookIntegrationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state WebhookIntegrationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, diags := encodeWebhookIntegrationRequest(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// PUT fully replaces the integration and, like bronto_slack_integration,
	// responds 200 with the updated resource rather than 204.
	httpResp, err := r.client.PutRequestWithBody(ctx, body, "/integrations/webhooks/"+state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while calling the update API. Error: "+err.Error(),
		)
		return
	}

	updated, err := bronto.DecodeWebhookIntegration(httpResp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while parsing the resource update response. JSON Error: "+err.Error(),
		)
		return
	}

	// Always reconcile tags, even if plan.Tags == state.Tags: SyncTags is a
	// no-op (zero API calls) when nothing's actually changed, but this way
	// a default_tags change on the provider also gets picked up whenever
	// this resource is updated for any other reason. currentTags MUST come
	// from state.TagsAll, not state.Tags - see tfutil.PopulateTags.
	currentTags, diags2 := tfutil.ExtractStringMap(ctx, state.TagsAll)
	resp.Diagnostics.Append(diags2...)
	desiredTags, diags3 := tfutil.DesiredTags(ctx, plan.Tags, r.client.DefaultTags)
	resp.Diagnostics.Append(diags3...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.SyncTags(ctx, "integration", updated.Id, currentTags, desiredTags); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while syncing webhook integration tags. Error: "+err.Error(),
		)
		return
	}

	// The PUT response predates the tag sync above, so read back the tags
	// actually applied.
	updated, err = r.getWebhookIntegration(ctx, updated.Id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while reading back the webhook integration after syncing tags. Error: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(populateModel(ctx, &plan, updated)...)
	resp.Diagnostics.Append(tfutil.PopulateTags(ctx, &plan.Tags, &plan.TagsAll, plan.Tags, updated.Tags, r.client.DefaultTags)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *webhookIntegrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data WebhookIntegrationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteRequest(ctx, "/integrations/webhooks/"+data.Id.ValueString())
	if err != nil && !bronto.IsNotFound(err) {
		resp.Diagnostics.AddError(
			"Unable to Delete Resource",
			"An unexpected error occurred while calling the delete API. Error: "+err.Error(),
		)
		return
	}
}

func (r *webhookIntegrationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func encodeWebhookIntegrationRequest(ctx context.Context, data WebhookIntegrationModel) ([]byte, diag.Diagnostics) {
	var diags diag.Diagnostics

	customHeaders, headerDiags := tfutil.ExtractStringMap(ctx, data.CustomHeaders)
	diags.Append(headerDiags...)
	if diags.HasError() {
		return nil, diags
	}

	reqBody := bronto.WebhookIntegrationRequest{
		Name:          data.Name.ValueString(),
		Url:           data.Url.ValueString(),
		CustomHeaders: customHeaders,
	}
	if !data.Payload.IsNull() && !data.Payload.IsUnknown() {
		reqBody.Payload = data.Payload.ValueString()
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		diags.AddError(
			"Unable to Encode Request",
			"An unexpected error occurred while encoding the webhook integration request. JSON Error: "+err.Error(),
		)
		return nil, diags
	}

	return body, diags
}

// populateModel maps an API webhook integration response onto the
// Terraform model, including the id, so it can be used from Create, Read,
// and Update.
func populateModel(ctx context.Context, data *WebhookIntegrationModel, w bronto.WebhookIntegration) diag.Diagnostics {
	var diags diag.Diagnostics

	data.Id = types.StringValue(w.Id)
	data.Name = types.StringValue(w.Name)
	data.Url = types.StringValue(w.Url)

	if w.Payload == "" {
		data.Payload = types.StringNull()
	} else {
		data.Payload = types.StringValue(w.Payload)
	}

	headersVal, headerDiags := tfutil.StringMapValue(ctx, w.CustomHeaders)
	diags.Append(headerDiags...)
	data.CustomHeaders = headersVal

	return diags
}

func (r *webhookIntegrationResource) getWebhookIntegration(ctx context.Context, id string) (bronto.WebhookIntegration, error) {
	httpResp, err := r.client.ReadRequest(ctx, "/integrations/webhooks/"+id)
	if err != nil {
		return bronto.WebhookIntegration{}, err
	}
	return bronto.DecodeWebhookIntegration(httpResp)
}
