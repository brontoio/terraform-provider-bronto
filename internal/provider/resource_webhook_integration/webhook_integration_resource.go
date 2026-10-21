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

	resp.Diagnostics.Append(populateModel(ctx, &plan, created)...)
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

	resp.Diagnostics.Append(populateModel(ctx, &plan, updated)...)
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

	customHeaders, headerDiags := extractStringMap(ctx, data.CustomHeaders)
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

	headersVal, headerDiags := stringMapValue(ctx, w.CustomHeaders)
	diags.Append(headerDiags...)
	data.CustomHeaders = headersVal

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

// stringMapValue converts a map[string]string into a types.Map, normalizing
// a nil map to an empty (but non-null) map first. types.MapValueFrom on its
// own maps a nil map to a null map, which would make this Optional+Computed
// attribute flip between null and an empty map depending on whether the API
// responds with an absent field or an empty object for the same "no
// values" state - the same kind of inconsistency that broke login_methods
// on the user resource.
func stringMapValue(ctx context.Context, values map[string]string) (types.Map, diag.Diagnostics) {
	if values == nil {
		values = map[string]string{}
	}
	return types.MapValueFrom(ctx, types.StringType, values)
}
