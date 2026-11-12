package resource_group

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
	_ resource.Resource                = &groupResource{}
	_ resource.ResourceWithConfigure   = &groupResource{}
	_ resource.ResourceWithImportState = &groupResource{}
)

func NewGroupResource() resource.Resource {
	return &groupResource{}
}

type groupResource struct {
	client *bronto.Client
}

func (r *groupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (r *groupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GroupResourceSchema(ctx)
}

func (r *groupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *groupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan GroupModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, diags := encodeGroupRequest(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.CreateRequest(ctx, body, "/groups")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while calling the create API. Error: "+err.Error(),
		)
		return
	}

	created, err := bronto.DecodeUserGroup(httpResp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while parsing the resource creation response. JSON Error: "+err.Error(),
		)
		return
	}

	// The create endpoint doesn't accept members; the only way to set them
	// is a follow-up add-members call.
	desiredMembers, diags2 := extractMembers(ctx, plan.Members)
	resp.Diagnostics.Append(diags2...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(desiredMembers) > 0 {
		if err := r.addMembers(ctx, created.GroupId, desiredMembers); err != nil {
			resp.Diagnostics.AddError(
				"Unable to Create Resource",
				"An unexpected error occurred while adding members to the new group. Error: "+err.Error(),
			)
			return
		}
	}

	members, err := r.getMembers(ctx, created.GroupId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while reading back the new group's members. Error: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(populateModel(ctx, &plan, created, members)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *groupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data GroupModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupId := data.Id.ValueString()

	httpResp, err := r.client.ReadRequest(ctx, "/groups/"+groupId)
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

	found, err := bronto.DecodeUserGroup(httpResp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Resource",
			"An unexpected error occurred while parsing the resource read response. JSON Error: "+err.Error(),
		)
		return
	}

	members, err := r.getMembers(ctx, groupId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Resource",
			"An unexpected error occurred while reading the group's members. Error: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(populateModel(ctx, &data, found, members)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *groupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan GroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state GroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupId := state.Id.ValueString()

	body, diags := encodeGroupRequest(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.UpdateRequest(ctx, body, "/groups/"+groupId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while calling the update API. Error: "+err.Error(),
		)
		return
	}

	updated, err := bronto.DecodeUserGroup(httpResp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while parsing the resource update response. JSON Error: "+err.Error(),
		)
		return
	}

	// currentMembers comes from state (populated from the API on the last
	// Read/Create), not re-fetched here, since the group's own membership
	// list doesn't drift from anything else this provider manages - unlike
	// tags_all, there's no default_tags-style side channel that could make
	// a stale state value under-count what's actually there.
	currentMembers, diags2 := extractMembers(ctx, state.Members)
	resp.Diagnostics.Append(diags2...)
	desiredMembers, diags3 := extractMembers(ctx, plan.Members)
	resp.Diagnostics.Append(diags3...)
	if resp.Diagnostics.HasError() {
		return
	}

	toAdd, toRemove := diffMembers(currentMembers, desiredMembers)

	if len(toAdd) > 0 {
		if err := r.addMembers(ctx, groupId, toAdd); err != nil {
			resp.Diagnostics.AddError(
				"Unable to Update Resource",
				"An unexpected error occurred while adding members to the group. Error: "+err.Error(),
			)
			return
		}
	}

	if len(toRemove) > 0 {
		if err := r.removeMembers(ctx, groupId, toRemove); err != nil {
			resp.Diagnostics.AddError(
				"Unable to Update Resource",
				"An unexpected error occurred while removing members from the group. Error: "+err.Error(),
			)
			return
		}
	}

	members, err := r.getMembers(ctx, groupId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while reading back the group's members. Error: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(populateModel(ctx, &plan, updated, members)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *groupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data GroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteRequest(ctx, "/groups/"+data.Id.ValueString())
	if err != nil && !bronto.IsNotFound(err) {
		resp.Diagnostics.AddError(
			"Unable to Delete Resource",
			"An unexpected error occurred while calling the delete API. Error: "+err.Error(),
		)
		return
	}
}

func (r *groupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *groupResource) getMembers(ctx context.Context, groupId string) ([]bronto.GroupMember, error) {
	httpResp, err := r.client.ReadRequest(ctx, "/groups/"+groupId+"/members")
	if err != nil {
		return nil, err
	}
	list, err := bronto.DecodeGroupMembersList(httpResp)
	if err != nil {
		return nil, err
	}
	return list.Members, nil
}

func (r *groupResource) addMembers(ctx context.Context, groupId string, members []bronto.GroupMemberRef) error {
	body, err := json.Marshal(bronto.GroupMembersRequest{Members: members})
	if err != nil {
		return err
	}
	return r.client.PostNoContentRequest(ctx, body, "/groups/"+groupId+"/members")
}

func (r *groupResource) removeMembers(ctx context.Context, groupId string, members []bronto.GroupMemberRef) error {
	body, err := json.Marshal(bronto.GroupMembersRequest{Members: members})
	if err != nil {
		return err
	}
	return r.client.DeleteRequestWithBody(ctx, body, "/groups/"+groupId+"/members")
}

func encodeGroupRequest(data GroupModel) ([]byte, diag.Diagnostics) {
	var diags diag.Diagnostics

	reqBody := bronto.GroupRequest{
		Name: data.Name.ValueString(),
	}
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		reqBody.Description = data.Description.ValueString()
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		diags.AddError(
			"Unable to Encode Request",
			"An unexpected error occurred while encoding the group request. JSON Error: "+err.Error(),
		)
		return nil, diags
	}

	return body, diags
}

