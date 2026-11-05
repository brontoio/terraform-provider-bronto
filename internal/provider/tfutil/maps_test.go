package tfutil

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestExtractStringMapNull(t *testing.T) {
	values, diags := ExtractStringMap(context.Background(), types.MapNull(types.StringType))
	if diags.HasError() {
		t.Fatalf("ExtractStringMap: %v", diags)
	}
	if values != nil {
		t.Errorf("ExtractStringMap(null) = %v, want nil", values)
	}
}

func TestStringMapValueNilIsEmptyNotNull(t *testing.T) {
	m, diags := StringMapValue(context.Background(), nil)
	if diags.HasError() {
		t.Fatalf("StringMapValue: %v", diags)
	}
	if m.IsNull() || len(m.Elements()) != 0 {
		t.Errorf("StringMapValue(nil) = %v, want an empty non-null map", m)
	}
}
