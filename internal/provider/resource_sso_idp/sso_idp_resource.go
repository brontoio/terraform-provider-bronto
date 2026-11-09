package resource_sso_idp

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
	_ resource.Resource                = &ssoIdpResource{}
	_ resource.ResourceWithConfigure   = &ssoIdpResource{}
	_ resource.ResourceWithImportState = &ssoIdpResource{}
)

func NewSsoIdpResource() resource.Resource {
	return &ssoIdpResource{}
}

type ssoIdpResource struct {
	client *bronto.Client
}

func (r *ssoIdpResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sso_idp"
}

func (r *ssoIdpResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = SsoIdpResourceSchema(ctx)
}

func (r *ssoIdpResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ssoIdpResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SsoIdpModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, diags := encodeCreateRequest(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.CreateRequest(ctx, body, "/sso/idps")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while calling the create API. Error: "+err.Error(),
		)
		return
	}

	created, err := bronto.DecodeIdp(httpResp)
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

func (r *ssoIdpResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SsoIdpModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.ReadRequest(ctx, "/sso/idps/"+data.Id.ValueString())
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

	found, err := bronto.DecodeIdp(httpResp)
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

func (r *ssoIdpResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SsoIdpModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state SsoIdpModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, diags := encodeUpdateRequest(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.UpdateRequest(ctx, body, "/sso/idps/"+state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while calling the update API. Error: "+err.Error(),
		)
		return
	}

	// The spec doesn't document a response body for this endpoint; assumed
	// to echo the updated Idp like every other SSO/Groups PATCH in this
	// API. If it instead responds with an empty body, this will surface as
	// a JSON decode error to fix once verified live.
	updated, err := bronto.DecodeIdp(httpResp)
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

func (r *ssoIdpResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SsoIdpModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteRequest(ctx, "/sso/idps/"+data.Id.ValueString())
	if err != nil && !bronto.IsNotFound(err) {
		resp.Diagnostics.AddError(
			"Unable to Delete Resource",
			"An unexpected error occurred while calling the delete API. Error: "+err.Error(),
		)
		return
	}
}

func (r *ssoIdpResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func encodeCreateRequest(ctx context.Context, data SsoIdpModel) ([]byte, diag.Diagnostics) {
	var diags diag.Diagnostics

	identifiers, identifierDiags := extractStringList(ctx, data.Identifiers)
	diags.Append(identifierDiags...)
	if diags.HasError() {
		return nil, diags
	}

	reqBody := bronto.CreateIdpRequest{
		Name:              data.Name.ValueString(),
		ProviderType:      data.ProviderType.ValueString(),
		Identifiers:       identifiers,
		AttributeMappings: extractAttributeMappings(data.AttributeMappings),
	}
	if !data.MetaCert.IsNull() && !data.MetaCert.IsUnknown() {
		reqBody.MetaCert = data.MetaCert.ValueString()
	}
	if !data.IssuerUrl.IsNull() && !data.IssuerUrl.IsUnknown() {
		reqBody.IssuerUrl = data.IssuerUrl.ValueString()
	}
	if !data.ClientId.IsNull() && !data.ClientId.IsUnknown() {
		reqBody.ClientId = data.ClientId.ValueString()
	}
	if !data.ClientSecret.IsNull() && !data.ClientSecret.IsUnknown() {
		reqBody.ClientSecret = data.ClientSecret.ValueString()
	}
	if !data.AcceptIdpInitiated.IsNull() && !data.AcceptIdpInitiated.IsUnknown() {
		v := data.AcceptIdpInitiated.ValueBool()
		reqBody.AcceptIdpInitiated = &v
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		diags.AddError(
			"Unable to Encode Request",
			"An unexpected error occurred while encoding the identity provider create request. JSON Error: "+err.Error(),
		)
		return nil, diags
	}

	return body, diags
}

func encodeUpdateRequest(ctx context.Context, data SsoIdpModel) ([]byte, diag.Diagnostics) {
	var diags diag.Diagnostics

	identifiers, identifierDiags := extractStringList(ctx, data.Identifiers)
	diags.Append(identifierDiags...)
	if diags.HasError() {
		return nil, diags
	}

	reqBody := bronto.UpdateIdpRequest{
		Identifiers:       identifiers,
		AttributeMappings: extractAttributeMappings(data.AttributeMappings),
	}
	if !data.MetaCert.IsNull() && !data.MetaCert.IsUnknown() {
		reqBody.MetaCert = data.MetaCert.ValueString()
	}
	if !data.AcceptIdpInitiated.IsNull() && !data.AcceptIdpInitiated.IsUnknown() {
		v := data.AcceptIdpInitiated.ValueBool()
		reqBody.AcceptIdpInitiated = &v
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		diags.AddError(
			"Unable to Encode Request",
			"An unexpected error occurred while encoding the identity provider update request. JSON Error: "+err.Error(),
		)
		return nil, diags
	}

	return body, diags
}

func extractAttributeMappings(models []AttributeMappingModel) []bronto.AttributeMapping {
	mappings := make([]bronto.AttributeMapping, 0, len(models))
	for _, m := range models {
		mappings = append(mappings, bronto.AttributeMapping{
			ProviderAttr: m.ProviderAttr.ValueString(),
			UserAttr:     m.UserAttr.ValueString(),
		})
	}
	return mappings
}

// populateModel maps an API Idp response onto the Terraform model. id is
// set to the idp's name, since the API has no separate generated id for
// this resource.
func populateModel(ctx context.Context, data *SsoIdpModel, idp bronto.Idp) diag.Diagnostics {
	var diags diag.Diagnostics

	data.Id = types.StringValue(idp.Name)
	data.Name = types.StringValue(idp.Name)
	data.ProviderType = types.StringValue(idp.ProviderType)
	data.AcceptIdpInitiated = types.BoolValue(idp.AcceptIdpInitiated)

	identifiersVal, identifierDiags := stringListValue(ctx, idp.Identifiers)
	diags.Append(identifierDiags...)
	data.Identifiers = identifiersVal

	if idp.MetaCert == "" {
		data.MetaCert = types.StringNull()
	} else {
		data.MetaCert = types.StringValue(idp.MetaCert)
	}

	if idp.IssuerUrl == "" {
		data.IssuerUrl = types.StringNull()
	} else {
		data.IssuerUrl = types.StringValue(idp.IssuerUrl)
	}

	if idp.ClientId == "" {
		data.ClientId = types.StringNull()
	} else {
		data.ClientId = types.StringValue(idp.ClientId)
	}

	if idp.ClientSecret == "" {
		data.ClientSecret = types.StringNull()
	} else {
		data.ClientSecret = types.StringValue(idp.ClientSecret)
	}

	if idp.AppClientId == "" {
		data.AppClientId = types.StringNull()
	} else {
		data.AppClientId = types.StringValue(idp.AppClientId)
	}

	mappings := make([]AttributeMappingModel, 0, len(idp.AttributeMappings))
	for _, m := range idp.AttributeMappings {
		mappings = append(mappings, AttributeMappingModel{
			ProviderAttr: types.StringValue(m.ProviderAttr),
			UserAttr:     types.StringValue(m.UserAttr),
		})
	}
	data.AttributeMappings = mappings

	return diags
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
// slice to an empty (but non-null) list first, matching the same
// null-vs-empty normalization used across the rest of this provider.
func stringListValue(ctx context.Context, values []string) (types.List, diag.Diagnostics) {
	if values == nil {
		values = []string{}
	}
	return types.ListValueFrom(ctx, types.StringType, values)
}
