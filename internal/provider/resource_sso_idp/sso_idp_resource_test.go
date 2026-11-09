package resource_sso_idp

import (
	"context"
	"testing"

	bronto "terraform-provider-bronto/internal"
)

func TestPopulateModelMapsAllFields(t *testing.T) {
	ctx := context.Background()

	idp := bronto.Idp{
		Name:               "Google",
		ProviderType:       "SAML",
		Identifiers:        []string{"bronto.io"},
		MetaCert:           "<xml/>",
		AcceptIdpInitiated: true,
		AppClientId:        "4s25sufdsie21brgq2kt5cs9dj",
		AttributeMappings: []bronto.AttributeMapping{
			{ProviderAttr: "givenname", UserAttr: "name"},
		},
	}

	var data SsoIdpModel
	diags := populateModel(ctx, &data, idp)
	if diags.HasError() {
		t.Fatalf("populateModel: %v", diags)
	}

	if data.Id.ValueString() != "Google" {
		t.Errorf("Id = %s, want it to equal Name", data.Id.ValueString())
	}
	if data.ProviderType.ValueString() != "SAML" {
		t.Errorf("ProviderType = %s", data.ProviderType.ValueString())
	}
	if !data.AcceptIdpInitiated.ValueBool() {
		t.Errorf("AcceptIdpInitiated = false, want true")
	}
	if len(data.AttributeMappings) != 1 || data.AttributeMappings[0].ProviderAttr.ValueString() != "givenname" {
		t.Errorf("AttributeMappings = %+v", data.AttributeMappings)
	}
}

func TestPopulateModelEmptyOptionalFieldsAreNull(t *testing.T) {
	idp := bronto.Idp{Name: "Okta", ProviderType: "OIDC"}

	var data SsoIdpModel
	diags := populateModel(context.Background(), &data, idp)
	if diags.HasError() {
		t.Fatalf("populateModel: %v", diags)
	}

	if !data.MetaCert.IsNull() {
		t.Errorf("MetaCert should be null when the API returns empty, got %q", data.MetaCert.ValueString())
	}
	if !data.ClientSecret.IsNull() {
		t.Errorf("ClientSecret should be null when the API returns empty")
	}
	if data.Identifiers.IsNull() {
		t.Errorf("Identifiers should be an empty list, not null, when the API returns none")
	}
}
