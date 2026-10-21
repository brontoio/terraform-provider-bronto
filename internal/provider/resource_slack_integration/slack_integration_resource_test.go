package resource_slack_integration

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	bronto "terraform-provider-bronto/internal"
)

func TestPopulateModelMapsAllFields(t *testing.T) {
	s := bronto.SlackIntegration{
		Id:          "slack-1",
		Name:        "api_alerts",
		WorkspaceId: "workspace-1",
		Channels: []bronto.SlackChannel{
			{Name: "production-alerts", IncludeGraphImage: true},
			{Name: "usage-alerts", IncludeGraphImage: false},
		},
	}

	var data SlackIntegrationModel
	diags := populateModel(&data, s)
	if diags.HasError() {
		t.Fatalf("populateModel: %v", diags)
	}

	if data.Id.ValueString() != "slack-1" {
		t.Errorf("Id = %s", data.Id.ValueString())
	}
	if data.Name.ValueString() != "api_alerts" {
		t.Errorf("Name = %s", data.Name.ValueString())
	}
	if data.WorkspaceId.ValueString() != "workspace-1" {
		t.Errorf("WorkspaceId = %s", data.WorkspaceId.ValueString())
	}
	if len(data.Channels) != 2 {
		t.Fatalf("Channels = %d entries, want 2", len(data.Channels))
	}
	if data.Channels[0].Name.ValueString() != "production-alerts" || !data.Channels[0].IncludeGraphImage.ValueBool() {
		t.Errorf("Channels[0] = %+v", data.Channels[0])
	}
	if data.Channels[1].Name.ValueString() != "usage-alerts" || data.Channels[1].IncludeGraphImage.ValueBool() {
		t.Errorf("Channels[1] = %+v", data.Channels[1])
	}
}

func TestEncodeSlackIntegrationRequest(t *testing.T) {
	data := SlackIntegrationModel{
		Name:        types.StringValue("api_alerts"),
		WorkspaceId: types.StringValue("workspace-1"),
		Channels: []ChannelModel{
			{Name: types.StringValue("production-alerts"), IncludeGraphImage: types.BoolValue(true)},
		},
	}

	body, diags := encodeSlackIntegrationRequest(data)
	if diags.HasError() {
		t.Fatalf("encodeSlackIntegrationRequest: %v", diags)
	}

	var decoded bronto.SlackIntegrationRequest
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("unmarshal request body: %v", err)
	}

	if decoded.Name != "api_alerts" {
		t.Errorf("Name = %s", decoded.Name)
	}
	if decoded.WorkspaceId != "workspace-1" {
		t.Errorf("WorkspaceId = %s", decoded.WorkspaceId)
	}
	if len(decoded.Channels) != 1 || decoded.Channels[0].Name != "production-alerts" || !decoded.Channels[0].IncludeGraphImage {
		t.Errorf("Channels = %+v", decoded.Channels)
	}
}
