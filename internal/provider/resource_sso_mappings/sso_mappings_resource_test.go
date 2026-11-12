package resource_sso_mappings

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	bronto "terraform-provider-bronto/internal"
)

func TestPopulateModelMapsAllFields(t *testing.T) {
	mappings := []bronto.SsoMapping{
		{ProviderKey: "member_of", ProviderValue: "Developers", AssignAttr: "Standard"},
		{ProviderKey: "member_of", ProviderValue: "Admins", AssignAttr: "Admin"},
	}

	var data SsoMappingsModel
	populateModel(&data, mappings)

	if data.Id.ValueString() != singletonId {
		t.Errorf("Id = %s, want %s", data.Id.ValueString(), singletonId)
	}
	if len(data.Mappings) != 2 {
		t.Fatalf("Mappings = %d entries, want 2", len(data.Mappings))
	}
	if data.Mappings[0].ProviderValue.ValueString() != "Developers" || data.Mappings[0].AssignAttr.ValueString() != "Standard" {
		t.Errorf("Mappings[0] = %+v", data.Mappings[0])
	}
}

func TestEncodeRequestRoundTrips(t *testing.T) {
	data := SsoMappingsModel{
		Mappings: []SsoMappingModel{
			{
				ProviderKey:   types.StringValue("member_of"),
				ProviderValue: types.StringValue("Developers"),
				AssignAttr:    types.StringValue("Standard"),
			},
		},
	}

	body, diags := encodeRequest(data)
	if diags.HasError() {
		t.Fatalf("encodeRequest: %v", diags)
	}

	decoded, err := bronto.DecodeSsoMappingsRequest(string(body))
	if err != nil {
		t.Fatalf("DecodeSsoMappingsRequest: %v", err)
	}
	if len(decoded.Mappings) != 1 || decoded.Mappings[0].ProviderKey != "member_of" {
		t.Errorf("decoded mappings = %+v", decoded.Mappings)
	}
}
