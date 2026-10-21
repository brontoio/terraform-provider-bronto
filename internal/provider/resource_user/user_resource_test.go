package resource_user

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	bronto "terraform-provider-bronto/internal"
)

func TestLoginMethodsRoundTrip(t *testing.T) {
	ctx := context.Background()
	r := &userResource{}

	api := []bronto.LoginMethod{
		{Name: "Password", Enabled: true},
		{Name: "Okta", Enabled: false},
	}

	set, diags := toLoginMethodsSet(ctx, api)
	if diags.HasError() {
		t.Fatalf("toLoginMethodsSet: %v", diags)
	}
	if set.IsNull() || set.IsUnknown() {
		t.Fatalf("expected a known, non-null set")
	}

	back, diags := r.extractLoginMethods(ctx, set)
	if diags.HasError() {
		t.Fatalf("extractLoginMethods: %v", diags)
	}

	if len(back) != len(api) {
		t.Fatalf("got %d login methods, want %d", len(back), len(api))
	}
	for i, m := range api {
		if back[i] != m {
			t.Errorf("login method %d = %+v, want %+v", i, back[i], m)
		}
	}
}

func TestLoginMethodsEmptySet(t *testing.T) {
	ctx := context.Background()

	set, diags := toLoginMethodsSet(ctx, nil)
	if diags.HasError() {
		t.Fatalf("toLoginMethodsSet: %v", diags)
	}
	if set.IsNull() {
		t.Fatalf("expected an empty but non-null set for no login methods")
	}
}

func TestExtractLoginMethodsNullSet(t *testing.T) {
	r := &userResource{}
	methods, diags := r.extractLoginMethods(context.Background(), types.SetNull(LoginMethodsValue{}.Type(context.Background())))
	if diags.HasError() {
		t.Fatalf("extractLoginMethods: %v", diags)
	}
	if len(methods) != 0 {
		t.Errorf("expected no login methods, got %v", methods)
	}
}

func TestDesiredActive(t *testing.T) {
	if got := desiredActive(types.BoolNull()); got != nil {
		t.Errorf("desiredActive(null) = %v, want nil", got)
	}
	if got := desiredActive(types.BoolUnknown()); got != nil {
		t.Errorf("desiredActive(unknown) = %v, want nil", got)
	}
	if got := desiredActive(types.BoolValue(true)); got == nil || !*got {
		t.Errorf("desiredActive(true) = %v, want pointer to true", got)
	}
	if got := desiredActive(types.BoolValue(false)); got == nil || *got {
		t.Errorf("desiredActive(false) = %v, want pointer to false", got)
	}
}

func TestPopulateModelMapsAllFields(t *testing.T) {
	ctx := context.Background()
	invitedAt := int64(1731370728)
	invitedBy := "admin@company.com"

	api := bronto.User{
		Id:        "user-1",
		FirstName: "Alice",
		LastName:  "Smith",
		Email:     "alice@company.com",
		Roles:     []string{"Admin", "ReadOnly"},
		Status:    bronto.UserStatusActive,
		Tags:      map[string]string{"region": "eu"},
		LastLogins: map[string]int64{
			"Password": 1731338939,
		},
		LoginMethods: []bronto.LoginMethod{{Name: "Password", Enabled: true}},
		InvitedAt:    &invitedAt,
		InvitedBy:    &invitedBy,
	}

	var data UserModel
	diags := populateModel(ctx, &data, api)
	if diags.HasError() {
		t.Fatalf("populateModel: %v", diags)
	}

	if data.Id.ValueString() != "user-1" {
		t.Errorf("Id = %s", data.Id.ValueString())
	}
	if !data.Active.ValueBool() {
		t.Errorf("Active = false, want true for status ACTIVE")
	}
	if data.Status.ValueString() != bronto.UserStatusActive {
		t.Errorf("Status = %s", data.Status.ValueString())
	}
	if data.InvitedAt.ValueInt64() != invitedAt {
		t.Errorf("InvitedAt = %d, want %d", data.InvitedAt.ValueInt64(), invitedAt)
	}
	if data.InvitedBy.ValueString() != invitedBy {
		t.Errorf("InvitedBy = %s, want %s", data.InvitedBy.ValueString(), invitedBy)
	}
	if data.LastLogins.IsNull() {
		t.Errorf("LastLogins should not be null")
	}
	if data.LoginMethods.IsNull() {
		t.Errorf("LoginMethods should not be null")
	}

	var roles []string
	data.Roles.ElementsAs(ctx, &roles, false)
	if len(roles) != 2 || roles[0] != "Admin" || roles[1] != "ReadOnly" {
		t.Errorf("Roles = %v", roles)
	}
}

func TestPopulateModelInactiveStatus(t *testing.T) {
	var data UserModel
	diags := populateModel(context.Background(), &data, bronto.User{Id: "u", Status: "DEACTIVATED"})
	if diags.HasError() {
		t.Fatalf("populateModel: %v", diags)
	}
	if data.Active.ValueBool() {
		t.Errorf("Active = true, want false for status DEACTIVATED")
	}
	if !data.InvitedAt.IsNull() {
		t.Errorf("InvitedAt should be null when the API omits it")
	}
	if !data.InvitedBy.IsNull() {
		t.Errorf("InvitedBy should be null when the API omits it")
	}
}

func TestPopulateTagsEchoesExplicitPlanValue(t *testing.T) {
	ctx := context.Background()
	r := &userResource{client: &bronto.Client{DefaultTags: map[string]string{"managed_by": "terraform"}}}

	planTags, _ := types.MapValueFrom(ctx, types.StringType, map[string]string{"team": "observability"})

	var data UserModel
	diags := r.populateTags(ctx, &data, planTags, map[string]string{"managed_by": "terraform", "team": "observability"})
	if diags.HasError() {
		t.Fatalf("populateTags: %v", diags)
	}

	if !data.Tags.Equal(planTags) {
		t.Errorf("Tags = %v, want it echoed back as the exact plan value %v", data.Tags, planTags)
	}

	var tagsAll map[string]string
	if diags := data.TagsAll.ElementsAs(ctx, &tagsAll, false); diags.HasError() {
		t.Fatalf("TagsAll.ElementsAs: %v", diags)
	}
	if tagsAll["managed_by"] != "terraform" || tagsAll["team"] != "observability" {
		t.Errorf("TagsAll = %v", tagsAll)
	}
}

func TestPopulateTagsDerivesOwnTagsWhenPlanUnset(t *testing.T) {
	ctx := context.Background()
	r := &userResource{client: &bronto.Client{DefaultTags: map[string]string{"managed_by": "terraform"}}}

	var data UserModel
	diags := r.populateTags(ctx, &data, types.MapNull(types.StringType), map[string]string{"managed_by": "terraform", "team": "observability"})
	if diags.HasError() {
		t.Fatalf("populateTags: %v", diags)
	}

	var tags map[string]string
	if diags := data.Tags.ElementsAs(ctx, &tags, false); diags.HasError() {
		t.Fatalf("Tags.ElementsAs: %v", diags)
	}
	if _, hasDefault := tags["managed_by"]; hasDefault {
		t.Errorf("Tags should exclude the default-derived key, got %v", tags)
	}
	if tags["team"] != "observability" {
		t.Errorf("Tags = %v, want team=observability", tags)
	}
}
