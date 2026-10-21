package datasource_dataset

import (
	"context"
	"testing"

	bronto "terraform-provider-bronto/internal"
)

func TestPopulateModelMapsAllFields(t *testing.T) {
	ctx := context.Background()

	createdAt := int64(1709251200000)

	d := bronto.Dataset{
		Id:                   "dataset-1",
		Collection:           "demo",
		DatasetName:          "firewall",
		IsSystemGenerated:    false,
		Tags:                 map[string]string{"region": "eu"},
		ParserId:             "parser-1",
		ParserMatchingExempt: true,
		Metadata: &bronto.DatasetMetadata{
			CreatedAt: &createdAt,
			CreatedBy: &bronto.Identity{Type: "USER_ID", Id: "user-1"},
		},
	}

	var data DatasetModel
	diags := populateModel(ctx, &data, d)
	if diags.HasError() {
		t.Fatalf("populateModel: %v", diags)
	}

	if data.Id.ValueString() != "dataset-1" {
		t.Errorf("Id = %s", data.Id.ValueString())
	}
	if data.Collection.ValueString() != "demo" {
		t.Errorf("Collection = %s", data.Collection.ValueString())
	}
	if data.Dataset.ValueString() != "firewall" {
		t.Errorf("Dataset = %s", data.Dataset.ValueString())
	}
	if data.ParserId.ValueString() != "parser-1" {
		t.Errorf("ParserId = %s", data.ParserId.ValueString())
	}
	if !data.ParserMatchingExempt.ValueBool() {
		t.Errorf("ParserMatchingExempt = %v, want true", data.ParserMatchingExempt.ValueBool())
	}

	var tags map[string]string
	if diags := data.Tags.ElementsAs(ctx, &tags, false); diags.HasError() {
		t.Fatalf("Tags.ElementsAs: %v", diags)
	}
	if tags["region"] != "eu" {
		t.Errorf("Tags = %v", tags)
	}

	if data.Metadata == nil {
		t.Fatalf("Metadata should not be nil")
	}
	if data.Metadata.CreatedAt.ValueInt64() != createdAt {
		t.Errorf("Metadata.CreatedAt = %d", data.Metadata.CreatedAt.ValueInt64())
	}
	if data.Metadata.CreatedBy == nil || data.Metadata.CreatedBy.Id.ValueString() != "user-1" {
		t.Errorf("Metadata.CreatedBy = %+v", data.Metadata.CreatedBy)
	}
	if data.Metadata.ModifiedBy != nil {
		t.Errorf("Metadata.ModifiedBy should be nil when absent, got %+v", data.Metadata.ModifiedBy)
	}
	if !data.Metadata.DeletedAt.IsNull() {
		t.Errorf("Metadata.DeletedAt should be null when absent")
	}
}

func TestPopulateModelEmptyParserIdIsNull(t *testing.T) {
	var data DatasetModel
	diags := populateModel(context.Background(), &data, bronto.Dataset{Id: "d", Collection: "c", DatasetName: "n"})
	if diags.HasError() {
		t.Fatalf("populateModel: %v", diags)
	}
	if !data.ParserId.IsNull() {
		t.Errorf("ParserId should be null when the API returns empty, got %q", data.ParserId.ValueString())
	}
	if data.Metadata != nil {
		t.Errorf("Metadata should be nil when the API returns none, got %+v", data.Metadata)
	}
}

func TestEscapeExprLiteral(t *testing.T) {
	if got := escapeExprLiteral("o'brien"); got != "o''brien" {
		t.Errorf("escapeExprLiteral = %q, want %q", got, "o''brien")
	}
	if got := escapeExprLiteral("plain"); got != "plain" {
		t.Errorf("escapeExprLiteral = %q, want %q", got, "plain")
	}
}
