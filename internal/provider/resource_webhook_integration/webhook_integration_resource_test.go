package resource_webhook_integration

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	bronto "terraform-provider-bronto/internal"
)

func TestPopulateModelMapsAllFields(t *testing.T) {
	ctx := context.Background()

	w := bronto.WebhookIntegration{
		Id:            "webhook-1",
		Name:          "my-webhook",
		Url:           "https://webhook.site/abc",
		CustomHeaders: map[string]string{"x-bronto-api-key": "xxxxx"},
		Payload:       `{"name":"$title"}`,
	}

	var data WebhookIntegrationModel
	diags := populateModel(ctx, &data, w)
	if diags.HasError() {
		t.Fatalf("populateModel: %v", diags)
	}

	if data.Id.ValueString() != "webhook-1" {
		t.Errorf("Id = %s", data.Id.ValueString())
	}
	if data.Name.ValueString() != "my-webhook" {
		t.Errorf("Name = %s", data.Name.ValueString())
	}
	if data.Url.ValueString() != "https://webhook.site/abc" {
		t.Errorf("Url = %s", data.Url.ValueString())
	}
	if data.Payload.ValueString() != `{"name":"$title"}` {
		t.Errorf("Payload = %s", data.Payload.ValueString())
	}

	var headers map[string]string
	if diags := data.CustomHeaders.ElementsAs(ctx, &headers, false); diags.HasError() {
		t.Fatalf("CustomHeaders.ElementsAs: %v", diags)
	}
	if headers["x-bronto-api-key"] != "xxxxx" {
		t.Errorf("CustomHeaders = %v", headers)
	}
}

func TestPopulateModelEmptyPayloadAndHeadersAreNullOrEmpty(t *testing.T) {
	ctx := context.Background()

	var data WebhookIntegrationModel
	diags := populateModel(ctx, &data, bronto.WebhookIntegration{Id: "w", Name: "n", Url: "https://example.com"})
	if diags.HasError() {
		t.Fatalf("populateModel: %v", diags)
	}

	if !data.Payload.IsNull() {
		t.Errorf("Payload should be null when the API returns empty, got %q", data.Payload.ValueString())
	}
	if data.CustomHeaders.IsNull() {
		t.Errorf("CustomHeaders should be an empty map, not null, when there are no headers")
	}
}

func TestEncodeWebhookIntegrationRequest(t *testing.T) {
	ctx := context.Background()

	headers, diags := types.MapValueFrom(ctx, types.StringType, map[string]string{"authorization": "Bearer xxx"})
	if diags.HasError() {
		t.Fatalf("types.MapValueFrom: %v", diags)
	}

	data := WebhookIntegrationModel{
		Name:          types.StringValue("my-webhook"),
		Url:           types.StringValue("https://webhook.site/abc"),
		CustomHeaders: headers,
		Payload:       types.StringValue(`{"name":"$title"}`),
	}

	body, diags := encodeWebhookIntegrationRequest(ctx, data)
	if diags.HasError() {
		t.Fatalf("encodeWebhookIntegrationRequest: %v", diags)
	}

	var decoded bronto.WebhookIntegrationRequest
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("unmarshal request body: %v", err)
	}

	if decoded.Name != "my-webhook" {
		t.Errorf("Name = %s", decoded.Name)
	}
	if decoded.Url != "https://webhook.site/abc" {
		t.Errorf("Url = %s", decoded.Url)
	}
	if decoded.Payload != `{"name":"$title"}` {
		t.Errorf("Payload = %s", decoded.Payload)
	}
	if decoded.CustomHeaders["authorization"] != "Bearer xxx" {
		t.Errorf("CustomHeaders = %v", decoded.CustomHeaders)
	}
}

func TestEncodeWebhookIntegrationRequestOmitsPayloadWhenNull(t *testing.T) {
	ctx := context.Background()

	data := WebhookIntegrationModel{
		Name:          types.StringValue("my-webhook"),
		Url:           types.StringValue("https://webhook.site/abc"),
		CustomHeaders: types.MapNull(types.StringType),
		Payload:       types.StringNull(),
	}

	body, diags := encodeWebhookIntegrationRequest(ctx, data)
	if diags.HasError() {
		t.Fatalf("encodeWebhookIntegrationRequest: %v", diags)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		t.Fatalf("unmarshal request body: %v", err)
	}
	if _, ok := raw["payload"]; ok {
		t.Errorf("expected payload to be omitted, got: %s", raw["payload"])
	}
}

func TestExtractStringMapNull(t *testing.T) {
	values, diags := extractStringMap(context.Background(), types.MapNull(types.StringType))
	if diags.HasError() {
		t.Fatalf("extractStringMap: %v", diags)
	}
	if values != nil {
		t.Errorf("expected nil for a null map, got %v", values)
	}
}
