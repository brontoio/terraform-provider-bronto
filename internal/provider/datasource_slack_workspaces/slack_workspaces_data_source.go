package datasource_slack_workspaces

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	bronto "terraform-provider-bronto/internal"
)

var (
	_ datasource.DataSource              = &slackWorkspacesDataSource{}
	_ datasource.DataSourceWithConfigure = &slackWorkspacesDataSource{}
)

func NewSlackWorkspacesDataSource() datasource.DataSource {
	return &slackWorkspacesDataSource{}
}

type slackWorkspacesDataSource struct {
	client *bronto.Client
}

func (d *slackWorkspacesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_slack_workspaces"
}

func (d *slackWorkspacesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = SlackWorkspacesDataSourceSchema(ctx)
}

func (d *slackWorkspacesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*bronto.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *bronto.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *slackWorkspacesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SlackWorkspacesModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := d.client.ReadRequest(ctx, "/integrations/slack/workspaces")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Slack Workspaces",
			"An unexpected error occurred while calling the read API. Error: "+err.Error(),
		)
		return
	}

	list, err := bronto.DecodeSlackWorkspacesList(httpResp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Slack Workspaces",
			"An unexpected error occurred while parsing the read response. JSON Error: "+err.Error(),
		)
		return
	}

	data.Id = types.StringValue("slack_workspaces")
	data.Workspaces = make([]WorkspaceModel, 0, len(list.SlackWorkspaces))
	for _, w := range list.SlackWorkspaces {
		data.Workspaces = append(data.Workspaces, WorkspaceModel{
			Id:   types.StringValue(w.Id),
			Name: types.StringValue(w.Name),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