// populateModel maps an API group response and its members onto the
// Terraform model, including the id, so it can be used from Create, Read,
// and Update.
func populateModel(ctx context.Context, data *GroupModel, g bronto.UserGroup, members []bronto.GroupMember) diag.Diagnostics {
	var diags diag.Diagnostics

	data.Id = types.StringValue(g.GroupId)
	data.Name = types.StringValue(g.Name)

	if g.Description == "" {
		data.Description = types.StringNull()
	} else {
		data.Description = types.StringValue(g.Description)
	}

	data.CreatedAt = types.Int64Value(g.CreatedAt)
	data.ModifiedAt = types.Int64Value(g.ModifiedAt)

	memberModels := make([]MemberModel, 0, len(members))
	for _, m := range members {
		memberModels = append(memberModels, MemberModel{
			MemberType: types.StringValue(m.MemberType),
			MemberId:   types.StringValue(m.MemberId),
		})
	}
	membersVal, memberDiags := types.ListValueFrom(ctx, memberObjectType, memberModels)
	diags.Append(memberDiags...)
	data.Members = membersVal

	return diags
}

func extractMembers(ctx context.Context, list types.List) ([]bronto.GroupMemberRef, diag.Diagnostics) {
	var diags diag.Diagnostics

	if list.IsNull() || list.IsUnknown() {
		return nil, diags
	}

	var models []MemberModel
	diags.Append(list.ElementsAs(ctx, &models, false)...)
	if diags.HasError() {
		return nil, diags
	}

	refs := make([]bronto.GroupMemberRef, 0, len(models))
	for _, m := range models {
		refs = append(refs, bronto.GroupMemberRef{
			MemberType: m.MemberType.ValueString(),
			MemberId:   m.MemberId.ValueString(),
		})
	}
	return refs, diags
}

// diffMembers compares current against desired group membership, keyed by
// (member_type, member_id), and returns the members to add and remove to
// reconcile current into desired.
func diffMembers(current, desired []bronto.GroupMemberRef) (toAdd, toRemove []bronto.GroupMemberRef) {
	key := func(m bronto.GroupMemberRef) string { return m.MemberType + ":" + m.MemberId }

	currentSet := make(map[string]bool, len(current))
	for _, m := range current {
		currentSet[key(m)] = true
	}
	desiredSet := make(map[string]bool, len(desired))
	for _, m := range desired {
		desiredSet[key(m)] = true
	}

	for _, m := range desired {
		if !currentSet[key(m)] {
			toAdd = append(toAdd, m)
		}
	}
	for _, m := range current {
		if !desiredSet[key(m)] {
			toRemove = append(toRemove, m)
		}
	}

	return toAdd, toRemove
}
