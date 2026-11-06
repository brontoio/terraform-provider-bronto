// Package tfutil holds conversions between Terraform framework values and
// plain Go values that are shared by more than one resource.
package tfutil

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ExtractStringMap converts a types.Map of strings into a map[string]string,
// returning nil for a null or unknown map.
func ExtractStringMap(ctx context.Context, m types.Map) (map[string]string, diag.Diagnostics) {
	var diags diag.Diagnostics

	if m.IsNull() || m.IsUnknown() {
		return nil, diags
	}

	var values map[string]string
	diags.Append(m.ElementsAs(ctx, &values, false)...)
	return values, diags
}

// StringMapValue converts a map[string]string into a types.Map, normalizing
// a nil map to an empty (but non-null) map first. types.MapValueFrom on its
// own maps a nil map to a null map, which would make an Optional+Computed
// attribute flip between null and an empty map depending on whether the API
// responds with an absent field or an empty object for the same "no
// values" state.
func StringMapValue(ctx context.Context, values map[string]string) (types.Map, diag.Diagnostics) {
	if values == nil {
		values = map[string]string{}
	}
	return types.MapValueFrom(ctx, types.StringType, values)
}
