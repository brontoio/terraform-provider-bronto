package datasource_dataset

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	bronto "terraform-provider-bronto/internal"
)

var (
	_ datasource.DataSource                     = &datasetDataSource{}
	_ datasource.DataSourceWithConfigure        = &datasetDataSource{}
	_ datasource.DataSourceWithConfigValidators = &datasetDataSource{}
)

func NewDatasetDataSource() datasource.DataSource {
	return &datasetDataSource{}
}

type datasetDataSource struct {
	client *bronto.Client
}

func (d *datasetDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dataset"
}

func (d *datasetDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = DatasetDataSourceSchema(ctx)
}

func (d *datasetDataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.RequiredTogether(
			path.MatchRoot("collection"),
			path.MatchRoot("dataset"),
		),
		datasourcevalidator.ExactlyOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("collection"),
		),
	}
}

func (d *datasetDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *datasetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DatasetModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var dataset bronto.Dataset
	var err error

	if !data.Id.IsNull() {
		dataset, err = d.findById(ctx, data.Id.ValueString())
		if err != nil {
			if bronto.IsNotFound(err) {
				resp.Diagnostics.AddAttributeError(
					path.Root("id"),
					"Dataset Not Found",
					fmt.Sprintf("No dataset was found with id %q.", data.Id.ValueString()),
				)
				return
			}
			resp.Diagnostics.AddError(
				"Unable to Read Dataset",
				"An unexpected error occurred while calling the read API. Error: "+err.Error(),
			)
			return
		}
	} else {
		dataset, err = d.findByName(ctx, data.Collection.ValueString(), data.Dataset.ValueString())
		if err != nil {
			resp.Diagnostics.AddAttributeError(path.Root("dataset"), "Unable to Find Dataset", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(populateModel(ctx, &data, dataset)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *datasetDataSource) findById(ctx context.Context, id string) (bronto.Dataset, error) {
	httpResp, err := d.client.ReadRequest(ctx, "/datasets/"+id)
	if err != nil {
		return bronto.Dataset{}, err
	}
	return bronto.DecodeDataset(httpResp)
}

// findByName lists datasets filtered server-side via from_expr and matches
// client-side, since the API has no get-dataset-by-name endpoint - the same
// approach bronto_user uses for its email lookup, since that API has no
// get-user-by-email endpoint either.
func (d *datasetDataSource) findByName(ctx context.Context, collection, dataset string) (bronto.Dataset, error) {
	expr := fmt.Sprintf("collection = '%s' AND dataset = '%s'", escapeExprLiteral(collection), escapeExprLiteral(dataset))

	query := url.Values{}
	query.Set("from_expr", expr)

	httpResp, err := d.client.ReadRequest(ctx, "/datasets?"+query.Encode())
	if err != nil {
		return bronto.Dataset{}, fmt.Errorf("listing datasets: %w", err)
	}

	list, err := bronto.DecodeDatasetsList(httpResp)
	if err != nil {
		return bronto.Dataset{}, fmt.Errorf("parsing datasets list response: %w", err)
	}

	if len(list.Datasets) == 0 {
		return bronto.Dataset{}, fmt.Errorf("no dataset found with collection %q and dataset %q", collection, dataset)
	}
	if len(list.Datasets) > 1 {
		return bronto.Dataset{}, fmt.Errorf("multiple datasets found with collection %q and dataset %q", collection, dataset)
	}
	return list.Datasets[0], nil
}

// escapeExprLiteral doubles any single quote in a value embedded in a
// from_expr string literal, so a collection/dataset name containing a quote
// can't break out of the quoted literal.
func escapeExprLiteral(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

// populateModel maps an API dataset response onto the data source's
// Terraform model, including the id/collection/dataset lookup keys.
func populateModel(ctx context.Context, data *DatasetModel, d bronto.Dataset) diag.Diagnostics {
	var diags diag.Diagnostics

	data.Id = types.StringValue(d.Id)
	data.Collection = types.StringValue(d.Collection)
	data.Dataset = types.StringValue(d.DatasetName)
	data.IsSystemGenerated = types.BoolValue(d.IsSystemGenerated)
	data.ParserMatchingExempt = types.BoolValue(d.ParserMatchingExempt)

	if d.ParserId == "" {
		data.ParserId = types.StringNull()
	} else {
		data.ParserId = types.StringValue(d.ParserId)
	}

	tagsVal, tagsDiags := types.MapValueFrom(ctx, types.StringType, d.Tags)
	diags.Append(tagsDiags...)
	data.Tags = tagsVal

	data.Metadata = toMetadataModel(d.Metadata)

	return diags
}

func toMetadataModel(m *bronto.DatasetMetadata) *MetadataModel {
	if m == nil {
		return nil
	}

	return &MetadataModel{
		CreatedAt:       int64PtrValue(m.CreatedAt),
		CreatedBy:       toIdentityModel(m.CreatedBy),
		ModifiedAt:      int64PtrValue(m.ModifiedAt),
		ModifiedBy:      toIdentityModel(m.ModifiedBy),
		DeletedAt:       int64PtrValue(m.DeletedAt),
		DeletedBy:       toIdentityModel(m.DeletedBy),
		LastHeartbeatAt: int64PtrValue(m.LastHeartbeatAt),
	}
}

func toIdentityModel(i *bronto.Identity) *IdentityModel {
	if i == nil {
		return nil
	}
	return &IdentityModel{
		Type: types.StringValue(i.Type),
		Id:   types.StringValue(i.Id),
	}
}

func int64PtrValue(v *int64) types.Int64 {
	if v == nil {
		return types.Int64Null()
	}
	return types.Int64Value(*v)
}
