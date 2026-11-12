package resource_group

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// memberTypes are the kinds of entity that can belong to a group, per the
// Groups API's member_type enum.
var memberTypes = []string{"USER", "GROUP"}

func GroupResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description:         "Manages a Bronto user group.",
		MarkdownDescription: "Manages a Bronto user group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The unique identifier for the group",
				MarkdownDescription: "The unique identifier for the group",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "The display name of the group",
				MarkdownDescription: "The display name of the group",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "A brief summary or explanation of the group's purpose",
				MarkdownDescription: "A brief summary or explanation of the group's purpose",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.Int64Attribute{
				Computed:            true,
				Description:         "The timestamp when the group was created, in epoch milliseconds",
				MarkdownDescription: "The timestamp when the group was created, in epoch milliseconds",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"modified_at": schema.Int64Attribute{
				Computed:            true,
				Description:         "The timestamp when the group was last modified, in epoch milliseconds",
				MarkdownDescription: "The timestamp when the group was last modified, in epoch milliseconds",
			},
			"members": schema.ListNestedAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Users or groups that belong to this group. Reconciled via individual add/remove API calls, not the group's own create/update endpoints, so each changed member costs a separate request.",
				MarkdownDescription: "Users or groups that belong to this group. Reconciled via individual add/remove API calls, not the group's own create/update endpoints, so each changed member costs a separate request.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"member_type": schema.StringAttribute{
							Required:            true,
							Description:         "The kind of entity this member is: USER or GROUP",
							MarkdownDescription: "The kind of entity this member is: `USER` or `GROUP`",
							Validators: []validator.String{
								stringvalidator.OneOf(memberTypes...),
							},
						},
						"member_id": schema.StringAttribute{
							Required:            true,
							Description:         "The id of the user or group that is a member of this group",
							MarkdownDescription: "The id of the user or group that is a member of this group",
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(100),
				},
			},
		},
	}
}

type MemberModel struct {
	MemberType types.String `tfsdk:"member_type"`
	MemberId   types.String `tfsdk:"member_id"`
}

type GroupModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	CreatedAt   types.Int64  `tfsdk:"created_at"`
	ModifiedAt  types.Int64  `tfsdk:"modified_at"`
	Members     types.List   `tfsdk:"members"`
}

var memberObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"member_type": types.StringType,
		"member_id":   types.StringType,
	},
}
