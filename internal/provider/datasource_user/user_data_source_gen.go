package datasource_user

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func UserDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description:         "Look up a Bronto user.",
		MarkdownDescription: "Look up a Bronto user.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The unique identifier for the user. Exactly one of id or email must be set.",
				MarkdownDescription: "The unique identifier for the user. Exactly one of `id` or `email` must be set.",
			},
			"email": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The user's email address. Exactly one of id or email must be set.",
				MarkdownDescription: "The user's email address. Exactly one of `id` or `email` must be set.",
			},
			"first_name": schema.StringAttribute{
				Computed:            true,
				Description:         "The user's first name",
				MarkdownDescription: "The user's first name",
			},
			"last_name": schema.StringAttribute{
				Computed:            true,
				Description:         "The user's last name",
				MarkdownDescription: "The user's last name",
			},
			"roles": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				Description:         "A list of role ids",
				MarkdownDescription: "A list of role ids",
			},
			"active": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether the user account is active.",
				MarkdownDescription: "Whether the user account is active.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				Description:         "Current status of the user",
				MarkdownDescription: "Current status of the user",
			},
			"tags": schema.MapAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				Description:         "A map of key value pairs associated with the resource",
				MarkdownDescription: "A map of key value pairs associated with the resource",
			},
			"last_logins": schema.MapAttribute{
				ElementType:         types.Int64Type,
				Computed:            true,
				Description:         "A map of login methods to their last login timestamps (Unix epoch).",
				MarkdownDescription: "A map of login methods to their last login timestamps (Unix epoch).",
			},
			"login_methods": schema.ListNestedAttribute{
				Computed:            true,
				Description:         "Per-user overrides for enabling or disabling specific login methods",
				MarkdownDescription: "Per-user overrides for enabling or disabling specific login methods",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:            true,
							Description:         "The name of the login method",
							MarkdownDescription: "The name of the login method",
						},
						"enabled": schema.BoolAttribute{
							Computed:            true,
							Description:         "Flag to enable or disable a login method override for this specific user.",
							MarkdownDescription: "Flag to enable or disable a login method override for this specific user.",
						},
					},
				},
			},
			"invited_at": schema.Int64Attribute{
				Computed:            true,
				Description:         "Unix timestamp (seconds) of the last invitation sent to the user",
				MarkdownDescription: "Unix timestamp (seconds) of the last invitation sent to the user",
			},
			"invited_by": schema.StringAttribute{
				Computed:            true,
				Description:         "Identifier for the user who sent the latest invitation",
				MarkdownDescription: "Identifier for the user who sent the latest invitation",
			},
		},
	}
}

type UserModel struct {
	Id           types.String `tfsdk:"id"`
	Email        types.String `tfsdk:"email"`
	FirstName    types.String `tfsdk:"first_name"`
	LastName     types.String `tfsdk:"last_name"`
	Roles        types.List   `tfsdk:"roles"`
	Active       types.Bool   `tfsdk:"active"`
	Status       types.String `tfsdk:"status"`
	Tags         types.Map    `tfsdk:"tags"`
	LastLogins   types.Map    `tfsdk:"last_logins"`
	LoginMethods types.List   `tfsdk:"login_methods"`
	InvitedAt    types.Int64  `tfsdk:"invited_at"`
	InvitedBy    types.String `tfsdk:"invited_by"`
}
