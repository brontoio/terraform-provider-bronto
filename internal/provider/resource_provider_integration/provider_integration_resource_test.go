package resource_provider_integration

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	bronto "terraform-provider-bronto/internal"
)

func TestPopulateModelMapsAllFieldsAndLeavesSecretAlone(t *testing.T) {
	p := bronto.ProviderIntegration{
		Id:        "integration-1",
		Name:      "PagerDuty",
		Provider:  "pagerduty",
		Url:       "https://events.pagerduty.com/v2/enqueue",
		HasSecret: true,
		AuthType:  "api_key",
	}

	data := ProviderIntegrationModel{
		Secret: types.StringValue("whatever-was-already-there"),
	}
	populateModel(&data, p)

	if data.Id.ValueString() != "integration-1" {
		t.Errorf("Id = %s", data.Id.ValueString())
	}
	if data.ProviderType.ValueString() != "pagerduty" {
		t.Errorf("ProviderType = %s", data.ProviderType.ValueString())
	}
	if !data.HasSecret.ValueBool() {
		t.Errorf("HasSecret = false, want true")
	}
	if data.Secret.ValueString() != "whatever-was-already-there" {
		t.Errorf("Secret = %q, populateModel must never touch it", data.Secret.ValueString())
	}
}
