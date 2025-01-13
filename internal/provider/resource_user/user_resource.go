package resource_user

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"strings"
	bronto "terraform-provider-bronto/internal"
)

var (
	_ resource.Resource              = &userResource{}
	_ resource.ResourceWithConfigure = &userResource{}
)

func NewUserResource() resource.Resource {
	return &userResource{}
}

type userResource struct {
	client *bronto.Client
}

type CreateUserApiModel struct {
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Email     string   `json:"email"`
	Roles     []string `json:"roles"`
}

type UpdateUserApiModel struct {
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Roles     []string `json:"roles"`
}

type UserApiModel struct {
	FirstName string   `json:"first_name"`
	Id        string   `json:"id"`
	LastName  string   `json:"last_name"`
	Email     string   `json:"email"`
	Roles     []string `json:"roles"`
}

func (r *userResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *userResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = UserResourceSchema(ctx)
}

func (r *userResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Error(ctx, "Resource configure")
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*bronto.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *hashicups.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

//func (r *userResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
//	var plan UserModel
//	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
//	if resp.Diagnostics.HasError() {
//		return
//	}
//
//	var state UserModel
//	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
//	if resp.Diagnostics.HasError() {
//		return
//	}
//
//	if !state.Email.Equal(plan.Email) {
//		resp.RequiresReplace = append(resp.RequiresReplace, path.Root("email"))
//		return
//	}
//}

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data UserModel

	// 1. Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// 2. Create HTTP Request
	httpReqBody, err := r.encodeJson(ctx, data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while creating the resource create request. "+
				"JSON Error: "+err.Error(),
		)
		return
	}

	// 3. Call API
	httpResp, err := r.client.CreateRequest(ctx, httpReqBody, "/users")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while calling create API. "+
				"Error: "+err.Error(),
		)
		return
	}

	// 4. Parse Response
	var createResp UserApiModel
	err = json.NewDecoder(strings.NewReader(httpResp)).Decode(&createResp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while parsing the resource creation response. "+
				"JSON Error: "+err.Error(),
		)
		return
	}

	// 5. Save Terraform state
	data.Id = types.StringValue(createResp.Id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UserModel

	// 1. Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// 2. Get Id
	userId := data.Id.ValueString()

	// 3. Call API
	httpResp, err := r.client.ReadRequest(ctx, "/users/"+userId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Resource",
			"An unexpected error occurred while calling read API. "+
				"Error: "+err.Error(),
		)
		return
	}

	// 4. Parse Response
	var readResp UserApiModel
	err = json.NewDecoder(strings.NewReader(httpResp)).Decode(&readResp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Resource",
			"An unexpected error occurred while parsing the resource read response. "+
				"JSON Error: "+err.Error(),
		)
		return
	}

	// 5. Save Terraform state
	data.FirstName = types.StringValue(readResp.FirstName)
	data.LastName = types.StringValue(readResp.LastName)
	data.Email = types.StringValue(readResp.Email)
	data.Roles = r.extractList(readResp.Roles)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan UserModel

	// 1. Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// 2. Get Id
	var state UserModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	userId := state.Id.ValueString()

	// 3. Create HTTP Request
	httpReqBody, err := r.encodeUpdateJson(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while creating the resource update request. "+
				"JSON Error: "+err.Error(),
		)
		return
	}

	// 4. Call API
	err = r.client.UpdateRequest(ctx, httpReqBody, "/users/"+userId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while calling create API. "+
				"Error: "+err.Error(),
		)
		return
	}

	// 5. Save Terraform state
	plan.Id = types.StringValue(userId)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data UserModel

	// 1. Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// 2. Get Id
	userId := data.Id.ValueString()

	// 3. Call API
	err := r.client.DeleteRequest(ctx, "/users/"+userId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Delete Resource",
			"An unexpected error occurred while calling deleting API. "+
				"Error: "+err.Error(),
		)
		return
	}

}

func (r *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *userResource) encodeJson(ctx context.Context, data UserModel) ([]byte, error) {
	var roles []string
	data.Roles.ElementsAs(ctx, &roles, false)
	// Convert from Terraform data model into API data model
	createReq := CreateUserApiModel{
		FirstName: data.FirstName.ValueString(),
		LastName:  data.LastName.ValueString(),
		Email:     data.Email.ValueString(),
		Roles:     roles,
	}

	httpReqBody, err := json.Marshal(createReq)
	return httpReqBody, err
}

func (r *userResource) encodeUpdateJson(ctx context.Context, data UserModel) ([]byte, error) {
	var roles []string
	data.Roles.ElementsAs(ctx, &roles, false)
	// Convert from Terraform data model into API data model
	updateReq := UpdateUserApiModel{
		FirstName: data.FirstName.ValueString(),
		LastName:  data.LastName.ValueString(),
		Roles:     roles,
	}

	httpReqBody, err := json.Marshal(updateReq)
	return httpReqBody, err
}

func (r *userResource) extractList(l []string) basetypes.ListValue {
	elements := make([]attr.Value, 0, len(l))
	for _, i := range l {
		elements = append(elements, types.StringValue(i))
	}
	list, _ := types.ListValue(types.StringType, elements)
	return list
}
