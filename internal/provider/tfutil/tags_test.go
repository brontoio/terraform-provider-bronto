package tfutil

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDesiredTagsMergesDefaultsWithOwnTagsWinning(t *testing.T) {
	ctx := context.Background()
	planTags, _ := types.MapValueFrom(ctx, types.StringType, map[string]string{"team": "core", "env": "prod"})

	desired, diags := DesiredTags(ctx, planTags, map[string]string{"managed_by": "terraform", "team": "default"})
	if diags.HasError() {
		t.Fatalf("DesiredTags: %v", diags)
	}

	want := map[string]string{"managed_by": "terraform", "team": "core", "env": "prod"}
	if len(desired) != len(want) {
		t.Fatalf("DesiredTags = %v, want %v", desired, want)
	}
	for k, v := range want {
		if desired[k] != v {
			t.Errorf("DesiredTags[%q] = %q, want %q", k, desired[k], v)
		}
	}
}

func TestDesiredTagsWithNullPlanIsJustDefaults(t *testing.T) {
	desired, diags := DesiredTags(context.Background(), types.MapNull(types.StringType), map[string]string{"managed_by": "terraform"})
	if diags.HasError() {
		t.Fatalf("DesiredTags: %v", diags)
	}
	if len(desired) != 1 || desired["managed_by"] != "terraform" {
		t.Errorf("DesiredTags = %v, want only managed_by=terraform", desired)
	}
}

func TestUpdateTagSyncCurrentTagsComesFromTagsAll(t *testing.T) {
	ctx := context.Background()
	defaultTags := map[string]string{"managed_by": "terraform"}

	stateTagsAll, _ := types.MapValueFrom(ctx, types.StringType, map[string]string{
		"managed_by": "terraform",
		"team":       "observability",
	})
	planTags := types.MapNull(types.StringType) // config doesn't declare its own tags

	currentTags, diags := ExtractStringMap(ctx, stateTagsAll)
	if diags.HasError() {
		t.Fatalf("ExtractStringMap(TagsAll): %v", diags)
	}
	desiredTags, diags := DesiredTags(ctx, planTags, defaultTags)
	if diags.HasError() {
		t.Fatalf("DesiredTags: %v", diags)
	}

	if currentTags["managed_by"] != desiredTags["managed_by"] {
		t.Fatalf("currentTags[managed_by] = %q, desiredTags[managed_by] = %q - SyncTags would wrongly try to re-create an existing tag",
			currentTags["managed_by"], desiredTags["managed_by"])
	}
}

func TestPopulateTagsEchoesExplicitPlanValue(t *testing.T) {
	ctx := context.Background()
	planTags, _ := types.MapValueFrom(ctx, types.StringType, map[string]string{"team": "observability"})

	var tags, tagsAll types.Map
	diags := PopulateTags(ctx, &tags, &tagsAll, planTags,
		map[string]string{"managed_by": "terraform", "team": "observability"},
		map[string]string{"managed_by": "terraform"})
	if diags.HasError() {
		t.Fatalf("PopulateTags: %v", diags)
	}

	if !tags.Equal(planTags) {
		t.Errorf("tags = %v, want it echoed back as the exact plan value %v", tags, planTags)
	}

	var all map[string]string
	if diags := tagsAll.ElementsAs(ctx, &all, false); diags.HasError() {
		t.Fatalf("tagsAll.ElementsAs: %v", diags)
	}
	if all["managed_by"] != "terraform" || all["team"] != "observability" {
		t.Errorf("tagsAll = %v", all)
	}
}

func TestPopulateTagsDerivesOwnTagsWhenPlanUnset(t *testing.T) {
	ctx := context.Background()

	var tags, tagsAll types.Map
	diags := PopulateTags(ctx, &tags, &tagsAll, types.MapNull(types.StringType),
		map[string]string{"managed_by": "terraform", "team": "observability"},
		map[string]string{"managed_by": "terraform"})
	if diags.HasError() {
		t.Fatalf("PopulateTags: %v", diags)
	}

	var own map[string]string
	if diags := tags.ElementsAs(ctx, &own, false); diags.HasError() {
		t.Fatalf("tags.ElementsAs: %v", diags)
	}
	if _, hasDefault := own["managed_by"]; hasDefault {
		t.Errorf("tags should exclude the default-derived key, got %v", own)
	}
	if own["team"] != "observability" {
		t.Errorf("tags = %v, want team=observability", own)
	}
}

func TestPopulateTagsNoApiTagsIsEmptyNotNull(t *testing.T) {
	ctx := context.Background()

	var tags, tagsAll types.Map
	diags := PopulateTags(ctx, &tags, &tagsAll, types.MapNull(types.StringType), nil, nil)
	if diags.HasError() {
		t.Fatalf("PopulateTags: %v", diags)
	}
	if tags.IsNull() || len(tags.Elements()) != 0 {
		t.Errorf("tags = %v, want an empty non-null map", tags)
	}
	if tagsAll.IsNull() || len(tagsAll.Elements()) != 0 {
		t.Errorf("tagsAll = %v, want an empty non-null map", tagsAll)
	}
}
