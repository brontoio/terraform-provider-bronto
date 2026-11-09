package resource_sso_mappings

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

// singletonId is the constant id for this resource: the API has one
// org-wide list of SSO mappings, not individually addressable resources.
const singletonId = "sso_mappings"

var (
	_ resource.Resource                = &ssoMappingsResource{}
	_ resource.ResourceWithConfigure   = &ssoMappingsResource{}
	_ resource.ResourceWithImportState = &ssoMappingsResource{}
)

func NewSsoMappingsResource() resource.Resource {
	return &ssoMappingsResource{}
}

type ssoMappingsResource struct {
	client *bronto.Client
}

func (r *ssoMappingsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sso_mappings"
}

func (r *ssoMappingsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = SsoMappingsResourceSchema(ctx)
}

func (r *ssoMappingsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ssoMappingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SsoMappingsModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, diags := encodeRequest(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.CreateRequest(ctx, body, "/sso/mappings")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while calling the create API. Error: "+err.Error(),
		)
		return
	}

	created, err := bronto.DecodeSsoMappingsRequest(httpResp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while parsing the resource creation response. JSON Error: "+err.Error(),
		)
		return
	}

	populateModel(&plan, created.Mappings)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ssoMappingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SsoMappingsModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.ReadRequest(ctx, "/sso/mappings")
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

	found, err := bronto.DecodeSsoMappingsRequest(httpResp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Resource",
			"An unexpected error occurred while parsing the resource read response. JSON Error: "+err.Error(),
		)
		return
	}

	populateModel(&data, found.Mappings)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ssoMappingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SsoMappingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, diags := encodeRequest(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// PUT fully replaces the mapping list. The spec doesn't document a
	// response body for this endpoint; assumed to echo the updated list
	// like create does. If it instead responds with an empty body, this
	// will surface as a JSON decode error to fix once verified live.
	httpResp, err := r.client.PutRequestWithBody(ctx, body, "/sso/mappings")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while calling the update API. Error: "+err.Error(),
		)
		return
	}

	updated, err := bronto.DecodeSsoMappingsRequest(httpResp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while parsing the resource update response. JSON Error: "+err.Error(),
		)
		return
	}

	populateModel(&plan, updated.Mappings)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ssoMappingsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// The spec documents no request body or parameters for DELETE
	// /sso/mappings, so this is assumed to clear every mapping for the
	// org - unverified live.
	err := r.client.DeleteRequest(ctx, "/sso/mappings")
	if err != nil && !bronto.IsNotFound(err) {
		resp.Diagnostics.AddError(
			"Unable to Delete Resource",
			"An unexpected error occurred while calling the delete API. Error: "+err.Error(),
		)
		return
	}
}

func (r *ssoMappingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func encodeRequest(data SsoMappingsModel) ([]byte, diag.Diagnostics) {
	var diags diag.Diagnostics

	mappings := make([]bronto.SsoMapping, 0, len(data.Mappings))
	for _, m := range data.Mappings {
		mappings = append(mappings, bronto.SsoMapping{
			ProviderKey:   m.ProviderKey.ValueString(),
			ProviderValue: m.ProviderValue.ValueString(),
			AssignAttr:    m.AssignAttr.ValueString(),
		})
	}

	body, err := json.Marshal(bronto.SsoMappingsRequest{Mappings: mappings})
	if err != nil {
		diags.AddError(
			"Unable to Encode Request",
			"An unexpected error occurred while encoding the SSO mappings request. JSON Error: "+err.Error(),
		)
		return nil, diags
	}

	return body, diags
}

// populateModel maps an API mappings list onto the Terraform model,
// including the constant id, so it can be used from Create, Read, and
// Update.
func populateModel(data *SsoMappingsModel, mappings []bronto.SsoMapping) {
	data.Id = types.StringValue(singletonId)

	models := make([]SsoMappingModel, 0, len(mappings))
	for _, m := range mappings {
		models = append(models, SsoMappingModel{
			ProviderKey:   types.StringValue(m.ProviderKey),
			ProviderValue: types.StringValue(m.ProviderValue),
			AssignAttr:    types.StringValue(m.AssignAttr),
		})
	}
	data.Mappings = models
}
