package tfutil

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	bronto "terraform-provider-bronto/internal"
)

// DesiredTags returns the full set of tags a resource should end up with:
// the provider's default_tags merged with the resource's own configured
// tags, the resource's own tags winning on key conflicts.
func DesiredTags(ctx context.Context, planTags types.Map, defaultTags map[string]string) (map[string]string, diag.Diagnostics) {
	specificTags, diags := ExtractStringMap(ctx, planTags)
	if diags.HasError() {
		return nil, diags
	}
	return bronto.MergeTags(defaultTags, specificTags), diags
}

// PopulateTags sets a resource's tags_all to the full tag set from the API,
// and its tags to either the plan's own explicit value (left untouched -
// it's already known to be exactly what's in config, and echoing it back is
// the only way to avoid "provider produced inconsistent result" when
// default_tags adds keys the config never declared) or, when the plan
// didn't declare any tags of its own (planTags null/unknown, as is always
// the case on Read, which has no plan to consult), a best-effort
// reconstruction of "the tags not explained by default_tags".
//
// When syncing tags on Update, the current tags must come from the state's
// tags_all (the full set last read from the API), not its tags (only "own"
// tags, excluding anything matched to a default). Diffing against tags
// would make every default-contributed key look new on every update, and
// SyncTags would try to re-create a tag that already exists (409 Conflict).
func PopulateTags(ctx context.Context, tags, tagsAll *types.Map, planTags types.Map, apiTags, defaultTags map[string]string) diag.Diagnostics {
	var diags diag.Diagnostics

	tagsAllVal, tagsAllDiags := StringMapValue(ctx, apiTags)
	diags.Append(tagsAllDiags...)
	*tagsAll = tagsAllVal

	if !planTags.IsNull() && !planTags.IsUnknown() {
		*tags = planTags
		return diags
	}

	tagsVal, tagsDiags := StringMapValue(ctx, bronto.OwnTags(apiTags, defaultTags))
	diags.Append(tagsDiags...)
	*tags = tagsVal
	return diags
}
