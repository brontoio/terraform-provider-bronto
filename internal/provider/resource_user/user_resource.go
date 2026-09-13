package resource_user

import (
	"context"
	"encoding/json"
	"fmt"
	bronto "terraform-provider-bronto/internal"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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
	FirstName    string                `json:"first_name"`
	LastName     string                `json:"last_name"`
	Roles        []string              `json:"roles"`
	LoginMethods *[]bronto.LoginMethod `json:"login_methods,omitempty"`
}

func (r *userResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *userResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = UserResourceSchema(ctx)
}

func (r *userResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Debug(ctx, "Configuring user resource")
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*bronto.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *bronto.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan UserModel

	// 1. Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// 2. Create HTTP Request
	httpReqBody, err := r.encodeCreateJson(ctx, plan)
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
	createResp, err := bronto.DecodeUser(httpResp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while parsing the resource creation response. "+
				"JSON Error: "+err.Error(),
		)
		return
	}

	// 5. The create endpoint doesn't accept login_methods, so if the config
	// sets them, apply them with an immediate follow-up update.
	if !plan.LoginMethods.IsNull() && !plan.LoginMethods.IsUnknown() {
		updateBody, diags := r.encodeUpdateJson(ctx, plan)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		httpResp, err = r.client.UpdateRequest(ctx, updateBody, "/users/"+createResp.Id)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Create Resource",
				"An unexpected error occurred while setting login methods on the new user. "+
					"Error: "+err.Error(),
			)
			return
		}

		createResp, err = bronto.DecodeUser(httpResp)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Create Resource",
				"An unexpected error occurred while parsing the login methods update response. "+
					"JSON Error: "+err.Error(),
			)
			return
		}
	}

	// 6. Reconcile the active/deactivated state, if the config expresses one.
	// New users are created with status INVITED, not ACTIVE, so active=true
	// needs reconciling here too, not just active=false.
	if desired := desiredActive(plan.Active); desired != nil {
		currentlyActive := createResp.Status == bronto.UserStatusActive
		if *desired != currentlyActive {
			createResp, err = r.setActive(ctx, createResp, *desired)
			if err != nil {
				action := "reactivating"
				if !*desired {
					action = "deactivating"
				}
				resp.Diagnostics.AddError(
					"Unable to Create Resource",
					fmt.Sprintf("An unexpected error occurred while %s the new user. Error: %s", action, err.Error()),
				)
				return
			}
		}
	}

	// 7. Save Terraform state
	resp.Diagnostics.Append(populateModel(ctx, &plan, createResp)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
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
		if bronto.IsNotFound(err) {
			// The user no longer exists upstream; drop it from state so the
			// next plan recreates it instead of failing.
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Read Resource",
			"An unexpected error occurred while calling read API. "+
				"Error: "+err.Error(),
		)
		return
	}

	// 4. Parse Response
	readResp, err := bronto.DecodeUser(httpResp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Resource",
			"An unexpected error occurred while parsing the resource read response. "+
				"JSON Error: "+err.Error(),
		)
		return
	}

	// 5. Save Terraform state
	resp.Diagnostics.Append(populateModel(ctx, &data, readResp)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan UserModel

	// 1. Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// 2. Get Id and prior state
	var state UserModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	userId := state.Id.ValueString()

	// 3. Create HTTP Request
	httpReqBody, diags := r.encodeUpdateJson(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// 4. Call API
	httpResp, err := r.client.UpdateRequest(ctx, httpReqBody, "/users/"+userId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while calling update API. "+
				"Error: "+err.Error(),
		)
		return
	}

	updateResp, err := bronto.DecodeUser(httpResp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while parsing the resource update response. "+
				"JSON Error: "+err.Error(),
		)
		return
	}

	// 5. Reconcile the active/deactivated state, if the config expresses one.
	if desired := desiredActive(plan.Active); desired != nil {
		currentlyActive := updateResp.Status == bronto.UserStatusActive
		if *desired != currentlyActive {
			updateResp, err = r.setActive(ctx, updateResp, *desired)
			if err != nil {
				action := "reactivating"
				if !*desired {
					action = "deactivating"
				}
				resp.Diagnostics.AddError(
					"Unable to Update Resource",
					fmt.Sprintf("An unexpected error occurred while %s the user. Error: %s", action, err.Error()),
				)
				return
			}
		}
	}

	// 6. Save Terraform state
	resp.Diagnostics.Append(populateModel(ctx, &plan, updateResp)...)
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
	if err != nil && !bronto.IsNotFound(err) {
		// A 404 means the user is already gone, which is the desired
		// end state, so treat delete as idempotent and succeed.
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

// setActive calls the deactivate or reactivate action endpoint and returns
// user with its status updated to match. The action endpoints only return
// {success, status}, not a full User object (despite what the OpenAPI spec
// documents), so the rest of user's fields are preserved from the caller
// rather than reconstructed from the response.
func (r *userResource) setActive(ctx context.Context, user bronto.User, active bool) (bronto.User, error) {
	action := "reactivate"
	if !active {
		action = "deactivate"
	}

	httpResp, err := r.client.PostRequest(ctx, nil, "/users/"+user.Id+"/"+action)
	if err != nil {
		return bronto.User{}, err
	}

	result, err := bronto.DecodeSetActiveResponse(httpResp)
	if err != nil {
		return bronto.User{}, err
	}

	user.Status = result.Status
	return user, nil
}

// desiredActive returns the user's desired active state from the plan, or
// nil if the config doesn't express a preference (attribute unset/unknown),
// in which case the active/deactivated state should be left untouched.
func desiredActive(planActive types.Bool) *bool {
	if planActive.IsNull() || planActive.IsUnknown() {
		return nil
	}
	v := planActive.ValueBool()
	return &v
}

func (r *userResource) encodeCreateJson(ctx context.Context, data UserModel) ([]byte, error) {
	var roles []string
	data.Roles.ElementsAs(ctx, &roles, false)
	// Convert from Terraform data model into API data model
	createReq := CreateUserApiModel{
		FirstName: data.FirstName.ValueString(),
		LastName:  data.LastName.ValueString(),
		Email:     data.Email.ValueString(),
		Roles:     roles,
	}

	return json.Marshal(createReq)
}

func (r *userResource) encodeUpdateJson(ctx context.Context, data UserModel) ([]byte, diag.Diagnostics) {
	var diags diag.Diagnostics

	var roles []string
	diags.Append(data.Roles.ElementsAs(ctx, &roles, false)...)
	if diags.HasError() {
		return nil, diags
	}

	loginMethods, loginMethodDiags := r.extractLoginMethods(ctx, data.LoginMethods)
	diags.Append(loginMethodDiags...)
	if diags.HasError() {
		return nil, diags
	}

	// Convert from Terraform data model into API data model
	updateReq := UpdateUserApiModel{
		FirstName: data.FirstName.ValueString(),
		LastName:  data.LastName.ValueString(),
		Roles:     roles,
	}
	if !data.LoginMethods.IsNull() && !data.LoginMethods.IsUnknown() {
		updateReq.LoginMethods = &loginMethods
	}

	body, err := json.Marshal(updateReq)
	if err != nil {
		diags.AddError(
			"Unable to Encode Update Request",
			"An unexpected error occurred while encoding the resource update request. "+
				"JSON Error: "+err.Error(),
		)
		return nil, diags
	}

	return body, diags
}

// extractLoginMethods converts the Terraform login_methods set into its API
// representation. It returns an empty (non-nil) slice for a null/unknown
// set so callers can tell "no login methods configured" apart from a JSON
// marshaling error.
func (r *userResource) extractLoginMethods(ctx context.Context, set types.Set) ([]bronto.LoginMethod, diag.Diagnostics) {
	var diags diag.Diagnostics

	if set.IsNull() || set.IsUnknown() {
		return []bronto.LoginMethod{}, diags
	}

	var values []LoginMethodsValue
	diags.Append(set.ElementsAs(ctx, &values, false)...)
	if diags.HasError() {
		return nil, diags
	}

	methods := make([]bronto.LoginMethod, 0, len(values))
	for _, v := range values {
		methods = append(methods, bronto.LoginMethod{
			Name:    v.Name.ValueString(),
			Enabled: v.Enabled.ValueBool(),
		})
	}
	return methods, diags
}

// toLoginMethodsSet converts the API's login_methods into a Terraform set of
// LoginMethodsValue. A set (rather than a list) is used because the backend
// doesn't preserve submission order, so the returned order can differ from
// whatever order was configured - a list would report that as spurious
// drift.
func toLoginMethodsSet(ctx context.Context, methods []bronto.LoginMethod) (types.Set, diag.Diagnostics) {
	values := make([]attr.Value, 0, len(methods))
	for _, m := range methods {
		values = append(values, LoginMethodsValue{
			Name:    types.StringValue(m.Name),
			Enabled: types.BoolValue(m.Enabled),
			state:   attr.ValueStateKnown,
		})
	}

	return types.SetValue(LoginMethodsValue{}.Type(ctx), values)
}

// populateModel maps an API user response onto the Terraform model,
// including the id, so it can be used from Create, Read, and Update.
func populateModel(ctx context.Context, data *UserModel, api bronto.User) diag.Diagnostics {
	var diags diag.Diagnostics

	data.Id = types.StringValue(api.Id)
	data.FirstName = types.StringValue(api.FirstName)
	data.LastName = types.StringValue(api.LastName)
	data.Email = types.StringValue(api.Email)
	data.Roles = extractStringList(api.Roles)
	data.Status = types.StringValue(api.Status)
	data.Active = types.BoolValue(api.Status == bronto.UserStatusActive)

	if api.InvitedAt != nil {
		data.InvitedAt = types.Int64Value(*api.InvitedAt)
	} else {
		data.InvitedAt = types.Int64Null()
	}

	if api.InvitedBy != nil {
		data.InvitedBy = types.StringValue(*api.InvitedBy)
	} else {
		data.InvitedBy = types.StringNull()
	}

	tagsVal, tagsDiags := types.MapValueFrom(ctx, types.StringType, api.Tags)
	diags.Append(tagsDiags...)
	data.Tags = tagsVal

	lastLoginsVal, lastLoginsDiags := types.MapValueFrom(ctx, types.Int64Type, api.LastLogins)
	diags.Append(lastLoginsDiags...)
	data.LastLogins = lastLoginsVal

	loginMethodsVal, loginMethodsDiags := toLoginMethodsSet(ctx, api.LoginMethods)
	diags.Append(loginMethodsDiags...)
	data.LoginMethods = loginMethodsVal

	return diags
}

func extractStringList(l []string) basetypes.ListValue {
	elements := make([]attr.Value, 0, len(l))
	for _, i := range l {
		elements = append(elements, types.StringValue(i))
	}
	list, _ := types.ListValue(types.StringType, elements)
	return list
}
