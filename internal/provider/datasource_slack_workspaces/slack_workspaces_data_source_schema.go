package datasource_slack_workspaces

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func SlackWorkspacesDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description:         "Lists Slack workspaces already connected to this org through the Bronto UI's OAuth flow.",
		MarkdownDescription: "Lists Slack workspaces already connected to this org through the Bronto UI's OAuth flow.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "Placeholder identifier for this data source.",
				MarkdownDescription: "Placeholder identifier for this data source.",
			},
			"workspaces": schema.ListNestedAttribute{
				Computed:            true,
				Description:         "The connected Slack workspaces.",
				MarkdownDescription: "The connected Slack workspaces.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							Description:         "The unique identifier of the workspace. This is an internal Bronto identifier, not Slack's own workspace/team ID.",
							MarkdownDescription: "The unique identifier of the workspace. This is an internal Bronto identifier, not Slack's own workspace/team ID.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							Description:         "The name of the Slack workspace.",
							MarkdownDescription: "The name of the Slack workspace.",
						},
					},
				},
			},
		},
	}
}

type WorkspaceModel struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type SlackWorkspacesModel struct {
	Id         types.String     `tfsdk:"id"`
	Workspaces []WorkspaceModel `tfsdk:"workspaces"`
}
