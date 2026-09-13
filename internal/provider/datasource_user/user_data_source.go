package datasource_user

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	bronto "terraform-provider-bronto/internal"
)

var (
	_ datasource.DataSource                     = &userDataSource{}
	_ datasource.DataSourceWithConfigure        = &userDataSource{}
	_ datasource.DataSourceWithConfigValidators = &userDataSource{}
)

func NewUserDataSource() datasource.DataSource {
	return &userDataSource{}
}

type userDataSource struct {
	client *bronto.Client
}

func (d *userDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *userDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = UserDataSourceSchema(ctx)
}

func (d *userDataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.ExactlyOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("email"),
		),
	}
}

func (d *userDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *userDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UserModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var user bronto.User
	var err error

	if !data.Id.IsNull() {
		user, err = d.findById(ctx, data.Id.ValueString())
		if err != nil {
			if bronto.IsNotFound(err) {
				resp.Diagnostics.AddAttributeError(
					path.Root("id"),
					"User Not Found",
					fmt.Sprintf("No user was found with id %q.", data.Id.ValueString()),
				)
				return
			}
			resp.Diagnostics.AddError(
				"Unable to Read User",
				"An unexpected error occurred while calling the read API. Error: "+err.Error(),
			)
			return
		}
	} else {
		user, err = d.findByEmail(ctx, data.Email.ValueString())
		if err != nil {
			resp.Diagnostics.AddAttributeError(path.Root("email"), "Unable to Find User", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(populateModel(ctx, &data, user)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *userDataSource) findById(ctx context.Context, id string) (bronto.User, error) {
	httpResp, err := d.client.ReadRequest(ctx, "/users/"+id)
	if err != nil {
		return bronto.User{}, err
	}
	return bronto.DecodeUser(httpResp)
}

// findByEmail lists all users and filters client-side, since the API has no
// get-user-by-email endpoint.
func (d *userDataSource) findByEmail(ctx context.Context, email string) (bronto.User, error) {
	httpResp, err := d.client.ReadRequest(ctx, "/users")
	if err != nil {
		return bronto.User{}, fmt.Errorf("listing users: %w", err)
	}

	list, err := bronto.DecodeUsersList(httpResp)
	if err != nil {
		return bronto.User{}, fmt.Errorf("parsing users list response: %w", err)
	}

	var match *bronto.User
	for i := range list.Users {
		if strings.EqualFold(list.Users[i].Email, email) {
			if match != nil {
				return bronto.User{}, fmt.Errorf("multiple users found with email %q", email)
			}
			match = &list.Users[i]
		}
	}
	if match == nil {
		return bronto.User{}, fmt.Errorf("no user found with email %q", email)
	}
	return *match, nil
}

// populateModel maps an API user response onto the data source's Terraform
// model, including the id/email lookup keys.
func populateModel(ctx context.Context, data *UserModel, user bronto.User) diag.Diagnostics {
	var diags diag.Diagnostics

	data.Id = types.StringValue(user.Id)
	data.Email = types.StringValue(user.Email)
	data.FirstName = types.StringValue(user.FirstName)
	data.LastName = types.StringValue(user.LastName)
	data.Status = types.StringValue(user.Status)
	data.Active = types.BoolValue(user.Status == bronto.UserStatusActive)

	rolesVal, rolesDiags := types.ListValueFrom(ctx, types.StringType, user.Roles)
	diags.Append(rolesDiags...)
	data.Roles = rolesVal

	tagsVal, tagsDiags := types.MapValueFrom(ctx, types.StringType, user.Tags)
	diags.Append(tagsDiags...)
	data.Tags = tagsVal

	lastLoginsVal, lastLoginsDiags := types.MapValueFrom(ctx, types.Int64Type, user.LastLogins)
	diags.Append(lastLoginsDiags...)
	data.LastLogins = lastLoginsVal

	if user.InvitedAt != nil {
		data.InvitedAt = types.Int64Value(*user.InvitedAt)
	} else {
		data.InvitedAt = types.Int64Null()
	}

	if user.InvitedBy != nil {
		data.InvitedBy = types.StringValue(*user.InvitedBy)
	} else {
		data.InvitedBy = types.StringNull()
	}

	loginMethodsVal, loginMethodsDiags := toLoginMethodsList(user.LoginMethods)
	diags.Append(loginMethodsDiags...)
	data.LoginMethods = loginMethodsVal

	return diags
}

var loginMethodAttrTypes = map[string]attr.Type{
	"name":    types.StringType,
	"enabled": types.BoolType,
}

func toLoginMethodsList(methods []bronto.LoginMethod) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	elements := make([]attr.Value, 0, len(methods))
	for _, m := range methods {
		obj, objDiags := types.ObjectValue(loginMethodAttrTypes, map[string]attr.Value{
			"name":    types.StringValue(m.Name),
			"enabled": types.BoolValue(m.Enabled),
		})
		diags.Append(objDiags...)
		elements = append(elements, obj)
	}

	list, listDiags := types.ListValue(types.ObjectType{AttrTypes: loginMethodAttrTypes}, elements)
	diags.Append(listDiags...)
	return list, diags
}
