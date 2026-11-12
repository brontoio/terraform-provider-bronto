package resource_sso_mappings

// This is a singleton: the API has one org-wide list of SSO mappings
// (GET/POST/PUT/DELETE /sso/mappings), not individually addressable
// mapping resources, so this resource manages the whole list at once. Its
// id is a constant, not server-generated.
//
// The list/create endpoints are flagged x-internal in the OpenAPI spec
// (update/delete aren't); whether that means they're unreachable with a
// normal API key hasn't been verified live yet. The delete endpoint's spec
// also documents no request body or parameters, so Delete is implemented
// as "clear every mapping" - unverified live.

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func SsoMappingsResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description:         "Manages the full set of Bronto SSO mappings (identity provider attribute values mapped to Bronto roles or groups) for the organization. This is a singleton: only one bronto_sso_mappings resource should be declared per org.",
		MarkdownDescription: "Manages the full set of Bronto SSO mappings (identity provider attribute values mapped to Bronto roles or groups) for the organization. This is a singleton: only one `bronto_sso_mappings` resource should be declared per org.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "A constant identifier for this singleton resource.",
				MarkdownDescription: "A constant identifier for this singleton resource.",
			},
			"mappings": schema.ListNestedAttribute{
				Required:            true,
				Description:         "The full list of SSO mappings for the organization.",
				MarkdownDescription: "The full list of SSO mappings for the organization.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"provider_key": schema.StringAttribute{
							Required:            true,
							Description:         "The identity provider attribute key to read the value from, e.g. member_of.",
							MarkdownDescription: "The identity provider attribute key to read the value from, e.g. `member_of`.",
						},
						"provider_value": schema.StringAttribute{
							Required:            true,
							Description:         "The identity provider attribute value to match, e.g. Developers.",
							MarkdownDescription: "The identity provider attribute value to match, e.g. `Developers`.",
						},
						"assign_attr": schema.StringAttribute{
							Required:            true,
							Description:         "The Bronto role or group to assign when provider_key/provider_value matches.",
							MarkdownDescription: "The Bronto role or group to assign when `provider_key`/`provider_value` matches.",
						},
					},
				},
			},
		},
	}
}

type SsoMappingModel struct {
	ProviderKey   types.String `tfsdk:"provider_key"`
	ProviderValue types.String `tfsdk:"provider_value"`
	AssignAttr    types.String `tfsdk:"assign_attr"`
}

type SsoMappingsModel struct {
	Id       types.String      `tfsdk:"id"`
	Mappings []SsoMappingModel `tfsdk:"mappings"`
}
