package resource_group

import (
	"context"
	"testing"

	bronto "terraform-provider-bronto/internal"
)

func TestPopulateModelMapsAllFields(t *testing.T) {
	ctx := context.Background()

	g := bronto.UserGroup{
		GroupId:     "group-1",
		Name:        "Platform Team",
		Description: "People who own platform infra",
		CreatedAt:   1710948395538,
		ModifiedAt:  1710948395999,
	}
	members := []bronto.GroupMember{
		{GroupId: "group-1", MemberId: "user-1", MemberType: "USER", ModifiedAt: 1},
		{GroupId: "group-1", MemberId: "group-2", MemberType: "GROUP", ModifiedAt: 2},
	}

	var data GroupModel
	diags := populateModel(ctx, &data, g, members)
	if diags.HasError() {
		t.Fatalf("populateModel: %v", diags)
	}

	if data.Id.ValueString() != "group-1" {
		t.Errorf("Id = %s", data.Id.ValueString())
	}
	if data.Name.ValueString() != "Platform Team" {
		t.Errorf("Name = %s", data.Name.ValueString())
	}
	if data.Description.ValueString() != "People who own platform infra" {
		t.Errorf("Description = %s", data.Description.ValueString())
	}
	if data.CreatedAt.ValueInt64() != 1710948395538 {
		t.Errorf("CreatedAt = %d", data.CreatedAt.ValueInt64())
	}

	var memberModels []MemberModel
	if diags := data.Members.ElementsAs(ctx, &memberModels, false); diags.HasError() {
		t.Fatalf("Members.ElementsAs: %v", diags)
	}
	if len(memberModels) != 2 {
		t.Fatalf("Members = %d entries, want 2", len(memberModels))
	}
	if memberModels[0].MemberType.ValueString() != "USER" || memberModels[0].MemberId.ValueString() != "user-1" {
		t.Errorf("Members[0] = %+v", memberModels[0])
	}
}

func TestPopulateModelEmptyDescriptionIsNull(t *testing.T) {
	var data GroupModel
	diags := populateModel(context.Background(), &data, bronto.UserGroup{GroupId: "g", Name: "n"}, nil)
	if diags.HasError() {
		t.Fatalf("populateModel: %v", diags)
	}
	if !data.Description.IsNull() {
		t.Errorf("Description should be null when the API returns empty, got %q", data.Description.ValueString())
	}
	if data.Members.IsNull() {
		t.Errorf("Members should be an empty list, not null, when the group has no members")
	}
}

func TestDiffMembersAddAndRemove(t *testing.T) {
	current := []bronto.GroupMemberRef{
		{MemberType: "USER", MemberId: "user-1"},
		{MemberType: "USER", MemberId: "user-2"},
	}
	desired := []bronto.GroupMemberRef{
		{MemberType: "USER", MemberId: "user-2"},
		{MemberType: "GROUP", MemberId: "group-3"},
	}

	toAdd, toRemove := diffMembers(current, desired)

	if len(toAdd) != 1 || toAdd[0].MemberType != "GROUP" || toAdd[0].MemberId != "group-3" {
		t.Errorf("toAdd = %v, want [{GROUP group-3}]", toAdd)
	}
	if len(toRemove) != 1 || toRemove[0].MemberType != "USER" || toRemove[0].MemberId != "user-1" {
		t.Errorf("toRemove = %v, want [{USER user-1}]", toRemove)
	}
}

func TestDiffMembersNoChange(t *testing.T) {
	members := []bronto.GroupMemberRef{{MemberType: "USER", MemberId: "user-1"}}
	toAdd, toRemove := diffMembers(members, members)
	if len(toAdd) != 0 || len(toRemove) != 0 {
		t.Errorf("toAdd = %v, toRemove = %v, want both empty for identical sets", toAdd, toRemove)
	}
}
