package resource_sso_idp

// This resource has no separate generated id: the API identifies an idp by
// its name (GET/PATCH/DELETE /sso/idps/{idp_name}), so name doubles as id
// and can't be changed after creation - only via destroy/recreate.
//
// The create and list endpoints (POST and GET /sso/idps) are flagged
// x-internal in the OpenAPI spec, unlike get/update/delete; whether that
// means they're unreachable with a normal API key hasn't been verified
// live yet.

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// providerTypes are the federated sign-in options accepted by Idp.provider_type.
var providerTypes = []string{"SAML", "OIDC"}

func SsoIdpResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description:         "Manages a Bronto SSO identity provider (SAML or OIDC).",
		MarkdownDescription: "Manages a Bronto SSO identity provider (SAML or OIDC).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The identity provider's name - this resource has no separate generated id, so id and name are always equal.",
				MarkdownDescription: "The identity provider's name - this resource has no separate generated id, so `id` and `name` are always equal.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "The name of the identity provider. Identifies the resource; cannot be changed after creation.",
				MarkdownDescription: "The name of the identity provider. Identifies the resource; cannot be changed after creation.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			// provider_type is the API's "provider" field, renamed here: a
			// schema attribute can't be called "provider" without
			// colliding with Terraform's reserved provider = ... resource
			// meta-argument.
			"provider_type": schema.StringAttribute{
				Required:            true,
				Description:         "The federated sign-in option: SAML or OIDC.",
				MarkdownDescription: "The federated sign-in option: `SAML` or `OIDC`.",
				Validators: []validator.String{
					stringvalidator.OneOf(providerTypes...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"identifiers": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "Email address domains that redirect a user to this identity provider for sign-in, e.g. [\"bronto.io\"].",
				MarkdownDescription: "Email address domains that redirect a user to this identity provider for sign-in, e.g. `[\"bronto.io\"]`.",
			},
			"meta_cert": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The SAML identity provider's metadata XML (its signing certificate and endpoints). Required for SAML providers.",
				MarkdownDescription: "The SAML identity provider's metadata XML (its signing certificate and endpoints). Required for SAML providers.",
			},
			"issuer_url": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The base URL of the OIDC identity provider, e.g. https://dev-123456.okta.com/oauth2/default. Required for OIDC providers.",
				MarkdownDescription: "The base URL of the OIDC identity provider, e.g. `https://dev-123456.okta.com/oauth2/default`. Required for OIDC providers.",
			},
			"client_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The client id this application was registered with on the OIDC identity provider. Required for OIDC providers.",
				MarkdownDescription: "The client id this application was registered with on the OIDC identity provider. Required for OIDC providers.",
			},
			"client_secret": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
				Description:         "The client secret used to authenticate with the OIDC identity provider. Required for OIDC providers.",
				MarkdownDescription: "The client secret used to authenticate with the OIDC identity provider. Required for OIDC providers.",
			},
			"accept_idp_initiated": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Description: "When enabled, sign-in sessions can be initiated from the identity provider rather than from Bronto. Not recommended: " +
					"this makes the integration more susceptible to spoofing and cross-site request forgery.",
				MarkdownDescription: "When enabled, sign-in sessions can be initiated from the identity provider rather than from Bronto. Not recommended: " +
					"this makes the integration more susceptible to spoofing and cross-site request forgery.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"app_client_id": schema.StringAttribute{
				Computed:            true,
				Description:         "The AWS Cognito application client id used to build this identity provider's sign-in start URL.",
				MarkdownDescription: "The AWS Cognito application client id used to build this identity provider's sign-in start URL.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"attribute_mappings": schema.ListNestedAttribute{
				Required:            true,
				Description:         "Maps identity provider attributes to Bronto user attributes.",
				MarkdownDescription: "Maps identity provider attributes to Bronto user attributes.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"provider_attr": schema.StringAttribute{
							Required:            true,
							Description:         "The attribute name from the identity provider to map from, e.g. givenname.",
							MarkdownDescription: "The attribute name from the identity provider to map from, e.g. `givenname`.",
						},
						"user_attr": schema.StringAttribute{
							Required:            true,
							Description:         "The Bronto user attribute to map to, e.g. name.",
							MarkdownDescription: "The Bronto user attribute to map to, e.g. `name`.",
						},
					},
				},
			},
		},
	}
}

type AttributeMappingModel struct {
	ProviderAttr types.String `tfsdk:"provider_attr"`
	UserAttr     types.String `tfsdk:"user_attr"`
}

type SsoIdpModel struct {
	Id                 types.String            `tfsdk:"id"`
	Name               types.String            `tfsdk:"name"`
	ProviderType       types.String            `tfsdk:"provider_type"`
	Identifiers        types.List              `tfsdk:"identifiers"`
	MetaCert           types.String            `tfsdk:"meta_cert"`
	IssuerUrl          types.String            `tfsdk:"issuer_url"`
	ClientId           types.String            `tfsdk:"client_id"`
	ClientSecret       types.String            `tfsdk:"client_secret"`
	AcceptIdpInitiated types.Bool              `tfsdk:"accept_idp_initiated"`
	AppClientId        types.String            `tfsdk:"app_client_id"`
	AttributeMappings  []AttributeMappingModel `tfsdk:"attribute_mappings"`
}
