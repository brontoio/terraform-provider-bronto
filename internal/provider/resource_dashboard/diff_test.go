package resource_dashboard

import (
	"reflect"
	"sort"
	"testing"
)

func sortedCopy(s []string) []string {
	out := append([]string(nil), s...)
	sort.Strings(out)
	return out
}

func TestDiffStringSets(t *testing.T) {
	cases := []struct {
		name       string
		current    []string
		desired    []string
		wantAdd    []string
		wantRemove []string
	}{
		{
			name:       "no change",
			current:    []string{"a", "b"},
			desired:    []string{"a", "b"},
			wantAdd:    nil,
			wantRemove: nil,
		},
		{
			name:       "add only",
			current:    []string{"a"},
			desired:    []string{"a", "b"},
			wantAdd:    []string{"b"},
			wantRemove: nil,
		},
		{
			name:       "remove only",
			current:    []string{"a", "b"},
			desired:    []string{"a"},
			wantAdd:    nil,
			wantRemove: []string{"b"},
		},
		{
			name:       "add and remove",
			current:    []string{"a", "b"},
			desired:    []string{"b", "c"},
			wantAdd:    []string{"c"},
			wantRemove: []string{"a"},
		},
		{
			name:       "empty to populated",
			current:    nil,
			desired:    []string{"a"},
			wantAdd:    []string{"a"},
			wantRemove: nil,
		},
		{
			name:       "populated to empty",
			current:    []string{"a"},
			desired:    nil,
			wantAdd:    nil,
			wantRemove: []string{"a"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotAdd, gotRemove := diffStringSets(tc.current, tc.desired)
			if !reflect.DeepEqual(sortedCopy(gotAdd), sortedCopy(tc.wantAdd)) {
				t.Errorf("toAdd = %v, want %v", gotAdd, tc.wantAdd)
			}
			if !reflect.DeepEqual(sortedCopy(gotRemove), sortedCopy(tc.wantRemove)) {
				t.Errorf("toRemove = %v, want %v", gotRemove, tc.wantRemove)
			}
		})
	}
}
