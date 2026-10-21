package internal

import (
	"encoding/json"
	"strings"
)

// Identity identifies the actor responsible for an operation, per the
// Datasets API's Identity schema.
type Identity struct {
	Type string `json:"type"`
	Id   string `json:"id"`
}

// DatasetMetadata is the common metadata attached to a Dataset.
type DatasetMetadata struct {
	CreatedAt  *int64    `json:"created_at,omitempty"`
	CreatedBy  *Identity `json:"created_by,omitempty"`
	ModifiedAt *int64    `json:"modified_at,omitempty"`
	ModifiedBy *Identity `json:"modified_by,omitempty"`
	DeletedAt  *int64    `json:"deleted_at,omitempty"`
	DeletedBy  *Identity `json:"deleted_by,omitempty"`
}

// Dataset is the wire representation of the Datasets API's Dataset schema.
type Dataset struct {
	Id                   string            `json:"id"`
	Collection           string            `json:"collection"`
	DatasetName          string            `json:"dataset"`
	IsSystemGenerated    bool              `json:"is_system_generated"`
	Tags                 map[string]string `json:"tags"`
	ParserId             string            `json:"parser_id,omitempty"`
	ParserMatchingExempt bool              `json:"parser_matching_exempt"`
	Metadata             *DatasetMetadata  `json:"metadata,omitempty"`
}

// DatasetsList is the response body of GET /datasets.
type DatasetsList struct {
	Datasets []Dataset `json:"datasets"`
}

// DecodeDataset decodes a single Dataset from a JSON response body.
func DecodeDataset(body string) (Dataset, error) {
	var d Dataset
	err := json.NewDecoder(strings.NewReader(body)).Decode(&d)
	return d, err
}

// DecodeDatasetsList decodes a {"datasets": [...]} response body.
func DecodeDatasetsList(body string) (DatasetsList, error) {
	var list DatasetsList
	err := json.NewDecoder(strings.NewReader(body)).Decode(&list)
	return list, err
}
